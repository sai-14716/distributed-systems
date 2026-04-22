#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

leader="$(
  python3 - <<'PY'
import json, urllib.request
ports = {"node1":19091,"node2":19092,"node3":19093,"node4":19094,"node5":19095}
for n,p in ports.items():
    try:
        with urllib.request.urlopen(f"http://127.0.0.1:{p}/state", timeout=0.6) as r:
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

echo "Stopping leader: $leader"
docker compose stop "$leader"

