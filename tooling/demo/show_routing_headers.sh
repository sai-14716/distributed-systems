#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

URL="${1:-http://127.0.0.1:8001/chat}"
N="${2:-5}"

echo "url=$URL n=$N"
for i in $(seq 1 "$N"); do
  echo "--- request $i ---"
  curl -s -D - -o /dev/null "$URL" \
    -H "X-Session-ID: demo-$i" \
    -H "X-Work-Ms: 50" \
    | awk -F': ' '
      tolower($1) == "x-lb-algorithm" {print $0}
      tolower($1) == "x-lb-pool-source" {print $0}
      tolower($1) == "x-lb-pool" {print $0}
      tolower($1) == "x-lb-candidates" {print $0}
      tolower($1) == "x-lb-backend" {print $0}
      tolower($1) == "x-lb-chosen-active" {print $0}
      tolower($1) == "x-lb-chosen-cpu-bucket" {print $0}
    ' | tr -d '\r'
  if [[ "$i" -lt "$N" ]]; then
    sleep 0.5
  fi
done

