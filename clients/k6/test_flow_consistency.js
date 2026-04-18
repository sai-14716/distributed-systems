// Test 1: Flow Consistency
// Sends a 512KB POST to /payload. Same X-Session-ID per VU.
// With Maglev: same session always hashes to same backend.
// Check: X-Server-ID in response is consistent across all iterations per VU.

import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        flow_consistency: {
            executor: 'per-vu-iterations',
            vus: 50,
            iterations: 10,
        },
    },
    thresholds: {
        checks: ['rate>0.99'],
        http_req_failed: ['rate<0.01'],
    },
};

const LB_BASE_URLS = (__ENV.LB_BASE_URLS || 'http://localhost:8081,http://localhost:8082,http://localhost:8083')
    .split(',')
    .map((v) => v.trim())
    .filter((v) => v.length > 0);
const PAYLOAD_SIZE = 512 * 1024; // 512KB
let expectedServer = '';

export default function () {
    const base = LB_BASE_URLS[(__VU - 1) % LB_BASE_URLS.length];
    const sessionID = `session-vu-${__VU}`;

    // Build a 512KB payload (multi-packet over TCP)
    const payload = 'A'.repeat(PAYLOAD_SIZE);

    const res = http.post(`${base}/payload`, payload, {
        headers: {
            'Content-Type': 'text/plain',
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Forwarded-For': `10.${__VU % 255}.${__ITER % 255}.1`,
        },
    });

    const body = JSON.parse(res.body || '{}');
    const serverID = res.headers['X-Server-Id'] || res.headers['X-Server-ID'] || body.server || '';

    // First iteration: record the selected backend; subsequent iterations must match.
    if (__ITER === 0) {
        check(res, {
            'status 200 on first': (r) => r.status === 200,
            'first response has backend id': () => !!serverID,
        });
        expectedServer = serverID || '';
    } else {
        check(res, {
            'same backend for session': (r) => r.status === 200 && !!serverID && serverID === expectedServer,
        });
    }

    check(body, {
        'received full payload': (b) => b.bytes_received === PAYLOAD_SIZE,
        'backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0,
    });
}
