#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

echo "Starting Machine C services (nodes: 4, 5 | backends: 8, 9, 10)..."
docker compose down -v
docker compose up -d --build
echo "Machine C is running."
