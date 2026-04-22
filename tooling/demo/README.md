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

