#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

python3 - <<'PY'
import json
import urllib.request
import sys
from pathlib import Path

# Add tooling to path for import
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "tooling" / "helper"))

from cluster_config import Config, fetch_json

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
