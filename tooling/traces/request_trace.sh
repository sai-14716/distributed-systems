#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

TRACE_ID="${1:-trace-$(date +%Y%m%dT%H%M%S)-$RANDOM}"
OUT_DIR="${TRACE_OUT_DIR:-$ROOT_DIR/logs}"
SINCE_WINDOW="${TRACE_SINCE_WINDOW:-2m}"
BUILD_FLAG="${TRACE_BUILD_FLAG:---build}"
TRACE_DIR="$OUT_DIR/$TRACE_ID"
mkdir -p "$TRACE_DIR"

TMP_K6_OUTPUT="$(mktemp)"
TMP_TRACE_LINES="$(mktemp)"
TMP_CLIENT_APP="$(mktemp)"
K6_FILE="$TRACE_DIR/k6_client.log"
CLIENT_PROXY_FILE="$TRACE_DIR/client_proxy.log"
LB_DATA_FILE="$TRACE_DIR/lb_data.log"
LB_CONTROL_FILE="$TRACE_DIR/lb_control.log"
BACKEND_FILE="$TRACE_DIR/backend.log"
METADATA_FILE="$TRACE_DIR/metadata.log"

echo "Running traced request with TRACE_ID=$TRACE_ID"

BUILD_ARG=()
if [[ -n "$BUILD_FLAG" ]]; then
  BUILD_ARG=("$BUILD_FLAG")
fi

docker compose --profile loadtest run "${BUILD_ARG[@]}" --rm \
  -e K6_SCRIPT=/app/k6/trace_request.js \
  -e TRACE_ID="$TRACE_ID" \
  k6-client 2>&1 | tee "$TMP_K6_OUTPUT"

echo "trace_id=$TRACE_ID" > "$METADATA_FILE"
echo "generated_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$METADATA_FILE"
echo "since_window=$SINCE_WINDOW" >> "$METADATA_FILE"

grep -E "(\\[client-app\\]|\\[client-proxy\\]).*trace=${TRACE_ID}|status is 200|backend id present|backend ack present|checks_succeeded|checks_failed" "$TMP_K6_OUTPUT" > "$TMP_CLIENT_APP" || true
if [[ -s "$TMP_CLIENT_APP" ]]; then
  cp "$TMP_CLIENT_APP" "$K6_FILE"
else
  cp "$TMP_K6_OUTPUT" "$K6_FILE"
fi

docker compose logs --since "$SINCE_WINDOW" \
  discovery node1 node2 backend-1 backend-2 backend-3 2>&1 \
  | grep "$TRACE_ID" > "$TMP_TRACE_LINES" || true

grep -E '\[client-proxy\]' "$TMP_TRACE_LINES" > "$CLIENT_PROXY_FILE" || true
grep -E '\[lb-data\]' "$TMP_TRACE_LINES" > "$LB_DATA_FILE" || true
grep -E '\[lb-control\]' "$TMP_TRACE_LINES" > "$LB_CONTROL_FILE" || true
grep -E '\[backend\]' "$TMP_TRACE_LINES" > "$BACKEND_FILE" || true

for f in "$CLIENT_PROXY_FILE" "$LB_DATA_FILE" "$LB_CONTROL_FILE" "$BACKEND_FILE"; do
  if [[ ! -s "$f" ]]; then
    rm -f "$f"
  fi
done

rm -f "$TMP_K6_OUTPUT"
rm -f "$TMP_TRACE_LINES"
rm -f "$TMP_CLIENT_APP"

echo "Saved trace logs in: $TRACE_DIR"
