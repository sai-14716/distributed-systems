// Test 6: Health Failover
// The Controller marks backend-1 unhealthy via IsHealthy=false (shared memory).
// Traffic must fully reroute to remaining backends.
// Uses the /admin/fail-backend endpoint on the LB to trigger the simulation.

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        before_failure: {
            executor: 'constant-arrival-rate',
            rate: 200,
            timeUnit: '1s',
            duration: '5s',
            preAllocatedVUs: 50,
            maxVUs: 100,
            startTime: '0s',
        },
        trigger_failure: {
            executor: 'shared-iterations',
            vus: 1,
            iterations: 1,
            startTime: '5s',
        },
        after_failure: {
            executor: 'constant-arrival-rate',
            rate: 200,
            timeUnit: '1s',
            duration: '10s',
            preAllocatedVUs: 50,
            maxVUs: 100,
            startTime: '6s',
        },
    },
    thresholds: {
        'http_req_failed': ['rate<0.05'],
        'checks{test:no_failed_backend}': ['rate>0.99'],
    },
};

const LB_PORT = '8081';

export function trigger_failure() {
    // Trigger backend-1 failure via LB admin endpoint
    const res = http.post(`http://localhost:${LB_PORT}/admin/fail-backend`, JSON.stringify({
        backend: 'backend-1',
        healthy: false,
    }), { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'failure triggered': (r) => r.status === 200 });
}

export default function () {
    const sessionID = `fail-test-vu-${__VU}`;
    const res = http.get(`http://localhost:${LB_PORT}/chat`, {
        headers: {
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Chat-ID': `chat-${__VU}`,
            'X-Forwarded-For': `10.1.${__VU % 254}.1`,
        },
    });

    const serverID = res.headers['X-Server-ID'];

    check(res, { 'request succeeds': (r) => r.status === 200 });

    // After failure triggered (startTime 6s), backend-1 (B1) should never respond
    check(null, {
        'test:no_failed_backend': () => serverID !== 'B1' || Date.now() < 5000,
    });
}
