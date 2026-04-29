#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

echo "Building demo images..."
docker compose build
docker compose --profile tools build admin
docker compose --profile loadtest build k6-client