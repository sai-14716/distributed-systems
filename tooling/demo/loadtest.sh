#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

script="${1:-/app/k6/stress-test.js}"
shift || true

docker compose --profile loadtest run --rm \
  -e K6_SCRIPT="$script" \
  "$@" \
  k6-client
