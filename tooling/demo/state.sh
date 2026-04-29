#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

export PYTHONPATH="$ROOT_DIR"

python3 - <<'PY'
import json
import urllib.request
from tooling.cluster_helper import get_config, get_url

def fetch(url: str):
    with urllib.request.urlopen(url, timeout=0.6) as r:
        return json.loads(r.read().decode("utf-8"))

cfg = get_config()
nodes = list(cfg.get("nodes", {}).keys())

print(f"{'node':<6} {'role':<10} {'leader':<6} {'term':<4} {'commit':<6} {'applied':<7} algo")
for n in nodes:
    try:
        url = get_url("nodes", n, cfg)
        st = fetch(f"{url}/state")
        c = st.get("config") or {}
        print(f"{n:<6} {st.get('role','?'):<10} {st.get('leader_id',''):<6} {st.get('term','')!s:<4} {st.get('commit_index','')!s:<6} {st.get('last_applied','')!s:<7} {c.get('algorithm','')}")
    except Exception as e:
        print(f"{n:<6} {'DOWN':<10} {'':<6} {'':<4} {'':<6} {'':<7} ({e.__class__.__name__})")
PY
