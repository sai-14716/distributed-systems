#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
. "$ROOT_DIR/tooling/lib/cluster_config.sh"

ALGO="${1:-wrr}"           # wrr | least-load
LB_URL="${2:-$(cluster_first_url load_balancers)}"
N="${3:-30}"
HOT_WORK_MS="${4:-5000}"
HOT_REQUESTS="${5:-8}"
HOT_BACKEND="${HOT_BACKEND:-$(python3 - <<'PY'
import json

with open("cluster_config.yaml", encoding="utf-8") as f:
    cfg = json.load(f)

backends = list((cfg.get("backends") or {}).keys())
print(backends[2] if len(backends) >= 3 else (backends[0] if backends else ""))
PY
)}"

docker compose --profile tools run --rm admin -targets "$(cluster_csv_urls nodes)" -algorithm "$ALGO" -timeout 30s >/dev/null

print_load_snapshot() {
  python3 - "$LB_URL" <<'PY'
import json
import sys
import urllib.request

url = sys.argv[1].rstrip("/") + "/admin/load-view"
try:
  with urllib.request.urlopen(url, timeout=1.0) as response:
    data = json.loads(response.read().decode("utf-8"))
except Exception as exc:
  print(f"  ERR load-view unavailable: {exc.__class__.__name__}")
  raise SystemExit(0)

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

echo "Warming up load via /payload (observing configured backend views)..."
heat_pids=()
for i in $(seq 1 "$HOT_REQUESTS"); do
  ( curl -s --max-time 6 -o /dev/null "$LB_URL/payload" \
    -H "X-Session-ID: warm-$i" \
    -H "X-Work-Ms: $HOT_WORK_MS" \
    --data-binary "x" || true ) &
  heat_pids+=("$!")
done

echo "Waiting a bit for LBs to observe CPU bucket changes..."
sleep 3

observed_bucket="$({ curl -s --max-time 1 "$LB_URL/admin/load-view" || true; } | HOT_BACKEND="$HOT_BACKEND" python3 -c 'import json, os, sys; data = json.load(sys.stdin); backend = data.get("backends", {}).get(os.environ.get("HOT_BACKEND", ""), {}); view = backend.get("view") or {}; print(view.get("bucket", "na"))' 2>/dev/null || printf 'na')"
echo "observed ${HOT_BACKEND} cpu bucket=${observed_bucket} (5% steps; higher means hotter)"
echo "current load view:"
print_load_snapshot

echo "Sending $N requests to / (uses all backends) with algo=$ALGO"
tmp="$(mktemp)"
for i in $(seq 1 "$N"); do
  { curl -s --max-time 1 -D - -o /dev/null "$LB_URL/" -H "X-Session-ID: demo-$i" -H "X-Work-Ms: 50" || true; } \
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

