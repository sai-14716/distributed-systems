import socket
import json
import threading
import time
from collections import deque

# Simulated Registry: Multiple LBs running on localhost but different ports
dns_registry = {
    "api.service.com": deque([
        {"ip": "127.0.0.1", "port": 8001, "id": "lb-node-01"},
        {"ip": "127.0.0.1", "port": 8002, "id": "lb-node-02"},
        {"ip": "127.0.0.1", "port": 8003, "id": "lb-node-03"},
        {"ip": "127.0.0.1", "port": 8004, "id": "lb-node-03"},
        {"ip": "127.0.0.1", "port": 8005, "id": "lb-node-03"}
    ])
}

node_health = {}

def periodic_health_check(interval=5):
    """Periodically checks if the load balancers are alive using a TCP connection."""
    while True:
        for service_name, nodes in dns_registry.items():
            for node in nodes:
                ip, port = node["ip"], node["port"]
                try:
                    with socket.create_connection((ip, port), timeout=0.5):
                        node_health[(ip, port)] = True
                except (socket.timeout, socket.error, ConnectionRefusedError):
                    node_health[(ip, port)] = False
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
        if node_health.get((node["ip"], node["port"]), True)  # Assume alive if not checked yet
    ]
    return alive_endpoints

def start_dns_server():
    # Start health check thread
    hc_thread = threading.Thread(target=periodic_health_check, args=(5,), daemon=True)
    hc_thread.start()

    server_sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

    server_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_sock.bind(('localhost', 5353))
    
    print("--- Simulation DNS Server Active (Port 5353) ---")
    
    while True:
        data, addr = server_sock.recvfrom(1024)
        query = data.decode('utf-8').strip()
        
        endpoints = resolve_service(query)
        
        if endpoints:
            print(f"[DNS] Resolved {query} -> Primary: {endpoints[0]['id']} ({endpoints[0]['port']})")
            response = json.dumps(endpoints).encode('utf-8')
        else:
            print(f"[DNS] Query Failed: {query} not found")
            response = json.dumps({"error": "NXDOMAIN"}).encode('utf-8')
            
        server_sock.sendto(response, addr)

if __name__ == "__main__":
    start_dns_server()