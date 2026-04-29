#!/usr/bin/env bash
set -euo pipefail

# Poll /admin/load-view from multiple LBs and print a backend-by-node health table.
#
# Usage:
#   sh tooling/demo/watch_health_propagation.sh [seconds]
#
# Each cell shows that node's current view for the backend:
#   <status> b<bucket> a<active> e<epoch>/<seq>

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

SECS="${1:-15}"

python3 <<'PY' "$SECS"
import json, urllib.request, sys, time, yaml
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
    
    def get_lb_url(self, node_name, endpoint=""):
        if node_name not in self.nodes: raise ValueError(f"Unknown node: {node_name}")
        node_info = self.nodes[node_name]
        laptop, port = node_info.get("laptop"), node_info.get("lb_port")
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

secs = int(sys.argv[1])

try:
    cfg = Config.load()
except Exception as e:
    print(f"ERROR: Failed to load config: {e}", file=sys.stderr)
    sys.exit(1)

# Build list of (node_name, lb_url) pairs
lbs = []
for node_name in cfg.get_node_names():
    lb_url = cfg.get_lb_url(node_name, "/admin/load-view")
    lbs.append((node_name, lb_url))

def fmt_int(v):
    try:
        return int(v)
    except Exception:
        return -1

def fmt_cell(backend_entry):
    if not backend_entry:
        return "NA"
    view = backend_entry.get("view") or {}
    if not view:
        return "NA"

    status = str(view.get("status", "?")).lower()
    if status == "up":
        st = "UP"
    elif status == "down":
        st = "DN"
    else:
        st = "??"

    bucket = fmt_int(view.get("bucket"))
    active = fmt_int(view.get("active"))
    epoch = fmt_int(view.get("epoch"))
    seq = fmt_int(view.get("seq"))    
    return f"{st} b{bucket:02d} a{active:02d} e{epoch}/{seq}"

def print_table(rows, nodes):
    colw = 20
    first = 12
    header = "backend".ljust(first) + "".join(n.ljust(colw) for n in nodes)
    print(header)
    print("-" * len(header))
    for backend_id in sorted(rows.keys()):
        line = backend_id.ljust(first)
        for n in nodes:
            line += rows[backend_id].get(n, "NA").ljust(colw)
        print(line)

end = time.time() + secs
while time.time() < end:
    nodes = [name for name, _ in lbs]
    snapshots = {}

    for name, url in lbs:
        try:
            snapshots[name] = fetch_json(url, timeout=1.5)
            if snapshots[name] is None:
                snapshots[name] = {"error": "timeout"}
        except Exception as e:
            snapshots[name] = {"error": f"ERR {e.__class__.__name__}"}

    backend_ids = set()
    for snap in snapshots.values():
        if snap is None or "error" in snap:
            continue
        backend_ids.update((snap.get("backends") or {}).keys())

    rows = {bid: {} for bid in backend_ids}
    for bid in backend_ids:
        for n in nodes:
            snap = snapshots.get(n, {})
            if snap is None or "error" in snap:
                rows[bid][n] = snap.get("error") if snap else "NA"
                continue
            entry = (snap.get("backends") or {}).get(bid)
            rows[bid][n] = fmt_cell(entry)

    print(time.strftime("%H:%M:%S"))
    if rows:
        print_table(rows, nodes)
    else:
        print("no backend data")
    print()
    time.sleep(1)
PY


