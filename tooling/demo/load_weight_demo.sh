#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

ALGO="${1:-wrr}"           # wrr | least-load
LB_URL="${2:-http://127.0.0.1:8001}"
N="${3:-30}"
HOT_WORK_MS="${4:-5000}"
HOT_REQUESTS="${5:-8}"

docker compose --profile tools run --rm admin -algorithm "$ALGO" -timeout 30s >/dev/null

print_load_snapshot() {
  python3 - "$LB_URL" <<'PY'
import json
import sys
import urllib.request

url = sys.argv[1].rstrip("/") + "/admin/load-view"
with urllib.request.urlopen(url, timeout=1.5) as response:
  data = json.loads(response.read().decode("utf-8"))

backends = data.get("backends", {})
for backend_id in sorted(backends):
  entry = backends.get(backend_id) or {}
  view = entry.get("view") or {}
  local_probe = entry.get("local_probe") or {}
  print(
    f"  {backend_id}: view_bucket={view.get('bucket', 'na')} "
    f"view_active={view.get('active', 'na')} view_status={view.get('status', 'na')} "
    f"owner={view.get('owner_id', 'na')} probes={local_probe.get('total', 'na')} fail={local_probe.get('fail', 'na')}"
  )
PY
}

echo "Warming up load on backend-3 via /payload (sticky rule routes payload -> backend-3)..."
heat_pids=()
for i in $(seq 1 "$HOT_REQUESTS"); do
  curl -s -o /dev/null "$LB_URL/payload" \
    -H "X-Session-ID: warm-$i" \
    -H "X-Work-Ms: $HOT_WORK_MS" \
    --data-binary "x" &
  heat_pids+=("$!")
done

echo "Waiting a bit for LBs to observe CPU bucket changes..."
sleep 3

observed_bucket="$(curl -s "$LB_URL/admin/load-view" | python3 -c 'import json, sys; data = json.load(sys.stdin); backend = data.get("backends", {}).get("backend-3", {}); view = backend.get("view") or {}; print(view.get("bucket", "na"))')"
echo "observed backend-3 cpu bucket=${observed_bucket} (5% steps; higher means hotter)"
echo "current load view:"
print_load_snapshot

echo "Sending $N requests to / (uses all backends) with algo=$ALGO"
tmp="$(mktemp)"
for i in $(seq 1 "$N"); do
  curl -s -D - -o /dev/null "$LB_URL/" -H "X-Session-ID: demo-$i" -H "X-Work-Ms: 50" \
    | awk 'tolower($1) == "x-lb-backend:" {print $2}' | tr -d '\r' >>"$tmp"
done

for pid in "${heat_pids[@]}"; do
  wait "$pid" || true
done

echo "backend distribution:"
sort "$tmp" | uniq -c | sort -nr

echo "load view after routing burst:"
print_load_snapshot

echo "Tip: inspect one request's candidates/CPU buckets with:"
echo "  sh tooling/demo/show_routing_headers.sh $LB_URL/ 1"

