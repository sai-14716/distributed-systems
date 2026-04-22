#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

CLEAN="${1:-}"
if [[ "$CLEAN" == "--clean" ]]; then
  docker compose down -v --remove-orphans
fi

docker compose up -d --build \
  discovery backend-1 backend-2 backend-3 \
  node1 node2 node3 node4 node5

echo
echo "Cluster is up."
echo "Control-plane (Raft) ports on host:"
echo "  node1 http://127.0.0.1:19091/state"
echo "  node2 http://127.0.0.1:19092/state"
echo "  node3 http://127.0.0.1:19093/state"
echo "  node4 http://127.0.0.1:19094/state"
echo "  node5 http://127.0.0.1:19095/state"

