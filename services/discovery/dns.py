import json
import os
import socket
import threading
import time
from collections import deque
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

def _default_registry():
    # Default matches docker-compose static IPs on the sdnet network.
    return {
        "api.service.com": deque([
            {"ip": "10.10.0.11", "port": 8000, "id": "node1"},
            {"ip": "10.10.0.12", "port": 8000, "id": "node2"},
            {"ip": "10.10.0.13", "port": 8000, "id": "node3"},
            {"ip": "10.10.0.14", "port": 8000, "id": "node4"},
            {"ip": "10.10.0.15", "port": 8000, "id": "node5"},
        ])
    }


def _load_registry_from_env():
    raw = os.getenv("DNS_REGISTRY_JSON", "").strip()
    if not raw:
        return None
    reg = json.loads(raw)
    out = {}
    for svc, endpoints in reg.items():
        if not isinstance(endpoints, list):
            continue
        out[svc] = deque(endpoints)
    return out or None


# Simulated Registry: Service -> rotating endpoint list
dns_registry = _load_registry_from_env() or _default_registry()

node_health = {}
node_probe_state = {}
state_lock = threading.Lock()


def _node_key(ip, port):
    return (ip, int(port))


def _record_probe(ip, port, healthy, error=None):
    now_ms = int(time.time() * 1000)
    key = _node_key(ip, port)
    with state_lock:
        node_health[key] = healthy
        node_probe_state[key] = {
            "last_probe_ms": now_ms,
            "last_probe_iso": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime(now_ms / 1000.0)),
            "healthy": healthy,
            "last_error": error or "",
        }


def _snapshot_state():
    with state_lock:
        services = {}
        for service_name, nodes in dns_registry.items():
            rendered_nodes = []
            for node in list(nodes):
                key = _node_key(node["ip"], node["port"])
                probe = node_probe_state.get(key, {})
                healthy = probe.get("healthy")
                if healthy is None:
                    healthy = node_health.get(key, True)
                rendered_nodes.append(
                    {
                        "id": node.get("id", ""),
                        "ip": node["ip"],
                        "port": int(node["port"]),
                        "healthy": bool(healthy),
                        "last_probe_ms": probe.get("last_probe_ms"),
                        "last_probe_iso": probe.get("last_probe_iso"),
                        "last_error": probe.get("last_error", ""),
                    }
                )
            services[service_name] = rendered_nodes
        return {
            "generated_at_ms": int(time.time() * 1000),
            "generated_at_iso": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "services": services,
        }


class DiscoveryDebugHandler(BaseHTTPRequestHandler):
    def log_message(self, _format, *_args):
        return

    def do_GET(self):
        if self.path not in {"/debug/state", "/debug/state/"}:
            return self._send_json(404, {"error": "NOT_FOUND"})
        return self._send_json(200, _snapshot_state())

    def _send_json(self, code, payload):
        raw = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

def periodic_health_check(interval=5):
    """Periodically checks if the load balancers are alive using a TCP connection."""
    while True:
        for service_name, nodes in dns_registry.items():
            for node in nodes:
                ip, port = node["ip"], node["port"]
                try:
                    with socket.create_connection((ip, port), timeout=0.5):
                        _record_probe(ip, port, True)
                except (socket.timeout, socket.error, ConnectionRefusedError) as e:
                    _record_probe(ip, port, False, str(e))
        time.sleep(interval)

def resolve_service(service_name):
    if service_name not in dns_registry:
        return None

    # Rotate the deque to provide Round Robin ordering
    # The first element becomes the "Primary" for this specific client request
    dns_registry[service_name].rotate(-1)
    
    # Return the full list of ALIVE nodes so the client can perform failover/retries
    alive_endpoints = [
        node for node in dns_registry[service_name]
        if node_health.get(_node_key(node["ip"], node["port"]), True)  # Assume alive if not checked yet
    ]
    return alive_endpoints

def start_discovery_server(host="127.0.0.1", port=6699):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind((host, port))

    print(f"--- Service Discovery Active (udp://{host}:{port}) ---")
    print("Services:", ", ".join(dns_registry.keys()))

    while True:
        data, addr = sock.recvfrom(2048)
        query = data.decode("utf-8", errors="replace").strip()
        endpoints = resolve_service(query)

        if endpoints:
            response = json.dumps(endpoints).encode("utf-8")
        else:
            response = json.dumps({"error": "NXDOMAIN"}).encode("utf-8")

        sock.sendto(response, addr)


def start_debug_server(host="0.0.0.0", port=6701):
    httpd = ThreadingHTTPServer((host, port), DiscoveryDebugHandler)
    print(f"--- Discovery Debug API Active (http://{host}:{port}/debug/state) ---")
    httpd.serve_forever()

if __name__ == "__main__":
    threading.Thread(target=periodic_health_check, args=(5,), daemon=True).start()
    debug_bind = os.getenv("DISCOVERY_DEBUG_BIND", "0.0.0.0")
    debug_port = int(os.getenv("DISCOVERY_DEBUG_PORT", "6701"))
    threading.Thread(target=start_debug_server, args=(debug_bind, debug_port), daemon=True).start()
    bind_host = os.getenv("DISCOVERY_BIND", "0.0.0.0")
    bind_port = int(os.getenv("DISCOVERY_PORT", "6699"))
    start_discovery_server(bind_host, bind_port)
