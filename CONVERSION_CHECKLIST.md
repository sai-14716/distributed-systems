# Distributed Conversion Checklist

Use this checklist to verify that your distributed conversion is complete and working.

## ✅ Pre-Setup Phase

- [ ] Clone/pull the repository
- [ ] Verify `cluster_config.yaml` exists in root directory
- [ ] Verify `docker-compose.yaml` is updated (only Laptop A services)
- [ ] Verify helper modules exist:
  - [ ] `tooling/helper/cluster_config.sh`
  - [ ] `tooling/helper/cluster_config.py`
- [ ] Verify docker-compose templates exist (optional):
  - [ ] `docker-compose.laptopb.yaml`
  - [ ] `docker-compose.laptopc.yaml`

## ✅ Configuration Phase

- [ ] Identify your machine IPs:
  - [ ] Laptop A IP: `_______________`
  - [ ] Laptop B IP: `_______________`
  - [ ] Laptop C IP: `_______________`
- [ ] Update `cluster_config.yaml` with these IPs
- [ ] Verify the file looks correct: `cat cluster_config.yaml | head -20`
- [ ] Verify YAML syntax is valid (no parsing errors)

## ✅ Laptop A Startup

- [ ] Ensure Docker Compose is installed and working
- [ ] Navigate to repository root: `cd /path/to/distributed-systems`
- [ ] Start Laptop A: `bash tooling/demo/up.sh`
- [ ] Verify output shows resolved endpoints (not hardcoded localhost)
- [ ] Verify all containers are running: `docker compose ps`
  - [ ] discovery (6699/udp, 6701)
  - [ ] node1 (19091, 8001)
  - [ ] node2 (19092, 8002)
  - [ ] backend-1 (8081)
  - [ ] backend-2 (8082)
  - [ ] backend-3 (8083)

## ✅ Functionality Testing

### Cluster State

- [ ] Check cluster state: `bash tooling/demo/state.sh`
- [ ] Output shows all nodes (at minimum node1 and node2)
- [ ] Output shows role (leader/follower) for each node
- [ ] No hardcoded localhost references in output

### Endpoint Accessibility

- [ ] Node1 state: `curl -s http://YOUR_IP:19091/state | jq .role`
  - Should output: `"leader"` or `"follower"`
- [ ] Node2 state: `curl -s http://YOUR_IP:19092/state | jq .role`
- [ ] Backend1: `curl -s http://YOUR_IP:8081/ | head -1`
  - Should return HTTP headers
- [ ] Discovery: `curl -s http://YOUR_IP:6701/ | jq .`
  - Should return JSON discovery state

### Load Balancer

- [ ] Check LB1 state: `curl -s http://YOUR_IP:8001/admin/load-view | jq .`
- [ ] Check LB2 state: `curl -s http://YOUR_IP:8002/admin/load-view | jq .`

## ✅ Script Refactoring Tests

- [ ] `bash tooling/demo/state.sh` works (no url parameter needed)
- [ ] `bash tooling/demo/watch_state.sh` works
- [ ] `bash tooling/demo/show_routing_headers.sh node1 /chat 3` works with new signature
- [ ] `bash tooling/demo/least_requests_demo.sh node1` works (node name, not url)
- [ ] `bash tooling/demo/load_weight_demo.sh wrr node1 10` works
- [ ] `bash tooling/demo/watch_health_propagation.sh 10` works
- [ ] `bash tooling/demo/watch_discovery_state.sh 5` works
- [ ] `bash tooling/demo/dynamic_backend_load.sh 30 0.5 1800 4` works

## ✅ Configuration Loading Tests

### Bash Helper

- [ ] `source tooling/helper/cluster_config.sh` loads without errors
- [ ] `get_node_url node1 /state` returns `http://YOUR_IP:19091/state`
- [ ] `get_backend_url backend-1` returns `http://YOUR_IP:8081`
- [ ] `get_lb_url node1 /chat` returns `http://YOUR_IP:8001/chat`
- [ ] `get_discovery_url http` returns `http://YOUR_IP:6701`
- [ ] `get_all_nodes` returns node list (at minimum node1 node2)
- [ ] `get_all_backends` returns backend list

### Python Helper

```bash
python3 -c "
import sys
sys.path.insert(0, 'tooling/helper')
from cluster_config import Config
cfg = Config.load()
print('Nodes:', cfg.get_node_names())
print('Node1 URL:', cfg.get_node_url('node1', '/state'))
print('Success!')
"
```

- [ ] Python script loads and prints correctly

## ✅ No Localhost Hardcoding

- [ ] Grep for hardcoded localhost: `grep -r "127.0.0.1" tooling/demo/ | wc -l`
  - Should return 0 (or only in docker-compose internal addresses)
- [ ] Grep for hardcoded service names: `grep -r "http://node[0-9]:" tooling/demo/ | wc -l`
  - Should return 0
- [ ] Port bindings don't include `127.0.0.1`:
  - `grep "127.0.0.1" docker-compose.yaml | wc -l` should be 0 or minimal

## ✅ Docker Compose Configuration

- [ ] Port bindings format: `PORT:CONTAINER_PORT` (not `127.0.0.1:PORT:CONTAINER_PORT`)
  - [ ] `docker compose config | grep "8081:8080"` works
- [ ] Environment variables use placeholders:
  - [ ] `docker compose config | grep "LAPTOP_A_IP"` shows references
- [ ] Networks section simplified (no static IPs)
- [ ] Only Laptop A services in main docker-compose.yaml:
  - [ ] node3, node4, node5 removed
  - [ ] backend-4 through backend-10 removed
  - [ ] k6-client and admin removed

## ✅ Optional: Multi-Machine Setup

If setting up laptops B and C:

### Laptop B

- [ ] `docker-compose.laptopb.yaml` exists with node3, node4, backends 4-7
- [ ] `LAPTOP_A_IP=... LAPTOP_B_IP=... LAPTOP_C_IP=... docker compose -f docker-compose.laptopb.yaml up -d` works
- [ ] Laptop A can query Laptop B nodes:
  - [ ] `curl -s http://LAPTOP_B_IP:19093/state | jq .role` returns leader/follower

### Laptop C

- [ ] `docker-compose.laptopc.yaml` exists with node5, backends 8-10
- [ ] Similar startup with environment variables works
- [ ] Full cluster communication verified

### Full Cluster Demo

- [ ] All 5 nodes are connected: `bash tooling/demo/state.sh` shows all nodes
- [ ] Lead election works: `bash tooling/demo/stop_leader.sh` stops leader, new one elected
- [ ] All 10 backends reachable from any laptop's LB

## ✅ Documentation

- [ ] `DISTRIBUTED_SETUP.md` exists and is readable
- [ ] `QUICK_REFERENCE.md` exists with command cheat sheet
- [ ] `CONVERSION_SUMMARY.md` exists documenting all changes
- [ ] `README_DISTRIBUTED.md` exists as entry point

## ✅ Cleanup & Finalization

- [ ] Removed any temporary test files
- [ ] Git status clean (no unexpected changes)
- [ ] All changes committed to version control
- [ ] Team notified of new configuration requirements

## ✅ Troubleshooting Verification

Test that error handling works:

- [ ] Stopping a backend doesn't crash scripts:
  - [ ] `docker compose stop backend-1`
  - [ ] `bash tooling/demo/state.sh` still runs (might show fewer backends)
  - [ ] `docker compose up -d backend-1` (restore it)
- [ ] Scripts handle missing cluster_config.yaml gracefully
  - [ ] Move config file temporarily: `mv cluster_config.yaml cluster_config.yaml.bak`
  - [ ] Try a script: `bash tooling/demo/state.sh`
  - [ ] Should show clear error message
  - [ ] Restore it: `mv cluster_config.yaml.bak cluster_config.yaml`

- [ ] Scripts timeout instead of hang:
  - [ ] Stop a critical service: `docker compose stop node1`
  - [ ] Run state query with timeout
  - [ ] Should timeout gracefully (not hang for minutes)

## ✅ Performance Verification

- [ ] Scripts complete in reasonable time (~1-2 seconds):
  - [ ] `time bash tooling/demo/state.sh`
  - [ ] Should be <5 seconds
- [ ] Discovery service responds quickly:
  - [ ] `curl -s http://YOUR_IP:6701/ | jq | head -20`
  - [ ] <1 second response time

## ✅ Final Sign-Off

- [ ] All checkmarks completed
- [ ] No known issues remaining
- [ ] Ready for production/testing deployment
- [ ] Documentation reviewed and understood
- [ ] Team trained on new commands and configuration

---

## Common Issues During Verification

| Issue                         | Check                                                                   |
| ----------------------------- | ----------------------------------------------------------------------- |
| cluster_config.yaml not found | `ls -la cluster_config.yaml`                                            |
| YAML parse error              | `python3 -c "import yaml; yaml.safe_load(open('cluster_config.yaml'))"` |
| Containers won't start        | `docker compose logs [service]`                                         |
| Scripts timeout               | Check `cluster_config.yaml` IPs are correct                             |
| Port already in use           | `lsof -i :PORT`                                                         |
| Scripts hang                  | Services likely unreachable - check firewall                            |

---

**Congratulations!** If you've completed all checkmarks, your distributed conversion is ready. 🎉
