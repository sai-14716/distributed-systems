#!/usr/bin/env bash
# Watch Raft election activity in real-time from local nodes (node3, node4)
# Shows: elections, leadership changes, vote requests, role changes
# Usage: bash tooling/demo/watch_raft_elections.sh

set -euo pipefail

echo "=== Raft Election Monitor (Local Nodes: node3) ==="
echo "Showing: elections, leader changes, vote requests, role transitions"
echo "Press Ctrl+C to stop"
echo ""

# Use sudo to avoid permission denied, watch both local nodes
# Filters based on actual log strings in internal/raft/raft.go
sudo docker compose logs -f node3  2>&1 | grep -E "(election started|became leader|election lost|request-vote|vote granted|higher term|node started role)" --color=always