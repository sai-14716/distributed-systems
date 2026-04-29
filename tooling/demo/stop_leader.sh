#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
export PYTHONPATH="$ROOT_DIR"

# Machine C local containers
LOCAL_NODES="node4 node5"

leader="$(
  python3 - <<'PY'
import json, urllib.request
from tooling.cluster_helper import get_config, get_url
cfg = get_config()
for n in sorted(cfg.get("nodes", {}).keys()):
    try:
        url = get_url("nodes", n, cfg)
        with urllib.request.urlopen(f"{url}/state", timeout=0.6) as r:
            st = json.loads(r.read().decode("utf-8"))
        if st.get("role") == "leader":
            print(n)
            raise SystemExit(0)
    except Exception:
        continue
raise SystemExit(1)
PY
)" || true

if [[ -z "${leader:-}" ]]; then
  echo "Could not determine leader (is the cluster up?)." >&2
  exit 1
fi

is_local=false
for ln in $LOCAL_NODES; do
  if [[ "$leader" == "$ln" ]]; then
    is_local=true
    break
  fi
done

if [[ "$is_local" == "true" ]]; then
  echo "Stopping leader: $leader"
  docker compose stop "$leader"
else
  echo "Leader is $leader (remote, not on this machine). Cannot stop from here."
  echo "Ask the operator of that machine to run: docker compose stop $leader"
fi
