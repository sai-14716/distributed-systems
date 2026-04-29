#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
. "$ROOT_DIR/tooling/lib/cluster_config.sh"

ALGO="${1:-}"
ITERATIONS="${2:-5}"
SESSION_ID="${3:-demo-fixed}"
SESSION_MODE="${4:-fixed}" # fixed | per_request
ROUTE_MODE="${5:-dns}"  # direct | dns

if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm|--no-config|-> [iterations] [session-id] [fixed|per_request] [direct|dns]" >&2
  exit 1
fi

if [[ "$ALGO" != "--no-config" && "$ALGO" != "-" ]]; then
  echo "$ROOT_DIR/tooling/demo/apply_algorithm.sh" "$ALGO"
  bash "$ROOT_DIR/tooling/demo/apply_algorithm.sh" "$ALGO"
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

# Define the pool of endpoints
ENDPOINTS=("/chat" "/encrypt" "/payload" "/health")

tmp_body="$(mktemp)"
trap 'rm -f "$tmp_body"' EXIT

echo "algorithm=$ALGO session_id=$SESSION_ID mode=$ROUTE_MODE lb_base=$LB_BASE"

for _ in $(seq 1 "$ITERATIONS"); do
  sid="$SESSION_ID"
  if [[ "$SESSION_MODE" == "per_request" ]]; then
    sid="${SESSION_ID}-${_}"
  fi
  
  # Pick a random endpoint for this specific request
  RANDOM_INDEX=$((RANDOM % ${#ENDPOINTS[@]}))
  TARGET_ENDPOINT="${ENDPOINTS[$RANDOM_INDEX]}"
  TARGET_URL="${LB_BASE%/}${TARGET_ENDPOINT}"

  response=$(
    { curl -sS --max-time 2 "${CURL_PROXY_ARGS[@]}" -D - -o "$tmp_body" -H "X-Session-ID: $sid" "$TARGET_URL" || true; } 2>&1
  )
  backend=$(
    printf '%s\n' "$response" | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r'
  )
  lb_node=$(
    printf '%s\n' "$response" | awk 'tolower($1) == "x-lb-node:" {print $2}' | tr -d '\r'
  )
  
  if [[ -n "$backend" ]]; then
    if [[ -n "$lb_node" ]]; then
      echo "[$TARGET_ENDPOINT] [LB=$lb_node] -> $backend"
    else
      echo "[$TARGET_ENDPOINT] $backend"
    fi
  else
    status="$(printf '%s\n' "$response" | awk 'toupper($1) == "HTTP/1.1" || toupper($1) == "HTTP/2" {print $2; exit}')"
    body="$(tr '\n' ' ' < "$tmp_body" | sed 's/[[:space:]]*$//')"
    if [[ -n "$body" ]]; then
      echo "[$TARGET_ENDPOINT] (missing X-LB-Backend status=${status:-curl-error} body=$body)"
    else
      echo "[$TARGET_ENDPOINT] (missing X-LB-Backend status=${status:-curl-error})"
    fi
  fi
done