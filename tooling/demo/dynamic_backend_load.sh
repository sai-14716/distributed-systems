#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Continuously rotate backend load so LB behavior can be observed live.
#
# Usage:
#   bash tooling/demo/dynamic_backend_load.sh [seconds] [phase_seconds] [base_work_ms] [peak_parallel]
#
# Example:
#   bash tooling/demo/dynamic_backend_load.sh 90 3 1800 8

TOTAL_SECS="${1:-60}"
PHASE_SECS="${2:-3}"
BASE_WORK_MS="${3:-1800}"
PEAK_PARALLEL="${4:-8}"

B1_URL="${BACKEND1_URL:-http://127.0.0.1:8081}"
B2_URL="${BACKEND2_URL:-http://127.0.0.1:8082}"
B3_URL="${BACKEND3_URL:-http://127.0.0.1:8083}"

if [[ "$TOTAL_SECS" -le 0 || "$PHASE_SECS" -le 0 || "$BASE_WORK_MS" -le 0 || "$PEAK_PARALLEL" -le 0 ]]; then
  echo "all numeric args must be > 0" >&2
  exit 1
fi

WARM_PARALLEL=$(( PEAK_PARALLEL / 2 ))
if [[ "$WARM_PARALLEL" -lt 1 ]]; then
  WARM_PARALLEL=1
fi

launch_load() {
  local name="$1"
  local url="$2"
  local count="$3"
  local work_ms="$4"
  local phase="$5"

  for i in $(seq 1 "$count"); do
    curl -s -o /dev/null "$url/payload" \
      -H "X-Session-ID: dyn-${name}-p${phase}-r${i}" \
      -H "X-Work-Ms: ${work_ms}" \
      --data-binary "x" &
  done
}

echo "dynamic load start: total=${TOTAL_SECS}s phase=${PHASE_SECS}s base_work_ms=${BASE_WORK_MS} peak_parallel=${PEAK_PARALLEL}"
echo "targets: b1=${B1_URL} b2=${B2_URL} b3=${B3_URL}"

declare -a pids=()
cleanup() {
  for p in "${pids[@]:-}"; do
    kill "$p" 2>/dev/null || true
  done
}
trap cleanup EXIT INT TERM

start_ts="$(date +%s)"
phase=0
while :; do
  now="$(date +%s)"
  elapsed=$(( now - start_ts ))
  if [[ "$elapsed" -ge "$TOTAL_SECS" ]]; then
    break
  fi

  hot=$(( phase % 3 ))
  c1="$WARM_PARALLEL"
  c2="$WARM_PARALLEL"
  c3="$WARM_PARALLEL"
  if [[ "$hot" -eq 0 ]]; then
    c1="$PEAK_PARALLEL"
  elif [[ "$hot" -eq 1 ]]; then
    c2="$PEAK_PARALLEL"
  else
    c3="$PEAK_PARALLEL"
  fi

  # Add slight per-phase jitter so bucket movement is easier to spot.
  work_ms=$(( BASE_WORK_MS + (phase % 4) * 200 ))

  echo "[$(date +%H:%M:%S)] phase=${phase} hot=backend-$((hot+1)) work_ms=${work_ms} parallel: b1=${c1} b2=${c2} b3=${c3}"

  launch_load "b1" "$B1_URL" "$c1" "$work_ms" "$phase"
  launch_load "b2" "$B2_URL" "$c2" "$work_ms" "$phase"
  launch_load "b3" "$B3_URL" "$c3" "$work_ms" "$phase"

  wait

  phase=$(( phase + 1 ))
  sleep "$PHASE_SECS"
done

echo "dynamic load done"
