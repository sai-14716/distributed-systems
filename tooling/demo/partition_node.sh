#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Machine C local containers only
LOCAL_NODES="node4 node5"

usage() {
  echo "Usage: bash tooling/demo/partition_node.sh isolate|restore <node>" >&2
  echo "  Local nodes on this machine: $LOCAL_NODES" >&2
  exit 2
}

action="${1:-}"
node="${2:-}"
[[ "$action" == "isolate" || "$action" == "restore" ]] || usage
[[ -n "$node" ]] || usage

# Verify the node is local to this machine
is_local=false
for ln in $LOCAL_NODES; do
  if [[ "$node" == "$ln" ]]; then
    is_local=true
    break
  fi
done

if [[ "$is_local" != "true" ]]; then
  echo "Error: $node is not a local container on this machine." >&2
  echo "This machine only controls: $LOCAL_NODES" >&2
  exit 1
fi

project="${COMPOSE_PROJECT_NAME:-distributed-systems}"
network="${DEMO_NETWORK:-${project}_sdnet}"

cid="$(docker compose ps -q "$node")"
if [[ -z "$cid" ]]; then
  echo "No running container for $node (did you run docker compose up?)." >&2
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
  echo "Reconnecting $node to $network (alias=$node)"
  docker network connect --alias "$node" "$network" "$cid"
  echo "Done. $node is back on the cluster network."
fi
