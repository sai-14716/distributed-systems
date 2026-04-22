#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

LB_NODE="${1:-node1}"
N="${2:-30}"

echo "=== 1) Per-request headers by algorithm ==="
for algo in round_robin maglev least-req wrr least-load; do
  echo
  echo "--- algorithm=$algo ---"
  docker compose --profile tools run --rm admin -algorithm "$algo" -timeout 30s >/dev/null
  bash tooling/demo/show_routing_headers.sh "$LB_NODE" "/chat" 1
done

echo
echo "=== 2) Least-Requests differentiation (in-flight avoidance) ==="
bash tooling/demo/least_requests_demo.sh "$LB_NODE"

echo
echo "=== 3) WRR vs Least-Load under a heated backend ==="
echo "--- wrr ---"
bash tooling/demo/load_weight_demo.sh wrr "$LB_NODE" "$N"

echo
echo "--- least-load ---"
bash tooling/demo/load_weight_demo.sh least-load "$LB_NODE" "$N"

echo
echo "Done. For health probe/propagation visibility, run:"
echo "  bash tooling/demo/watch_health_propagation.sh 20"
