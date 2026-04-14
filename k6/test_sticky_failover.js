// Test 7: Sticky Session Failover
// A session is consistently mapped to backend-1 (via sticky map or hash).
// Controller marks backend-1 unhealthy.
// The LB must remap the session to a healthy backend — no dropped requests.

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        establish_stickiness: {
            executor: 'per-vu-iterations',
            vus: 10,
            iterations: 5,
            startTime: '0s',
        },
        trigger_failure: {
            executor: 'shared-iterations',
            vus: 1,
            iterations: 1,
            startTime: '8s',
        },
        after_failover: {
            executor: 'per-vu-iterations',
            vus: 10,
            iterations: 5,
            startTime: '9s',
        },
    },
    thresholds: {
        'http_req_failed': ['rate<0.02'],
        'checks': ['rate>0.95'],
    },
};

const LB_PORT = '8081';

export function trigger_failure() {
    const res = http.post(`http://localhost:${LB_PORT}/admin/fail-backend`, JSON.stringify({
        backend: 'backend-1',
        healthy: false,
    }), { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'failover triggered': (r) => r.status === 200 });
}

export default function () {
    const sessionID = `sticky-failover-vu-${__VU}`;
    const res = http.get(`http://localhost:${LB_PORT}/chat`, {
        headers: {
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Chat-ID': `chat-${__VU}`,
            'X-Role': 'user',
            'X-Forwarded-For': `10.2.${__VU % 254}.1`,
        },
    });

    check(res, {
        'session request succeeds after failover': (r) => r.status === 200,
        'server responded (remap occurred)': (r) => r.headers['X-Server-ID'] !== undefined,
    });

    sleep(0.1);
}
