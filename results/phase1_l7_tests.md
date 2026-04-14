# Phase 1.6 - L7 Test Results
Generated: 2026-04-14 19:22:20

---
### test_flow_consistency -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_flow_consistency.js
        output: -

     scenarios: (100.00%) 1 scenario, 50 max VUs, 10m30s max duration (incl. graceful stop):
              * flow_consistency: 10 iterations for each of 50 VUs (maxDuration: 10m0s, gracefulStop: 30s)


running (00m01.0s), 50/50 VUs, 209 complete and 0 interrupted iterations
flow_consistency   [  42% ] 50 VUs  00m01.0s/10m0s  209/500 iters, 10 per VU

running (00m02.0s), 00/50 VUs, 500 complete and 0 interrupted iterations
flow_consistency Γ£ô [ 100% ] 50 VUs  00m02.0s/10m0s  500/500 iters, 10 per VU


  Γûê THRESHOLDS 

    checks{test:same_server}
    Γ£ô 'rate==1.0' rate=0.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 1000   507.455022/s
    checks_succeeded...: 55.00% 550 out of 1000
    checks_failed......: 45.00% 450 out of 1000

    Γ£ô status 200 on first
    Γ£ô received full payload
    Γ£ù test:same_server
      Γå│  0% ΓÇö Γ£ô 0 / Γ£ù 450

    HTTP
    http_req_duration..............: avg=181.71ms min=56.87ms med=182.09ms max=393.44ms p(90)=193.06ms p(95)=235.79ms
      { expected_response:true }...: avg=181.71ms min=56.87ms med=182.09ms max=393.44ms p(90)=193.06ms p(95)=235.79ms
    http_req_failed................: 0.00%  0 out of 500
    http_reqs......................: 500    253.727511/s

    EXECUTION
    iteration_duration.............: avg=189.9ms  min=59.73ms med=184.84ms max=475.7ms  p(90)=196.3ms  p(95)=289.01ms
    iterations.....................: 500    253.727511/s
    vus............................: 50     min=50       max=50
    vus_max........................: 50     min=50       max=50

    NETWORK
    data_received..................: 82 kB  42 kB/s
    data_sent......................: 262 MB 133 MB/s




running (00m02.0s), 00/50 VUs, 500 complete and 0 interrupted iterations
flow_consistency Γ£ô [ 100% ] 50 VUs  00m02.0s/10m0s  500/500 iters, 10 per VU
```

---
### test_http11_hol -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_http11_hol.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 50s max duration (incl. graceful stop):
              * http11_hol: 20 looping VUs for 20s (gracefulStop: 30s)


running (01.0s), 20/20 VUs, 46 complete and 0 interrupted iterations
http11_hol   [   5% ] 20 VUs  01.0s/20s

running (02.0s), 20/20 VUs, 120 complete and 0 interrupted iterations
http11_hol   [  10% ] 20 VUs  02.0s/20s

running (03.0s), 20/20 VUs, 180 complete and 0 interrupted iterations
http11_hol   [  15% ] 20 VUs  03.0s/20s

running (04.0s), 20/20 VUs, 240 complete and 0 interrupted iterations
http11_hol   [  20% ] 20 VUs  04.0s/20s

running (05.0s), 20/20 VUs, 300 complete and 0 interrupted iterations
http11_hol   [  25% ] 20 VUs  05.0s/20s

running (06.0s), 20/20 VUs, 360 complete and 0 interrupted iterations
http11_hol   [  30% ] 20 VUs  06.0s/20s

running (07.0s), 20/20 VUs, 426 complete and 0 interrupted iterations
http11_hol   [  35% ] 20 VUs  07.0s/20s

running (08.0s), 20/20 VUs, 495 complete and 0 interrupted iterations
http11_hol   [  40% ] 20 VUs  08.0s/20s

running (09.0s), 20/20 VUs, 560 complete and 0 interrupted iterations
http11_hol   [  45% ] 20 VUs  09.0s/20s

running (10.0s), 20/20 VUs, 620 complete and 0 interrupted iterations
http11_hol   [  50% ] 20 VUs  10.0s/20s

running (11.0s), 20/20 VUs, 680 complete and 0 interrupted iterations
http11_hol   [  55% ] 20 VUs  11.0s/20s

running (12.0s), 20/20 VUs, 746 complete and 0 interrupted iterations
http11_hol   [  60% ] 20 VUs  12.0s/20s

running (13.0s), 20/20 VUs, 810 complete and 0 interrupted iterations
http11_hol   [  65% ] 20 VUs  13.0s/20s

running (14.0s), 20/20 VUs, 878 complete and 0 interrupted iterations
http11_hol   [  70% ] 20 VUs  14.0s/20s

running (15.0s), 20/20 VUs, 940 complete and 0 interrupted iterations
http11_hol   [  75% ] 20 VUs  15.0s/20s

running (16.0s), 20/20 VUs, 1000 complete and 0 interrupted iterations
http11_hol   [  80% ] 20 VUs  16.0s/20s

running (17.0s), 20/20 VUs, 1060 complete and 0 interrupted iterations
http11_hol   [  85% ] 20 VUs  17.0s/20s

running (18.0s), 20/20 VUs, 1125 complete and 0 interrupted iterations
http11_hol   [  90% ] 20 VUs  18.0s/20s

running (19.0s), 20/20 VUs, 1187 complete and 0 interrupted iterations
http11_hol   [  95% ] 20 VUs  19.0s/20s

running (20.0s), 20/20 VUs, 1255 complete and 0 interrupted iterations
http11_hol   [ 100% ] 20 VUs  20.0s/20s


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.95' rate=99.97%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 7650   376.906638/s
    checks_succeeded...: 99.97% 7648 out of 7650
    checks_failed......: 0.02%  2 out of 7650

    Γ£ô chat1 status 200
    Γ£ô chat2 status 200
    Γ£ô chat3 status 200
    Γ£ù chat1 took >= 200ms (HoL slow backend)
      Γå│  99% ΓÇö Γ£ô 1273 / Γ£ù 2
    Γ£ô HTTP/1.1 ordering: chat2 after chat1
    Γ£ô HTTP/1.1 ordering: chat3 after chat2

    HTTP
    http_req_duration..............: avg=71.65ms  min=0s       med=5.13ms   max=231.03ms p(90)=206.35ms p(95)=207.62ms
      { expected_response:true }...: avg=71.65ms  min=0s       med=5.13ms   max=231.03ms p(90)=206.35ms p(95)=207.62ms
    http_req_failed................: 0.00%  0 out of 3825
    http_reqs......................: 3825   188.453319/s

    EXECUTION
    iteration_duration.............: avg=316.53ms min=300.67ms med=315.58ms max=356.37ms p(90)=321.35ms p(95)=327.9ms 
    iterations.....................: 1275   62.817773/s
    vus............................: 20     min=20        max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 658 kB 32 kB/s
    data_sent......................: 805 kB 40 kB/s




running (20.3s), 00/20 VUs, 1275 complete and 0 interrupted iterations
http11_hol Γ£ô [ 100% ] 20 VUs  20s
```

---
### test_http2_mux -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_http2_mux.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 50s max duration (incl. graceful stop):
              * http2_multiplex: 20 looping VUs for 20s (gracefulStop: 30s)


running (01.0s), 20/20 VUs, 80 complete and 0 interrupted iterations
http2_multiplex   [   5% ] 20 VUs  01.0s/20s

running (02.0s), 20/20 VUs, 180 complete and 0 interrupted iterations
http2_multiplex   [  10% ] 20 VUs  02.0s/20s

running (03.0s), 20/20 VUs, 278 complete and 0 interrupted iterations
http2_multiplex   [  15% ] 20 VUs  03.0s/20s

running (04.0s), 20/20 VUs, 372 complete and 0 interrupted iterations
http2_multiplex   [  20% ] 20 VUs  04.0s/20s

running (05.0s), 20/20 VUs, 465 complete and 0 interrupted iterations
http2_multiplex   [  25% ] 20 VUs  05.0s/20s

running (06.0s), 20/20 VUs, 560 complete and 0 interrupted iterations
http2_multiplex   [  30% ] 20 VUs  06.0s/20s

running (07.0s), 20/20 VUs, 660 complete and 0 interrupted iterations
http2_multiplex   [  35% ] 20 VUs  07.0s/20s

running (08.0s), 20/20 VUs, 760 complete and 0 interrupted iterations
http2_multiplex   [  40% ] 20 VUs  08.0s/20s

running (09.0s), 20/20 VUs, 854 complete and 0 interrupted iterations
http2_multiplex   [  45% ] 20 VUs  09.0s/20s

running (10.0s), 20/20 VUs, 952 complete and 0 interrupted iterations
http2_multiplex   [  50% ] 20 VUs  10.0s/20s

running (11.0s), 20/20 VUs, 1045 complete and 0 interrupted iterations
http2_multiplex   [  55% ] 20 VUs  11.0s/20s

running (12.0s), 20/20 VUs, 1140 complete and 0 interrupted iterations
http2_multiplex   [  60% ] 20 VUs  12.0s/20s

running (13.0s), 20/20 VUs, 1240 complete and 0 interrupted iterations
http2_multiplex   [  65% ] 20 VUs  13.0s/20s

running (14.0s), 20/20 VUs, 1340 complete and 0 interrupted iterations
http2_multiplex   [  70% ] 20 VUs  14.0s/20s

running (15.0s), 20/20 VUs, 1437 complete and 0 interrupted iterations
http2_multiplex   [  75% ] 20 VUs  15.0s/20s

running (16.0s), 20/20 VUs, 1530 complete and 0 interrupted iterations
http2_multiplex   [  80% ] 20 VUs  16.0s/20s

running (17.0s), 20/20 VUs, 1626 complete and 0 interrupted iterations
http2_multiplex   [  85% ] 20 VUs  17.0s/20s

running (18.0s), 20/20 VUs, 1724 complete and 0 interrupted iterations
http2_multiplex   [  90% ] 20 VUs  18.0s/20s

running (19.0s), 20/20 VUs, 1820 complete and 0 interrupted iterations
http2_multiplex   [  95% ] 20 VUs  19.0s/20s

running (20.0s), 20/20 VUs, 1920 complete and 0 interrupted iterations
http2_multiplex   [ 100% ] 20 VUs  20.0s/20s


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.95' rate=100.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 9700    480.564212/s
    checks_succeeded...: 100.00% 9700 out of 9700
    checks_failed......: 0.00%   0 out of 9700

    Γ£ô chat1 stream 200
    Γ£ô chat2 stream 200
    Γ£ô chat3 stream 200
    Γ£ô all 3 streams succeeded
    Γ£ô HTTP/2: fast streams not blocked by slow stream

    HTTP
    http_req_duration..............: avg=72.45ms  min=0s       med=6.13ms   max=241.92ms p(90)=206.76ms p(95)=208ms   
      { expected_response:true }...: avg=72.45ms  min=0s       med=6.13ms   max=241.92ms p(90)=206.76ms p(95)=208ms   
    http_req_failed................: 0.00%  0 out of 5820
    http_reqs......................: 5820   288.338527/s

    EXECUTION
    iteration_duration.............: avg=207.36ms min=195.06ms med=206.47ms max=263.42ms p(90)=210.11ms p(95)=213.24ms
    iterations.....................: 1940   96.112842/s
    vus............................: 20     min=20        max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 1.0 MB 50 kB/s
    data_sent......................: 1.1 MB 54 kB/s




running (20.2s), 00/20 VUs, 1940 complete and 0 interrupted iterations
http2_multiplex Γ£ô [ 100% ] 20 VUs  20s
```

---
### test_sticky_rules -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_sticky_rules.js
        output: -

     scenarios: (100.00%) 1 scenario, 30 max VUs, 10m30s max duration (incl. graceful stop):
              * sticky_sessions: 10 iterations for each of 30 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  Γûê THRESHOLDS 

    checks{test:sticky}
    Γ£ô 'rate>0.95' rate=0.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 1200   7201.307757/s
    checks_succeeded...: 50.00% 600 out of 1200
    checks_failed......: 50.00% 600 out of 1200

    Γ£ô chat 200
    Γ£ô payload 200
    Γ£ù test:sticky /chat in backend-1 or backend-2
      Γå│  0% ΓÇö Γ£ô 0 / Γ£ù 300
    Γ£ù test:sticky /payload in backend-3
      Γå│  0% ΓÇö Γ£ô 0 / Γ£ù 300

    HTTP
    http_req_duration..............: avg=6.65ms  min=745.6┬╡s med=4.49ms  max=43.95ms p(90)=13.77ms p(95)=20.09ms
      { expected_response:true }...: avg=6.65ms  min=745.6┬╡s med=4.49ms  max=43.95ms p(90)=13.77ms p(95)=20.09ms
    http_req_failed................: 0.00%  0 out of 600
    http_reqs......................: 600    3600.653879/s

    EXECUTION
    iteration_duration.............: avg=15.09ms min=2.21ms  med=10.52ms max=69.63ms p(90)=34.01ms p(95)=42.91ms
    iterations.....................: 300    1800.326939/s

    NETWORK
    data_received..................: 100 kB 601 kB/s
    data_sent......................: 124 kB 745 kB/s




running (00m00.2s), 00/30 VUs, 300 complete and 0 interrupted iterations
sticky_sessions Γ£ô [ 100% ] 30 VUs  00m00.2s/10m0s  300/300 iters, 10 per VU
```

---
### test_realistic_load -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_realistic_load.js
        output: -

     scenarios: (100.00%) 1 scenario, 500 max VUs, 50s max duration (incl. graceful stop):
              * realistic_load: 1000.00 iterations/s for 20s (maxVUs: 200-500, gracefulStop: 30s)


running (01.0s), 012/200 VUs, 961 complete and 0 interrupted iterations
realistic_load   [   5% ] 012/200 VUs  01.0s/20s  1000.00 iters/s

running (02.0s), 014/200 VUs, 1959 complete and 0 interrupted iterations
realistic_load   [  10% ] 014/200 VUs  02.0s/20s  1000.00 iters/s

running (03.0s), 011/200 VUs, 2961 complete and 0 interrupted iterations
realistic_load   [  15% ] 012/200 VUs  03.0s/20s  1000.00 iters/s

running (04.0s), 012/200 VUs, 3961 complete and 0 interrupted iterations
realistic_load   [  20% ] 011/200 VUs  04.0s/20s  1000.00 iters/s

running (05.0s), 014/200 VUs, 4958 complete and 0 interrupted iterations
realistic_load   [  25% ] 015/200 VUs  05.0s/20s  1000.00 iters/s

running (06.0s), 011/200 VUs, 5961 complete and 0 interrupted iterations
realistic_load   [  30% ] 011/200 VUs  06.0s/20s  1000.00 iters/s

running (07.0s), 012/200 VUs, 6961 complete and 0 interrupted iterations
realistic_load   [  35% ] 012/200 VUs  07.0s/20s  1000.00 iters/s

running (08.0s), 012/200 VUs, 7960 complete and 0 interrupted iterations
realistic_load   [  40% ] 012/200 VUs  08.0s/20s  1000.00 iters/s

running (09.0s), 013/200 VUs, 8960 complete and 0 interrupted iterations
realistic_load   [  45% ] 013/200 VUs  09.0s/20s  1000.00 iters/s

running (10.0s), 013/200 VUs, 9959 complete and 0 interrupted iterations
realistic_load   [  50% ] 014/200 VUs  10.0s/20s  1000.00 iters/s

running (11.0s), 012/200 VUs, 10961 complete and 0 interrupted iterations
realistic_load   [  55% ] 012/200 VUs  11.0s/20s  1000.00 iters/s

running (12.0s), 014/200 VUs, 11959 complete and 0 interrupted iterations
realistic_load   [  60% ] 013/200 VUs  12.0s/20s  1000.00 iters/s

running (13.0s), 011/200 VUs, 12962 complete and 0 interrupted iterations
realistic_load   [  65% ] 011/200 VUs  13.0s/20s  1000.00 iters/s

running (14.0s), 012/200 VUs, 13961 complete and 0 interrupted iterations
realistic_load   [  70% ] 012/200 VUs  14.0s/20s  1000.00 iters/s

running (15.0s), 013/200 VUs, 14959 complete and 0 interrupted iterations
realistic_load   [  75% ] 013/200 VUs  15.0s/20s  1000.00 iters/s

running (16.0s), 012/200 VUs, 15961 complete and 0 interrupted iterations
realistic_load   [  80% ] 012/200 VUs  16.0s/20s  1000.00 iters/s

running (17.0s), 012/200 VUs, 16961 complete and 0 interrupted iterations
realistic_load   [  85% ] 012/200 VUs  17.0s/20s  1000.00 iters/s

running (18.0s), 012/200 VUs, 17961 complete and 0 interrupted iterations
realistic_load   [  90% ] 012/200 VUs  18.0s/20s  1000.00 iters/s

running (19.0s), 019/200 VUs, 18954 complete and 0 interrupted iterations
realistic_load   [  95% ] 019/200 VUs  19.0s/20s  1000.00 iters/s

running (20.0s), 015/200 VUs, 19958 complete and 0 interrupted iterations
realistic_load   [ 100% ] 015/200 VUs  20.0s/20s  1000.00 iters/s


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.99' rate=100.00%

    http_req_duration
    Γ£ô 'p(95)<300' p(95)=7.1ms

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 23989   1198.630107/s
    checks_succeeded...: 100.00% 23989 out of 23989
    checks_failed......: 0.00%   0 out of 23989

    Γ£ô chat 200
    Γ£ô admin chat 200
    Γ£ô payload 200
    Γ£ô bytes received correct

    HTTP
    http_req_duration..............: avg=3.08ms  min=506.49┬╡s med=2.37ms  max=88.3ms   p(90)=4.78ms  p(95)=7.1ms  
      { expected_response:true }...: avg=3.08ms  min=506.49┬╡s med=2.37ms  max=88.3ms   p(90)=4.78ms  p(95)=7.1ms  
    http_req_failed................: 0.00%  0 out of 20001
    http_reqs......................: 20001  999.366408/s

    EXECUTION
    iteration_duration.............: avg=13.65ms min=10.51ms  med=12.91ms max=100.49ms p(90)=15.76ms p(95)=18.15ms
    iterations.....................: 20001  999.366408/s
    vus............................: 14     min=11         max=19 
    vus_max........................: 200    min=200        max=200

    NETWORK
    data_received..................: 3.4 MB 171 kB/s
    data_sent......................: 266 MB 13 MB/s




running (20.0s), 000/200 VUs, 20001 complete and 0 interrupted iterations
realistic_load Γ£ô [ 100% ] 000/200 VUs  20s  1000.00 iters/s
```

---
### test_health_failover -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_health_failover.js
        output: -

     scenarios: (100.00%) 3 scenarios, 201 max VUs, 10m35s max duration (incl. graceful stop):
              * before_failure: 200.00 iterations/s for 5s (maxVUs: 50-100, gracefulStop: 30s)
              * trigger_failure: 1 iterations shared among 1 VUs (maxDuration: 10m0s, startTime: 5s, gracefulStop: 30s)
              * after_failure: 200.00 iterations/s for 10s (maxVUs: 50-100, startTime: 6s, gracefulStop: 30s)


running (00m01.0s), 000/101 VUs, 194 complete and 0 interrupted iterations
before_failure    [  19% ] 000/050 VUs  1.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      4.0s    
after_failure   ΓÇó [   0% ] waiting      5.0s    

running (00m02.0s), 001/101 VUs, 394 complete and 0 interrupted iterations
before_failure    [  39% ] 001/050 VUs  2.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      3.0s    
after_failure   ΓÇó [   0% ] waiting      4.0s    

running (00m03.0s), 001/101 VUs, 594 complete and 0 interrupted iterations
before_failure    [  59% ] 001/050 VUs  3.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      2.0s    
after_failure   ΓÇó [   0% ] waiting      3.0s    

running (00m04.0s), 001/101 VUs, 794 complete and 0 interrupted iterations
before_failure    [  79% ] 001/050 VUs  4.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      1.0s    
after_failure   ΓÇó [   0% ] waiting      2.0s    

running (00m05.0s), 001/101 VUs, 994 complete and 0 interrupted iterations
before_failure    [  99% ] 001/050 VUs  5.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      0.0s    
after_failure   ΓÇó [   0% ] waiting      1.0s    

running (00m06.0s), 000/101 VUs, 1002 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure   ΓÇó [   0% ] waiting      0.0s           

running (00m07.0s), 000/101 VUs, 1196 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  10% ] 000/050 VUs  01.0s/10s       200.00 iters/s

running (00m08.0s), 001/101 VUs, 1395 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  20% ] 002/050 VUs  02.0s/10s       200.00 iters/s

running (00m09.0s), 000/101 VUs, 1596 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  30% ] 000/050 VUs  03.0s/10s       200.00 iters/s

running (00m10.0s), 002/101 VUs, 1794 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  40% ] 002/050 VUs  04.0s/10s       200.00 iters/s

running (00m11.0s), 000/101 VUs, 1996 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  50% ] 000/050 VUs  05.0s/10s       200.00 iters/s

running (00m12.0s), 002/101 VUs, 2195 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  60% ] 002/050 VUs  06.0s/10s       200.00 iters/s

running (00m13.0s), 001/101 VUs, 2395 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  70% ] 001/050 VUs  07.0s/10s       200.00 iters/s

running (00m14.0s), 000/101 VUs, 2596 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  80% ] 000/050 VUs  08.0s/10s       200.00 iters/s

running (00m15.0s), 000/101 VUs, 2796 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [  90% ] 000/050 VUs  09.0s/10s       200.00 iters/s

running (00m16.0s), 001/101 VUs, 2995 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure     [ 100% ] 001/050 VUs  10.0s/10s       200.00 iters/s


  Γûê THRESHOLDS 

    checks{test:no_failed_backend}
    Γ£ô 'rate>0.99' rate=0.00%

    http_req_failed
    Γ£ô 'rate<0.05' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 6006    375.214082/s
    checks_succeeded...: 100.00% 6006 out of 6006
    checks_failed......: 0.00%   0 out of 6006

    Γ£ô request succeeds
    Γ£ô test:no_failed_backend

    HTTP
    http_req_duration..............: avg=3.5ms  min=517.4┬╡s med=2.88ms max=24.35ms p(90)=6.32ms p(95)=8.13ms
      { expected_response:true }...: avg=3.5ms  min=517.4┬╡s med=2.88ms max=24.35ms p(90)=6.32ms p(95)=8.13ms
    http_req_failed................: 0.00%  0 out of 3003
    http_reqs......................: 3003   187.607041/s

    EXECUTION
    iteration_duration.............: avg=3.71ms min=517.4┬╡s med=3ms    max=25.61ms p(90)=6.79ms p(95)=8.38ms
    iterations.....................: 3003   187.607041/s
    vus............................: 1      min=0         max=2  
    vus_max........................: 101    min=101       max=101

    NETWORK
    data_received..................: 522 kB 33 kB/s
    data_sent......................: 510 kB 32 kB/s




running (00m16.0s), 000/101 VUs, 3003 complete and 0 interrupted iterations
before_failure  Γ£ô [ 100% ] 000/050 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        00m00.0s/10m0s  1/1 shared iters
after_failure   Γ£ô [ 100% ] 000/050 VUs  10s             200.00 iters/s
```

---
### test_sticky_failover -- FAILED (exit 99)

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: .\k6\test_sticky_failover.js
        output: -

     scenarios: (100.00%) 3 scenarios, 21 max VUs, 10m39s max duration (incl. graceful stop):
              * establish_stickiness: 5 iterations for each of 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)
              * trigger_failure: 1 iterations shared among 1 VUs (maxDuration: 10m0s, startTime: 8s, gracefulStop: 30s)
              * after_failover: 5 iterations for each of 10 VUs (maxDuration: 10m0s, startTime: 9s, gracefulStop: 30s)


running (00m01.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  7.0s           
after_failover       ΓÇó [   0% ] waiting  8.0s           

running (00m02.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  6.0s           
after_failover       ΓÇó [   0% ] waiting  7.0s           

running (00m03.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  5.0s           
after_failover       ΓÇó [   0% ] waiting  6.0s           

running (00m04.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  4.0s           
after_failover       ΓÇó [   0% ] waiting  5.0s           

running (00m05.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  3.0s           
after_failover       ΓÇó [   0% ] waiting  4.0s           

running (00m06.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  2.0s           
after_failover       ΓÇó [   0% ] waiting  3.0s           

running (00m07.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  1.0s           
after_failover       ΓÇó [   0% ] waiting  2.0s           

running (00m08.0s), 00/21 VUs, 50 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  0.0s           
after_failover       ΓÇó [   0% ] waiting  1.0s           

running (00m09.0s), 00/21 VUs, 51 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs   00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs    00m00.1s/10m0s  1/1 shared iters
after_failover       ΓÇó [   0% ] waiting  0.0s           


  Γûê THRESHOLDS 

    checks
    Γ£ù 'rate>0.95' rate=50.00%

    http_req_failed
    Γ£ô 'rate<0.02' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 202    21.15682/s
    checks_succeeded...: 50.00% 101 out of 202
    checks_failed......: 50.00% 101 out of 202

    Γ£ô session request succeeds after failover
    Γ£ù server responded (remap occurred)
      Γå│  0% ΓÇö Γ£ô 0 / Γ£ù 101

    HTTP
    http_req_duration..............: avg=5.49ms   min=2.69ms   med=4.63ms   max=11.7ms   p(90)=8.49ms   p(95)=9.59ms  
      { expected_response:true }...: avg=5.49ms   min=2.69ms   med=4.63ms   max=11.7ms   p(90)=8.49ms   p(95)=9.59ms  
    http_req_failed................: 0.00% 0 out of 101
    http_reqs......................: 101   10.57841/s

    EXECUTION
    iteration_duration.............: avg=107.17ms min=103.16ms med=105.64ms max=117.38ms p(90)=112.32ms p(95)=114.69ms
    iterations.....................: 101   10.57841/s
    vus............................: 0     min=0        max=0 
    vus_max........................: 21    min=21       max=21

    NETWORK
    data_received..................: 18 kB 1.8 kB/s
    data_sent......................: 19 kB 2.0 kB/s




running (00m09.5s), 00/21 VUs, 101 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs  00m00.5s/10m0s  50/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   00m00.1s/10m0s  1/1 shared iters
after_failover       Γ£ô [ 100% ] 10 VUs  00m00.5s/10m0s  50/50 iters, 5 per VU
time="2026-04-14T19:23:53+05:30" level=error msg="thresholds on metrics 'checks' have been crossed"
```

