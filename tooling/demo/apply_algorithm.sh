#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

ALGO="${1:-}"
if [[ -z "$ALGO" ]]; then
  echo "usage: $0 <algorithm>" >&2
  exit 1
fi

docker compose --profile tools run --rm admin -algorithm "$ALGO" -timeout 30s >/dev/null