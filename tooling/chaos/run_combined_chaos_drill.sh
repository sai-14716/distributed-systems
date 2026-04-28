#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
. "$ROOT_DIR/tooling/lib/cluster_config.sh"

DRILL_ID="${1:-chaos-$(date +%Y%m%dT%H%M%S)}"
TARGETS="${TARGETS:-$(cluster_csv_urls nodes)}"
CONFIG_ALGO_1="${CONFIG_ALGO_1:-wrr}"
CONFIG_ALGO_2="${CONFIG_ALGO_2:-least-req}"
STOP_BACKEND="${STOP_BACKEND:-$(python3 - <<'PY'
import json

with open("cluster_config.yaml", encoding="utf-8") as f:
    cfg = json.load(f)

backends = list((cfg.get("backends") or {}).keys())
print(backends[1] if len(backends) > 1 else (backends[0] if backends else ""))
PY
)}"
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

  if [[ -n "$STOP_BACKEND" ]]; then
    docker compose up -d "$STOP_BACKEND" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT

find_leader() {
  python3 - <<'PY'
import json
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed

with open("cluster_config.yaml") as f:
    cfg = json.load(f)

laptops = cfg["laptops"]

def resolve(service_type, name):
    info = cfg[service_type][name]
    ip = laptops[info["laptop"]]
    return ip, info

def fetch_role(name):
    ip, info = resolve("nodes", name)
    try:
        with urllib.request.urlopen(f"http://{ip}:{info['port']}/state", timeout=0.6) as r:
            state = json.loads(r.read().decode("utf-8"))
        return name, state.get("role")
    except Exception:
        return name, None

nodes = list(cfg["nodes"].keys())
for _ in range(30):
    with ThreadPoolExecutor(max_workers=min(16, max(1, len(nodes)))) as pool:
        futures = [pool.submit(fetch_role, node) for node in nodes]
        for fut in as_completed(futures):
            node, role = fut.result()
            if role == "leader":
                print(node)
                raise SystemExit(0)
    time.sleep(0.3)
raise SystemExit(1)
PY
}

log "bringing up cluster"
configured_services="$(python3 - <<'PY'
import json

with open("cluster_config.yaml", encoding="utf-8") as f:
    cfg = json.load(f)

services = ["discovery"]
services.extend((cfg.get("backends") or {}).keys())
services.extend((cfg.get("nodes") or {}).keys())
print(" ".join(services))
PY
)"
available_services="$(docker compose config --services)"
services_to_start=()
for service in $configured_services; do
  if printf '%s\n' "$available_services" | grep -Fxq "$service"; then
    services_to_start+=("$service")
  fi
done

if [[ "${#services_to_start[@]}" -eq 0 ]]; then
  log "no configured cluster services are present in this compose file"
  exit 1
fi

docker compose up -d --build "${services_to_start[@]}" \
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
      -e LB_BASE_URLS="$(cluster_csv_urls load_balancers)" \
      k6-client
  ) >"$DRILL_DIR/load.log" 2>&1 &
  LOAD_PID="$!"
fi

log "sending config update while dropping leader"
(
  docker compose --profile tools run --rm admin \
    -targets "$TARGETS" \
    -algorithm "$CONFIG_ALGO_1" \
    -probe-interval-ms 700
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
docker compose --profile tools run --rm admin \
  -targets "$TARGETS" \
  -algorithm "$CONFIG_ALGO_2" \
  -probe-interval-ms 900 \
  >"$DRILL_DIR/config_update_after_failover.log" 2>&1

log "dropping backend: $STOP_BACKEND"
if [[ -n "$STOP_BACKEND" ]]; then
  docker compose stop "$STOP_BACKEND" >"$DRILL_DIR/backend_stop.log" 2>&1
else
  log "no backend configured to drop"
fi

TRACE_ID="$DRILL_ID-backend-drop"
log "running traced request with backend down: $TRACE_ID"
TRACE_BUILD_FLAG="" ./tooling/traces/request_trace.sh "$TRACE_ID" >"$DRILL_DIR/trace_request.log" 2>&1

if [[ -n "$LOAD_PID" ]]; then
  wait "$LOAD_PID" || true
fi

log "drill complete"
log "artifacts: $DRILL_DIR"
log "trace logs: $TRACE_OUT_DIR/$TRACE_ID"
