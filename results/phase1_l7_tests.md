# Phase 1.6 - L7 Test Results
Generated: 2026-04-19 02:03:41

---
### test_flow_consistency -- PASSED

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_flow_consistency.js
        output: -

     scenarios: (100.00%) 1 scenario, 50 max VUs, 10m30s max duration (incl. graceful stop):
              * flow_consistency: 10 iterations for each of 50 VUs (maxDuration: 10m0s, gracefulStop: 30s)


running (00m00.9s), 50/50 VUs, 0 complete and 0 interrupted iterations
flow_consistency   [   0% ] 50 VUs  00m01.0s/10m0s  000/500 iters, 10 per VU

running (00m01.9s), 50/50 VUs, 9 complete and 0 interrupted iterations
flow_consistency   [   2% ] 50 VUs  00m01.9s/10m0s  009/500 iters, 10 per VU

running (00m03.0s), 50/50 VUs, 18 complete and 0 interrupted iterations
flow_consistency   [   4% ] 50 VUs  00m03.0s/10m0s  018/500 iters, 10 per VU

running (00m04.0s), 50/50 VUs, 28 complete and 0 interrupted iterations
flow_consistency   [   6% ] 50 VUs  00m04.0s/10m0s  028/500 iters, 10 per VU

running (00m05.0s), 50/50 VUs, 43 complete and 0 interrupted iterations
flow_consistency   [   9% ] 50 VUs  00m05.0s/10m0s  043/500 iters, 10 per VU

running (00m06.0s), 50/50 VUs, 57 complete and 0 interrupted iterations
flow_consistency   [  11% ] 50 VUs  00m06.0s/10m0s  057/500 iters, 10 per VU
What is the net
running (00m06.9s), 50/50 VUs, 80 complete and 0 interrupted iterations
flow_consistency   [  16% ] 50 VUs  00m06.9s/10m0s  080/500 iters, 10 per VU

running (00m07.9s), 50/50 VUs, 127 complete and 0 interrupted iterations
flow_consistency   [  25% ] 50 VUs  00m07.9s/10m0s  127/500 iters, 10 per VU

running (00m08.9s), 50/50 VUs, 176 complete and 0 interrupted iterations
flow_consistency   [  35% ] 50 VUs  00m08.9s/10m0s  176/500 iters, 10 per VU

running (00m09.9s), 50/50 VUs, 229 complete and 0 interrupted iterations
flow_consistency   [  46% ] 50 VUs  00m09.9s/10m0s  229/500 iters, 10 per VU

running (00m10.9s), 50/50 VUs, 280 complete and 0 interrupted iterations
flow_consistency   [  56% ] 50 VUs  00m10.9s/10m0s  280/500 iters, 10 per VU

running (00m12.0s), 50/50 VUs, 330 complete and 0 interrupted iterations
flow_consistency   [  66% ] 50 VUs  00m12.0s/10m0s  330/500 iters, 10 per VU

running (00m12.9s), 50/50 VUs, 378 complete and 0 interrupted iterations
flow_consistency   [  76% ] 50 VUs  00m13.0s/10m0s  378/500 iters, 10 per VU

running (00m13.9s), 50/50 VUs, 428 complete and 0 interrupted iterations
flow_consistency   [  86% ] 50 VUs  00m13.9s/10m0s  428/500 iters, 10 per VU

running (00m14.9s), 27/50 VUs, 473 complete and 0 interrupted iterations
flow_consistency   [  95% ] 50 VUs  00m14.9s/10m0s  473/500 iters, 10 per VU

running (00m15.8s), 00/50 VUs, 500 complete and 0 interrupted iterations
flow_consistency Γ£ô [ 100% ] 50 VUs  00m15.8s/10m0s  500/500 iters, 10 per VU


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.99' rate=100.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 1550    98.102316/s
    checks_succeeded...: 100.00% 1550 out of 1550
    checks_failed......: 0.00%   0 out of 1550

    Γ£ô status 200 on first
    Γ£ô first response has backend id
    Γ£ô received full payload
    Γ£ô backend ack present
    Γ£ô same backend for session

    HTTP
    http_req_duration..............: avg=1.46s min=834.14ms med=1s    max=6.46s p(90)=3.13s p(95)=3.64s
      { expected_response:true }...: avg=1.46s min=834.14ms med=1s    max=6.46s p(90)=3.13s p(95)=3.64s
    http_req_failed................: 0.00%  0 out of 500
    http_reqs......................: 500    31.645908/s

    EXECUTION
    iteration_duration.............: avg=1.49s min=861.65ms med=1.02s max=6.58s p(90)=3.16s p(95)=3.73s
    iterations.....................: 500    31.645908/s
    vus............................: 27     min=27       max=50
    vus_max........................: 50     min=50       max=50

    NETWORK
    data_received..................: 262 MB 17 MB/s
    data_sent......................: 262 MB 17 MB/s




running (00m15.8s), 00/50 VUs, 500 complete and 0 interrupted iterations
flow_consistency Γ£ô [ 100% ] 50 VUs  00m15.8s/10m0s  500/500 iters, 10 per VU
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
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_http11_hol.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 50s max duration (incl. graceful stop):
              * http11_hol: 20 looping VUs for 20s (gracefulStop: 30s)


running (01.0s), 20/20 VUs, 40 complete and 0 interrupted iterations
http11_hol   [   5% ] 20 VUs  01.0s/20s

running (02.0s), 20/20 VUs, 102 complete and 0 interrupted iterations
http11_hol   [  10% ] 20 VUs  02.0s/20s

running (03.0s), 20/20 VUs, 166 complete and 0 interrupted iterations
http11_hol   [  15% ] 20 VUs  03.0s/20s

running (04.0s), 20/20 VUs, 230 complete and 0 interrupted iterations
http11_hol   [  20% ] 20 VUs  04.0s/20s

running (05.0s), 20/20 VUs, 291 complete and 0 interrupted iterations
http11_hol   [  25% ] 20 VUs  05.0s/20s

running (06.0s), 20/20 VUs, 360 complete and 0 interrupted iterations
http11_hol   [  30% ] 20 VUs  06.0s/20s

running (07.0s), 20/20 VUs, 414 complete and 0 interrupted iterations
http11_hol   [  35% ] 20 VUs  07.0s/20s

running (08.0s), 20/20 VUs, 482 complete and 0 interrupted iterations
http11_hol   [  40% ] 20 VUs  08.0s/20s

running (09.0s), 20/20 VUs, 535 complete and 0 interrupted iterations
http11_hol   [  45% ] 20 VUs  09.0s/20s

running (10.0s), 20/20 VUs, 586 complete and 0 interrupted iterations
http11_hol   [  50% ] 20 VUs  10.0s/20s

running (11.0s), 20/20 VUs, 649 complete and 0 interrupted iterations
http11_hol   [  55% ] 20 VUs  11.0s/20s

running (12.0s), 20/20 VUs, 706 complete and 0 interrupted iterations
http11_hol   [  60% ] 20 VUs  12.0s/20s

running (13.0s), 20/20 VUs, 767 complete and 0 interrupted iterations
http11_hol   [  65% ] 20 VUs  13.0s/20s

running (14.0s), 20/20 VUs, 826 complete and 0 interrupted iterations
http11_hol   [  70% ] 20 VUs  14.0s/20s

running (15.0s), 20/20 VUs, 883 complete and 0 interrupted iterations
http11_hol   [  75% ] 20 VUs  15.0s/20s

running (16.0s), 20/20 VUs, 943 complete and 0 interrupted iterations
http11_hol   [  80% ] 20 VUs  16.0s/20s

running (17.0s), 20/20 VUs, 1001 complete and 0 interrupted iterations
http11_hol   [  85% ] 20 VUs  17.0s/20s

running (18.0s), 20/20 VUs, 1063 complete and 0 interrupted iterations
http11_hol   [  90% ] 20 VUs  18.0s/20s

running (19.0s), 20/20 VUs, 1123 complete and 0 interrupted iterations
http11_hol   [  95% ] 20 VUs  19.0s/20s

running (20.0s), 20/20 VUs, 1186 complete and 0 interrupted iterations
http11_hol   [ 100% ] 20 VUs  20.0s/20s


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.95' rate=100.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 10890   536.072234/s
    checks_succeeded...: 100.00% 10890 out of 10890
    checks_failed......: 0.00%   0 out of 10890

    Γ£ô chat1 status 200
    Γ£ô chat2 status 200
    Γ£ô chat3 status 200
    Γ£ô chat1 backend ack present
    Γ£ô chat2 backend ack present
    Γ£ô chat3 backend ack present
    Γ£ô chat1 took >= 200ms (HoL slow backend)
    Γ£ô HTTP/1.1 ordering: chat2 after chat1
    Γ£ô HTTP/1.1 ordering: chat3 after chat2

    HTTP
    http_req_duration..............: avg=76.86ms  min=1.09ms   med=10.35ms  max=335.03ms p(90)=210.93ms p(95)=214.11ms
      { expected_response:true }...: avg=76.86ms  min=1.09ms   med=10.35ms  max=335.03ms p(90)=210.93ms p(95)=214.11ms
    http_req_failed................: 0.00%  0 out of 3630
    http_reqs......................: 3630   178.690745/s

    EXECUTION
    iteration_duration.............: avg=333.76ms min=307.51ms med=326.96ms max=586.11ms p(90)=352.4ms  p(95)=371.04ms
    iterations.....................: 1210   59.563582/s
    vus............................: 20     min=20        max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 1.3 MB 66 kB/s
    data_sent......................: 764 kB 38 kB/s




running (20.3s), 00/20 VUs, 1210 complete and 0 interrupted iterations
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
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_http2_mux.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 50s max duration (incl. graceful stop):
              * http2_multiplex: 20 looping VUs for 20s (gracefulStop: 30s)


running (01.0s), 20/20 VUs, 80 complete and 0 interrupted iterations
http2_multiplex   [   5% ] 20 VUs  01.0s/20s

running (02.0s), 20/20 VUs, 180 complete and 0 interrupted iterations
http2_multiplex   [  10% ] 20 VUs  02.0s/20s

running (03.0s), 20/20 VUs, 276 complete and 0 interrupted iterations
http2_multiplex   [  15% ] 20 VUs  03.0s/20s

running (04.0s), 20/20 VUs, 363 complete and 0 interrupted iterations
http2_multiplex   [  20% ] 20 VUs  04.0s/20s

running (05.0s), 20/20 VUs, 460 complete and 0 interrupted iterations
http2_multiplex   [  25% ] 20 VUs  05.0s/20s

running (06.0s), 20/20 VUs, 553 complete and 0 interrupted iterations
http2_multiplex   [  30% ] 20 VUs  06.0s/20s

running (07.0s), 20/20 VUs, 643 complete and 0 interrupted iterations
http2_multiplex   [  35% ] 20 VUs  07.0s/20s

running (08.0s), 20/20 VUs, 736 complete and 0 interrupted iterations
http2_multiplex   [  40% ] 20 VUs  08.0s/20s

running (09.0s), 20/20 VUs, 827 complete and 0 interrupted iterations
http2_multiplex   [  45% ] 20 VUs  09.0s/20s

running (10.0s), 20/20 VUs, 922 complete and 0 interrupted iterations
http2_multiplex   [  50% ] 20 VUs  10.0s/20s

running (11.0s), 20/20 VUs, 1021 complete and 0 interrupted iterations
http2_multiplex   [  55% ] 20 VUs  11.0s/20s

running (12.0s), 20/20 VUs, 1114 complete and 0 interrupted iterations
http2_multiplex   [  60% ] 20 VUs  12.0s/20s

running (13.0s), 20/20 VUs, 1208 complete and 0 interrupted iterations
http2_multiplex   [  65% ] 20 VUs  13.0s/20s

running (14.0s), 20/20 VUs, 1307 complete and 0 interrupted iterations
http2_multiplex   [  70% ] 20 VUs  14.0s/20s

running (15.0s), 20/20 VUs, 1407 complete and 0 interrupted iterations
http2_multiplex   [  75% ] 20 VUs  15.0s/20s

running (16.0s), 20/20 VUs, 1505 complete and 0 interrupted iterations
http2_multiplex   [  80% ] 20 VUs  16.0s/20s

running (17.0s), 20/20 VUs, 1594 complete and 0 interrupted iterations
http2_multiplex   [  85% ] 20 VUs  17.0s/20s

running (18.0s), 20/20 VUs, 1683 complete and 0 interrupted iterations
http2_multiplex   [  90% ] 20 VUs  18.0s/20s

running (19.0s), 20/20 VUs, 1776 complete and 0 interrupted iterations
http2_multiplex   [  95% ] 20 VUs  19.0s/20s

running (20.0s), 20/20 VUs, 1867 complete and 0 interrupted iterations
http2_multiplex   [ 100% ] 20 VUs  20.0s/20s


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.95' rate=100.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 15096   746.518294/s
    checks_succeeded...: 100.00% 15096 out of 15096
    checks_failed......: 0.00%   0 out of 15096

    Γ£ô chat1 stream 200
    Γ£ô chat2 stream 200
    Γ£ô chat3 stream 200
    Γ£ô chat1 backend ack present
    Γ£ô chat2 backend ack present
    Γ£ô chat3 backend ack present
    Γ£ô all 3 streams succeeded
    Γ£ô HTTP/2: fast streams not blocked by slow stream

    HTTP
    http_req_duration..............: avg=77.11ms  min=1.71ms  med=12.08ms  max=304.15ms p(90)=211.99ms p(95)=215.62ms
      { expected_response:true }...: avg=77.11ms  min=1.71ms  med=12.08ms  max=304.15ms p(90)=211.99ms p(95)=215.62ms
    http_req_failed................: 0.00%  0 out of 5661
    http_reqs......................: 5661   279.94436/s

    EXECUTION
    iteration_duration.............: avg=212.84ms min=203.2ms med=210.21ms max=305.23ms p(90)=223.3ms  p(95)=231.46ms
    iterations.....................: 1887   93.314787/s
    vus............................: 20     min=20        max=20
    vus_max........................: 20     min=20        max=20

    NETWORK
    data_received..................: 2.1 MB 103 kB/s
    data_sent......................: 1.1 MB 53 kB/s




running (20.2s), 00/20 VUs, 1887 complete and 0 interrupted iterations
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
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_sticky_rules.js
        output: -

     scenarios: (100.00%) 1 scenario, 30 max VUs, 10m30s max duration (incl. graceful stop):
              * sticky_sessions: 10 iterations for each of 30 VUs (maxDuration: 10m0s, gracefulStop: 30s)



  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.95' rate=100.00%

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.00%


  Γûê TOTAL RESULTS 

    checks_total.......: 1800    2239.66841/s
    checks_succeeded...: 100.00% 1800 out of 1800
    checks_failed......: 0.00%   0 out of 1800

    Γ£ô chat 200
    Γ£ô payload 200
    Γ£ô chat backend ack present
    Γ£ô payload backend ack present
    Γ£ô sticky /chat in backend-1 or backend-2
    Γ£ô sticky /payload in backend-3

    HTTP
    http_req_duration..............: avg=34.29ms min=4.03ms med=28.68ms max=129.69ms p(90)=69.7ms   p(95)=83.12ms 
      { expected_response:true }...: avg=34.29ms min=4.03ms med=28.68ms max=129.69ms p(90)=69.7ms   p(95)=83.12ms 
    http_req_failed................: 0.00%  0 out of 600
    http_reqs......................: 600    746.556137/s

    EXECUTION
    iteration_duration.............: avg=71.09ms min=18.2ms med=62.91ms max=167.99ms p(90)=113.39ms p(95)=136.81ms
    iterations.....................: 300    373.278068/s

    NETWORK
    data_received..................: 239 kB 298 kB/s
    data_sent......................: 124 kB 154 kB/s




running (00m00.8s), 00/30 VUs, 300 complete and 0 interrupted iterations
sticky_sessions Γ£ô [ 100% ] 30 VUs  00m00.8s/10m0s  300/300 iters, 10 per VU
```

---
### test_realistic_load -- FAILED (exit 99)

```

         /\      Grafana   /ΓÇ╛ΓÇ╛/  
    /\  /  \     |\  __   /  /   
   /  \/    \    | |/ /  /   ΓÇ╛ΓÇ╛\ 
  /          \   |   (  |  (ΓÇ╛)  |
 / __________ \  |_|\_\  \_____/ 


     execution: local
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_realistic_load.js
        output: -

     scenarios: (100.00%) 1 scenario, 500 max VUs, 50s max duration (incl. graceful stop):
              * realistic_load: 1000.00 iterations/s for 20s (maxVUs: 200-500, gracefulStop: 30s)


running (00.9s), 368/368 VUs, 204 complete and 0 interrupted iterations
realistic_load   [   4% ] 368/368 VUs  00.9s/20s  1000.00 iters/s
time="2026-04-19T02:04:45+05:30" level=warning msg="Insufficient VUs, reached 500 active VUs and cannot initialize more" executor=constant-arrival-rate scenario=realistic_load

running (01.9s), 500/500 VUs, 517 complete and 0 interrupted iterations
realistic_load   [   9% ] 500/500 VUs  01.9s/20s  1000.00 iters/s

running (02.9s), 497/500 VUs, 929 complete and 0 interrupted iterations
realistic_load   [  14% ] 497/500 VUs  02.9s/20s  1000.00 iters/s

running (03.9s), 475/500 VUs, 1554 complete and 0 interrupted iterations
realistic_load   [  19% ] 475/500 VUs  03.9s/20s  1000.00 iters/s

running (04.9s), 499/500 VUs, 2151 complete and 0 interrupted iterations
realistic_load   [  25% ] 499/500 VUs  04.9s/20s  1000.00 iters/s
time="2026-04-19T02:04:49+05:30" level=warning msg="Request Failed" error="unexpected EOF"
time="2026-04-19T02:04:49+05:30" level=error msg="SyntaxError: unexpected EOF\n\tat parse (native)\n\tat default (file:///C:/Users/srika/Documents/distributed-systems/clients/k6/test_realistic_load.js:80:39(171))\n" executor=constant-arrival-rate scenario=realistic_load source=stacktrace

running (05.9s), 500/500 VUs, 2611 complete and 0 interrupted iterations
realistic_load   [  30% ] 500/500 VUs  05.9s/20s  1000.00 iters/s

running (06.9s), 500/500 VUs, 2960 complete and 0 interrupted iterations
realistic_load   [  34% ] 500/500 VUs  06.9s/20s  1000.00 iters/s

running (07.9s), 499/500 VUs, 3540 complete and 0 interrupted iterations
realistic_load   [  39% ] 499/500 VUs  07.9s/20s  1000.00 iters/s

running (08.9s), 486/500 VUs, 4096 complete and 0 interrupted iterations
realistic_load   [  44% ] 486/500 VUs  08.9s/20s  1000.00 iters/s

running (09.9s), 500/500 VUs, 4765 complete and 0 interrupted iterations
realistic_load   [  49% ] 500/500 VUs  09.9s/20s  1000.00 iters/s

running (10.9s), 495/500 VUs, 5423 complete and 0 interrupted iterations
realistic_load   [  54% ] 495/500 VUs  10.9s/20s  1000.00 iters/s

running (11.9s), 499/500 VUs, 6105 complete and 0 interrupted iterations
realistic_load   [  59% ] 499/500 VUs  11.9s/20s  1000.00 iters/s

running (12.9s), 485/500 VUs, 6794 complete and 0 interrupted iterations
realistic_load   [  64% ] 485/500 VUs  12.9s/20s  1000.00 iters/s

running (13.9s), 495/500 VUs, 7483 complete and 0 interrupted iterations
realistic_load   [  69% ] 494/500 VUs  13.9s/20s  1000.00 iters/s

running (14.9s), 499/500 VUs, 7787 complete and 0 interrupted iterations
realistic_load   [  75% ] 499/500 VUs  14.9s/20s  1000.00 iters/s

running (15.9s), 491/500 VUs, 7978 complete and 0 interrupted iterations
realistic_load   [  80% ] 491/500 VUs  15.9s/20s  1000.00 iters/s

running (16.9s), 499/500 VUs, 8175 complete and 0 interrupted iterations
realistic_load   [  84% ] 499/500 VUs  16.9s/20s  1000.00 iters/s

running (17.9s), 500/500 VUs, 8370 complete and 0 interrupted iterations
realistic_load   [  89% ] 500/500 VUs  17.9s/20s  1000.00 iters/s

running (18.9s), 500/500 VUs, 8569 complete and 0 interrupted iterations
realistic_load   [  94% ] 500/500 VUs  18.9s/20s  1000.00 iters/s

running (19.9s), 500/500 VUs, 8794 complete and 0 interrupted iterations
realistic_load   [  99% ] 500/500 VUs  19.9s/20s  1000.00 iters/s

running (20.9s), 288/500 VUs, 9021 complete and 0 interrupted iterations
realistic_load Γåô [ 100% ] 498/500 VUs  20s  1000.00 iters/s

running (21.9s), 051/500 VUs, 9258 complete and 0 interrupted iterations
realistic_load Γåô [ 100% ] 498/500 VUs  20s  1000.00 iters/s


  Γûê THRESHOLDS 

    checks
    Γ£ù 'rate>0.99' rate=93.19%

    http_req_duration
    Γ£ù 'p(95)<300' p(95)=3.44s

    http_req_failed
    Γ£ô 'rate<0.01' rate=0.01%


  Γûê TOTAL RESULTS 

    checks_total.......: 27925  1264.568955/s
    checks_succeeded...: 93.19% 26026 out of 27925
    checks_failed......: 6.80%  1899 out of 27925

    Γ£ù payload 200
      Γå│  99% ΓÇö Γ£ô 1898 / Γ£ù 1
    Γ£ô admin chat 200
    Γ£ô admin backend id present
    Γ£ô admin backend ack present
    Γ£ù bytes received correct
      Γå│  0% ΓÇö Γ£ô 0 / Γ£ù 1898
    Γ£ô payload backend ack present
    Γ£ô chat 200
    Γ£ô chat backend id present
    Γ£ô chat backend ack present

    HTTP
    http_req_duration..............: avg=1.08s min=7.75ms  med=300.05ms max=21.43s p(90)=1.12s p(95)=3.44s
      { expected_response:true }...: avg=1.08s min=7.75ms  med=300.05ms max=21.43s p(90)=1.12s p(95)=3.44s
    http_req_failed................: 0.01%  1 out of 9309
    http_reqs......................: 9309   421.553175/s

    EXECUTION
    dropped_iterations.............: 10688  484.000465/s
    iteration_duration.............: avg=1.09s min=17.87ms med=312.68ms max=21.45s p(90)=1.14s p(95)=3.47s
    iterations.....................: 9309   421.553175/s
    vus............................: 51     min=51        max=500
    vus_max........................: 500    min=369       max=500

    NETWORK
    data_received..................: 128 MB 5.8 MB/s
    data_sent......................: 126 MB 5.7 MB/s




running (22.1s), 000/500 VUs, 9309 complete and 0 interrupted iterations
realistic_load Γ£ô [ 100% ] 000/500 VUs  20s  1000.00 iters/s
time="2026-04-19T02:05:07+05:30" level=error msg="thresholds on metrics 'checks, http_req_duration' have been crossed"
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
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_health_failover.js
        output: -

     scenarios: (100.00%) 3 scenarios, 201 max VUs, 10m35s max duration (incl. graceful stop):
              * before_failure: 200.00 iterations/s for 5s (maxVUs: 50-100, gracefulStop: 30s)
              * trigger_failure: 1 iterations shared among 1 VUs (maxDuration: 10m0s, startTime: 5s, gracefulStop: 30s)
              * after_failure: 200.00 iterations/s for 10s (maxVUs: 50-100, startTime: 6s, gracefulStop: 30s)


running (00m01.0s), 001/101 VUs, 193 complete and 0 interrupted iterations
before_failure    [  19% ] 001/050 VUs  1.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      4.0s    
after_failure   ΓÇó [   0% ] waiting      5.0s    

running (00m02.0s), 003/101 VUs, 391 complete and 0 interrupted iterations
before_failure    [  39% ] 003/050 VUs  2.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      3.0s    
after_failure   ΓÇó [   0% ] waiting      4.0s    

running (00m03.0s), 001/101 VUs, 593 complete and 0 interrupted iterations
before_failure    [  59% ] 001/050 VUs  3.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      2.0s    
after_failure   ΓÇó [   0% ] waiting      3.0s    

running (00m04.0s), 001/101 VUs, 793 complete and 0 interrupted iterations
before_failure    [  79% ] 001/050 VUs  4.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      1.0s    
after_failure   ΓÇó [   0% ] waiting      2.0s    

running (00m05.0s), 071/122 VUs, 899 complete and 0 interrupted iterations
before_failure    [  99% ] 071/071 VUs  5.0s/5s  200.00 iters/s
trigger_failure ΓÇó [   0% ] waiting      0.0s    
after_failure   ΓÇó [   0% ] waiting      1.0s    

running (00m06.0s), 073/124 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m01.0s/10m0s  0/1 shared iters
after_failure   ΓÇó [   0% ] waiting      0.0s           
time="2026-04-19T02:05:14+05:30" level=warning msg="Insufficient VUs, reached 100 active VUs and cannot initialize more" executor=constant-arrival-rate scenario=after_failure

running (00m07.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m02.0s/10m0s  0/1 shared iters
after_failure     [  10% ] 100/100 VUs  01.0s/10s       200.00 iters/s

running (00m08.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m03.0s/10m0s  0/1 shared iters
after_failure     [  20% ] 100/100 VUs  02.0s/10s       200.00 iters/s

running (00m09.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m04.0s/10m0s  0/1 shared iters
after_failure     [  30% ] 100/100 VUs  03.0s/10s       200.00 iters/s

running (00m10.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m05.0s/10m0s  0/1 shared iters
after_failure     [  40% ] 100/100 VUs  04.0s/10s       200.00 iters/s

running (00m11.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m06.0s/10m0s  0/1 shared iters
after_failure     [  50% ] 100/100 VUs  05.0s/10s       200.00 iters/s

running (00m12.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m07.0s/10m0s  0/1 shared iters
after_failure     [  60% ] 100/100 VUs  06.0s/10s       200.00 iters/s

running (00m13.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m08.0s/10m0s  0/1 shared iters
after_failure     [  70% ] 100/100 VUs  07.0s/10s       200.00 iters/s

running (00m14.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m09.0s/10m0s  0/1 shared iters
after_failure     [  80% ] 100/100 VUs  08.0s/10s       200.00 iters/s

running (00m15.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m10.0s/10m0s  0/1 shared iters
after_failure     [  90% ] 100/100 VUs  09.0s/10s       200.00 iters/s

running (00m16.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m11.0s/10m0s  0/1 shared iters
after_failure     [ 100% ] 100/100 VUs  10.0s/10s       200.00 iters/s

running (00m17.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m12.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m18.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m13.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m19.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m14.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m20.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m15.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m21.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m16.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m22.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m17.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m23.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m18.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m24.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m19.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m25.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m20.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m26.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m21.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m27.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m22.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m28.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m23.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m29.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m24.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m30.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m25.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m31.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m26.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m32.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m27.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m33.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m28.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m34.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m29.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m35.0s), 173/174 VUs, 899 complete and 0 interrupted iterations
before_failure  Γåô [ 100% ] 072/072 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m30.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m36.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m31.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m37.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m32.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m38.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m33.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m39.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m34.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m40.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m35.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m41.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m36.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m42.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m37.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m43.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m38.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m44.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m39.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m45.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m40.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m46.0s), 101/174 VUs, 899 complete and 72 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m41.0s/10m0s  0/1 shared iters
after_failure   Γåô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m47.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m42.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m48.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m43.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m49.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m44.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m50.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m45.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m51.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m46.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m52.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m47.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m53.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m48.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m54.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m49.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m55.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m50.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m56.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m51.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m57.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m52.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m58.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m53.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (00m59.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m54.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m00.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m55.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m01.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m56.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m02.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m57.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m03.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m58.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m04.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        00m59.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s

running (01m05.0s), 001/174 VUs, 899 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure   [   0% ] 1 VUs        01m00.0s/10m0s  0/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s
time="2026-04-19T02:06:12+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"


  Γûê THRESHOLDS 

    checks
    Γ£ô 'rate>0.99' rate=99.91%

    http_req_failed
    Γ£ô 'rate<0.05' rate=0.11%


  Γûê TOTAL RESULTS 

    checks_total.......: 3600   55.366881/s
    checks_succeeded...: 99.91% 3597 out of 3600
    checks_failed......: 0.08%  3 out of 3600

    Γ£ù request succeeds
      Γå│  99% ΓÇö Γ£ô 899 / Γ£ù 1
    Γ£ù backend reply present
      Γå│  99% ΓÇö Γ£ô 899 / Γ£ù 1
    Γ£ù backend ack present
      Γå│  99% ΓÇö Γ£ô 899 / Γ£ù 1
    Γ£ô no failed backend after failover

    HTTP
    http_req_duration..............: avg=72.16ms min=2.05ms med=4.36ms max=59.99s p(90)=9.89ms  p(95)=12.11ms
      { expected_response:true }...: avg=5.51ms  min=2.05ms med=4.36ms max=21.6ms p(90)=9.86ms  p(95)=12.09ms
    http_req_failed................: 0.11%  1 out of 900
    http_reqs......................: 900    13.84172/s

    EXECUTION
    dropped_iterations.............: 1930   29.6828/s
    iteration_duration.............: avg=72.6ms  min=2.2ms  med=4.77ms max=1m0s   p(90)=10.54ms p(95)=12.99ms
    iterations.....................: 900    13.84172/s
    vus............................: 1      min=1        max=173
    vus_max........................: 174    min=101      max=174

    NETWORK
    data_received..................: 333 kB 5.1 kB/s
    data_sent......................: 182 kB 2.8 kB/s




running (01m05.0s), 000/174 VUs, 900 complete and 172 interrupted iterations
before_failure  Γ£ô [ 100% ] 072/073 VUs  5s              200.00 iters/s
trigger_failure Γ£ô [ 100% ] 1 VUs        01m00.0s/10m0s  1/1 shared iters
after_failure   Γ£ô [ 100% ] 100/100 VUs  10s             200.00 iters/s
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
        script: C:\Users\srika\Documents\distributed-systems\clients\k6\test_sticky_failover.js
        output: -

     scenarios: (100.00%) 3 scenarios, 21 max VUs, 10m39s max duration (incl. graceful stop):
              * establish_stickiness: 5 iterations for each of 10 VUs (maxDuration: 10m0s, gracefulStop: 30s)
              * trigger_failure: 1 iterations shared among 1 VUs (maxDuration: 10m0s, startTime: 8s, gracefulStop: 30s)
              * after_failover: 5 iterations for each of 10 VUs (maxDuration: 10m0s, startTime: 9s, gracefulStop: 30s)


running (00m01.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m01.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  7.0s           
after_failover       ΓÇó [   0% ] waiting  8.0s           

running (00m02.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m02.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  6.0s           
after_failover       ΓÇó [   0% ] waiting  7.0s           

running (00m03.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m03.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  5.0s           
after_failover       ΓÇó [   0% ] waiting  6.0s           

running (00m04.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m04.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  4.0s           
after_failover       ΓÇó [   0% ] waiting  5.0s           

running (00m05.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m05.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  3.0s           
after_failover       ΓÇó [   0% ] waiting  4.0s           

running (00m06.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m06.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  2.0s           
after_failover       ΓÇó [   0% ] waiting  3.0s           

running (00m07.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m07.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  1.0s           
after_failover       ΓÇó [   0% ] waiting  2.0s           

running (00m08.0s), 10/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m08.0s/10m0s  00/50 iters, 5 per VU
trigger_failure      ΓÇó [   0% ] waiting  0.0s           
after_failover       ΓÇó [   0% ] waiting  1.0s           

running (00m09.0s), 11/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs   00m09.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs    00m01.0s/10m0s  0/1 shared iters
after_failover       ΓÇó [   0% ] waiting  0.0s           

running (00m10.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m10.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m02.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m01.0s/10m0s  00/50 iters, 5 per VU

running (00m11.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m11.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m03.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m02.0s/10m0s  00/50 iters, 5 per VU

running (00m12.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m12.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m04.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m03.0s/10m0s  00/50 iters, 5 per VU

running (00m13.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m13.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m05.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m04.0s/10m0s  00/50 iters, 5 per VU

running (00m14.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m14.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m06.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m05.0s/10m0s  00/50 iters, 5 per VU

running (00m15.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m15.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m07.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m06.0s/10m0s  00/50 iters, 5 per VU

running (00m16.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m16.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m08.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m07.0s/10m0s  00/50 iters, 5 per VU

running (00m17.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m17.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m09.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m08.0s/10m0s  00/50 iters, 5 per VU

running (00m18.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m18.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m10.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m09.0s/10m0s  00/50 iters, 5 per VU

running (00m19.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m19.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m11.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m10.0s/10m0s  00/50 iters, 5 per VU

running (00m20.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m20.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m12.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m11.0s/10m0s  00/50 iters, 5 per VU

running (00m21.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m21.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m13.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m12.0s/10m0s  00/50 iters, 5 per VU

running (00m22.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m22.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m14.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m13.0s/10m0s  00/50 iters, 5 per VU

running (00m23.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m23.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m15.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m14.0s/10m0s  00/50 iters, 5 per VU

running (00m24.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m24.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m16.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m15.0s/10m0s  00/50 iters, 5 per VU

running (00m25.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m25.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m17.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m16.0s/10m0s  00/50 iters, 5 per VU

running (00m26.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m26.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m18.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m17.0s/10m0s  00/50 iters, 5 per VU

running (00m27.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m27.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m19.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m18.0s/10m0s  00/50 iters, 5 per VU

running (00m28.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m28.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m20.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m19.0s/10m0s  00/50 iters, 5 per VU

running (00m29.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m29.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m21.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m20.0s/10m0s  00/50 iters, 5 per VU

running (00m30.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m30.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m22.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m21.0s/10m0s  00/50 iters, 5 per VU

running (00m31.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m31.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m23.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m22.0s/10m0s  00/50 iters, 5 per VU

running (00m32.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m32.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m24.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m23.0s/10m0s  00/50 iters, 5 per VU

running (00m33.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m33.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m25.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m24.0s/10m0s  00/50 iters, 5 per VU

running (00m34.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m34.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m26.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m25.0s/10m0s  00/50 iters, 5 per VU

running (00m35.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m35.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m27.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m26.0s/10m0s  00/50 iters, 5 per VU

running (00m36.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m36.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m28.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m27.0s/10m0s  00/50 iters, 5 per VU

running (00m37.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m37.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m29.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m28.0s/10m0s  00/50 iters, 5 per VU

running (00m38.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m38.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m30.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m29.0s/10m0s  00/50 iters, 5 per VU

running (00m39.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m39.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m31.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m30.0s/10m0s  00/50 iters, 5 per VU

running (00m40.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m40.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m32.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m31.0s/10m0s  00/50 iters, 5 per VU

running (00m41.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m41.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m33.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m32.0s/10m0s  00/50 iters, 5 per VU

running (00m42.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m42.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m34.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m33.0s/10m0s  00/50 iters, 5 per VU

running (00m43.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m43.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m35.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m34.0s/10m0s  00/50 iters, 5 per VU

running (00m44.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m44.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m36.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m35.0s/10m0s  00/50 iters, 5 per VU

running (00m45.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m45.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m37.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m36.0s/10m0s  00/50 iters, 5 per VU

running (00m46.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m46.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m38.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m37.0s/10m0s  00/50 iters, 5 per VU

running (00m47.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m47.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m39.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m38.0s/10m0s  00/50 iters, 5 per VU

running (00m48.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m48.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m40.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m39.0s/10m0s  00/50 iters, 5 per VU

running (00m49.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m49.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m41.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m40.0s/10m0s  00/50 iters, 5 per VU

running (00m50.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m50.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m42.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m41.0s/10m0s  00/50 iters, 5 per VU

running (00m51.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m51.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m43.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m42.0s/10m0s  00/50 iters, 5 per VU

running (00m52.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m52.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m44.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m43.0s/10m0s  00/50 iters, 5 per VU

running (00m53.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m53.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m45.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m44.0s/10m0s  00/50 iters, 5 per VU

running (00m54.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m54.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m46.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m45.0s/10m0s  00/50 iters, 5 per VU

running (00m55.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m55.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m47.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m46.0s/10m0s  00/50 iters, 5 per VU

running (00m56.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m56.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m48.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m47.0s/10m0s  00/50 iters, 5 per VU

running (00m57.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m57.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m49.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m48.0s/10m0s  00/50 iters, 5 per VU

running (00m58.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m58.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m50.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m49.0s/10m0s  00/50 iters, 5 per VU

running (00m59.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  00m59.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m51.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m50.0s/10m0s  00/50 iters, 5 per VU

running (01m00.0s), 21/21 VUs, 0 complete and 0 interrupted iterations
establish_stickiness   [   0% ] 10 VUs  01m00.0s/10m0s  00/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m52.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m51.0s/10m0s  00/50 iters, 5 per VU
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:13+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"

running (01m01.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m01.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m53.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m52.0s/10m0s  00/50 iters, 5 per VU

running (01m02.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m02.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m54.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m53.0s/10m0s  00/50 iters, 5 per VU

running (01m03.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m03.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m55.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m54.0s/10m0s  00/50 iters, 5 per VU

running (01m04.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m04.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m56.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m55.0s/10m0s  00/50 iters, 5 per VU

running (01m05.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m05.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m57.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m56.0s/10m0s  00/50 iters, 5 per VU

running (01m06.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m06.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m58.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m57.0s/10m0s  00/50 iters, 5 per VU

running (01m07.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m07.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   00m59.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m58.0s/10m0s  00/50 iters, 5 per VU

running (01m08.0s), 21/21 VUs, 10 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m08.0s/10m0s  10/50 iters, 5 per VU
trigger_failure        [   0% ] 1 VUs   01m00.0s/10m0s  0/1 shared iters
after_failover         [   0% ] 10 VUs  00m59.0s/10m0s  00/50 iters, 5 per VU
time="2026-04-19T02:07:21+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"

running (01m09.0s), 20/21 VUs, 11 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m09.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [   0% ] 10 VUs  01m00.0s/10m0s  00/50 iters, 5 per VU
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"
time="2026-04-19T02:07:22+05:30" level=warning msg="Request Failed" error="Get \"http://localhost:8081/chat\": request timeout"

running (01m10.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m10.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m01.0s/10m0s  10/50 iters, 5 per VU

running (01m11.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m11.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m02.0s/10m0s  10/50 iters, 5 per VU

running (01m12.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m12.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m03.0s/10m0s  10/50 iters, 5 per VU

running (01m13.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m13.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m04.0s/10m0s  10/50 iters, 5 per VU

running (01m14.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m14.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m05.0s/10m0s  10/50 iters, 5 per VU

running (01m15.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m15.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m06.0s/10m0s  10/50 iters, 5 per VU

running (01m16.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m16.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m07.0s/10m0s  10/50 iters, 5 per VU

running (01m17.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m17.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m08.0s/10m0s  10/50 iters, 5 per VU

running (01m18.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m18.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m09.0s/10m0s  10/50 iters, 5 per VU

running (01m19.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m19.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m10.0s/10m0s  10/50 iters, 5 per VU

running (01m20.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m20.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m11.0s/10m0s  10/50 iters, 5 per VU

running (01m21.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m21.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m12.0s/10m0s  10/50 iters, 5 per VU

running (01m22.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m22.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m13.0s/10m0s  10/50 iters, 5 per VU

running (01m23.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m23.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m14.0s/10m0s  10/50 iters, 5 per VU

running (01m24.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m24.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m15.0s/10m0s  10/50 iters, 5 per VU

running (01m25.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m25.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m16.0s/10m0s  10/50 iters, 5 per VU

running (01m26.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m26.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m17.0s/10m0s  10/50 iters, 5 per VU

running (01m27.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m27.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m18.0s/10m0s  10/50 iters, 5 per VU

running (01m28.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m28.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m19.0s/10m0s  10/50 iters, 5 per VU

running (01m29.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m29.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m20.0s/10m0s  10/50 iters, 5 per VU

running (01m30.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m30.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m21.0s/10m0s  10/50 iters, 5 per VU

running (01m31.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m31.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m22.0s/10m0s  10/50 iters, 5 per VU

running (01m32.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m32.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m23.0s/10m0s  10/50 iters, 5 per VU

running (01m33.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m33.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m24.0s/10m0s  10/50 iters, 5 per VU

running (01m34.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m34.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m25.0s/10m0s  10/50 iters, 5 per VU

running (01m35.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m35.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m26.0s/10m0s  10/50 iters, 5 per VU

running (01m36.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m36.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m27.0s/10m0s  10/50 iters, 5 per VU

running (01m37.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m37.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m28.0s/10m0s  10/50 iters, 5 per VU

running (01m38.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m38.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m29.0s/10m0s  10/50 iters, 5 per VU

running (01m39.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m39.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m30.0s/10m0s  10/50 iters, 5 per VU

running (01m40.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m40.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m31.0s/10m0s  10/50 iters, 5 per VU

running (01m41.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m41.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m32.0s/10m0s  10/50 iters, 5 per VU

running (01m42.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m42.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m33.0s/10m0s  10/50 iters, 5 per VU

running (01m43.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m43.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m34.0s/10m0s  10/50 iters, 5 per VU

running (01m44.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m44.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m35.0s/10m0s  10/50 iters, 5 per VU

running (01m45.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m45.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m36.0s/10m0s  10/50 iters, 5 per VU

running (01m46.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m46.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m37.0s/10m0s  10/50 iters, 5 per VU

running (01m47.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m47.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m38.0s/10m0s  10/50 iters, 5 per VU

running (01m48.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m48.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m39.0s/10m0s  10/50 iters, 5 per VU

running (01m49.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m49.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m40.0s/10m0s  10/50 iters, 5 per VU

running (01m50.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m50.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m41.0s/10m0s  10/50 iters, 5 per VU

running (01m51.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m51.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m42.0s/10m0s  10/50 iters, 5 per VU

running (01m52.0s), 20/21 VUs, 21 complete and 0 interrupted iterations
establish_stickiness   [  20% ] 10 VUs  01m52.0s/10m0s  10/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover         [  20% ] 10 VUs  01m43.0s/10m0s  10/50 iters, 5 per VU

running (01m53.0s), 00/21 VUs, 101 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs  01m53.0s/10m0s  50/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover       Γ£ô [ 100% ] 10 VUs  01m44.0s/10m0s  50/50 iters, 5 per VU


  Γûê THRESHOLDS 

    checks
    Γ£ù 'rate>0.95' rate=79.45%

    http_req_failed
    Γ£ù 'rate<0.02' rate=20.79%


  Γûê TOTAL RESULTS 

    checks_total.......: 404    3.575928/s
    checks_succeeded...: 79.45% 321 out of 404
    checks_failed......: 20.54% 83 out of 404

    Γ£ù session request succeeds
      Γå│  79% ΓÇö Γ£ô 80 / Γ£ù 21
    Γ£ù backend id present
      Γå│  79% ΓÇö Γ£ô 80 / Γ£ù 21
    Γ£ù backend ack present
      Γå│  79% ΓÇö Γ£ô 80 / Γ£ù 21
    Γ£ù remapped off backend-1 after failover
      Γå│  80% ΓÇö Γ£ô 81 / Γ£ù 20

    HTTP
    http_req_duration..............: avg=21.97s min=4.37ms   med=12.03ms  max=1m0s   p(90)=59.99s p(95)=59.99s
      { expected_response:true }...: avg=11.98s min=4.37ms   med=11ms     max=52.43s p(90)=52.42s p(95)=52.42s
    http_req_failed................: 20.79% 21 out of 101
    http_reqs......................: 101    0.893982/s

    EXECUTION
    iteration_duration.............: avg=22.07s min=105.47ms med=115.89ms max=1m0s   p(90)=1m0s   p(95)=1m0s  
    iterations.....................: 101    0.893982/s
    vus............................: 20     min=10        max=21
    vus_max........................: 21     min=21        max=21

    NETWORK
    data_received..................: 30 kB  262 B/s
    data_sent......................: 19 kB  168 B/s




running (01m53.0s), 00/21 VUs, 101 complete and 0 interrupted iterations
establish_stickiness Γ£ô [ 100% ] 10 VUs  01m53.0s/10m0s  50/50 iters, 5 per VU
trigger_failure      Γ£ô [ 100% ] 1 VUs   01m00.1s/10m0s  1/1 shared iters
after_failover       Γ£ô [ 100% ] 10 VUs  01m44.0s/10m0s  50/50 iters, 5 per VU
time="2026-04-19T02:08:06+05:30" level=error msg="thresholds on metrics 'checks, http_req_failed' have been crossed"
```

