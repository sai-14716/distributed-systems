#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

leader="$(
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
    
    def get_node_names(self):
        return sorted(self.nodes.keys())

def fetch_json(url, timeout=0.6):
    try:
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=timeout) as response:
            return json.loads(response.read().decode('utf-8'))
    except:
        return None

try:
    cfg = Config.load()
except Exception as e:
    print(f"ERROR: Failed to load config: {e}", file=sys.stderr)
    sys.exit(1)

# Check nodes on this laptop (Laptop A) first
for node_name in cfg.get_node_names():
    try:
        url = cfg.get_node_url(node_name, "/state")
        st = fetch_json(url, timeout=0.6)
        if st and st.get("role") == "leader":
            print(node_name)
            sys.exit(0)
    except Exception:
        continue

# If no local leader, check other nodes
# (This only works if running from a host that can reach all laptops)
for node_name in cfg.get_node_names():
    try:
        url = cfg.get_node_url(node_name, "/state")
        st = fetch_json(url, timeout=1.0)
        if st and st.get("role") == "leader":
            print(node_name)
            sys.exit(0)
    except Exception:
        continue

sys.exit(1)
PY
)" || true

if [[ -z "${leader:-}" ]]; then
  echo "Could not determine leader (is the cluster up? Check cluster_config.yaml)." >&2
  exit 1
fi

echo "Stopping leader: $leader"
docker compose stop "$leader"


