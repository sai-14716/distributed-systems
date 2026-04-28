#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

source tooling/helper/cluster_config.sh

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