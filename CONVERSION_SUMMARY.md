# Distributed Conversion Summary

This document summarizes all changes made to convert the single-machine Docker Compose setup to a distributed multi-machine architecture.

## File Changes Overview

### New Files Created

1. **cluster_config.yaml**
   - Central configuration file defining the topology for all 3 laptops
   - Specifies which services run on which laptops
   - Contains all IP addresses, ports, and service mappings
   - Single source of truth for topology

2. **tooling/helper/cluster_config.sh**
   - Bash helper module for cluster configuration
   - Provides functions to resolve service endpoints dynamically
   - Functions: `get_node_url()`, `get_backend_url()`, `get_lb_url()`, `get_discovery_url()`
   - Sources YAML config and exports environment variables
   - Used by all bash scripts

3. **tooling/helper/cluster_config.py**
   - Python helper module for cluster configuration
   - Loads cluster_config.yaml and provides endpoint resolution
   - Config class with methods: `get_node_url()`, `get_backend_url()`, etc.
   - Includes helper functions: `fetch_json()`, `fetch_text()`
   - Used by all Python scripts

4. **docker-compose.laptopb.yaml**
   - Docker Compose template for Laptop B
   - Contains node3, node4, and backends 4-7
   - Uses environment variables for cross-machine IP configuration
   - Demonstrates the pattern for other laptops

5. **docker-compose.laptopc.yaml**
   - Docker Compose template for Laptop C
   - Contains node5 and backends 8-10
   - Uses environment variables for cross-machine IP configuration

6. **DISTRIBUTED_SETUP.md**
   - Comprehensive guide for setting up and using the distributed cluster
   - Instructions for configuring IPs, starting services, running demos
   - Troubleshooting guide
   - Architecture overview

7. **CONVERSION_SUMMARY.md** (this file)
   - Documents all changes made during the conversion

### Modified Files

#### docker-compose.yaml (Primary - Laptop A)

**Removed services:**

- node3, node4, node5
- backend-4, backend-5, backend-6, backend-7, backend-8, backend-9, backend-10
- k6-client (loadtest profile)
- admin (tools profile)

**Modified services:**

- All services now bind to `0.0.0.0` instead of `127.0.0.1`
- Node port bindings: Changed from `127.0.0.1:PORT:CONTAINER_PORT` to `PORT:CONTAINER_PORT`
  - node1: `19091:19090` instead of `127.0.0.1:19091:19090`
  - node2: `19092:19090` instead of `127.0.0.1:19092:19090`
- Environment variables updated to use machine IPs:
  - `SELF_URL`: Changed from `http://nodeX:19090` to `http://${LAPTOP_A_IP}:19091`
  - `PEERS`: Changed from `http://nodeX:19090` to full IP:port format
- Removed static IPv4 assignments
- Simplified networks section (removed subnet configuration)
- Reduced volumes to only node1_data and node2_data

#### tooling/demo/state.sh

**Changes:**

- Added cluster config loading: `source tooling/helper/cluster_config.sh`
- Replaced hardcoded localhost port list with dynamic config loading
- Uses `Config.load()` to get node list from cluster_config.yaml
- Uses `cfg.get_node_url(node_name, "/state")` to resolve endpoints
- Removed hardcoded `ports = {"node1": 19091, ...}` dictionary
- Now handles error cases gracefully (timeout, connection refused)

#### tooling/demo/up.sh

**Changes:**

- Added cluster config loading
- Only starts Laptop A services: `discovery backend-1 backend-2 backend-3 node1 node2`
- Removed node3-5 and backend-4-10 from startup
- Displays resolved endpoint URLs from config instead of hardcoded localhost
- Shows proper endpoint discovery for LBs and backends

#### tooling/demo/show_routing_headers.sh

**Changes:**

- Added cluster config loading
- Changed signature: now takes `[lb_node] [endpoint] [count]` instead of `[url] [count]`
- Uses `get_lb_url()` to resolve LB endpoint from node name
- Script now topology-aware

#### tooling/demo/routing_behaviour.sh

**Changes:**

- Updated to accept `[algo] [iterations] [session_id] [fixed|per_request] [lb_node]`
- Uses `get_lb_url()` to resolve LB endpoint
- Added note about admin tool requiring update for distributed mode

#### tooling/demo/dynamic_backend_load.sh

**Changes:**

- Added cluster config loading
- Removed hardcoded backend port configuration (8081-8090)
- Uses `get_all_backends()` to get backend list from config
- Dynamically builds backend URLs from config
- Removed BACKEND_COUNT, BACKEND_PORT_BASE, BACKEND_URLS environment variables

#### tooling/demo/stop_leader.sh

**Changes:**

- Added cluster config loading
- Replaced hardcoded node list with `cfg.get_node_names()`
- Uses `cfg.get_node_url()` to construct endpoint URLs
- Queries all configured nodes to find leader
- Graceful fallback if cluster is partially unreachable

#### tooling/demo/partition_node.sh

**Changes:**

- Added cluster config loading
- Removed hardcoded `node_ip()` function with static IPs (10.10.0.11, etc.)
- Uses `get_node_url()` to validate node exists in config
- Simplified network disconnect/reconnect (no longer passing explicit IPs)

#### tooling/demo/watch_health_propagation.sh

**Changes:**

- Added cluster config loading
- Replaced hardcoded LB list with dynamic config loading
- Uses `cfg.get_lb_url()` for each node
- Automatically discovers all nodes from config
- Uses `fetch_json()` helper instead of urllib directly

#### tooling/demo/least_requests_demo.sh

**Changes:**

- Added cluster config loading
- Changed parameter from `[url]` to `[lb_node]`
- Uses `get_lb_url()` to resolve endpoint

#### tooling/demo/load_weight_demo.sh

**Changes:**

- Added cluster config loading
- Changed parameters: `[algo] [lb_node] [requests] ...` instead of `[algo] [url] ...`
- Uses `get_lb_url()` to resolve endpoint

#### tooling/demo/least_requests_with_dynamic_load.sh

**Changes:**

- Added cluster config loading
- Updated parameter parsing to use lb_node instead of url
- Updated show_routing_headers call to pass node name

#### tooling/demo/wrr_with_dynamic_load.sh

**Changes:**

- Added cluster config loading
- Updated parameter parsing to use lb_node instead of url
- Updated show_routing_headers call to pass node name

#### tooling/demo/run_all_algos_demo.sh

**Changes:**

- Added cluster config loading
- Changed parameters: `[lb_node] [requests]` instead of `[url] [requests]`
- Updated all script calls to pass lb_node instead of url

#### tooling/demo/watch_discovery_state.sh

**Changes:**

- Added cluster config loading
- Changed parameter from `[seconds] [url]` to just `[seconds]`
- Uses `get_discovery_url("http")` to resolve discovery endpoint
- Simplified usage: no need to specify hostname/port manually

#### tooling/demo/compare_ring_assignment.sh

**Changes:**

- Added cluster config loading
- Changed parameter: `[lb_node]` instead of `[url]`
- Uses `get_lb_url()` to resolve endpoint

#### tooling/traces/periodic_trace_collector.sh

**Changes:**

- Updated docker compose logs command to only reference Laptop A services
- Changed from all 10 backends to just 3: `backend-1 backend-2 backend-3`
- Removed reference to nodes 3-5

#### tooling/traces/request_trace.sh

**Changes:**

- Updated docker compose logs command to only reference Laptop A services
- Removed references to nodes 3-5 and backends 4-10

#### tooling/chaos/run_combined_chaos_drill.sh

**Changes:**

- Added cluster config loading
- Removed TARGETS variable (no longer needed)
- Updated `find_leader()` to use only Laptop A nodes: `(node1 node2)`
- Changes leader detection to use `curl` instead of `docker compose exec`
- Uses `get_node_url()` for endpoint resolution
- Updated cluster startup to only include Laptop A services

## Key Configuration Parameters

### cluster_config.yaml Structure

```yaml
laptops:
  <letter>: <ip_address>

nodes:
  <node_name>:
    laptop: <letter>
    raft_port: <port>
    lb_port: <port>

backends:
  <backend_name>:
    laptop: <letter>
    port: <port>

discovery:
  laptop: <letter>
  udp_port: <port>
  http_port: <port>
```

### Environment Variable Replacements in docker-compose

Original hardcoded patterns replaced:

- `http://node1:19090` → `http://${LAPTOP_A_IP}:19091`
- `http://127.0.0.1:8081` → `http://${LAPTOP_A_IP}:8081`
- Static service name references → Dynamic IP:port resolution

### Port Binding Changes

All port bindings changed from:

```yaml
ports:
  - "127.0.0.1:PORT:CONTAINER_PORT"
```

To:

```yaml
ports:
  - "PORT:CONTAINER_PORT"
```

This allows ports to be accessible from any IP interface, not just localhost.

## Behavioral Changes

### Before

- All services accessed via `127.0.0.1`
- Service names used for network communication
- Docker internal networking (bridge)
- Single-machine testing only
- Scripts had hardcoded port mappings

### After

- All services accessed via real machine IP addresses
- IP:port used for all inter-service communication
- Real network communication between machines
- Multi-machine distributed testing capability
- Scripts dynamically resolve endpoints from config

## Breaking Changes

1. **Script Parameters**: Many scripts changed parameter signatures
   - `show_routing_headers.sh [url] [n]` → `[lb_node] [endpoint] [n]`
   - `state.sh` no longer takes parameters (reads all nodes from config)
   - Other scripts similarly updated

2. **Port Exposure**: Services no longer bind only to 127.0.0.1
   - This is intentional to support cross-machine communication
   - Firewall configuration may be needed

3. **Environment Variables**:
   - Scripts no longer accept full URLs as parameters
   - Instead, they accept service/node names and resolve from config

## Testing and Validation

To verify the conversion:

1. **Laptop A only**: `bash tooling/demo/up.sh` and `bash tooling/demo/state.sh`
2. **Script execution**: All demos should work referencing node names instead of URLs
3. **Configuration**: Edit `cluster_config.yaml` and reload scripts; behavior should change accordingly
4. **Port bindings**: `docker compose ps` should show ports accessible on all interfaces

## Migration Notes for Users

If you have custom scripts:

### Bash Scripts

Replace:

```bash
curl -s http://127.0.0.1:19091/state
```

With:

```bash
source tooling/helper/cluster_config.sh
url=$(get_node_url "node1" "/state")
curl -s "$url"
```

### Python Scripts

Replace:

```python
url = "http://127.0.0.1:19091/state"
```

With:

```python
from tooling.helper.cluster_config import Config
cfg = Config.load()
url = cfg.get_node_url("node1", "/state")
```

## Future Enhancements

Potential improvements to consider:

1. **Admin tool update**: Make admin tool distributed-aware
2. **Automatic IP detection**: Scripts could auto-detect local machine IP
3. **DNS integration**: Use discovery service for all lookups
4. **Load balancer configuration**: Make LB peer discovery dynamic
5. **Multi-region clusters**: Extend topology model to support multiple regions
