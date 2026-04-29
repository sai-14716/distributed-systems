#!/usr/bin/env bash
set -euo pipefail

# Poll the discovery debug endpoint and print per-node health/probe state.
#
# Usage:
#   bash tooling/demo/watch_discovery_state.sh [seconds]
#
# Uses discovery URL from cluster_config.yaml

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config into env variables
_load_config() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml
import sys
cfg = yaml.safe_load(open(sys.argv[1]))
for laptop, ip in cfg.get('laptops', {}).items():
  print(f"export LAPTOP_{laptop}={ip}")
for node_name, node_info in cfg.get('nodes', {}).items():
  print(f"export NODE_{node_name}_LAPTOP={node_info.get('laptop', '')}")
  print(f"export NODE_{node_name}_RAFT_PORT={node_info.get('raft_port', '')}")
  print(f"export NODE_{node_name}_LB_PORT={node_info.get('lb_port', '')}")
disc = cfg.get('discovery', {})
print(f"export DISCOVERY_LAPTOP={disc.get('laptop', '')}")
print(f"export DISCOVERY_HTTP_PORT={disc.get('http_port', '')}")
PY
}
eval "$(_load_config)"

# Resolve discovery service URL
get_discovery_url() {
  local protocol="${1:-http}"
  local laptop="$DISCOVERY_LAPTOP"
  [[ -z "$laptop" ]] && { echo "ERROR: Discovery service not configured" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop for discovery: $laptop" >&2; return 1; }
  case "$protocol" in
    http|HTTP) echo "http://${ip}:${DISCOVERY_HTTP_PORT}" ;;
    udp|UDP) echo "${ip}:${DISCOVERY_UDP_PORT}" ;;
    *) echo "ERROR: Unknown discovery protocol: $protocol" >&2; return 1 ;;
  esac
}

SECS="${1:-20}"
DISCOVERY_URL="$(get_discovery_url http)/debug/state"

python3 <<'PY' "$SECS" "$DISCOVERY_URL"
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
