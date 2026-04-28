#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

python3 - <<'PY'
import json, urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

with open("cluster_config.yaml") as f:
    cfg = json.load(f)

def resolve(service_type, name):
    info = cfg[service_type][name]
    ip = cfg["laptops"][info["laptop"]]
    return ip, info

nodes = list(cfg["nodes"].keys())

def fetch(url: str):
    with urllib.request.urlopen(url, timeout=0.6) as r:
        return json.loads(r.read().decode("utf-8"))

def fetch_node(n):
    ip, info = resolve("nodes", n)
    try:
        st = fetch(f"http://{ip}:{info['port']}/state")
        node_cfg = st.get("config") or {}
        return n, st, node_cfg, None
    except Exception as e:
        return n, None, None, e

print(f"{'node':<6} {'role':<10} {'leader':<6} {'term':<4} {'commit':<6} {'applied':<7} algo")
results = {}
with ThreadPoolExecutor(max_workers=min(16, max(1, len(nodes)))) as pool:
    futures = {pool.submit(fetch_node, n): n for n in nodes}
    for fut in as_completed(futures):
        n, st, node_cfg, err = fut.result()
        results[n] = (st, node_cfg, err)

for n in nodes:
    st, node_cfg, err = results[n]
    if err is not None:
        e = err
        print(f"{n:<6} {'DOWN':<10} {'':<6} {'':<4} {'':<6} {'':<7} ({e.__class__.__name__})")
    else:
        print(f"{n:<6} {st.get('role','?'):<10} {st.get('leader_id',''):<6} {st.get('term','')!s:<4} {st.get('commit_index','')!s:<6} {st.get('last_applied','')!s:<7} {node_cfg.get('algorithm','')}")
PY
