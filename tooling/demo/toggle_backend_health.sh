#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

source tooling/helper/cluster_config.sh

BACKEND_NAME="${1:-backend-3}"
WAIT_SECS="${2:-5}"

echo "Stopping ${BACKEND_NAME}..."
docker compose stop "$BACKEND_NAME" >/dev/null

sleep "$WAIT_SECS"

echo "Starting ${BACKEND_NAME}..."
docker compose up -d "$BACKEND_NAME" >/dev/null