Architecture runtime in each node container:
- Data plane process (`/bin/lb`) listens on `:8000` for client traffic.
- Control plane process (`/bin/raft-node`) listens on `:19090` for Raft/admin APIs.
- Control plane pushes committed config locally to data plane via `127.0.0.1:18080`.

Live demo guide: see `docs/DEMO.md`.

Notes:
- `admin` and `k6-client` are behind Compose profiles (`tools`, `loadtest`) so `docker compose up` won’t accidentally start them.
- Each node’s Raft control plane is exposed on localhost for observability:
  - node1 `127.0.0.1:19091`, node2 `:19092`, node3 `:19093`, node4 `:19094`, node5 `:19095`

Start full stack for client tests:
docker compose up --build discovery backend-1 backend-2 backend-3 backend-4 backend-5 backend-6 backend-7 backend-8 backend-9 backend-10 node1 node2 node3 node4 node5

Push/load-balancer config through Raft control plane:
docker compose --profile tools run --rm admin -algorithm maglev
docker compose --profile tools run --rm admin -algorithm wrr
docker compose --profile tools run --rm admin -algorithm least-req
docker compose --profile tools run --rm admin -algorithm least-load

Run one client load container (defaults to stress test):
docker compose --profile loadtest up --build k6-client

Backend load simulation + monitoring:
- Backends do CPU-bound work per request (`WORK_MS`, default 50ms) and expose `/internal/load` with 10% CPU buckets.
- Each LB owns a subset of backends (rendezvous hashing; primary+secondary) and gossips load deltas to every other LB via `/internal/lb/gossip`.
- Load info is considered stale after 5s unless refreshed (owners refresh at least every 5s).

Visualizing what happens (logs + JSONL):
- LB logs include key=value events: `[lb-load] event=probe|publish|takeover|reclaim ...` and `[lb-gossip] event=send|recv ...`
- Convert live docker logs to JSONL with:
  - `docker compose logs -f node1 node2 node3 node4 node5 | python3 tooling/visualize/parse_lb_events.py > /tmp/lb_events.jsonl`
  - then inspect with `jq` (e.g. `jq -c 'select(.event==\"publish\")' /tmp/lb_events.jsonl | head`)

Run multiple client containers:
docker compose --profile loadtest up --build --scale k6-client=10 k6-client

Run a specific k6 test file:
docker compose --profile loadtest run --rm -e K6_SCRIPT=/app/k6/test_flow_consistency.js k6-client
docker compose --profile loadtest run --rm -e K6_SCRIPT=/app/k6/test_health_failover.js k6-client
docker compose --profile loadtest run --rm -e K6_SCRIPT=/app/k6/test_http2_mux.js k6-client

Run direct multi-node dataplane tests inside Docker (bypasses discovery proxy):
docker compose --profile loadtest run --rm \
	-e DISABLE_PROXY=1 \
	-e K6_SCRIPT=/app/k6/test_flow_consistency.js \
	-e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000 \
	k6-client

Quick dataplane status check on nodes:
curl http://localhost:8001/admin/status
curl http://localhost:8002/admin/status
curl http://localhost:8003/admin/status
curl http://localhost:8001/admin/load-view

Reset and restart cleanly:
docker compose down -v
docker compose up --build discovery backend-1 backend-2 backend-3 backend-4 backend-5 backend-6 backend-7 backend-8 backend-9 backend-10 node1 node2 node3 node4 node5


##Exactly how to run it:

Clean start
docker compose down -v

Start cluster (discovery, backends, 5 node containers with both processes)
docker compose up -d --build discovery backend-1 backend-2 backend-3 backend-4 backend-5 backend-6 backend-7 backend-8 backend-9 backend-10 node1 node2 node3 node4 node5

Optional: set LB algorithm through Raft control-plane
docker compose --profile tools run --rm admin -algorithm maglev
or
docker compose --profile tools run --rm admin -algorithm wrr
or
docker compose --profile tools run --rm admin -algorithm least-req

Run default client load test (discovery-proxy path, stress test)
docker compose --profile loadtest up --build --scale k6-client=10 k6-client

Run specific client tests (direct mode, no discovery proxy interception)
Use node dataplane endpoints as targets:
docker compose --profile loadtest run --rm -e DISABLE_PROXY=1 -e K6_SCRIPT=/app/k6/test_http11_hol.js -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000 k6-client

Example for realistic load:
docker compose --profile loadtest run --rm -e DISABLE_PROXY=1 -e K6_SCRIPT=/app/k6/test_realistic_load.js -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000 k6-client

Example for health failover:
docker compose --profile loadtest run --rm -e DISABLE_PROXY=1 -e K6_SCRIPT=/app/k6/test_health_failover.js -e LB_BASE_URL=http://node1:8000 -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000 k6-client

Quick verification:

Check dataplane status/config on node ports:
curl http://localhost:8001/admin/status
curl http://localhost:8002/admin/status
Send a dataplane request:
curl http://localhost:8001/chat

## Trace Collection (Local)

Generate a single traced request and store correlated logs locally:

./tooling/traces/request_trace.sh

Provide your own trace ID:

./tooling/traces/request_trace.sh demo-trace-123

Output is written to:

logs/<TRACE_ID>/

The script captures:
- client-proxy trace output from the request run
- LB data/control traces
- backend request traces

Generated files:
- metadata.log
- k6_client.log
- client_proxy.log
- lb_data.log
- lb_control.log
- backend.log

Optional continuous collection (recommended during load tests):

./tooling/traces/periodic_trace_collector.sh 30

This captures trace snapshots every 30 seconds into:

logs/live/<TIMESTAMP>/

Generated files per snapshot:
- client_proxy.log
- lb_data.log
- lb_control.log
- backend.log

Stop it with Ctrl+C.

## Fault Tolerance and Scale Test Playbook

Use these drills to validate control-plane resilience, data-plane failover, and request handling under load.

### 0) One-time setup

```bash
docker compose down -v
docker compose up -d --build discovery backend-1 backend-2 backend-3 backend-4 backend-5 backend-6 backend-7 backend-8 backend-9 backend-10 node1 node2 node3 node4 node5
```

Optional: start periodic log snapshots while testing.

```bash
./tooling/traces/periodic_trace_collector.sh 15
```

### 1) Drop leader while sending a config update

Goal: confirm Raft elects a new leader and config updates still converge.

1. Find current leader.

```bash
for n in node1 node2 node3 node4 node5; do
	echo -n "$n -> "
	docker compose exec -T "$n" sh -lc "wget -qO- http://127.0.0.1:19090/state" | grep -o '"role":"[^"]*"\|"leader_id":"[^"]*"'
done
```

2. Export the current leader (replace if needed based on output above).

```bash
LEADER=node1
```

3. Trigger config update and stop the leader at nearly the same time.

```bash
(docker compose --profile tools run --rm admin -algorithm wrr -probe-interval-ms 700 -health-threshold 0.70 &) \
; sleep 0.2 \
; docker compose stop "$LEADER" \
; wait
```

4. Verify a new leader is elected.

```bash
for n in node1 node2 node3 node4 node5; do
	echo -n "$n -> "
	docker compose exec -T "$n" sh -lc "wget -qO- http://127.0.0.1:19090/state" | grep -o '"role":"[^"]*"\|"leader_id":"[^"]*"'
done
```

5. Re-run one config update to confirm convergence after failover.

```bash
docker compose --profile tools run --rm admin -algorithm least-req -probe-interval-ms 900 -health-threshold 0.80
```

Expected result:
- one node reports role=leader
- admin command exits successfully with config convergence

### 2) Drop a backend and send requests immediately

Goal: confirm requests still succeed and client receives backend replies (ack).

1. Start request trace collector for one request.

```bash
./tooling/traces/request_trace.sh backend-drop-$(date +%s)
```

2. Stop one backend.

```bash
docker compose stop backend-2
```

3. Send another traced request after backend drop.

```bash
TRACE_ID=backend-drop-after-$(date +%s)
./tooling/traces/request_trace.sh "$TRACE_ID"
```

4. Validate client still receives ack and status 200.

```bash
LATEST_DIR=$(ls -dt logs/backend-drop-after-* | head -n 1)
cat "$LATEST_DIR/k6_client.log"
cat "$LATEST_DIR/lb_data.log"
cat "$LATEST_DIR/backend.log"
```

Expected result:
- client checks show status 200, backend id present, backend ack present
- lb_data shows chosen_backend is healthy backend
- backend log shows request served by remaining backend(s)

5. Recover backend.

```bash
docker compose up -d backend-2
```

### 3) Scale load and verify stability

Goal: exercise sustained throughput and confirm low failure rates.

1. Run direct multi-node realistic load.

```bash
docker compose --profile loadtest run --rm \
	-e DISABLE_PROXY=1 \
	-e K6_SCRIPT=/app/k6/test_realistic_load.js \
	-e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000,http://node4:8000,http://node5:8000 \
	k6-client
```

2. Increase concurrent clients via scaling.

```bash
docker compose --profile loadtest up --build --scale k6-client=10 k6-client
```

3. Watch node and backend health while load runs.

```bash
curl http://localhost:8001/admin/status
curl http://localhost:8002/admin/status
curl http://localhost:8003/admin/status
```

Expected result:
- k6 thresholds pass (low http_req_failed, checks near 100%)
- no sustained 5xx spikes during load

### 4) Combined chaos drill (recommended)

Run this full sequence in order:

1. Scale load test running.
2. Drop current leader while submitting config update.
3. Drop one backend during active traffic.
4. Run traced request and inspect logs under logs/<TRACE_ID>/.

Success criteria:
- control plane recovers leader election
- config updates eventually converge
- data plane keeps serving requests
- client receives backend ack messages end-to-end

### 5) One-command combined chaos drill

Run the full sequence automatically (leader drop during config update, backend drop, traced request, optional background load):

```bash
./tooling/chaos/run_combined_chaos_drill.sh
```

Optional custom run id:

```bash
./tooling/chaos/run_combined_chaos_drill.sh my-chaos-run
```

Useful environment overrides:

```bash
RUN_LOAD=0 STOP_BACKEND=backend-3 CONFIG_ALGO_1=maglev CONFIG_ALGO_2=least-req ./tooling/chaos/run_combined_chaos_drill.sh
```

Outputs:
- drill artifact logs: logs/<DRILL_ID>/
- traced request logs: logs/<DRILL_ID>-backend-drop/

Key files to inspect:
- logs/<DRILL_ID>/config_update_during_leader_drop.log
- logs/<DRILL_ID>/config_update_after_failover.log
- logs/<DRILL_ID>-backend-drop/k6_client.log
- logs/<DRILL_ID>-backend-drop/lb_data.log
- logs/<DRILL_ID>-backend-drop/backend.log
