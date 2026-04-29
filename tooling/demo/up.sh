#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config using embedded Python config class
python3 <<'CONFIGPY'
import sys, yaml
from pathlib import Path

def find_config_file():
    pwd = Path.cwd()
    for _ in range(5):
        candidate = pwd / "cluster_config.yaml"
        if candidate.exists(): return candidate
        pwd = pwd.parent
    raise FileNotFoundError("cluster_config.yaml not found")

with open(find_config_file()) as f:
    config = yaml.safe_load(f)

laptops = config.get("laptops", {})
for name in ["A", "B", "C"]:
    env_var = f"LAPTOP_{name}_IP"
    if name in laptops:
        print(f"export {env_var}='{laptops[name]}'")
CONFIGPY
eval "$(python3 <<'CONFIGPY'
import sys, yaml
from pathlib import Path

def find_config_file():
    pwd = Path.cwd()
    for _ in range(5):
        candidate = pwd / "cluster_config.yaml"
        if candidate.exists(): return candidate
        pwd = pwd.parent
    raise FileNotFoundError("cluster_config.yaml not found")

with open(find_config_file()) as f:
    config = yaml.safe_load(f)

laptops = config.get("laptops", {})
for name in ["A", "B", "C"]:
    env_var = f"LAPTOP_{name}_IP"
    if name in laptops:
        print(f"export {env_var}='{laptops[name]}'")
CONFIGPY
)"

# Compose expects LAPTOP_*_IP variables; derive them from cluster_config.yaml.
if [[ -z "${LAPTOP_A_IP:-}" || -z "${LAPTOP_B_IP:-}" || -z "${LAPTOP_C_IP:-}" ]]; then
  echo "ERROR: Missing laptop IPs in cluster_config.yaml (expected laptops A, B, C)" >&2
  exit 1
fi

CLEAN="${1:-}"
if [[ "$CLEAN" == "--clean" ]]; then
  docker compose down -v --remove-orphans
fi

echo "Starting Laptop A services..."
docker compose up -d --build \
  backend-1 backend-2 backend-3 \
  node1 node2

echo
echo "Laptop A is now running."
echo

# Display endpoints using embedded config class
python3 <<'PY'
import json, urllib.request, sys, yaml
from pathlib import Path

def find_config_file():
    pwd = Path.cwd()
    for _ in range(5):
        candidate = pwd / "cluster_config.yaml"
        if candidate.exists(): return candidate
        pwd = pwd.parent
    raise FileNotFoundError("cluster_config.yaml not found")

class Config:
    def __init__(self, data):
        self.config = data
        self.laptops = data.get("laptops", {})
        self.nodes = data.get("nodes", {})
        self.backends = data.get("backends", {})
        self.discovery = data.get("discovery", {})
    
    @staticmethod
    def load():
        with open(find_config_file()) as f:
            return Config(yaml.safe_load(f))
    
    def get_node_url(self, node_name, endpoint=""):
        if node_name not in self.nodes: raise ValueError(f"Unknown node: {node_name}")
        node_info = self.nodes[node_name]
        laptop, port = node_info.get("laptop"), node_info.get("raft_port")
        if not laptop or not port: raise ValueError(f"Invalid node: {node_name}")
        ip = self.laptops.get(laptop)
        if not ip: raise ValueError(f"Unknown laptop: {laptop}")
        return f"http://{ip}:{port}{endpoint}"
    
    def get_lb_url(self, node_name, endpoint=""):
        if node_name not in self.nodes: raise ValueError(f"Unknown node: {node_name}")
        node_info = self.nodes[node_name]
        laptop, port = node_info.get("laptop"), node_info.get("lb_port")
        if not laptop or not port: raise ValueError(f"Invalid node: {node_name}")
        ip = self.laptops.get(laptop)
        if not ip: raise ValueError(f"Unknown laptop: {laptop}")
        return f"http://{ip}:{port}{endpoint}"
    
    def get_backend_url(self, backend_name, endpoint=""):
        if backend_name not in self.backends: raise ValueError(f"Unknown backend: {backend_name}")
        backend_info = self.backends[backend_name]
        laptop, port = backend_info.get("laptop"), backend_info.get("port")
        if not laptop or not port: raise ValueError(f"Invalid backend: {backend_name}")
        ip = self.laptops.get(laptop)
        if not ip: raise ValueError(f"Unknown laptop: {laptop}")
        return f"http://{ip}:{port}{endpoint}"

cfg = Config.load()

print("Control-plane (Raft) endpoints:")
for node_name in ["node1", "node2"]:
    url = cfg.get_node_url(node_name, "/state")
    print(f"  {node_name}: {url}")

print()
print("Load Balancer endpoints:")
for node_name in ["node1", "node2"]:
    url = cfg.get_lb_url(node_name, "/chat")
    print(f"  {node_name}: {url}")

print()
print("Backend endpoints (on Laptop A):")
for backend_name in ["backend-1", "backend-2", "backend-3"]:
    url = cfg.get_backend_url(backend_name)
    print(f"  {backend_name}: {url}")

print()
print("Note: To connect other laptops (B, C), update cluster_config.yaml with their IPs")
print("and start services on those machines with their own docker-compose configurations.")
PY


