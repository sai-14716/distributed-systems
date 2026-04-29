#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

_load_config() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml, sys
cfg = yaml.safe_load(open(sys.argv[1]))
for laptop, ip in cfg.get('laptops', {}).items():
  print(f"export LAPTOP_{laptop}={ip}")
for node_name, node_info in cfg.get('nodes', {}).items():
  print(f"export NODE_{node_name}_LAPTOP={node_info.get('laptop', '')}")
  print(f"export NODE_{node_name}_LB_PORT={node_info.get('lb_port', '')}")
  print(f"export NODE_{node_name}_RAFT_PORT={node_info.get('raft_port', '')}")
PY
}
eval "$(_load_config)"

get_lb_url() {
  eval "echo http://\${LAPTOP_\${NODE_$1_LAPTOP}}:\${NODE_$1_LB_PORT}${2:-}"
}

submit_config() {
  eval "curl -sS -X POST http://\${LAPTOP_\${NODE_${4:-node1}_LAPTOP}}:\${NODE_${4:-node1}_RAFT_PORT}/admin/submit -H 'Content-Type: application/json' -d '{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$1\",\"probe_interval_ms\":${2:-1000},\"health_threshold\":${3:-0.8}}}' -w '\\nStatus: %{http_code}\\n' 2>&1"
}

LB_NODE="${1:-node1}"
LB_URL="$(get_lb_url "$LB_NODE")"
REQUESTS="${2:-15}"
LOAD_SECS="${3:-60}"
PHASE_SECS="${4:-0.5}"
BASE_WORK_MS="${5:-1800}"
PEAK_PARALLEL="${6:-8}"

if [[ "$REQUESTS" -le 0 || "$LOAD_SECS" -le 0 ]]; then
  echo "REQUESTS and LOAD_SECS must be > 0" >&2
  exit 1
fi

echo "setting algorithm to wrr"
submit_config "wrr" 1000 0.8 >/dev/null 2>&1

echo "starting dynamic backend load in background"
bash tooling/demo/dynamic_backend_load.sh "$LOAD_SECS" "$PHASE_SECS" "$BASE_WORK_MS" "$PEAK_PARALLEL" &
load_pid=$!

cleanup() {
  if kill -0 "$load_pid" 2>/dev/null; then
    kill "$load_pid" 2>/dev/null || true
    wait "$load_pid" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

sleep 1

echo "showing $REQUESTS routed requests while dynamic load is running"
bash tooling/demo/show_routing_headers.sh "$LB_NODE" "/" "$REQUESTS"

echo "wrr + dynamic load demo done"
