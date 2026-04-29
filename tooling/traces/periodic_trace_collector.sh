#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

INTERVAL_SEC="${1:-30}"
OUT_DIR="${TRACE_OUT_DIR:-$ROOT_DIR/logs/live}"
mkdir -p "$OUT_DIR"

echo "Starting periodic trace collector (interval=${INTERVAL_SEC}s)"
echo "Writing snapshots to: $OUT_DIR"

while true; do
  TS="$(date -u +%Y%m%dT%H%M%SZ)"
  SNAPSHOT_DIR="$OUT_DIR/$TS"
  mkdir -p "$SNAPSHOT_DIR"

  TMP_LINES="$(mktemp)"
  CLIENT_PROXY_FILE="$SNAPSHOT_DIR/client_proxy.log"
  LB_DATA_FILE="$SNAPSHOT_DIR/lb_data.log"
  LB_CONTROL_FILE="$SNAPSHOT_DIR/lb_control.log"
  BACKEND_FILE="$SNAPSHOT_DIR/backend.log"

  docker compose logs --since "${INTERVAL_SEC}s" \
    discovery lb-1 lb-2 lb-3 node1 node2 node3 node4 node5 backend-1 backend-2 backend-3 2>&1 \
    | grep -E '\[client-proxy\]|\[lb-data\]|\[lb-control\]|\[backend\]' > "$TMP_LINES" || true

  grep -E '\[client-proxy\]' "$TMP_LINES" > "$CLIENT_PROXY_FILE" || true
  grep -E '\[lb-data\]' "$TMP_LINES" > "$LB_DATA_FILE" || true
  grep -E '\[lb-control\]' "$TMP_LINES" > "$LB_CONTROL_FILE" || true
  grep -E '\[backend\]' "$TMP_LINES" > "$BACKEND_FILE" || true

  rm -f "$TMP_LINES"

  for f in "$CLIENT_PROXY_FILE" "$LB_DATA_FILE" "$LB_CONTROL_FILE" "$BACKEND_FILE"; do
    if [[ ! -s "$f" ]]; then
      rm -f "$f"
    fi
  done

  if [[ -z "$(ls -A "$SNAPSHOT_DIR")" ]]; then
    rmdir "$SNAPSHOT_DIR"
  else
    echo "Captured: $SNAPSHOT_DIR"
  fi

  sleep "$INTERVAL_SEC"
done
