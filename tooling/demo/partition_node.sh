#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config and helper functions
source tooling/helper/cluster_config.sh

usage() {
  echo "Usage: bash tooling/demo/partition_node.sh isolate|restore <node_name>" >&2
  echo "Example: bash tooling/demo/partition_node.sh isolate node1" >&2
  exit 2
}

action="${1:-}"
node="${2:-}"
[[ "$action" == "isolate" || "$action" == "restore" ]] || usage
[[ -n "$node" ]] || usage

# Validate node exists in config
if ! get_node_url "$node" >/dev/null 2>&1; then
  echo "ERROR: Unknown node: $node" >&2
  exit 1
fi

project="${COMPOSE_PROJECT_NAME:-distributed-systems}"
network="${DEMO_NETWORK:-${project}_sdnet}"

cid="$(docker compose ps -q "$node")"
if [[ -z "$cid" ]]; then
  echo "No running container for $node (did you run 'bash tooling/demo/up.sh'?)." >&2
  exit 1
fi

is_connected() {
  docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{println $k}}{{end}}' "$cid" 2>/dev/null \
    | grep -Fxq "$network"
}

if [[ "$action" == "isolate" ]]; then
  if ! is_connected; then
    echo "$node is already disconnected from $network"
    exit 0
  fi
  echo "Disconnecting $node from $network"
  docker network disconnect "$network" "$cid"
  echo "Done. $node is now partitioned from the cluster."
else
  if is_connected; then
    echo "$node is already connected to $network"
    exit 0
  fi
  echo "Reconnecting $node to $network"
  docker network connect "$network" "$cid"
  echo "Done. $node is back on the cluster network."
fi
