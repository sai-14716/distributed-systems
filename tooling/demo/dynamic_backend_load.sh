#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config into env variables
_load_config() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml
import sys
cfg = yaml.safe_load(open(sys.argv[1]))
for laptop, ip in cfg.get('laptops', {}).items():
  print(f"export LAPTOP_{laptop}={ip}")
for backend_name, backend_info in cfg.get('backends', {}).items():
  backend_key = ''.join(ch if (ch.isalnum() or ch == '_') else '_' for ch in backend_name)
  print(f"export BACKEND_{backend_key}_LAPTOP={backend_info.get('laptop', '')}")
  print(f"export BACKEND_{backend_key}_PORT={backend_info.get('port', '')}")
PY
}
eval "$(_load_config)"

# Helper to convert backend names to shell-safe variable keys
_to_var_key() {
  local input="$1"
  echo "${input//[^a-zA-Z0-9_]/_}"
}

# Resolve backend URL
get_backend_url() {
  local backend_name="$1" endpoint="${2:-}"
  local backend_key
  backend_key="$(_to_var_key "$backend_name")"
  local laptop_var="BACKEND_${backend_key}_LAPTOP" port_var="BACKEND_${backend_key}_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown backend: $backend_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
}

# Get all backend names from config
get_all_backends() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml
import sys
cfg = yaml.safe_load(open(sys.argv[1]))
for backend_name in sorted(cfg.get('backends', {}).keys()):
  print(backend_name)
PY
}

# Continuously rotate backend load so LB behavior can be observed live.
#
# Usage:
#   bash tooling/demo/dynamic_backend_load.sh [seconds] [phase_seconds] [base_work_ms] [peak_parallel]
#
# Uses backends from cluster_config.yaml

TOTAL_SECS="${1:-60}"
PHASE_SECS="${2:-0.5}"
BASE_WORK_MS="${3:-1800}"
PEAK_PARALLEL="${4:-8}"

if [[ "$TOTAL_SECS" -le 0 || "$PHASE_SECS" -le 0 || "$BASE_WORK_MS" -le 0 || "$PEAK_PARALLEL" -le 0 ]]; then
  echo "all numeric args must be > 0" >&2
  exit 1
fi

WARM_PARALLEL=$(( PEAK_PARALLEL / 2 ))
if [[ "$WARM_PARALLEL" -lt 1 ]]; then
  WARM_PARALLEL=1
fi

# Get all backends from config
declare -a BACKEND_URLS
for backend_name in $(get_all_backends); do
  BACKEND_URLS+=("$(get_backend_url "$backend_name")")
done

backend_count="${#BACKEND_URLS[@]}"
if [[ "$backend_count" -le 0 ]]; then
  echo "no backends configured in cluster_config.yaml" >&2
  exit 1
fi

launch_load() {
  local backend_name="$1"
  local url="$2"
  local count="$3"
  local work_ms="$4"
  local phase="$5"

  for i in $(seq 1 "$count"); do
    curl -s -o /dev/null "$url/payload" \
      -H "X-Session-ID: dyn-${backend_name}-p${phase}-r${i}" \
      -H "X-Work-Ms: ${work_ms}" \
      --data-binary "x" &
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
