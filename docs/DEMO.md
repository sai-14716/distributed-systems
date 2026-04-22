# Live demo runbook

Goal: start the system with **no client tests running**, then trigger failures + load on demand while showing what’s happening.

## 0) Pre-demo prep

Build everything (including the optional tool/loadtest images):

```bash
docker compose build
docker compose --profile tools build admin
docker compose --profile loadtest build k6-client
```

Clean start:

```bash
docker compose down -v
```

## 1) Start the cluster (no tests)

```bash
bash tooling/demo/up.sh --clean
```

Observability checks (from host):

```bash
bash tooling/demo/state.sh
curl -s http://127.0.0.1:8001/admin/status
curl -s http://127.0.0.1:8001/admin/load-view
curl -s http://127.0.0.1:8081/health
```

If you want a live “leader election monitor”, run this in a second terminal:

```bash
bash tooling/demo/watch_state.sh
```

## 2) Show config changes replicated via Raft

Switch LB algorithm (write goes through Raft and is applied on every node’s dataplane):

```bash
docker compose --profile tools run --rm admin -algorithm wrr
docker compose --profile tools run --rm admin -algorithm least-req
```

Curl equivalent from host (detect leader first, then submit):

```bash
LEADER_PORT=$(for p in 19091 19092 19093 19094 19095; do
  role=$(curl -sf "http://127.0.0.1:${p}/state" | jq -r '.role // empty') || continue
  if [ "$role" = "leader" ]; then echo "$p"; break; fi
done)

curl -sS -X POST "http://127.0.0.1:${LEADER_PORT}/admin/submit" \
  -H 'Content-Type: application/json' \
  -d '{"type":"set_config","data":{"algorithm":"wrr","probe_interval_ms":1000,"health_threshold":0.8}}'
```

Verify the dataplane is using the new config:

```bash
curl -s http://127.0.0.1:8001/admin/status
```

Show routing behavior changed (simplest visual check):

```bash
bash tooling/demo/routing_behaviour.sh round_robin
bash tooling/demo/routing_behaviour.sh maglev
```

If both runs show the same backend, try a different key (for example `demo-fixed-2`), since hashing can map some keys to the same backend under both algorithms.

## 2.1) Make algorithm choice visible per request

Show LB decision headers for real requests:

```bash
bash tooling/demo/show_routing_headers.sh http://127.0.0.1:8001/chat 3
```

This prints:
- algorithm in use
- pool source and pool members considered
- candidates with active requests + CPU buckets
- chosen backend

Demonstrate that algorithms behave differently under controlled conditions:

```bash
# Least-Requests: short request should avoid the backend with more active requests
bash tooling/demo/least_requests_demo.sh http://127.0.0.1:8001

# WRR vs Least-Load: induce heat on backend-3, compare distributions
bash tooling/demo/load_weight_demo.sh wrr http://127.0.0.1:8001 50
bash tooling/demo/load_weight_demo.sh least-load http://127.0.0.1:8001 50
```

Each run prints the current load view first and again after the burst, so you can see the CPU buckets that the LB is routing on.

Single-command version (runs all algorithm visibility demos):

```bash
bash tooling/demo/run_all_algos_demo.sh http://127.0.0.1:8001 30
```

## 2.2) Make health probing + propagation visible

Watch probe and propagation activity from every LB node in real time:

```bash
bash tooling/demo/watch_health_propagation.sh 20
```

Run this while creating load/failure conditions in another terminal, for example:

```bash
# Heat backend-3 through sticky /payload rule
for i in $(seq 1 10); do
  curl -s -o /dev/null http://127.0.0.1:8001/payload -H "X-Work-Ms: 1200" --data-binary x
done

# Or force health transition
docker compose stop backend-3
sleep 5
docker compose up -d backend-3
```

What to point out live:
- `probe_total` increments continuously (active probing)
- `fail` increments during outages
- `cpu_bucket` moves in 5% steps, so hot load is easier to see live
- backend `view` epoch/seq/status changes on owner
- `recv_from=...` updates across peers (propagation)

## 3) Drop an LB (node) and show control-plane recovery

Stop the current Raft leader (forces re-election):

```bash
bash tooling/demo/stop_leader.sh
```

Expected: within a couple seconds, `bash tooling/demo/watch_state.sh` shows a new leader.

Bring the node back:

```bash
docker compose up -d <stopped-node>
```

## 4) Drop a backend and show dataplane failover

Stop one backend:

```bash
docker compose stop backend-2
```

Send a few requests through different LBs (should still succeed):

```bash
curl -s http://127.0.0.1:8001/chat
curl -s http://127.0.0.1:8003/chat
curl -s http://127.0.0.1:8005/chat
```

Recover backend:

```bash
docker compose up -d backend-2
```

## 5) Partition a node (simulate a network partition)

Isolate a follower from the `sdnet` network:

```bash
bash tooling/demo/partition_node.sh isolate node5
```

Expected: remaining majority continues; the isolated node typically flips to `candidate` (can’t reach quorum), while the majority keeps a stable leader.

Restore the node:

```bash
bash tooling/demo/partition_node.sh restore node5
```

## 6) Run load tests on-demand (during the demo)

Single realistic load test (direct mode, bypass discovery proxy):

```bash
docker compose --profile loadtest run --rm \
  -e DISABLE_PROXY=1 \
  -e K6_SCRIPT=/app/k6/test_realistic_load.js \
  -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000,http://node4:8000,http://node5:8000 \
  k6-client
```

Scale load generators:

```bash
docker compose --profile loadtest up --build --scale k6-client=10 k6-client
```

## Reset between runs

```bash
docker compose down -v
```
