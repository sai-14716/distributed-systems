# Distributed Multi-Machine Cluster Setup Guide

## Overview

This repository has been converted from a single-machine Docker Compose setup to a distributed multi-machine setup. The cluster is now designed to run across 3 laptops (A, B, C) with real IP-based communication instead of Docker-internal networking.

## Architecture Changes

### Previous Setup (Single Machine)

- All services on localhost
- Docker internal networking with container names
- Hardcoded ports on 127.0.0.1
- Limited scalability and no distributed testing

### New Setup (Multi-Machine)

- **Laptop A**: node1, node2, backend-1, backend-2, backend-3, discovery service
- **Laptop B**: node3, node4, backend-4, backend-5, backend-6, backend-7
- **Laptop C**: node5, backend-8, backend-9, backend-10
- All inter-node communication uses real machine IPs
- Dynamic topology configuration via `cluster_config.yaml`
- Scripts are topology-aware and work from any machine

## Quick Start

### 1. Configure Your Cluster Topology

Edit `cluster_config.yaml` and set the laptop IP addresses:

```yaml
laptops:
  A: "192.168.1.100" # Get the actual IP of your Laptop A (use `hostname -I` or `ipconfig`)
  B: "192.168.1.101" # Laptop B's IP
  C: "192.168.1.102" # Laptop C's IP
```

Find your machine's IP address:

- **Linux/Mac**: `hostname -I` or `ifconfig`
- **Windows**: `ipconfig` (look for IPv4 Address)

### 2. Start Laptop A

On **Laptop A**, run:

```bash
bash tooling/demo/up.sh
```

This will:

- Start discovery service on port 6699 (UDP) and 6701 (HTTP)
- Start node1 and node2 (Raft consensus nodes) on ports 19091 and 19092
- Start 3 backends on ports 8081, 8082, 8083
- Show you the endpoints for all services

Expected output:

```
Laptop A is now running.

Control-plane (Raft) endpoints:
  node1: http://192.168.1.100:19091/state
  node2: http://192.168.1.100:19092/state

Load Balancer endpoints:
  node1: http://192.168.1.100:8001/chat
  node2: http://192.168.1.100:8002/chat

Backend endpoints (on Laptop A):
  backend-1: http://192.168.1.100:8081
  backend-2: http://192.168.1.100:8082
  backend-3: http://192.168.1.100:8083

Discovery service:
  HTTP: http://192.168.1.100:6701
  UDP: 192.168.1.100:6699
```

### 3. Configure Other Laptops (B and C)

Clone this repository on **Laptop B** and **Laptop C**, then:

```bash
# Edit cluster_config.yaml with the same IPs as Laptop A
# Then create laptop-specific docker-compose files (instructions below)
```

## Understanding the Configuration

### cluster_config.yaml

This file is the single source of truth for the cluster topology:

```yaml
laptops:
  A: <ip> # Machine running node1, node2, discovery, backend-1/2/3
  B: <ip> # Machine running node3, node4, backend-4/5/6/7
  C: <ip> # Machine running node5, backend-8/9/10

nodes:
  node1: { laptop: A, raft_port: 19091, lb_port: 8001 }
  # ... etc

backends:
  backend-1: { laptop: A, port: 8081 }
  # ... etc

discovery:
  laptop: A
  udp_port: 6699
  http_port: 6701
```

All scripts automatically load this configuration and resolve endpoints accordingly.

## Running Demo Scripts

All scripts have been updated to use the cluster configuration. They work from any machine and resolve endpoints dynamically.

### View Cluster State

```bash
bash tooling/demo/state.sh
```

Shows current Raft cluster state (role, leader, term, etc.) for all nodes.

### Watch State Changes

```bash
bash tooling/demo/watch_state.sh
```

Continuously polls and displays cluster state with 1-second updates.

### Test Routing Behavior

```bash
# Show load balancer routing headers for 5 requests
bash tooling/demo/show_routing_headers.sh [node] [endpoint] [count]

# Examples:
bash tooling/demo/show_routing_headers.sh node1 /chat 5
bash tooling/demo/show_routing_headers.sh node2 /payload 10
```

### Demonstrate Load Balancing Algorithms

```bash
# Test round-robin, least-requests, weighted-round-robin, etc.
bash tooling/demo/run_all_algos_demo.sh [lb-node] [num-requests]

# Examples:
bash tooling/demo/run_all_algos_demo.sh node1 30
bash tooling/demo/run_all_algos_demo.sh node2 50
```

### Test Health Propagation

```bash
# Watch how backend health changes propagate across load balancers
bash tooling/demo/watch_health_propagation.sh [seconds]

# Example:
bash tooling/demo/watch_health_propagation.sh 20
```

### Monitor Discovery Service

```bash
# View discovery service's understanding of cluster health
bash tooling/demo/watch_discovery_state.sh [seconds]

# Example:
bash tooling/demo/watch_discovery_state.sh 15
```

### Test Partition Tolerance

```bash
# Isolate a node from the network (partition)
bash tooling/demo/partition_node.sh isolate node1

# Restore the node
bash tooling/demo/partition_node.sh restore node1
```

### Dynamic Load Generation

```bash
# Generate rotating load on all backends
bash tooling/demo/dynamic_backend_load.sh [total_secs] [phase_secs] [work_ms] [peak_parallel]

# Example: 60-second test with 0.5-second phases
bash tooling/demo/dynamic_backend_load.sh 60 0.5 1800 8
```

## Architecture for Multi-Machine Deployment

### Laptop B Configuration (node3, node4, backends 4-7)

Create `docker-compose.laptopb.yaml` (or modify template):

```yaml
services:
  backend-4:
    build:
      context: .
      dockerfile: services/backend/Dockerfile
    environment:
      - SERVER_ID=B4
    ports:
      - "8084:8080" # Exposed on all IPs
    networks:
      - sdnet

  backend-5, backend-6, backend-7: # ... similar pattern

  node3:
    build:
      context: .
      dockerfile: cmd/node/Dockerfile
    environment:
      NODE_ID: node3
      HTTP_ADDR: :19090
      SELF_URL: http://${LAPTOP_B_IP}:19093
      PEERS: http://${LAPTOP_A_IP}:19091,http://${LAPTOP_A_IP}:19092,http://${LAPTOP_B_IP}:19093,http://${LAPTOP_C_IP}:19094,http://${LAPTOP_C_IP}:19095
      LB_ADDR: :8000
      LB_CONTROL_ADDR: 127.0.0.1:18080
    ports:
      - "8003:8000"
      - "19093:19090" # Exposed on all IPs
    networks:
      - sdnet

  node4: # ... similar to node3

networks:
  sdnet:
    driver: bridge
```

Then start with environment variables:

```bash
LAPTOP_A_IP=192.168.1.100 \
LAPTOP_B_IP=192.168.1.101 \
LAPTOP_C_IP=192.168.1.102 \
docker compose -f docker-compose.laptopb.yaml up -d
```

### Laptop C Configuration (node5, backends 8-10)

Similar to Laptop B but with node5 and backends 8-10.

## Port Reference

| Service    | Laptop | Type | Port(s)  | Purpose                 |
| ---------- | ------ | ---- | -------- | ----------------------- |
| node1      | A      | Raft | 19091    | Consensus protocol      |
| node1      | A      | LB   | 8001     | Load balancer listener  |
| node2      | A      | Raft | 19092    | Consensus protocol      |
| node2      | A      | LB   | 8002     | Load balancer listener  |
| node3      | B      | Raft | 19093    | Consensus protocol      |
| node3      | B      | LB   | 8003     | Load balancer listener  |
| node4      | C      | Raft | 19094    | Consensus protocol      |
| node4      | C      | LB   | 8004     | Load balancer listener  |
| node5      | C      | Raft | 19095    | Consensus protocol      |
| node5      | C      | LB   | 8005     | Load balancer listener  |
| backend-1  | A      | HTTP | 8081     | Request handling        |
| backend-2  | A      | HTTP | 8082     | Request handling        |
| backend-3  | A      | HTTP | 8083     | Request handling        |
| backend-4  | B      | HTTP | 8084     | Request handling        |
| ...        | ...    | ...  | ...      | ...                     |
| backend-10 | C      | HTTP | 8090     | Request handling        |
| discovery  | A      | DNS  | 6699/UDP | Service discovery       |
| discovery  | A      | HTTP | 6701     | Discovery state queries |

## How Scripts Use cluster_config.yaml

### Bash Scripts

All bash scripts source `tooling/helper/cluster_config.sh` which provides functions:

```bash
source tooling/helper/cluster_config.sh

# Get a node's URL
get_node_url "node1" "/state"           # -> http://IP:19091/state

# Get a backend's URL
get_backend_url "backend-3"              # -> http://IP:8083

# Get LB URL
get_lb_url "node1" "/chat"               # -> http://IP:8001/chat

# Get discovery URL
get_discovery_url "http"                 # -> http://IP:6701
get_discovery_url "udp"                  # -> IP:6699

# Get all service names
get_all_nodes                            # -> node1 node2 ...
get_all_backends                         # -> backend-1 backend-2 ...
```

### Python Scripts

Python scripts use the `cluster_config` module:

```python
from tooling.helper.cluster_config import Config

cfg = Config.load()

# Get URLs
url = cfg.get_node_url("node1", "/state")
url = cfg.get_backend_url("backend-1")
url = cfg.get_lb_url("node1", "/chat")
url = cfg.get_discovery_url("http")

# Get service names
for node_name in cfg.get_node_names():
    ...
```

## Troubleshooting

### "cluster_config.yaml not found"

**Solution**: Ensure `cluster_config.yaml` exists in the repository root:

```bash
ls -la cluster_config.yaml
```

### Can't reach other laptops from scripts

**Solution**: Check that:

1. `cluster_config.yaml` has correct IP addresses
2. All laptops are on the same network (ping test)
3. Firewall allows ports (19091-19095, 8001-8090, 6699, 6701)
4. Docker containers bind to 0.0.0.0 (not 127.0.0.1)

### "Unknown node: node3" when running on Laptop A only

**Solution**: This is expected. Laptop A only has node1 and node2. Other nodes are on different laptops. The scripts will report them as "down" or unreachable, which is correct for a single-machine setup. To test the full cluster, run the scripts after starting services on all laptops.

### Port already in use

**Solution**: Check if another container/process is using the port:

```bash
lsof -i :19091   # Find what's using port 19091
docker compose down -v  # Bring down all containers
```

## Key Design Principles

1. **No localhost assumptions**: All container-to-container communication uses real machine IPs
2. **Configuration-driven**: All topology comes from `cluster_config.yaml`
3. **Topology-aware scripts**: Scripts read config and resolve endpoints dynamically
4. **Graceful degradation**: Missing/down services don't crash scripts
5. **Timeout discipline**: Short timeouts (0.6-1.5s) prevent hanging
6. **Port exposure**: All ports exposed to `0.0.0.0` (not just `127.0.0.1`)

## Running Full Cluster Demos

### Example: Test across all 3 laptops

1. **Laptop A**: `bash tooling/demo/up.sh`
2. **Laptop B**: Start node3, node4, backends 4-7 (using laptop-specific docker-compose)
3. **Laptop C**: Start node5, backends 8-10 (using laptop-specific docker-compose)
4. **Any machine**: `bash tooling/demo/state.sh` (queries all nodes)
5. **Any machine**: `bash tooling/demo/watch_health_propagation.sh 20` (watches all LBs)

## Next Steps

1. Update `cluster_config.yaml` with your actual laptop IPs
2. Start services on Laptop A: `bash tooling/demo/up.sh`
3. Test cluster state: `bash tooling/demo/state.sh`
4. Test routing: `bash tooling/demo/show_routing_headers.sh node1 /chat 5`
5. (Optional) Set up Laptops B and C with their respective services

## Support

For issues or questions:

- Check `cluster_config.yaml` is correctly configured
- Verify all services are running: `docker compose ps`
- Check logs: `docker compose logs -f node1`
- Validate network connectivity: `ping <laptop-ip>`
