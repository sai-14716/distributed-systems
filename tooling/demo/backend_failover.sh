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
PY
}
eval "$(_load_config)"

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

BACKEND_NAME="${1:-backend-2}"
shift || true

LB_NODES=("$@")
if [[ $# -eq 0 ]]; then
  LB_NODES=(node1 node2 node3)
fi

echo "Stopping ${BACKEND_NAME}..."
docker compose stop "$BACKEND_NAME" >/dev/null

for lb_node in "${LB_NODES[@]}"; do
  url="$(get_lb_url "$lb_node" "/chat")"
  code="$(curl -s -o /dev/null -w '%{http_code}' "$url")"
  echo "${lb_node}: ${code}"
done

echo "Recovering ${BACKEND_NAME}..."
docker compose up -d "$BACKEND_NAME" >/dev/null