#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config into env variables
_load_config() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml, sys
cfg = yaml.safe_load(open(sys.argv[1]))
for laptop, ip in cfg.get('laptops', {}).items():
  print(f"export LAPTOP_{laptop}={ip}")
for node_name, node_info in cfg.get('nodes', {}).items():
  print(f"export NODE_{node_name}_LAPTOP={node_info.get('laptop', '')}")
  print(f"export NODE_{node_name}_RAFT_PORT={node_info.get('raft_port', '')}")
  print(f"export NODE_{node_name}_LB_PORT={node_info.get('lb_port', '')}")
PY
}
eval "$(_load_config)"

# Resolve LB URL
get_lb_url() {
  local node_name="$1" endpoint="${2:-}"
  eval "local ip=\${LAPTOP_\${NODE_${node_name}_LAPTOP}-}" "local port=\${NODE_${node_name}_LB_PORT-}"
  echo "http://${ip}:${port}${endpoint}"
}

# Submit config  
submit_config() {
  local algorithm="${1:-}" target_node="${4:-node1}"
  eval "local ip=\${LAPTOP_\${NODE_${target_node}_LAPTOP}-}" "local port=\${NODE_${target_node}_RAFT_PORT-}"
  curl -sS -X POST "http://${ip}:${port}/admin/submit" -H 'Content-Type: application/json' \
    -d "{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$algorithm\",\"probe_interval_ms\":${2:-1000},\"health_threshold\":${3:-0.8}}}" -w "\nStatus: %{http_code}\n" 2>&1
}

# Use first LB (node1) by default
LB_NODE="${1:-node1}"
LB_URL="$(get_lb_url "$LB_NODE")"

submit_config "least-req" 1000 0.8 >/dev/null 2>&1

tmpdir="$(mktemp -d)"
hdr_a="$tmpdir/a.hdr"
hdr_b="$tmpdir/b.hdr"

echo "Starting a long request A (keeps one backend's active-request count at 1)..."
curl -s -D "$hdr_a" -o /dev/null "$LB_URL/chat" -H "X-Session-ID: demo-A" -H "X-Work-Ms: 5000" &
pid=$!

sleep 0.2

echo "Sending short request B while A is in-flight (least-req should prefer the backend with fewer active requests)..."
curl -s -D "$hdr_b" -o /dev/null "$LB_URL/chat" -H "X-Session-ID: demo-B" -H "X-Work-Ms: 50"

wait "$pid" || true

backend_a="$(awk 'tolower($1) == "x-lb-backend:" {print $2}' "$hdr_a" | tr -d '\r')"
backend_b="$(awk 'tolower($1) == "x-lb-backend:" {print $2}' "$hdr_b" | tr -d '\r')"
pool_a="$(awk 'tolower($1) == "x-lb-candidates:" {print $2}' "$hdr_a" | tr -d '\r')"
pool_b="$(awk 'tolower($1) == "x-lb-candidates:" {print $2}' "$hdr_b" | tr -d '\r')"

echo "A backend=$backend_a"
echo "A candidates=$pool_a"
echo "B backend=$backend_b"
echo "B candidates=$pool_b"
if [[ -n "$backend_a" && -n "$backend_b" && "$backend_a" != "$backend_b" ]]; then
  echo "OK: B avoided A's backend while A was in-flight"
else
  echo "NOTE: did not observe avoidance (try again; depends on timing / pool size / health)"
fi

