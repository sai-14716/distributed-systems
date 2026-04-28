#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

source tooling/helper/cluster_config.sh

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