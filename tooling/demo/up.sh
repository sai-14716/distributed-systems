#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Load cluster config
source tooling/helper/cluster_config.sh

# Compose expects LAPTOP_*_IP variables; derive them from cluster_config.yaml.
if [[ -z "${LAPTOP_A:-}" || -z "${LAPTOP_B:-}" || -z "${LAPTOP_C:-}" ]]; then
  echo "ERROR: Missing laptop IPs in cluster_config.yaml (expected laptops A, B, C)" >&2
  exit 1
fi
export LAPTOP_A_IP="$LAPTOP_A"
export LAPTOP_B_IP="$LAPTOP_B"
export LAPTOP_C_IP="$LAPTOP_C"

CLEAN="${1:-}"
if [[ "$CLEAN" == "--clean" ]]; then
  docker compose down -v --remove-orphans
fi

echo "Starting Laptop A services..."
docker compose up -d --build \
  backend-1 backend-2 backend-3 \
  node1 node2

echo
echo "Laptop A is now running."
echo

# Display endpoints using config
python3 - <<'PY'
import sys
sys.path.insert(0, "tooling/helper")
from cluster_config import Config

cfg = Config.load()

print("Control-plane (Raft) endpoints:")
for node_name in ["node1", "node2"]:
    url = cfg.get_node_url(node_name, "/state")
    print(f"  {node_name}: {url}")

print()
print("Load Balancer endpoints:")
for node_name in ["node1", "node2"]:
    url = cfg.get_lb_url(node_name, "/chat")
    print(f"  {node_name}: {url}")

print()
print("Backend endpoints (on Laptop A):")
for backend_name in ["backend-1", "backend-2", "backend-3"]:
    url = cfg.get_backend_url(backend_name)
    print(f"  {backend_name}: {url}")

print()
# print("Discovery service:")
# http_url = cfg.get_discovery_url("http")
# udp_addr = cfg.get_discovery_url("udp")
# print(f"  HTTP: {http_url}")
# print(f"  UDP: {udp_addr}")

print()
print("Note: To connect other laptops (B, C), update cluster_config.yaml with their IPs")
print("and start services on those machines with their own docker-compose configurations.")
PY

