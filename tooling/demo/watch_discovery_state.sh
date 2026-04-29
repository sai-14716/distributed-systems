#!/usr/bin/env bash
set -euo pipefail

# Poll the discovery debug endpoint and print per-node health/probe state.
#
# Usage:
#   bash tooling/demo/watch_discovery_state.sh [seconds] [url]
#
# Examples:
#   bash tooling/demo/watch_discovery_state.sh 20
#   bash tooling/demo/watch_discovery_state.sh 30 http://127.0.0.1:6701/debug/state

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

export PYTHONPATH="$ROOT_DIR"
DISC_BASE=$(python3 tooling/cluster_helper.py get_discovery_http)

SECS="${1:-20}"
STATE_URL="${2:-$DISC_BASE/debug/state}"

python3 - <<'PY' "$SECS" "$STATE_URL"
import json
import sys
import time
import urllib.request

secs = int(sys.argv[1])
url = sys.argv[2]


def fetch(endpoint):
    with urllib.request.urlopen(endpoint, timeout=1.5) as r:
        return json.loads(r.read().decode("utf-8"))


def fmt_node(node):
    nid = node.get("id", "-")
    ip = node.get("ip", "-")
    port = node.get("port", "-")
    healthy = "up" if node.get("healthy") else "down"
    last_probe = node.get("last_probe_iso") or "never"
    last_error = node.get("last_error") or ""
    err = "" if not last_error else f" err={last_error}"
    return f"{nid:<6} {ip}:{port:<5} health={healthy:<4} last_probe={last_probe}{err}"


end = time.time() + secs
while time.time() < end:
    print(time.strftime("%H:%M:%S"))
    try:
        state = fetch(url)
    except Exception as e:
        print(f"  ERR fetch failed: {e.__class__.__name__}: {e}")
        print()
        time.sleep(1)
        continue

    services = state.get("services", {})
    if not services:
        print("  no services in discovery state")
        print()
        time.sleep(1)
        continue

    for service_name in sorted(services.keys()):
        print(f"  service={service_name}")
        for node in services.get(service_name, []):
            print("    " + fmt_node(node))
    print()
    time.sleep(1)
PY
