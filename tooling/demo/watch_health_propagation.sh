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

python3 - <<'PY' "$SECS"
import json
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

secs = int(sys.argv[1])

with open("cluster_config.yaml") as f:
    cfg = json.load(f)

laptops = cfg["laptops"]

def resolve(service_type, name):
    info = cfg[service_type][name]
    ip = laptops[info["laptop"]]
    return ip, info

lbs = []
for name in cfg.get("load_balancers", {}).keys():
    ip, info = resolve("load_balancers", name)
    lbs.append((name, f"http://{ip}:{info['port']}/admin/load-view"))

def fetch(url):
    with urllib.request.urlopen(url, timeout=1.0) as r:
        return json.loads(r.read().decode("utf-8"))

def fetch_lb(name, url):
    try:
        return name, fetch(url)
    except Exception as e:
        return name, {"error": f"ERR {e.__class__.__name__}"}

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

    with ThreadPoolExecutor(max_workers=min(16, max(1, len(lbs)))) as pool:
        futures = [pool.submit(fetch_lb, name, url) for name, url in lbs]
        for fut in as_completed(futures):
            name, snap = fut.result()
            snapshots[name] = snap

    backend_ids = set()
    for snap in snapshots.values():
        if "error" in snap:
            continue
        backend_ids.update((snap.get("backends") or {}).keys())

    rows = {bid: {} for bid in backend_ids}
    for bid in backend_ids:
        for n in nodes:
            snap = snapshots.get(n, {})
            if "error" in snap:
                rows[bid][n] = snap["error"]
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

