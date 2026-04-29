import argparse
import http.client
import json
import os
import queue
import socket
import threading
import time
import uuid
import yaml
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlsplit

# ── SSE broadcast bus ──────────────────────────────────────────────────────────
_sse_clients: list = []
_sse_lock = threading.Lock()

def _sse_broadcast(event: dict):
    payload = "data: " + json.dumps(event) + "\n\n"
    with _sse_lock:
        dead = []
        for q in _sse_clients:
            try:
                q.put_nowait(payload)
            except queue.Full:
                dead.append(q)
        for d in dead:
            _sse_clients.remove(d)

# ── DNS ────────────────────────────────────────────────────────────────────────

DNS_SERVER = ("", 6699)

HOP_BY_HOP_HEADERS = {
    "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
    "te", "trailers", "transfer-encoding", "upgrade", "proxy-connection",
}

# ── Rotating messages sent to /chat ───────────────────────────────────────────
_MESSAGES = [
    "hello world",
    "the quick brown fox jumps over the lazy dog",
    "distributed systems are fun",
    "service discovery routes your traffic",
    "load balancers keep things balanced",
    "round robin picks the next node",
    "every request tells a story",
    "chaos engineering builds resilience",
]

# ── EndpointCache ──────────────────────────────────────────────────────────────
class EndpointCache:
    def __init__(self, ttl_ms: int = 500):
        self.ttl_ms = ttl_ms
        self._lock  = threading.Lock()
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


# ── Proxy handler ──────────────────────────────────────────────────────────────
class DiscoveryForwardProxy(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"
    cache: EndpointCache = EndpointCache(ttl_ms=500)
    max_tries: int = 3

    def log_message(self, fmt, *args):
        return  # silence default access log

    # ── routing ────────────────────────────────────────────────────────────────
    def do_GET(self):
        if self.path == "/events":
            return self._sse_stream()
        if self.path in ("/", "/dashboard"):
            return self._serve_dashboard()
        if self.path == "/health":
            return self._send_json(200, {"ok": True})
        self._handle()

    def do_POST(self):
        if self.path == "/send":
            return self._trigger_send()
        self._handle()

    def do_PUT(self):    self._handle()
    def do_DELETE(self): self._handle()
    def do_PATCH(self):  self._handle()
    def do_HEAD(self):   self._handle()
    def do_CONNECT(self): self.send_error(501, "CONNECT not supported")

    # ── SSE stream ─────────────────────────────────────────────────────────────
    def _sse_stream(self):
        q = queue.Queue(maxsize=50)
        with _sse_lock:
            _sse_clients.append(q)
        self.send_response(200)
        self.send_header("Content-Type", "text/event-stream")
        self.send_header("Cache-Control", "no-cache")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Connection", "keep-alive")
        self.end_headers()
        try:
            while True:
                try:
                    msg = q.get(timeout=15)
                    self.wfile.write(msg.encode())
                    self.wfile.flush()
                except queue.Empty:
                    self.wfile.write(b": ping\n\n")
                    self.wfile.flush()
        except Exception:
            pass
        finally:
            with _sse_lock:
                if q in _sse_clients:
                    _sse_clients.remove(q)

    # ── dashboard HTML ─────────────────────────────────────────────────────────
    def _serve_dashboard(self):
        here = os.path.dirname(os.path.abspath(__file__))
        html_path = os.path.join(here, "dashboard.html")
        try:
            with open(html_path, "rb") as f:
                body = f.read()
        except FileNotFoundError:
            return self._send_json(404, {"error": "dashboard.html not found"})
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    # ── /send: fire a POST with a rotating message ─────────────────────────────
    def _trigger_send(self):
        msg = _MESSAGES[int(time.time()) % len(_MESSAGES)]
        ep = "/chat"

        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length > 0:
                raw_body = self.rfile.read(length)
                parsed = json.loads(raw_body)
                if "message" in parsed and parsed["message"].strip():
                    msg = parsed["message"]
                if "endpoint" in parsed and parsed["endpoint"].strip():
                    ep = parsed["endpoint"]
        except Exception:
            pass

        body = msg.encode("utf-8")

        def _go():
            try:
                import urllib.parse
                qs = urllib.parse.urlencode({"q": msg})
                connect_host = os.getenv("DISCOVERY_CLIENT_CONNECT_HOST") or socket.gethostbyname(socket.gethostname())
                conn = http.client.HTTPConnection(connect_host, _PROXY_PORT, timeout=5)
                conn.request(
                    "GET", f"http://api.service.com:8000{ep}?{qs}",
                    headers={
                        "Host":           "api.service.com:8000",
                    },
                )
                r = conn.getresponse()
                r.read()
                conn.close()
            except Exception:
                pass
        threading.Thread(target=_go, daemon=True).start()

        raw = json.dumps({"ok": True, "message": msg}).encode()
        self.send_response(202)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(raw)

    # ── core proxy ─────────────────────────────────────────────────────────────
    def _handle(self):
        dest_host  = ""
        dest_port  = None
        origin_path = self.path or "/"

        # Proxy-form absolute URI
        if origin_path.startswith("http://") or origin_path.startswith("https://"):
            u = urlsplit(origin_path)
            dest_host   = u.hostname or ""
            dest_port   = u.port
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

        trace_id   = self.headers.get("X-Trace-ID", "").strip() or str(uuid.uuid4())
        req_start  = time.time()
        session_id = self.headers.get("X-Session-ID", "")
        chat_id    = self.headers.get("X-Chat-ID", "")
        role       = self.headers.get("X-Role", "")

        # ── Step 1 ────────────────────────────────────────────────────────────
        print(f"[proxy] trace={trace_id} {self.command} {dest_host}{origin_path}", flush=True)
        _sse_broadcast({
            "step": 1, "type": "request",
            "trace": trace_id, "method": self.command,
            "service": dest_host, "path": origin_path,
            "session": session_id,
            "body_preview": body[:120].decode("utf-8", errors="replace") if body else "",
        })

        # ── Step 2: DNS ───────────────────────────────────────────────────────
        try:
            endpoints, rr_idx = self._get_endpoints(dest_host)
        except Exception as exc:
            _sse_broadcast({"step": 2, "type": "dns_error", "trace": trace_id, "error": str(exc)})
            return self._send_json(502, {"error": "DNS_FAILED", "detail": str(exc)})

        tries = min(self.max_tries, len(endpoints))
        chosen_initial_idx = rr_idx % len(endpoints)

        _sse_broadcast({
            "step": 2, "type": "dns",
            "trace": trace_id,
            "endpoints": endpoints,
            "selected_idx": chosen_initial_idx,
        })

        # ── Steps 3 & 4 ───────────────────────────────────────────────────────
        last_exc = None
        for attempt in range(tries):
            node = endpoints[rr_idx % len(endpoints)]
            rr_idx = (rr_idx + 1) % len(endpoints)
            endpoint_id = node.get("id", f"{node['ip']}:{node['port']}")

            _sse_broadcast({
                "step": 3, "type": "forward",
                "trace": trace_id, "attempt": attempt, "lb": endpoint_id,
            })

            try:
                result = self._forward(dest_host, dest_port, origin_path, body, node, trace_id)
                self.cache.bump_rr(dest_host, rr_idx)
                elapsed_ms = int((time.time() - req_start) * 1000)

                print(f"[proxy] trace={trace_id} lb={endpoint_id} backend={result['backend']} status={result['status']} {elapsed_ms}ms", flush=True)
                _sse_broadcast({
                    "step": 4, "type": "response",
                    "trace": trace_id,
                    "status":     result["status"],
                    "latency_ms": elapsed_ms,
                    "lb_node":    endpoint_id,
                    "lb_backend": result["lb_backend"],
                    "backend":    result["backend"],
                    "ack":        result["ack"],
                    "body":       result["body"],
                })
                return

            except Exception as exc:
                last_exc = exc
                print(f"[proxy] trace={trace_id} lb={endpoint_id} RETRY error={exc}", flush=True)
                _sse_broadcast({
                    "step": 3, "type": "retry",
                    "trace": trace_id, "attempt": attempt,
                    "lb": endpoint_id, "error": str(exc),
                })
                continue

        _sse_broadcast({"step": 5, "type": "failed", "trace": trace_id, "error": str(last_exc)})
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
                if lk in HOP_BY_HOP_HEADERS or lk == "host":
                    continue
                headers[k] = v

            headers["Host"]                 = service_name
            headers["X-Discovery-Service"]  = service_name
            headers["X-Discovery-Endpoint"] = endpoint_id
            headers["X-Trace-ID"]           = trace_id

            conn.request(self.command, origin_path, body=body if body else None, headers=headers)
            resp = conn.getresponse()
            raw  = resp.read()

            ack = "missing"
            if raw:
                try:
                    parsed = json.loads(raw.decode("utf-8", errors="replace"))
                    ack = str(parsed.get("ack", "missing")) if isinstance(parsed, dict) else "missing"
                except Exception:
                    ack = "invalid_json"

            lb_backend = resp.getheader("X-LB-Backend", "unknown")
            backend    = resp.getheader("X-Server-ID",  "unknown")

            # Decode body for dashboard (up to 2 KB)
            try:
                body_text = raw[:2048].decode("utf-8", errors="replace") if raw else ""
            except Exception:
                body_text = ""

            self.send_response(resp.status, resp.reason)
            self.send_header("X-Discovery-Service",  service_name)
            self.send_header("X-Discovery-Endpoint", endpoint_id)
            self.send_header("X-Trace-ID",           trace_id)
            self.send_header("Access-Control-Allow-Origin", "*")
            for hk, hv in resp.getheaders():
                if hk.lower() in HOP_BY_HOP_HEADERS:
                    continue
                self.send_header(hk, hv)
            self.end_headers()
            if self.command != "HEAD":
                self.wfile.write(raw)

            return {
                "status":     resp.status,
                "lb_backend": lb_backend,
                "backend":    backend,
                "ack":        ack,
                "body":       body_text,
            }
        finally:
            conn.close()

    def _send_json(self, code, payload):
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type",   "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(raw)


# ── entry point ────────────────────────────────────────────────────────────────
_PROXY_PORT = 6700

def main():
    global _PROXY_PORT

    parser = argparse.ArgumentParser(description="Service discovery client (HTTP forward proxy).")
    parser.add_argument("--listen",       default="0.0.0.0:6700")
    parser.add_argument("--dns",          default="")
    parser.add_argument("--config",       default=os.getenv("CLUSTER_CONFIG", "cluster_config.yaml"))
    parser.add_argument("--cache-ttl-ms", type=int, default=500)
    parser.add_argument("--max-tries",    type=int, default=3)
    args = parser.parse_args()

    host, port_s = args.listen.rsplit(":", 1)
    dns_addr = args.dns
    if not dns_addr:
        with open(args.config, encoding="utf-8") as f:
            # yaml.safe_load handles both the true YAML and the JSON files seamlessly
            cfg = yaml.safe_load(f) 
            
        laptops = cfg.get("laptops", {})
        disc = cfg.get("discovery", {})
        
        laptop_id = disc.get("laptop")
        udp_port = disc.get("udp_port")
        
        if laptop_id in laptops and udp_port:
            dns_addr = f"{laptops[laptop_id]}:{udp_port}"
        else:
            raise ValueError(f"Invalid config: missing laptop '{laptop_id}' or udp_port in {args.config}")
    dns_host, dns_ps = dns_addr.rsplit(":", 1)
    _PROXY_PORT = int(port_s)

    global DNS_SERVER
    DNS_SERVER = (dns_host, int(dns_ps))

    DiscoveryForwardProxy.cache     = EndpointCache(ttl_ms=args.cache_ttl_ms)
    DiscoveryForwardProxy.max_tries = args.max_tries

    httpd = ThreadingHTTPServer((host, _PROXY_PORT), DiscoveryForwardProxy)
    print(f"--- Discovery Client Proxy  http://{host}:{_PROXY_PORT} ---")
    print(f"--- Dashboard              http://{host}:{_PROXY_PORT}/dashboard ---")
    print(f"--- DNS                    udp://{dns_host}:{dns_ps} ---")
    httpd.serve_forever()


if __name__ == "__main__":
    main()