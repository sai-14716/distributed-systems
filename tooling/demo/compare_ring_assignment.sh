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
for node_name, node_info in cfg.get('nodes', {}).items():
  print(f"export NODE_{node_name}_LAPTOP={node_info.get('laptop', '')}")
  print(f"export NODE_{node_name}_LB_PORT={node_info.get('lb_port', '')}")
PY
}
eval "$(_load_config)"

# Resolve LB node URL
get_lb_url() {
  local node_name="$1" endpoint="${2:-}"
  local laptop_var="NODE_${node_name}_LAPTOP" port_var="NODE_${node_name}_LB_PORT"
  eval "local laptop=\${${laptop_var}-}" "local port=\${${port_var}-}"
  [[ -z "$laptop" || -z "$port" ]] && { echo "ERROR: Unknown node: $node_name" >&2; return 1; }
  local ip_var="LAPTOP_${laptop}"
  eval "local ip=\${${ip_var}-}"
  [[ -z "$ip" ]] && { echo "ERROR: Unknown laptop: $laptop" >&2; return 1; }
  echo "http://${ip}:${port}${endpoint}"
}

MODE="${1:-real}"
LB_NODE="${2:-node1}"
NODE_TO_TOGGLE="${3:-node5}"
OUT_DIR="${4:-/tmp}"
WAIT_SECS="${WAIT_SECS:-5}"

# Show help if --help or -h
if [[ "$MODE" == "-h" || "$MODE" == "--help" ]]; then
  cat <<'HELP'
Compare ring assignments before/after node removal

USAGE: compare_ring_assignment.sh [MODE] [LB_NODE] [NODE_TO_TOGGLE] [OUT_DIR]

MODES:
  local      Fetch from localhost LB nodes (docker compose running locally)
  real       Fetch from remote cluster IPs (multi-laptop distributed setup)
  simulate   Simulate ring changes using the Python script (no services needed)

EXAMPLES:
  # Local mode (docker compose up on this machine)
  ./compare_ring_assignment.sh local node1 node2

  # Real mode (requires remote cluster nodes running)
  ./compare_ring_assignment.sh real node1 node5

  # Simulate mode (uses Python script for demonstration)
  ./compare_ring_assignment.sh simulate

DEFAULT ARGS:
  MODE: real
  LB_NODE: node1
  NODE_TO_TOGGLE: node5
  OUT_DIR: /tmp

ENVIRONMENT:
  WAIT_SECS: Wait time between operations (default: 5)
HELP
  exit 0
fi

if [[ "$MODE" == "simulate" ]]; then
  # ========================================
  # SIMULATE MODE: Use Python script
  # ========================================
  python3 tooling/demo/compare_ring_assignment.py \
    --before-peers node1,node2,node3,node4,node5 \
    --after-peers node1,node2,node3,node4 \
    --replicas 50
  exit 0
fi

if [[ "$MODE" == "local" ]]; then
  # ========================================
  # LOCAL MODE: Connect to localhost LB nodes
  # ========================================
  # Map node name to localhost port
  get_local_lb_url() {
    local node_name="$1"
    case "$node_name" in
      node1) echo "http://localhost:8001" ;;
      node2) echo "http://localhost:8002" ;;
      node3) echo "http://localhost:8003" ;;
      node4) echo "http://localhost:8004" ;;
      node5) echo "http://localhost:8005" ;;
      *) echo "ERROR: Unknown node: $node_name" >&2; return 1 ;;
    esac
  }

  # Get backends from a specific node
  get_node_backends() {
    local node_url="$1"
    curl -sf --connect-timeout 2 "$node_url/admin/backends" 2>/dev/null | \
      jq -r '.backends[]' 2>/dev/null | sort
  }

  # Query all nodes
  query_all_nodes() {
    for node in node1 node2 node3 node4 node5; do
      local url=$(get_local_lb_url "$node") || continue
      local backends=$(get_node_backends "$url")
      if [[ -n "$backends" ]]; then
        echo "# $node"
        echo "$backends"
      fi
    done
  }

  ########################################
  # BASELINE
  ########################################
  echo "[ring-demo] (local) fetching backend assignments from all nodes..."
  echo
  before_output=$OUT_DIR/backends.before.txt
  query_all_nodes > "$before_output"
  cat "$before_output"

  ########################################
  # STOP NODE
  ########################################
  echo
  echo "[ring-demo] stopping $NODE_TO_TOGGLE..."
  docker compose stop "$NODE_TO_TOGGLE" >/dev/null
  sleep "$WAIT_SECS"

  echo "[ring-demo] after stop..."
  echo
  after_stop_output=$OUT_DIR/backends.after-stop.txt
  query_all_nodes > "$after_stop_output"
  cat "$after_stop_output"

  ########################################
  # RECOVER NODE
  ########################################
  echo
  echo "[ring-demo] recovering $NODE_TO_TOGGLE..."
  docker compose up -d "$NODE_TO_TOGGLE" >/dev/null
  sleep "$WAIT_SECS"

  echo "[ring-demo] after recover..."
  echo
  after_recover_output=$OUT_DIR/backends.after-recover.txt
  query_all_nodes > "$after_recover_output"
  cat "$after_recover_output"

  ########################################
  # SUMMARY
  ########################################
  echo
  echo "[ring-demo] Comparison files:"
  echo "  $before_output"
  echo "  $after_stop_output"
  echo "  $after_recover_output"
  exit 0
fi

# REAL MODE (default)
echo "[ring-demo] REAL mode: querying actual node backend assignments..."
echo "  (requires docker services with updated /admin/backends endpoint)"
echo

LB_URL="$(get_lb_url "$LB_NODE")" || exit 1
MODE_DESC="remote cluster"

# Get backends from a specific node URL
get_node_backends_real() {
  local node_url="$1"
  curl -sf --connect-timeout 3 "$node_url/admin/backends" 2>/dev/null | \
    jq -r '.backends[]' 2>/dev/null | sort
}

# Query all nodes in real mode
query_all_nodes_real() {
  # For real mode, get all node names from env or hardcode
  local all_nodes=(node1 node2 node3 node4 node5)
  for node in "${all_nodes[@]}"; do
    local url=$(get_lb_url "$node" "" 2>/dev/null) || continue
    local backends=$(get_node_backends_real "$url")
    if [[ -n "$backends" ]]; then
      echo "# $node"
      echo "$backends"
    fi
  done
}

mkdir -p "$OUT_DIR"

########################################
# BASELINE
########################################
echo "[ring-demo] baseline: backend assignments from all nodes"
before_output="$OUT_DIR/backends.before.txt"
query_all_nodes_real > "$before_output"
cat "$before_output"

########################################
# STOP NODE
########################################
echo
echo "[ring-demo] stopping $NODE_TO_TOGGLE"
# Use ssh or direct command depending on setup
# For now, assume local docker for testing
if command -v docker &> /dev/null; then
  docker compose stop "$NODE_TO_TOGGLE" >/dev/null 2>&1 || echo "  (docker stop not available in REAL mode)"
fi
sleep "$WAIT_SECS"

echo "[ring-demo] after stop..."
echo
after_stop_output="$OUT_DIR/backends.after-stop.txt"
query_all_nodes_real > "$after_stop_output"
cat "$after_stop_output"

########################################
# RECOVER NODE
########################################
echo
echo "[ring-demo] recovering $NODE_TO_TOGGLE"
if command -v docker &> /dev/null; then
  docker compose up -d "$NODE_TO_TOGGLE" >/dev/null 2>&1 || echo "  (docker up not available in REAL mode)"
fi
sleep "$WAIT_SECS"

echo "[ring-demo] after recover..."
echo
after_recover_output="$OUT_DIR/backends.after-recover.txt"
query_all_nodes_real > "$after_recover_output"
cat "$after_recover_output"

########################################
# SUMMARY
########################################
echo
echo "[ring-demo] Comparison files:"
echo "  $before_output"
echo "  $after_stop_output"
echo "  $after_recover_output"