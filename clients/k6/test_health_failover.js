// Test 6: Health Failover
// The Controller marks backend-1 unhealthy via IsHealthy=false (shared memory).
// Traffic must fully reroute to remaining backends.
// Uses the /admin/fail-backend endpoint on the LB to trigger the simulation.

import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

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
        checks: ['rate>0.99'],
    },
};

const LB_BASE_URLS = (__ENV.LB_BASE_URLS || 'http://localhost:8081,http://localhost:8082,http://localhost:8083')
    .split(',')
    .map((v) => v.trim())
    .filter((v) => v.length > 0);
const LB_BASE_URL = __ENV.LB_BASE_URL || LB_BASE_URLS[0];

export function trigger_failure() {
    // Trigger backend-1 failure via LB admin endpoint
    const res = http.post(`${LB_BASE_URL}/admin/fail-backend`, JSON.stringify({
        backend: 'backend-1',
        healthy: false,
    }), { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'failure triggered': (r) => r.status === 200 });
}

export default function () {
    const sessionID = `fail-test-vu-${__VU}`;
    const res = http.get(`${LB_BASE_URL}/chat`, {
        headers: {
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Chat-ID': `chat-${__VU}`,
            'X-Forwarded-For': `10.1.${__VU % 254}.1`,
        },
    });

    const body = JSON.parse(res.body || '{}');
    const serverID = res.headers['X-Server-Id'] || res.headers['X-Server-ID'] || body.server || '';

    check(res, { 'request succeeds': (r) => r.status === 200 });
    check(null, {
        'backend reply present': () => !!serverID,
        'backend ack present': () => typeof body.ack === 'string' && body.ack.length > 0,
    });

    // Only assert backend-1 exclusion in the post-failure scenario.
    const afterFailure = exec.scenario.name === 'after_failure';
    check(null, {
        'no failed backend after failover': () => !afterFailure || (serverID !== 'B1' && serverID !== 'backend-1'),
    });
}
