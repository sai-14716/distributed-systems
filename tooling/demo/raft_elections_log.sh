#!/usr/bin/env bash
# Watch Raft election activity in real-time from local nodes (node1, node2)
# Shows: elections, leadership changes, vote requests, role changes
# Usage: bash tooling/demo/watch_raft_elections.sh

set -euo pipefail

echo "=== Raft Election Monitor (Local Nodes Only) ==="
echo "Showing: elections, leader changes, vote requests, role transitions"
echo "Press Ctrl+C to stop"
echo ""

# Use follow mode to stream logs as they appear
docker compose logs -f node4 node5 2>&1 | grep -E "(election started|became leader|election lost|request-vote|role change|higher term)" --color=always