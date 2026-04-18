#!/bin/sh
set -eu

/bin/lb &
LB_PID=$!

/bin/raft-node &
RAFT_PID=$!

shutdown() {
  kill "$LB_PID" 2>/dev/null || true
  kill "$RAFT_PID" 2>/dev/null || true
  wait "$LB_PID" 2>/dev/null || true
  wait "$RAFT_PID" 2>/dev/null || true
}

trap 'shutdown; exit 0' INT TERM

while true; do
  if ! kill -0 "$LB_PID" 2>/dev/null; then
    echo "[entrypoint] lb process exited"
    shutdown
    exit 1
  fi
  if ! kill -0 "$RAFT_PID" 2>/dev/null; then
    echo "[entrypoint] raft-node process exited"
    shutdown
    exit 1
  fi
  sleep 1
done
