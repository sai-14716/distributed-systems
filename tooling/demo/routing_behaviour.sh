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

# Note: admin tool runs locally in docker, so it still uses localhost colons
# TODO: Update admin tool to accept IP:port for distributed setup
echo "Configuring algorithm on local LB control endpoints..."
echo "(This requires admin tool update for distributed mode)"

# Get the LB URL for testing
LB_URL=$(get_lb_url "$LB_NODE" "/chat")

echo "Testing algorithm=$ALGO on $LB_NODE with LB_URL=$LB_URL"
for i in $(seq 1 "$ITERATIONS"); do
  sid="$SESSION_ID"
  if [[ "$SESSION_MODE" == "per_request" ]]; then
    sid="${SESSION_ID}-${i}"
  fi
  backend=$(
    curl -s -D - -o /dev/null -H "X-Session-ID: $sid" "$LB_URL" \
      | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r'
  )
  if [[ -n "$backend" ]]; then
    echo "$backend"
  else
    echo "(missing X-Lb-Backend)"
  fi
done
