#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

leader="$(
  python3 - <<'PY'
import sys
sys.path.insert(0, "tooling/helper")
from cluster_config import Config, fetch_json

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

