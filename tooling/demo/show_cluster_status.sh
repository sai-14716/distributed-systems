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
  print(f"export NODE_{node_name}_LB_PORT={node_info.get('lb_port', '')}")
for backend_name, backend_info in cfg.get('backends', {}).items():
  backend_key = ''.join(ch if (ch.isalnum() or ch == '_') else '_' for ch in backend_name)
  print(f"export BACKEND_{backend_key}_LAPTOP={backend_info.get('laptop', '')}")
  print(f"export BACKEND_{backend_key}_PORT={backend_info.get('port', '')}")
PY
}
eval "$(_load_config)"

# Helper to convert backend names to shell-safe variable keys
_to_var_key() {
  local input="$1"
  echo "${input//[^a-zA-Z0-9_]/_}"
}

# Resolve LB node URL
get_lb_url() {
  local node_name="$1" endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP" port_var="NODE_${node_name}_LB_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown node: $node_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
}

# Resolve backend URL
get_backend_url() {
  local backend_name="$1" endpoint="${2:-}"
  local backend_key
  backend_key="$(_to_var_key "$backend_name")"
  local laptop_var="BACKEND_${backend_key}_LAPTOP" port_var="BACKEND_${backend_key}_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown backend: $backend_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
}

LB_NODE="${1:-node1}"
BACKEND_NAME="${2:-backend-1}"

echo "== Raft state =="
bash tooling/demo/state.sh

echo
echo "== LB status for ${LB_NODE} =="
curl -s "$(get_lb_url "$LB_NODE" "/admin/status")" | jq --arg backend "$BACKEND_NAME" '{config: .config, backend: .load_view.backends[$backend]}'

echo
echo "== Backend health for ${BACKEND_NAME} =="
curl -s "$(get_backend_url "$BACKEND_NAME" "/health")" | jq '.'