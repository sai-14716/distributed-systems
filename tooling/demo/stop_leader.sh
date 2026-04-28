#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

leader="$(
python3 - <<'PY'
import json, urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

with open("cluster_config.yaml") as f:
    cfg = json.load(f)

def resolve(service_type, name):
    info = cfg[service_type][name]
    ip = cfg["laptops"][info["laptop"]]
    return ip, info

def fetch_role(n):
    ip, info = resolve("nodes", n)
    try:
        with urllib.request.urlopen(f"http://{ip}:{info['port']}/state", timeout=0.6) as r:
            st = json.loads(r.read().decode("utf-8"))
        return n, st.get("role")
    except Exception:
        return n, None

nodes = list(cfg["nodes"].keys())
with ThreadPoolExecutor(max_workers=min(16, max(1, len(nodes)))) as pool:
    futures = [pool.submit(fetch_role, n) for n in nodes]
    for fut in as_completed(futures):
        n, role = fut.result()
        if role == "leader":
            print(n)
            raise SystemExit(0)
raise SystemExit(1)
PY
)" || true

if [[ -z "${leader:-}" ]]; then
  echo "Could not determine leader (is the cluster up?)." >&2
  exit 1
fi

echo "Stopping leader: $leader"
docker compose stop "$leader"

