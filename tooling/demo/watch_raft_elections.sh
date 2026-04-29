#!/usr/bin/env bash
# Enhanced Raft state monitor - polls and displays node states and recent activity
# Usage: bash tooling/demo/watch_raft_elections_detailed.sh [interval_secs]

set -euo pipefail

INTERVAL="${1:-2}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

export PYTHONPATH="$ROOT_DIR"

# Get the actual distributed URLs for node1 and node2
NODE1_URL=$(python3 tooling/cluster_helper.py get_url nodes node1)
NODE2_URL=$(python3 tooling/cluster_helper.py get_url nodes node2)

echo "=== Raft State Monitor (node1 & node2) - Updates every ${INTERVAL}s ==="
echo "Press Ctrl+C to stop"
echo ""

while true; do
  clear
  echo "=== Raft State at $(date '+%H:%M:%S') ==="
  echo ""
  
  # Get both node states in parallel using their actual cluster IPs
  NODE1_STATE=$(curl -s $NODE1_URL/state 2>/dev/null || echo '{"error":"unreachable"}')
  NODE2_STATE=$(curl -s $NODE2_URL/state 2>/dev/null || echo '{"error":"unreachable"}')
  
  echo "NODE1 (19091):"
  echo "$NODE1_STATE" | jq '{id, role, term, leader_id, commit_index: .commitIndex, last_applied: .lastApplied}'
  echo ""
  
  echo "NODE2 (19092):"
  echo "$NODE2_STATE" | jq '{id, role, term, leader_id, commit_index: .commitIndex, last_applied: .lastApplied}'
  echo ""
  
  echo "--- Recent Raft Activity (last 10 events) ---"
  docker compose logs node1 node2 --tail 50 2>/dev/null | grep -E "(election started|became leader|election lost|request-vote saw higher|higher term)" | tail -10
  
  echo ""
  echo "Waiting ${INTERVAL}s for next update... (Ctrl+C to exit)"
  sleep "$INTERVAL"
done