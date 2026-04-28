#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

ALGO="${1:-}"
ITERATIONS="${2:-5}"
SESSION_ID="${3:-demo-fixed}"
SESSION_MODE="${4:-fixed}" # fixed | per_request
LB_NODE="${5:-node1}"       # Which LB node to use

if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm> [iterations] [session-id] [fixed|per_request] [lb-node]" >&2
  exit 1
fi

echo "Configuring algorithm on local LB control endpoints..."
if ! docker compose --profile tools run --rm admin -algorithm "$ALGO" -timeout 30s >/dev/null; then
  echo "(algorithm update skipped: admin command failed)" >&2
fi

# Get the LB URL for testing
LB_URL=$(get_lb_url "$LB_NODE" "/chat")

echo "Testing algorithm=$ALGO on $LB_NODE with LB_URL=$LB_URL"
for i in $(seq 1 "$ITERATIONS"); do
  sid="$SESSION_ID"
  if [[ "$SESSION_MODE" == "per_request" ]]; then
    sid="${SESSION_ID}-${i}"
  fi
  response_headers=$(curl -s -D - -o /dev/null -H "X-Session-ID: $sid" "$LB_URL")
  status=$(printf '%s' "$response_headers" | awk 'toupper($1) == "HTTP/1.1" || toupper($1) == "HTTP/2" {code=$2} END {print code}')
  backend=$(printf '%s' "$response_headers" | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r')
  if [[ -n "$backend" ]]; then
    echo "$backend"
  else
    echo "(missing X-Lb-Backend, status=${status:-unknown})"
  fi
done
