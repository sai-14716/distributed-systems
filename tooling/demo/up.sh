#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

CLEAN="${1:-}"
if [[ "$CLEAN" == "--clean" ]]; then
  docker compose down -v --remove-orphans
fi

# Machine C only: node4, node5, backend-8, backend-9, backend-10
docker compose up -d --build \
  backend-8 backend-9 backend-10 \
  node4 node5

echo
echo "Machine C is up (node4, node5, backend-8, backend-9, backend-10)."
echo "Control-plane (Raft) ports:"
export PYTHONPATH="$ROOT_DIR"
for n in node4 node5; do
  url=$(python3 tooling/cluster_helper.py get_url nodes "$n")
  echo "  $n ${url}/state"
done
