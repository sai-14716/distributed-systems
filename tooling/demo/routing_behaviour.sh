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

# Resolve node URL from cluster config
get_node_url() {
  local node_name="$1" endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP" port_var="NODE_${node_name}_RAFT_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown node: $node_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
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

# Submit config to Raft cluster
submit_config() {
  local algorithm="${1:-}"
  [[ -z "$algorithm" ]] && { echo "ERROR: algorithm required" >&2; return 1; }
  local probe_interval_ms="${2:-1000}" health_threshold="${3:-0.8}" target_node="${4:-node1}"
  local target_url=$(get_node_url "$target_node" "/admin/submit") || return 1
  echo "[config] Sending to $target_node: algorithm=$algorithm" >&2
  curl -sS -X POST "$target_url" -H 'Content-Type: application/json' \
    -d "{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$algorithm\",\"probe_interval_ms\":$probe_interval_ms,\"health_threshold\":$health_threshold}}" \
    -w "\nStatus: %{http_code}\n" 2>&1
}

ALGO="${1:-}"
ITERATIONS="${2:-5}"
SESSION_ID="${3:-demo-fixed}"
SESSION_MODE="${4:-fixed}" # fixed | per_request
ROUTE_MODE="${5:-direct}"  # direct | dns

if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm|--no-config|-> [iterations] [session-id] [fixed|per_request] [direct|dns]" >&2
  exit 1
fi

echo "Configuring algorithm via Raft submit..."
if ! submit_config "$ALGO" 1000 0.8 >/dev/null 2>&1; then
  echo "(algorithm update skipped: failed to reach leader)" >&2
fi

LB_BASE=""
if [[ "$ROUTE_MODE" == "dns" ]]; then
  LB_BASE="http://api.service.com:8000"
  CURL_PROXY_ARGS=("--proxy" "${DISCOVERY_PROXY_URL:-http://127.0.0.1:6700}")
else
  CURL_PROXY_ARGS=()
  for lb in $(cluster_names load_balancers); do
    candidate="$(cluster_url load_balancers "$lb")"
    if curl -sf --max-time 1 "$candidate/admin/status" >/dev/null; then
      LB_BASE="$candidate"
      break
    fi
  done
  if [[ -z "$LB_BASE" ]]; then
    LB_BASE="$(cluster_first_url load_balancers)"
  fi
fi

LB_URL="${LB_BASE%/}/chat"
tmp_body="$(mktemp)"
trap 'rm -f "$tmp_body"' EXIT

echo "algorithm=$ALGO session_id=$SESSION_ID mode=$ROUTE_MODE lb=$LB_URL"
for _ in $(seq 1 "$ITERATIONS"); do
  sid="$SESSION_ID"
  if [[ "$SESSION_MODE" == "per_request" ]]; then
    sid="${SESSION_ID}-${_}"
  fi
  response=$(
    { curl -sS --max-time 2 "${CURL_PROXY_ARGS[@]}" -D - -o "$tmp_body" -H "X-Session-ID: $sid" "$LB_URL" || true; } 2>&1
  )
  backend=$(
    printf '%s\n' "$response" | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r'
  )
  lb_node=$(
    printf '%s\n' "$response" | awk 'tolower($1) == "x-lb-node:" {print $2}' | tr -d '\r'
  )
  if [[ -n "$backend" ]]; then
    if [[ -n "$lb_node" ]]; then
      echo "[LB=$lb_node] -> $backend"
    else
      echo "$backend"
    fi
  else
    status="$(printf '%s\n' "$response" | awk 'toupper($1) == "HTTP/1.1" || toupper($1) == "HTTP/2" {print $2; exit}')"
    body="$(tr '\n' ' ' < "$tmp_body" | sed 's/[[:space:]]*$//')"
    if [[ -n "$body" ]]; then
      echo "(missing X-LB-Backend status=${status:-curl-error} body=$body)"
    else
      echo "(missing X-LB-Backend status=${status:-curl-error})"
    fi
  fi
done