#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

LB_NODE="${1:-node1}"
LB_URL="$(get_lb_url "$LB_NODE")"
REQUESTS="${2:-15}"
LOAD_SECS="${3:-60}"
PHASE_SECS="${4:-1.5}"
BASE_WORK_MS="${5:-1800}"
PEAK_PARALLEL="${6:-8}"

if [[ "$REQUESTS" -le 0 || "$LOAD_SECS" -le 0 ]]; then
  echo "REQUESTS and LOAD_SECS must be > 0" >&2
  exit 1
fi

echo "setting algorithm to least-req"
submit_config "least-req" 1000 0.8 >/dev/null 2>&1

echo "starting dynamic backend load in background"
bash tooling/demo/dynamic_backend_load.sh "$LOAD_SECS" "$PHASE_SECS" "$BASE_WORK_MS" "$PEAK_PARALLEL" &
load_pid=$!

cleanup() {
  if kill -0 "$load_pid" 2>/dev/null; then
    kill "$load_pid" 2>/dev/null || true
    wait "$load_pid" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

sleep 1

echo "showing $REQUESTS routed requests while dynamic load is running"
bash tooling/demo/show_routing_headers.sh "$LB_NODE" "/" "$REQUESTS"

echo "least-req + dynamic load demo done"
