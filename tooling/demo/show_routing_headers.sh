#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

# Use first LB (node1) by default, or override with argument
LB_NODE="${1:-node1}"
ENDPOINT="${2:-/chat}"
N="${3:-5}"

# Resolve the LB URL from config
LB_URL=$(get_lb_url "$LB_NODE" "$ENDPOINT")

echo "Testing LB routing headers against: $LB_URL"
echo "Requests: $N"
echo

for i in $(seq 1 "$N"); do
  echo "--- request $i ---"
  curl -s -D - -o /dev/null "$LB_URL" \
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

