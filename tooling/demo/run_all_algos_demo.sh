#!/usr/bin/env bash
set -euo pipefail

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
PY
}
eval "$(_load_config)"

# Resolve node URL
get_node_url() {
  local node_name="$1" endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP" port_var="NODE_${node_name}_RAFT_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown node: $node_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  echo "http://${ip}:${port}${endpoint}"
}

# Resolve LB URL
get_lb_url() {
  local node_name="$1" endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP" port_var="NODE_${node_name}_LB_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  eval "local ip=\${LAPTOP_${laptop}-}"
  echo "http://${ip}:${port}${endpoint}"
}

# Submit config
submit_config() {
  local algorithm="${1:-}" probe_interval_ms="${2:-1000}" health_threshold="${3:-0.8}" target_node="${4:-node1}"
  local target_url=$(get_node_url "$target_node" "/admin/submit") || return 1
  curl -sS -X POST "$target_url" -H 'Content-Type: application/json' \
    -d "{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$algorithm\",\"probe_interval_ms\":$probe_interval_ms,\"health_threshold\":$health_threshold}}" -w "\nStatus: %{http_code}\n" 2>&1
}

# This script needs show_routing_headers which is in tooling/demo/
show_routing_headers() {
  bash tooling/demo/show_routing_headers.sh "$@"
}

LB_NODE="${1:-node1}"
N="${2:-30}"

echo "=== 1) Per-request headers by algorithm ==="
for algo in round_robin maglev least-req wrr least-load; do
  echo
  echo "--- algorithm=$algo ---"
  submit_config "$algo" 1000 0.8 >/dev/null 2>&1
  show_routing_headers "$LB_NODE" "/chat" 1
done

echo
echo "=== 2) Least-Requests differentiation (in-flight avoidance) ==="
bash tooling/demo/least_requests_demo.sh "$LB_NODE"

echo
echo "=== 3) WRR vs Least-Load under a heated backend ==="
echo "--- wrr ---"
bash tooling/demo/load_weight_demo.sh wrr "$LB_NODE" "$N"

echo
echo "--- least-load ---"
bash tooling/demo/load_weight_demo.sh least-load "$LB_NODE" "$N"

echo
echo "Done. For health probe/propagation visibility, run:"
echo "  bash tooling/demo/watch_health_propagation.sh 20"
