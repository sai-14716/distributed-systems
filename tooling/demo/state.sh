#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

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
    except: return None

try:
    cfg = Config.load()
except Exception as e:
    print(f"ERROR: Failed to load config: {e}", file=sys.stderr)
    sys.exit(1)

nodes = cfg.get_node_names()
print(f"{'node':<6} {'role':<10} {'leader':<6} {'term':<4} {'commit':<6} {'applied':<7} algo")
for n in nodes:
    try:
        url = cfg.get_node_url(n, "/state")
        st = fetch_json(url, timeout=0.6)
        if st is None:
            print(f"{n:<6} {'DOWN':<10} {'':<6} {'':<4} {'':<6} {'':<7} (timeout)")
            continue
        cfg_info = st.get("config") or {}
        print(f"{n:<6} {st.get('role','?'):<10} {st.get('leader_id',''):<6} {st.get('term','')!s:<4} {st.get('commit_index','')!s:<6} {st.get('last_applied','')!s:<7} {cfg_info.get('algorithm','')}")
    except Exception as e:
        print(f"{n:<6} {'DOWN':<10} {'':<6} {'':<4} {'':<6} {'':<7} ({e.__class__.__name__})")
PY
