# Quick Reference: Distributed Cluster Commands

## Setup

```bash
# 1. Configure your IPs (edit cluster_config.yaml)
edit cluster_config.yaml
# Set LAPTOP_A, LAPTOP_B, LAPTOP_C IPs

# 2. Start Laptop A
bash tooling/demo/up.sh

# 3. (Optional) Start Laptop B
LAPTOP_A_IP=... LAPTOP_B_IP=... LAPTOP_C_IP=... \
  docker compose -f docker-compose.laptopb.yaml up -d

# 4. (Optional) Start Laptop C
LAPTOP_A_IP=... LAPTOP_B_IP=... LAPTOP_C_IP=... \
  docker compose -f docker-compose.laptopc.yaml up -d
```

## Cluster Status

```bash
# Show current cluster state
bash tooling/demo/state.sh

# Watch state continuously
bash tooling/demo/watch_state.sh

# Stop watching (Ctrl+C)
```

## Testing Load Balancing

```bash
# Show routing headers for 5 requests
bash tooling/demo/show_routing_headers.sh node1 /chat 5

# Test least-requests avoidance
bash tooling/demo/least_requests_demo.sh node1

# Test WRR (weighted round-robin)
bash tooling/demo/load_weight_demo.sh wrr node1 30

# Run all algorithm demos
bash tooling/demo/run_all_algos_demo.sh node1 30
```

## Health and Discovery

```bash
# Watch health propagation across LBs
bash tooling/demo/watch_health_propagation.sh 20

# Watch discovery service state
bash tooling/demo/watch_discovery_state.sh 15

# Generate dynamic load on backends
bash tooling/demo/dynamic_backend_load.sh 60 0.5 1800 8
```

## Chaos Testing

```bash
# Isolate a node from network
bash tooling/demo/partition_node.sh isolate node1

# Restore the node
bash tooling/demo/partition_node.sh restore node1

# Stop current leader
bash tooling/demo/stop_leader.sh
```

## Direct Service Access

From **Laptop A** (or any machine with network access):

```bash
# Check node1 state
curl -s http://192.168.1.100:19091/state | jq

# Check node2 state
curl -s http://192.168.1.100:19092/state | jq

# Hit load balancer
curl -s http://192.168.1.100:8001/chat -H "X-Session-ID: test"

# Check backend
curl -s http://192.168.1.100:8081/

# Query discovery service
curl -s http://192.168.1.100:6701/ | jq
```

## Configuration

```bash
# View topology config
cat cluster_config.yaml

# Update IPs (must match ALL machines)
edit cluster_config.yaml
# Change: LAPTOP_A_IP, LAPTOP_B_IP, LAPTOP_C_IP
```

## Helper Functions (Bash)

```bash
source tooling/helper/cluster_config.sh

# Get specific node's URL
get_node_url "node1" "/state"

# Get load balancer URL
get_lb_url "node2" "/chat"

# Get backend URL
get_backend_url "backend-1"

# Get discovery URL
get_discovery_url "http"
get_discovery_url "udp"

# List all nodes
get_all_nodes

# List all backends
get_all_backends
```

## Helper Functions (Python)

```python
from tooling.helper.cluster_config import Config

cfg = Config.load()

# Get URLs
cfg.get_node_url("node1", "/state")
cfg.get_lb_url("node2", "/chat")
cfg.get_backend_url("backend-1")
cfg.get_discovery_url("http")

# List services
cfg.get_node_names()
cfg.get_backend_names()
```

## Debugging

```bash
# Check all containers running
docker compose ps

# View logs
docker compose logs -f node1
docker compose logs -f backend-1

# Check if service is up
curl -s -I http://192.168.1.100:8001/ | head -1

# Verify network connectivity
ping 192.168.1.100
nc -zv 192.168.1.100 19091
```

## Common Issues

| Issue                           | Solution                                           |
| ------------------------------- | -------------------------------------------------- |
| "cluster_config.yaml not found" | `ls cluster_config.yaml` in repo root              |
| Can't reach other laptops       | Check IPs in cluster_config.yaml match reality     |
| Port already in use             | `docker compose down -v`                           |
| Services down                   | `docker compose up -d`                             |
| Scripts say nodes are DOWN      | Normal if only running Laptop A; add other laptops |

## Full Cluster Demo

Run these in order on different machines:

```bash
# Laptop A (start cluster)
bash tooling/demo/up.sh

# Laptop B
docker compose -f docker-compose.laptopb.yaml up -d

# Laptop C
docker compose -f docker-compose.laptopc.yaml up -d

# Any machine (view state)
bash tooling/demo/state.sh

# Any machine (test routing)
bash tooling/demo/show_routing_headers.sh node1 /chat 10

# Any machine (watch health)
bash tooling/demo/watch_health_propagation.sh 20
```

## Notes

- All scripts work from **any** machine in the cluster
- Configuration is **shared** across all machines (edit cluster_config.yaml on all or use shared storage)
- **Firewall**: Ensure ports 19091-19095, 8001-8090, 6699, 6701 are open
- **IPs must be static** for reliable operation
- Services bind to `0.0.0.0` (all interfaces)
