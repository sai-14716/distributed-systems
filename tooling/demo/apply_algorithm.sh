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
PY
}
eval "$(_load_config)"

# Resolve node URL from cluster config
get_node_url() {
  local node_name="$1"
  local endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP"
  local port_var="NODE_${node_name}_RAFT_PORT"
  eval "local laptop=\${${laptop_var}-}"
  eval "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown node: $node_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
}

# Submit config to Raft cluster
submit_config() {
  local algorithm="${1:-}"
  [[ -z "$algorithm" ]] && { echo "ERROR: algorithm required" >&2; return 1; }
  local probe_interval_ms="${2:-1000}"
  local health_threshold="${3:-0.8}"
  local target_node="${4:-node1}"
  
  local target_url=$(get_node_url "$target_node" "/admin/submit") || return 1
  echo "[config] Sending to $target_node: algorithm=$algorithm probe_interval_ms=$probe_interval_ms health_threshold=$health_threshold" >&2
  
  curl -sS -X POST "$target_url" \
    -H 'Content-Type: application/json' \
    -d "{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$algorithm\",\"probe_interval_ms\":$probe_interval_ms,\"health_threshold\":$health_threshold}}" \
    -w "\nStatus: %{http_code}\n" 2>&1
}

ALGO="${1:-}"
if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm>" >&2
  exit 1
fi

submit_config "$ALGO" 1000 0.8