#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
. "$ROOT_DIR/tooling/lib/cluster_config.sh"

# Continuously rotate backend load so LB behavior can be observed live.
#
# Usage:
#   bash tooling/demo/dynamic_backend_load.sh [seconds] [phase_seconds] [base_work_ms] [peak_parallel]
#
# Optional env overrides:
#   BACKEND_URLS="http://<backend-ip>:<backend-port>,http://<backend-ip>:<backend-port>,..."

TOTAL_SECS="${1:-60}"
PHASE_SECS="${2:-0.5}"
BASE_WORK_MS="${3:-1800}"
PEAK_PARALLEL="${4:-8}"
BACKEND_URLS_CSV="${BACKEND_URLS:-}"

if [[ "$TOTAL_SECS" -le 0 || "$PHASE_SECS" -le 0 || "$BASE_WORK_MS" -le 0 || "$PEAK_PARALLEL" -le 0 ]]; then
  echo "all numeric args must be > 0" >&2
  exit 1
fi
WARM_PARALLEL=$(( PEAK_PARALLEL / 2 ))
if [[ "$WARM_PARALLEL" -lt 1 ]]; then
  WARM_PARALLEL=1
fi

declare -a BACKEND_URLS
if [[ -n "$BACKEND_URLS_CSV" ]]; then
  IFS=',' read -r -a BACKEND_URLS <<<"$BACKEND_URLS_CSV"
else
  IFS=',' read -r -a BACKEND_URLS <<<"$(cluster_csv_urls backends)"
fi

backend_count="${#BACKEND_URLS[@]}"
if [[ "$backend_count" -le 0 ]]; then
  echo "no backends configured" >&2
  exit 1
fi

launch_load() {
  local backend_name="$1"
  local url="$2"
  local count="$3"
  local work_ms="$4"
  local phase="$5"
  local timeout_secs=$(( (work_ms + 999) / 1000 + 1 ))

  for i in $(seq 1 "$count"); do
    ( curl -s --max-time "$timeout_secs" -o /dev/null "$url/payload" \
      -H "X-Session-ID: dyn-${backend_name}-p${phase}-r${i}" \
      -H "X-Work-Ms: ${work_ms}" \
      --data-binary "x" || true ) &
  done
}

echo "dynamic load start: total=${TOTAL_SECS}s phase=${PHASE_SECS}s base_work_ms=${BASE_WORK_MS} peak_parallel=${PEAK_PARALLEL}"
echo "targets (${backend_count}): ${BACKEND_URLS[*]}"

start_ts="$(date +%s)"
phase=0
while :; do
  now="$(date +%s)"
  elapsed=$(( now - start_ts ))
  if [[ "$elapsed" -ge "$TOTAL_SECS" ]]; then
    break
  fi

  hot_idx=$(( phase % backend_count ))
  work_ms=$(( BASE_WORK_MS + (phase % 4) * 200 ))

  echo "[$(date +%H:%M:%S)] phase=${phase} hot=backend-$((hot_idx + 1)) work_ms=${work_ms}"

  for idx in "${!BACKEND_URLS[@]}"; do
    url="${BACKEND_URLS[$idx]}"
    parallel="$WARM_PARALLEL"
    if [[ "$idx" -eq "$hot_idx" ]]; then
      parallel="$PEAK_PARALLEL"
    fi
    echo "  backend-$((idx + 1)) parallel=${parallel} url=${url}"
    launch_load "b$((idx + 1))" "$url" "$parallel" "$work_ms" "$phase"
  done

  wait

  phase=$(( phase + 1 ))
  sleep "$PHASE_SECS"
done

echo "dynamic load done"
