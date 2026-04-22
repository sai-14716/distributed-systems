import argparse
import http.client
import json
import socket
import threading
import time
import uuid
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlsplit

DNS_SERVER = ("127.0.0.1", 6699)

HOP_BY_HOP_HEADERS = {
    "connection",
    "keep-alive",
    "proxy-authenticate",
    "proxy-authorization",
    "te",
    "trailers",
    "transfer-encoding",
    "upgrade",
    "proxy-connection",
}


class EndpointCache:
    def __init__(self, ttl_ms: int = 500):
        self.ttl_ms = ttl_ms
        self._lock = threading.Lock()
        self._cache = {}  # service -> (expires_at_ms, endpoints, rr_idx)

    def get(self, service: str):
        now = int(time.time() * 1000)
        with self._lock:
            hit = self._cache.get(service)
            if not hit:
                return None
            exp, endpoints, rr_idx = hit
            if exp <= now or not endpoints:
                return None
            return (endpoints, rr_idx)

    def get_stale(self, service: str):
        with self._lock:
            hit = self._cache.get(service)
            if not hit:
                return None
            exp, endpoints, rr_idx = hit
            if not endpoints:
                return None
            return (endpoints, rr_idx)

    def put(self, service: str, endpoints):
        now = int(time.time() * 1000)
        with self._lock:
            self._cache[service] = (now + self.ttl_ms, endpoints, 0)

    def bump_rr(self, service: str, rr_idx: int):
        with self._lock:
            hit = self._cache.get(service)
            if not hit:
                return
            exp, endpoints, _ = hit
            self._cache[service] = (exp, endpoints, rr_idx)


def resolve(service_name: str, timeout: float = 0.2):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(timeout)
    try:
        sock.sendto(service_name.encode("utf-8"), DNS_SERVER)
        data, _ = sock.recvfrom(8192)
    finally:
        sock.close()

    response = json.loads(data.decode("utf-8", errors="replace"))
    if isinstance(response, dict) and "error" in response:
        raise RuntimeError(response["error"])
    if not isinstance(response, list) or not response:
        raise RuntimeError("no endpoints")
    return response


class DiscoveryForwardProxy(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    cache: EndpointCache = EndpointCache(ttl_ms=500)
    max_tries: int = 3

    def log_message(self, fmt, *args):
        return

    def do_CONNECT(self):
        self.send_error(501, "CONNECT not supported")

    def do_GET(self):
        self._handle()

    def do_POST(self):
        self._handle()

    def do_PUT(self):
        self._handle()

    def do_DELETE(self):
        self._handle()

    def do_PATCH(self):
        self._handle()

    def do_HEAD(self):
        self._handle()

    def _handle(self):
        if self.path == "/health":
            return self._send_json(200, {"ok": True})

        dest_host = ""
        dest_port = None
        origin_path = self.path or "/"

        # Proxy-form absolute URI
        if origin_path.startswith("http://") or origin_path.startswith("https://"):
            u = urlsplit(origin_path)
            dest_host = u.hostname or ""
            dest_port = u.port
            origin_path = u.path or "/"
            if u.query:
                origin_path = f"{origin_path}?{u.query}"

        # Origin-form fallback
        if not dest_host:
            host_hdr = self.headers.get("Host", "").strip()
            if host_hdr:
                dest_host = host_hdr.split(":", 1)[0]
                if ":" in host_hdr:
                    try:
                        dest_port = int(host_hdr.split(":", 1)[1])
                    except ValueError:
                        dest_port = None

        if not dest_host:
            return self._send_json(400, {"error": "MISSING_HOST"})

        body = b""
        if self.command in {"POST", "PUT", "PATCH"}:
            try:
                length = int(self.headers.get("Content-Length", "0"))
            except ValueError:
                length = 0
            if length > 0:
                body = self.rfile.read(length)

        trace_id = self.headers.get("X-Trace-ID", "").strip() or str(uuid.uuid4())
        req_start = time.time()
        session_id = self.headers.get("X-Session-ID", "")
        chat_id = self.headers.get("X-Chat-ID", "")
        role = self.headers.get("X-Role", "")

        print(
            f"[client-proxy] event=request_sent trace={trace_id} method={self.command} service={dest_host} path={origin_path} session={session_id} chat={chat_id} role={role}",
            flush=True,
        )

        endpoints, rr_idx = self._get_endpoints(dest_host)
        tries = min(self.max_tries, len(endpoints))

        last_exc = None
        for _ in range(tries):
            node = endpoints[rr_idx % len(endpoints)]
            rr_idx = (rr_idx + 1) % len(endpoints)
            try:
                result = self._forward(dest_host, dest_port, origin_path, body, node, trace_id)
                self.cache.bump_rr(dest_host, rr_idx)
                elapsed_ms = int((time.time() - req_start) * 1000)
                endpoint_id = node.get("id", f"{node['ip']}:{node['port']}")
                print(
                    f"[client-proxy] event=response_received trace={trace_id} method={self.command} service={dest_host} path={origin_path} endpoint={endpoint_id} status={result['status']} lb_backend={result['lb_backend']} backend={result['backend']} ack=\"{result['ack']}\" latency_ms={elapsed_ms}",
                    flush=True,
                )
                return
            except Exception as exc:
                last_exc = exc
                endpoint_id = node.get("id", f"{node['ip']}:{node['port']}")
                print(
                    f"[client-proxy] trace={trace_id} method={self.command} service={dest_host} path={origin_path} endpoint={endpoint_id} status=retry error={exc}",
                    flush=True,
                )
                continue

        return self._send_json(502, {"error": "NO_BACKEND_AVAILABLE", "detail": str(last_exc) if last_exc else ""})

    def _get_endpoints(self, service: str):
        hit = self.cache.get(service)
        if hit:
            return hit
        try:
            endpoints = resolve(service)
        except Exception:
            stale = self.cache.get_stale(service)
            if stale:
                return stale
            raise
        else:
            self.cache.put(service, endpoints)
            return (endpoints, 0)

    def _forward(self, service_name, _dest_port, origin_path, body, node, trace_id):
        ip, port = node["ip"], int(node["port"])
        endpoint_id = node.get("id", f"{ip}:{port}")

        conn = http.client.HTTPConnection(ip, port, timeout=0.8)
        try:
            headers = {}
            for k, v in self.headers.items():
                lk = k.lower()
                if lk in HOP_BY_HOP_HEADERS:
                    continue
                if lk == "host":
                    continue
                headers[k] = v

            headers["Host"] = service_name
            headers["X-Discovery-Service"] = service_name
            headers["X-Discovery-Endpoint"] = endpoint_id
            headers["X-Trace-ID"] = trace_id
            headers["X-Packet-Path"] = "Client -> ServiceDiscovery"

            conn.request(self.command, origin_path, body=body if body else None, headers=headers)
            resp = conn.getresponse()
            raw = resp.read()
            ack = "missing"
            
            packet_path = resp.getheader("X-Packet-Path", "")
            if packet_path:
                packet_path += " -> ServiceDiscovery -> Client"

            if raw:
                try:
                    parsed = json.loads(raw.decode("utf-8", errors="replace"))
                    if isinstance(parsed, dict):
                        ack = str(parsed.get("ack", "missing"))
                        if packet_path:
                            parsed["packet_path"] = packet_path
                            raw = json.dumps(parsed).encode("utf-8")
                except Exception:
                    ack = "invalid_json"

            lb_backend = resp.getheader("X-LB-Backend", "unknown")
            backend = resp.getheader("X-Server-ID", "unknown")

            self.send_response(resp.status, resp.reason)
            self.send_header("X-Discovery-Service", service_name)
            self.send_header("X-Discovery-Endpoint", endpoint_id)
            self.send_header("X-Trace-ID", trace_id)
            if packet_path:
                self.send_header("X-Packet-Path", packet_path)
            for hk, hv in resp.getheaders():
                if hk.lower() in HOP_BY_HOP_HEADERS:
                    continue
                self.send_header(hk, hv)
            self.end_headers()

            if self.command != "HEAD":
                self.wfile.write(raw)

            return {
                "status": resp.status,
                "lb_backend": lb_backend,
                "backend": backend,
                "ack": ack,
            }
        finally:
            conn.close()

    def _send_json(self, code, payload):
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)


def main():
    parser = argparse.ArgumentParser(description="Service discovery client (HTTP forward proxy).")
    parser.add_argument("--listen", default="127.0.0.1:6700", help="proxy listen addr (host:port)")
    parser.add_argument("--dns", default="127.0.0.1:6699", help="discovery UDP addr (host:port)")
    parser.add_argument("--cache-ttl-ms", type=int, default=500)
    parser.add_argument("--max-tries", type=int, default=3)
    args = parser.parse_args()

    host, port_s = args.listen.rsplit(":", 1)
    port = int(port_s)
    dns_host, dns_port_s = args.dns.rsplit(":", 1)
    dns_port = int(dns_port_s)

    global DNS_SERVER
    DNS_SERVER = (dns_host, dns_port)

    DiscoveryForwardProxy.cache = EndpointCache(ttl_ms=args.cache_ttl_ms)
    DiscoveryForwardProxy.max_tries = args.max_tries

    httpd = ThreadingHTTPServer((host, port), DiscoveryForwardProxy)
    print(f"--- Discovery Client Proxy Active (http://{host}:{port}) ---")
    print(f"Discovery server: udp://{dns_host}:{dns_port} (service->endpoints)")
    httpd.serve_forever()


if __name__ == "__main__":
    main()
