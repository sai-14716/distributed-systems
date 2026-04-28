#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"
. "$ROOT_DIR/tooling/lib/cluster_config.sh"

CLEAN="${1:-}"
if [[ "$CLEAN" == "--clean" ]]; then
  docker compose down -v --remove-orphans
fi

configured_services="$(python3 - <<'PY'
import json

with open("cluster_config.yaml", encoding="utf-8") as f:
    cfg = json.load(f)

services = ["discovery"]
services.extend((cfg.get("backends") or {}).keys())
services.extend((cfg.get("nodes") or {}).keys())
print(" ".join(services))
PY
)"
available_services="$(docker compose config --services)"
services_to_start=()
for service in $configured_services; do
  if printf '%s\n' "$available_services" | grep -Fxq "$service"; then
    services_to_start+=("$service")
  fi
done

if [[ "${#services_to_start[@]}" -eq 0 ]]; then
  echo "No configured cluster services are present in this compose file." >&2
  exit 1
fi

docker compose up -d --build "${services_to_start[@]}"

echo
echo "Cluster is up."
echo "Control-plane (Raft) ports on host:"
for node in $(cluster_names nodes); do
  echo "  $node $(cluster_url nodes "$node" state)"
done
