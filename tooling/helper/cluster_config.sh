#!/usr/bin/env bash
# Distributed Cluster Config Helper
# This module loads cluster_config.yaml and provides functions to resolve endpoints.
# 
# Usage in bash scripts:
#   source tooling/helper/cluster_config.sh
#   get_node_url "node1" "/state"       # -> http://IP:port/state
#   get_backend_url "backend-3"          # -> http://IP:port/
#   get_discovery_url "http"            # -> http://IP:6701

set -euo pipefail

# Detect root directory
CLUSTER_CONFIG_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_CONFIG_FILE="${CLUSTER_CONFIG_ROOT}/cluster_config.yaml"

if [[ ! -f "$CLUSTER_CONFIG_FILE" ]]; then
  echo "ERROR: cluster_config.yaml not found at $CLUSTER_CONFIG_FILE" >&2
  exit 1
fi

# Python helper to extract YAML and resolve endpoints
# This is a workaround since bash doesn't have native YAML support
declare -A _CONFIG_CACHE

_load_config() {
  if [[ ${#_CONFIG_CACHE[@]} -gt 0 ]]; then
    return
  fi

  python3 - <<'PY'
import yaml
import sys

try:
  with open(sys.argv[1]) as f:
    cfg = yaml.safe_load(f)
  
  # Extract laptops
  for laptop, ip in cfg.get('laptops', {}).items():
    print(f"LAPTOP_{laptop}={ip}")
  
  # Extract nodes
  for node_name, node_info in cfg.get('nodes', {}).items():
    laptop = node_info.get('laptop', '')
    raft_port = node_info.get('raft_port', '')
    lb_port = node_info.get('lb_port', '')
    print(f"NODE_{node_name}_LAPTOP={laptop}")
    print(f"NODE_{node_name}_RAFT_PORT={raft_port}")
    print(f"NODE_{node_name}_LB_PORT={lb_port}")
  
  # Extract backends
  for backend_name, backend_info in cfg.get('backends', {}).items():
    laptop = backend_info.get('laptop', '')
    port = backend_info.get('port', '')
    print(f"BACKEND_{backend_name}_LAPTOP={laptop}")
    print(f"BACKEND_{backend_name}_PORT={port}")
  
  # Extract discovery
  disc = cfg.get('discovery', {})
  print(f"DISCOVERY_LAPTOP={disc.get('laptop', '')}")
  print(f"DISCOVERY_UDP_PORT={disc.get('udp_port', '')}")
  print(f"DISCOVERY_HTTP_PORT={disc.get('http_port', '')}")

except Exception as e:
  print(f"ERROR: Failed to load config: {e}", file=sys.stderr)
  sys.exit(1)
PY
}

# Source the environment variables from config
eval "$(_load_config)"

# Resolve node URL
# Usage: get_node_url <node_name> [endpoint]
# Example: get_node_url node1 /state
get_node_url() {
  local node_name="$1"
  local endpoint="${2:-}"
  
  local laptop_var="NODE_${node_name}_LAPTOP"
  local port_var="NODE_${node_name}_RAFT_PORT"
  
  local laptop="${!laptop_var:-}"
  local port="${!port_var:-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown node: $node_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip="${!ip_var:-}"
  
  if [[ -z "$ip" ]]; then
    echo "ERROR: Unknown laptop: $laptop" >&2
    return 1
  fi
  
  echo "http://${ip}:${port}${endpoint}"
}

# Resolve backend URL
# Usage: get_backend_url <backend_name> [endpoint]
# Example: get_backend_url backend-1 /metrics
get_backend_url() {
  local backend_name="$1"
  local endpoint="${2:-}"
  
  local laptop_var="BACKEND_${backend_name}_LAPTOP"
  local port_var="BACKEND_${backend_name}_PORT"
  
  local laptop="${!laptop_var:-}"
  local port="${!port_var:-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown backend: $backend_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip="${!ip_var:-}"
  
  if [[ -z "$ip" ]]; then
    echo "ERROR: Unknown laptop: $laptop" >&2
    return 1
  fi
  
  echo "http://${ip}:${port}${endpoint}"
}

# Resolve load balancer URL (LB is on same node as control plane)
# Usage: get_lb_url <node_name> [endpoint]
get_lb_url() {
  local node_name="$1"
  local endpoint="${2:-}"
  
  local laptop_var="NODE_${node_name}_LAPTOP"
  local port_var="NODE_${node_name}_LB_PORT"
  
  local laptop="${!laptop_var:-}"
  local port="${!port_var:-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown node/LB: $node_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip="${!ip_var:-}"
  
  if [[ -z "$ip" ]]; then
    echo "ERROR: Unknown laptop: $laptop" >&2
    return 1
  fi
  
  echo "http://${ip}:${port}${endpoint}"
}

# Resolve discovery service URL
# Usage: get_discovery_url [protocol]
# Example: get_discovery_url http  # -> http://IP:6701
# Example: get_discovery_url udp   # -> IP:6699
get_discovery_url() {
  local protocol="${1:-http}"
  
  local laptop="$DISCOVERY_LAPTOP"
  
  if [[ -z "$laptop" ]]; then
    echo "ERROR: Discovery service not configured" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip="${!ip_var:-}"
  
  if [[ -z "$ip" ]]; then
    echo "ERROR: Unknown laptop for discovery: $laptop" >&2
    return 1
  fi
  
  case "$protocol" in
    http|HTTP)
      echo "http://${ip}:${DISCOVERY_HTTP_PORT}"
      ;;
    udp|UDP)
      echo "${ip}:${DISCOVERY_UDP_PORT}"
      ;;
    *)
      echo "ERROR: Unknown discovery protocol: $protocol" >&2
      return 1
      ;;
  esac
}

# Get all node names from config
# Usage: get_all_nodes
get_all_nodes() {
  python3 - "$CLUSTER_CONFIG_FILE" <<'PY'
import yaml
import sys
cfg = yaml.safe_load(open(sys.argv[1]))
for node_name in sorted(cfg.get('nodes', {}).keys()):
  print(node_name)
PY
}

# Get all backend names from config
# Usage: get_all_backends
get_all_backends() {
  python3 - "$CLUSTER_CONFIG_FILE" <<'PY'
import yaml
import sys
cfg = yaml.safe_load(open(sys.argv[1]))
for backend_name in sorted(cfg.get('backends', {}).keys()):
  print(backend_name)
PY
}

# Fetch a URL with timeout and error handling
# Usage: fetch_url <url> [timeout_seconds]
# Returns: HTTP response body on success, empty string on failure
fetch_url() {
  local url="$1"
  local timeout="${2:-0.6}"
  
  curl -s --connect-timeout "$timeout" --max-time "$timeout" "$url" 2>/dev/null || echo ""
}

# Fetch JSON and parse with jq (requires jq installed)
# Usage: fetch_json <url> [jq_filter]
# Example: fetch_json "http://node1:19091/state" ".role"
fetch_json() {
  local url="$1"
  local jq_filter="${2:-.}"
  
  fetch_url "$url" | jq -r "$jq_filter" 2>/dev/null || echo ""
}

export -f get_node_url
export -f get_backend_url
export -f get_lb_url
export -f get_discovery_url
export -f get_all_nodes
export -f get_all_backends
export -f fetch_url
export -f fetch_json
