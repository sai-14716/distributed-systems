# Demo helper scripts

These are optional convenience wrappers around `docker compose` to make the live demo easier.

Assumptions:
- You run them from the repo root (or they’ll `cd` to it).
- Your Compose project name is the default (`distributed-systems`). If you use a different name, set `COMPOSE_PROJECT_NAME`.

Common flows:

- Bring the cluster up (no load tests):
  - `bash tooling/demo/up.sh --clean`
  - `bash tooling/demo/up.sh`

- Show Raft state (roles/leader/term):
  - `bash tooling/demo/state.sh`
  - `bash tooling/demo/watch_state.sh`

- Stop the current leader (forces re-election):
  - `bash tooling/demo/stop_leader.sh`

- “Partition” a node (disconnect it from the `sdnet` network):
  - `bash tooling/demo/partition_node.sh isolate node3`
  - `bash tooling/demo/partition_node.sh restore node3`

- Run a k6 test on-demand:
  - `bash tooling/demo/loadtest.sh /app/k6/test_realistic_load.js`
  - `bash tooling/demo/loadtest.sh /app/k6/test_health_failover.js`

- Show LB routing decision headers (pool, candidates, chosen backend):
  - `bash tooling/demo/show_routing_headers.sh http://127.0.0.1:8001/chat 3`

- Demonstrate least-requests with in-flight avoidance:
  - `bash tooling/demo/least_requests_demo.sh http://127.0.0.1:8001`

- Compare WRR vs least-load under an induced hot backend:
  - `bash tooling/demo/load_weight_demo.sh wrr http://127.0.0.1:8001 50`
  - `bash tooling/demo/load_weight_demo.sh least-load http://127.0.0.1:8001 50`

- Watch health probing and propagation activity across all LB nodes:
  - `bash tooling/demo/watch_health_propagation.sh 20`

- Watch discovery DNS state (nodes, health, last probe, last error):
  - `bash tooling/demo/watch_discovery_state.sh 20`
  - `bash tooling/demo/watch_discovery_state.sh 30 http://127.0.0.1:6701/debug/state`

- Run all algorithm visibility demos in one shot:
  - `bash tooling/demo/run_all_algos_demo.sh http://127.0.0.1:8001 30`

