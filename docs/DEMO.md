bash tooling/demo/build_demo_images.sh
docker compose down -v

bash tooling/demo/up.sh --clean
bash tooling/demo/state.sh
bash tooling/demo/show_cluster_status.sh node1 backend-1
bash tooling/demo/watch_state.sh
bash tooling/demo/watch_raft_elections.sh
bash tooling/demo/watch_raft_elections_detailed.sh

bash tooling/demo/apply_algorithm.sh wrr
bash tooling/demo/apply_algorithm.sh least-req
bash tooling/demo/show_cluster_status.sh node1 backend-1
bash tooling/demo/routing_behaviour.sh round_robin node1
bash tooling/demo/routing_behaviour.sh maglev node1

bash tooling/demo/show_routing_headers.sh node1 /chat 3
bash tooling/demo/least_requests_demo.sh node1
bash tooling/demo/dynamic_backend_load.sh 90 3 1800 8
bash tooling/demo/least_requests_with_dynamic_load.sh node1 20
bash tooling/demo/wrr_with_dynamic_load.sh node1 20
bash tooling/demo/show_routing_headers.sh node1 / 5
bash tooling/demo/run_all_algos_demo.sh node1 30

bash tooling/demo/compare_ring_assignment.sh
bash tooling/demo/watch_health_propagation.sh 20
bash tooling/demo/toggle_backend_health.sh backend-3
bash tooling/demo/stop_leader.sh
bash tooling/demo/restart_node.sh node5
bash tooling/demo/backend_failover.sh backend-2 node1 node2 node3
bash tooling/demo/partition_node.sh isolate node5
bash tooling/demo/partition_node.sh restore node5

bash tooling/demo/loadtest.sh /app/k6/test_realistic_load.js -e DISABLE_PROXY=1 -e LB_BASE_URLS=http://node1:8000,http://node2:8000,http://node3:8000,http://node4:8000,http://node5:8000
docker compose --profile loadtest up --build --scale k6-client=10 k6-client

docker compose down -v
