#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

NODE_NAME="${1:-}"
if [[ -z "$NODE_NAME" ]]; then
  echo "usage: $0 <node-name>" >&2
  exit 1
fi

docker compose up -d "$NODE_NAME" >/dev/null