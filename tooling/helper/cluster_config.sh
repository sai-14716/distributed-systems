#!/usr/bin/env bash
# Distributed Cluster Config Helper
# This module loads cluster_config.yaml and provides functions to resolve endpoints.
# 
# Usage in bash scripts:
#   source tooling/helper/cluster_config.sh
#   get_node_url "node1" "/state"       # -> http://IP:port/state
#   get_backend_url "backend-3"          # -> http://IP:port/
#   get_discovery_url "http"            # -> http://IP:6701

set -eu
set -o pipefail 2>/dev/null || true

# Detect root directory by walking up from the current directory.
# This keeps the helper usable when sourced from either bash or zsh.
_cluster_config_find_root() {
  local root="$PWD"
  while [[ "$root" != "/" ]]; do
    if [[ -f "$root/cluster_config.yaml" ]]; then
      printf '%s\n' "$root"
      return 0
    fi
    root="$(cd "$root/.." && pwd)"
  done
  return 1
}

CLUSTER_CONFIG_ROOT="$(_cluster_config_find_root)"
CLUSTER_CONFIG_FILE="${CLUSTER_CONFIG_ROOT}/cluster_config.yaml"

if [[ ! -f "$CLUSTER_CONFIG_FILE" ]]; then
  echo "ERROR: cluster_config.yaml not found at $CLUSTER_CONFIG_FILE" >&2
  exit 1
fi

_load_config() {
  if [[ -n "${_CONFIG_LOADED:-}" ]]; then
    return
  fi

  python3 - "$CLUSTER_CONFIG_FILE" <<'PY'
import yaml
import sys

try:
  with open(sys.argv[1]) as f:
    cfg = yaml.safe_load(f)
  
  # Extract laptops
  for laptop, ip in cfg.get('laptops', {}).items():
    print(f"export LAPTOP_{laptop}={ip}")
    print(f"export LAPTOP_{laptop}_IP={ip}")
  
  # Extract nodes
  for node_name, node_info in cfg.get('nodes', {}).items():
    laptop = node_info.get('laptop', '')
    raft_port = node_info.get('raft_port', '')
    lb_port = node_info.get('lb_port', '')
    print(f"export NODE_{node_name}_LAPTOP={laptop}")
    print(f"export NODE_{node_name}_RAFT_PORT={raft_port}")
    print(f"export NODE_{node_name}_LB_PORT={lb_port}")
  
  # Extract backends (normalize names for shell-safe variable keys)
  for backend_name, backend_info in cfg.get('backends', {}).items():
    backend_key = ''.join(ch if (ch.isalnum() or ch == '_') else '_' for ch in backend_name)
    laptop = backend_info.get('laptop', '')
    port = backend_info.get('port', '')
    print(f"export BACKEND_{backend_key}_LAPTOP={laptop}")
    print(f"export BACKEND_{backend_key}_PORT={port}")
  
  # Extract discovery
  disc = cfg.get('discovery', {})
  print(f"export DISCOVERY_LAPTOP={disc.get('laptop', '')}")
  print(f"export DISCOVERY_UDP_PORT={disc.get('udp_port', '')}")
  print(f"export DISCOVERY_HTTP_PORT={disc.get('http_port', '')}")

except Exception as e:
  print(f"ERROR: Failed to load config: {e}", file=sys.stderr)
  sys.exit(1)
PY
}

# Source the environment variables from config
eval "$(_load_config)"
_CONFIG_LOADED=1

_to_var_key() {
  local input="$1"
  echo "${input//[^a-zA-Z0-9_]/_}"
}

# Resolve node URL
# Usage: get_node_url <node_name> [endpoint]
# Example: get_node_url node1 /state
get_node_url() {
  local node_name="$1"
  local endpoint="${2:-}"
  
  local laptop_var="NODE_${node_name}_LAPTOP"
  local port_var="NODE_${node_name}_RAFT_PORT"
  
  local laptop port
  eval "laptop=\${${laptop_var}-}"
  eval "port=\${${port_var}-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown node: $node_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip
  eval "ip=\${${ip_var}-}"
  
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
  local backend_key
  backend_key="$(_to_var_key "$backend_name")"
  
  local laptop_var="BACKEND_${backend_key}_LAPTOP"
  local port_var="BACKEND_${backend_key}_PORT"
  
  local laptop port
  eval "laptop=\${${laptop_var}-}"
  eval "port=\${${port_var}-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown backend: $backend_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip
  eval "ip=\${${ip_var}-}"
  
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
  
  local laptop port
  eval "laptop=\${${laptop_var}-}"
  eval "port=\${${port_var}-}"
  
  if [[ -z "$laptop" || -z "$port" ]]; then
    echo "ERROR: Unknown node/LB: $node_name" >&2
    return 1
  fi
  
  local ip_var="LAPTOP_${laptop}"
  local ip
  eval "ip=\${${ip_var}-}"
  
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
  local ip
  eval "ip=\${${ip_var}-}"
  
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

# Find the active leader node among all configured nodes (for debugging)
# Usage: find_leader 
# Returns: leader node name (e.g., "node1")
find_leader() {
  local timeout=1
  for node in $(get_all_nodes); do
    local state_url
    state_url="$(get_node_url "$node" "/state")"
    local role
    role="$(fetch_json "$state_url" ".role" 2>/dev/null)" || continue
    if [[ "$role" == "Leader" ]]; then
      echo "$node"
      return 0
    fi
  done
  return 1
}

# Submit config change to the Raft cluster
# Sends to node1 by default; routing automatically forwards to leader if needed
# Usage: submit_config <algorithm> [probe_interval_ms] [health_threshold] [node]
# Example: submit_config "wrr" 1000 0.8 [node1]
submit_config() {
  local algorithm="${1:-}"
  local probe_interval_ms="${2:-1000}"
  local health_threshold="${3:-0.8}"
  local target_node="${4:-node1}"  # Any node works due to request routing
  
  if [[ -z "$algorithm" ]]; then
    echo "ERROR: algorithm required" >&2
    return 1
  fi
  
  # Get target node URL (routing forwards to leader if needed)
  local target_url
  target_url=$(get_node_url "$target_node" "/admin/submit") || return 1
  
  # Submit config via curl
  echo "[config] Sending to $target_node: algorithm=$algorithm probe_interval_ms=$probe_interval_ms health_threshold=$health_threshold" >&2
  
  curl -sS -X POST "$target_url" \
    -H 'Content-Type: application/json' \
    -d "{\"type\":\"set_config\",\"data\":{\"algorithm\":\"$algorithm\",\"probe_interval_ms\":$probe_interval_ms,\"health_threshold\":$health_threshold}}" \
    -w "\nStatus: %{http_code}\n" 2>&1
}


