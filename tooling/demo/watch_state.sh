#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

while true; do
  date
  bash tooling/demo/state.sh
  echo
  sleep "${DEMO_POLL_SECS:-1}"
done
