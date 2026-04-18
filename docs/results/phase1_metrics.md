# Phase 1 and Phase 1.5 -- Load Balancer Metrics Report

Generated: 2026-04-14 10:55:51

---

## System Overview

| Property           | Value                        |
|--------------------|------------------------------|
| Load Balancers     | 1 (lb-1)                     |
| Backend Servers    | 3 (backend-1, 2, 3)          |
| Load (k6)          | 1,000 requests/sec for 10s   |
| Total Requests     | ~10,000 per algorithm run     |
| LB Algorithm       | Configurable via ALGO= env   |

---

### Phase 1 -- Round Robin (rr)

Cycles backends atomically in order 1->2->3->1. Stateless. Expected: even ~33/33/33% split.

| Backend   | Total Requests | Share (%) |
|-----------|----------------|-----------|
| backend-1   | 3333 | 33.3% |
| backend-2   | 3334 | 33.3% |
| backend-3   | 3334 | 33.3% |

<details><summary>Raw Controller Log Snapshot</summary>

```
2026/04/14 05:24:52 [Controller] Running health checks (Mocked). Active connections read from shared mem:
2026/04/14 05:24:52   -> backend-1 [http://backend-1:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3333, Weight: 0.35
2026/04/14 05:24:52   -> backend-2 [http://backend-2:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3334, Weight: 0.52
2026/04/14 05:24:52   -> backend-3 [http://backend-3:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3334, Weight: 0.66
```

</details>

---

### Phase 1.5 -- Weighted Round Robin (wrr)

Uses Controller dynamic Normal(0,1) weights to probabilistically route. Expected: uneven, weight-driven split.

| Backend   | Total Requests | Share (%) |
|-----------|----------------|-----------|
| backend-1   | 2833 | 28.3% |
| backend-2   | 3689 | 36.9% |
| backend-3   | 3478 | 34.8% |

<details><summary>Raw Controller Log Snapshot</summary>

```
2026/04/14 05:25:11 [Controller] Running health checks (Mocked). Active connections read from shared mem:
2026/04/14 05:25:11   -> backend-1 [http://backend-1:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 2833, Weight: 0.27
2026/04/14 05:25:11   -> backend-2 [http://backend-2:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3689, Weight: 1.46
2026/04/14 05:25:11   -> backend-3 [http://backend-3:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3478, Weight: 1.25
```

</details>

---

### Phase 1.5 -- Least Requests (least-req)

Routes to backend with fewest active connections. In fast local env, ActiveReqs drops to 0 instantly, so backend-1 dominates.

| Backend   | Total Requests | Share (%) |
|-----------|----------------|-----------|
| backend-1   | 7499 | 75% |
| backend-2   | 2008 | 20.1% |
| backend-3   | 493 | 4.9% |

<details><summary>Raw Controller Log Snapshot</summary>

```
2026/04/14 05:25:30 [Controller] Running health checks (Mocked). Active connections read from shared mem:
2026/04/14 05:25:30   -> backend-1 [http://backend-1:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 7499, Weight: 0.05
2026/04/14 05:25:30   -> backend-2 [http://backend-2:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 2008, Weight: 1.12
2026/04/14 05:25:30   -> backend-3 [http://backend-3:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 493, Weight: 1.59
```

</details>

---

### Phase 1.5 -- Maglev Consistent Hashing (maglev)

Consistent hashing via permutation table (M=251). L7 key: X-Client-VU HTTP header injected by k6. Expected: stable per-VU partitions.

| Backend   | Total Requests | Share (%) |
|-----------|----------------|-----------|
| backend-1   | 3333 | 33.3% |
| backend-2   | 3334 | 33.3% |
| backend-3   | 3333 | 33.3% |

<details><summary>Raw Controller Log Snapshot</summary>

```
2026/04/14 05:25:48 [Controller] Running health checks (Mocked). Active connections read from shared mem:
2026/04/14 05:25:48   -> backend-1 [http://backend-1:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3333, Weight: 0.76
2026/04/14 05:25:48   -> backend-2 [http://backend-2:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3334, Weight: 1.74
2026/04/14 05:25:48   -> backend-3 [http://backend-3:8080] Healthy: true, ActiveReqs: 0, TotalReqs: 3333, Weight: 0.38
```

</details>

---

## Summary

| Algorithm     | Distribution      | Key Insight                                                              |
|---------------|-------------------|--------------------------------------------------------------------------|
| Round Robin   | ~33% / 33% / 33%  | Perfect even distribution, stateless                                     |
| WRR           | Weight-driven     | Mirrors Controller's N(0,1) weights, dynamic per-second                  |
| Least Req     | Heavily skewed    | In fast local backends, ActiveReqs=0 almost always, backend-1 dominates  |
| Maglev        | ~33% / 33% / 33%  | Consistent per-VU routing, stable hash partitions                        |

*Results captured by k6/collect_metrics.ps1 -- 1000 RPS / 10s per algorithm run.*
