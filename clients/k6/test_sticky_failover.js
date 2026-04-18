// Test 7: Sticky Session Failover
// A session is consistently mapped to backend-1 (via sticky map or hash).
// Controller marks backend-1 unhealthy.
// The LB must remap the session to a healthy backend — no dropped requests.

import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

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

const LB_BASE_URLS = (__ENV.LB_BASE_URLS || 'http://localhost:8081,http://localhost:8082,http://localhost:8083')
    .split(',')
    .map((v) => v.trim())
    .filter((v) => v.length > 0);
const LB_BASE_URL = __ENV.LB_BASE_URL || LB_BASE_URLS[0];

export function trigger_failure() {
    const res = http.post(`${LB_BASE_URL}/admin/fail-backend`, JSON.stringify({
        backend: 'backend-1',
        healthy: false,
    }), { headers: { 'Content-Type': 'application/json' } });
    check(res, { 'failover triggered': (r) => r.status === 200 });
}

export default function () {
    const sessionID = `sticky-failover-vu-${__VU}`;
    const res = http.get(`${LB_BASE_URL}/chat`, {
        headers: {
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Chat-ID': `chat-${__VU}`,
            'X-Role': 'user',
            'X-Forwarded-For': `10.2.${__VU % 254}.1`,
        },
    });

    const body = JSON.parse(res.body || '{}');
    const serverID = res.headers['X-Server-Id'] || res.headers['X-Server-ID'] || body.server || '';
    const afterFailover = exec.scenario.name === 'after_failover';

    check(res, {
        'session request succeeds': (r) => r.status === 200,
        'backend id present': () => !!serverID,
        'backend ack present': () => typeof body.ack === 'string' && body.ack.length > 0,
        'remapped off backend-1 after failover': () =>
            !afterFailover || (serverID !== 'B1' && serverID !== 'backend-1'),
    });

    sleep(0.1);
}
