#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

_load_config() {
  python3 - "$ROOT_DIR/cluster_config.yaml" <<'PY'
import yaml, sys
cfg = yaml.safe_load(open(sys.argv[1]))
for laptop, ip in cfg.get('laptops', {}).items():
  print(f"export LAPTOP_{laptop}={ip}")
for node_name, node_info in cfg.get('nodes', {}).items():
  print(f"export NODE_{node_name}_LAPTOP={node_info.get('laptop', '')}")
  print(f"export NODE_{node_name}_RAFT_PORT={node_info.get('raft_port', '')}")
PY
}
eval "$(_load_config)"

submit_config() {
  eval "curl -sS -X POST http://\${LAPTOP_\${NODE_${4:-node1}_LAPTOP}}:\${NODE_${4:-node1}_RAFT_PORT}/admin/submit -H 'Content-Type: application/json' -d '{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$1\",\"probe_interval_ms\":${2:-1000},\"health_threshold\":${3:-0.8}}}' -w '\\nStatus: %{http_code}\\n' 2>&1"
}

DRILL_ID="${1:-chaos-$(date +%Y%m%dT%H%M%S)}"
CONFIG_ALGO_1="${CONFIG_ALGO_1:-wrr}"
CONFIG_ALGO_2="${CONFIG_ALGO_2:-least-req}"
STOP_BACKEND="${STOP_BACKEND:-backend-2}"
RUN_LOAD="${RUN_LOAD:-1}"
TRACE_OUT_DIR="${TRACE_OUT_DIR:-$ROOT_DIR/logs}"
DRILL_DIR="$TRACE_OUT_DIR/$DRILL_ID"

mkdir -p "$DRILL_DIR"

LEADER_BEFORE=""
LOAD_PID=""

log() {
  printf '[chaos-drill] %s\n' "$*"
}

cleanup() {
  if [[ -n "$LOAD_PID" ]] && kill -0 "$LOAD_PID" 2>/dev/null; then
    log "stopping background load process"
    kill "$LOAD_PID" 2>/dev/null || true
    wait "$LOAD_PID" 2>/dev/null || true
  fi

  if [[ -n "$LEADER_BEFORE" ]]; then
    docker compose up -d "$LEADER_BEFORE" >/dev/null 2>&1 || true
  fi

  docker compose up -d "$STOP_BACKEND" >/dev/null 2>&1 || true
}

trap cleanup EXIT

find_leader() {
  local nodes=(node1 node2)
  local state
  for _ in {1..30}; do
    for n in "${nodes[@]}"; do
      state="$(docker compose exec -T "$n" sh -lc "curl -s http://127.0.0.1:19090/state" 2>/dev/null || true)"
      if [[ "$state" == *'"role":"leader"'* ]]; then
        printf '%s' "$n"
        return 0
      fi
    done
    sleep 0.3
  done
  return 1
}

log "bringing up Laptop A cluster"
docker compose up -d --build discovery backend-1 backend-2 backend-3 node1 node2 \
  >"$DRILL_DIR/cluster_up.log" 2>&1

LEADER_BEFORE="$(find_leader || true)"
if [[ -z "$LEADER_BEFORE" ]]; then
  log "could not determine current leader"
  exit 1
fi
log "detected current leader: $LEADER_BEFORE"

if [[ "$RUN_LOAD" == "1" ]]; then
  log "starting background realistic load"
  (
    docker compose --profile loadtest run --rm \
      -e DISABLE_PROXY=1 \
      -e K6_SCRIPT=/app/k6/test_realistic_load.js \
      -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000,http://node4:8000,http://node5:8000 \
      k6-client
  ) >"$DRILL_DIR/load.log" 2>&1 &
  LOAD_PID="$!"
fi

log "sending config update while dropping leader"
(
  submit_config "$CONFIG_ALGO_1" 700 0.8
) >"$DRILL_DIR/config_update_during_leader_drop.log" 2>&1 &
ADMIN_PID="$!"

sleep 0.2
docker compose stop "$LEADER_BEFORE" >"$DRILL_DIR/leader_stop.log" 2>&1

if wait "$ADMIN_PID"; then
  log "config update command during leader drop completed"
else
  log "config update command during leader drop failed (expected in some timing windows)"
fi

NEW_LEADER="$(find_leader || true)"
if [[ -z "$NEW_LEADER" ]]; then
  log "new leader was not elected in time"
  exit 1
fi
log "new leader detected: $NEW_LEADER"

log "running post-failover config update for convergence"
submit_config "$CONFIG_ALGO_2" 900 0.8 \
  >"$DRILL_DIR/config_update_after_failover.log" 2>&1

log "dropping backend: $STOP_BACKEND"
docker compose stop "$STOP_BACKEND" >"$DRILL_DIR/backend_stop.log" 2>&1

TRACE_ID="$DRILL_ID-backend-drop"
log "running traced request with backend down: $TRACE_ID"
TRACE_BUILD_FLAG="" ./tooling/traces/request_trace.sh "$TRACE_ID" >"$DRILL_DIR/trace_request.log" 2>&1

if [[ -n "$LOAD_PID" ]]; then
  wait "$LOAD_PID" || true
fi

log "drill complete"
log "artifacts: $DRILL_DIR"
log "trace logs: $TRACE_OUT_DIR/$TRACE_ID"
