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

# Use first LB (node1) by default, or override with argument
LB_NODE="${1:-node1}"
ENDPOINT="${2:-/chat}"
N="${3:-5}"

# Resolve the LB URL from config
LB_URL=$(get_lb_url "$LB_NODE" "$ENDPOINT")

echo "Testing LB routing headers against: $LB_URL"
echo "Requests: $N"
echo

for i in $(seq 1 "$N"); do
  echo "--- request $i ---"
  curl -s -D - -o /dev/null "$LB_URL" \
    -H "X-Session-ID: demo-$i" \
    -H "X-Work-Ms: 50" \
    | awk -F': ' '
      tolower($1) == "x-lb-algorithm" {print $0}
      tolower($1) == "x-lb-pool-source" {print $0}
      tolower($1) == "x-lb-pool" {print $0}
      tolower($1) == "x-lb-candidates" {print $0}
      tolower($1) == "x-lb-backend" {print $0}
      tolower($1) == "x-lb-chosen-active" {print $0}
      tolower($1) == "x-lb-chosen-cpu-bucket" {print $0}
    ' | tr -d '\r'
  if [[ "$i" -lt "$N" ]]; then
    sleep 0.5
  fi
done

