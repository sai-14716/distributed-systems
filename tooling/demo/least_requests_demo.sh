#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

# Use first LB (node1) by default, or override with argument
LB_NODE="${1:-node1}"
LB_URL="$(get_lb_url "$LB_NODE")"

submit_config "least-req" 1000 0.8 >/dev/null 2>&1

tmpdir="$(mktemp -d)"
hdr_a="$tmpdir/a.hdr"
hdr_b="$tmpdir/b.hdr"

echo "Starting a long request A (keeps one backend's active-request count at 1)..."
curl -s -D "$hdr_a" -o /dev/null "$LB_URL/chat" -H "X-Session-ID: demo-A" -H "X-Work-Ms: 5000" &
pid=$!

sleep 0.2

echo "Sending short request B while A is in-flight (least-req should prefer the backend with fewer active requests)..."
curl -s -D "$hdr_b" -o /dev/null "$LB_URL/chat" -H "X-Session-ID: demo-B" -H "X-Work-Ms: 50"

wait "$pid" || true

backend_a="$(awk 'tolower($1) == "x-lb-backend:" {print $2}' "$hdr_a" | tr -d '\r')"
backend_b="$(awk 'tolower($1) == "x-lb-backend:" {print $2}' "$hdr_b" | tr -d '\r')"
pool_a="$(awk 'tolower($1) == "x-lb-candidates:" {print $2}' "$hdr_a" | tr -d '\r')"
pool_b="$(awk 'tolower($1) == "x-lb-candidates:" {print $2}' "$hdr_b" | tr -d '\r')"

echo "A backend=$backend_a"
echo "A candidates=$pool_a"
echo "B backend=$backend_b"
echo "B candidates=$pool_b"
if [[ -n "$backend_a" && -n "$backend_b" && "$backend_a" != "$backend_b" ]]; then
  echo "OK: B avoided A's backend while A was in-flight"
else
  echo "NOTE: did not observe avoidance (try again; depends on timing / pool size / health)"
fi

