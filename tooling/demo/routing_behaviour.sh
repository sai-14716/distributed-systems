#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
export PYTHONPATH="$ROOT_DIR"

ALGO="${1:-}"
ITERATIONS="${2:-5}"
SESSION_ID="${3:-demo-fixed}"
SESSION_MODE="${4:-fixed}" # fixed | per_request

if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm> [iterations] [session-id] [fixed|per_request]" >&2
  exit 1
fi

LB_BASE_URL="${LB_BASE_URL:-$(python3 tooling/cluster_helper.py get_url load_balancers node4)}"

docker compose --profile tools run --rm admin -algorithm "$ALGO" -timeout 30s

echo "algorithm=$ALGO session_id=$SESSION_ID"
for _ in $(seq 1 "$ITERATIONS"); do
  sid="$SESSION_ID"
  if [[ "$SESSION_MODE" == "per_request" ]]; then
    sid="${SESSION_ID}-${_}"
  fi
  backend=$(
    curl -s -D - -o /dev/null -H "X-Session-ID: $sid" "$LB_BASE_URL/chat" \
      | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r'
  )
  if [[ -n "$backend" ]]; then
    echo "$backend"
  else
    echo "(missing X-Lb-Backend)"
  fi
done
