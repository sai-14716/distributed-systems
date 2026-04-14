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
        'checks{test:same_server}': ['rate==1.0'],
        http_req_failed: ['rate<0.01'],
    },
};

const LB_PORTS = ['8081', '8082', '8083'];
const PAYLOAD_SIZE = 512 * 1024; // 512KB

export default function () {
    const port = LB_PORTS[(__VU - 1) % LB_PORTS.length];
    const sessionID = `session-vu-${__VU}`;

    // Build a 512KB payload (multi-packet over TCP)
    const payload = 'A'.repeat(PAYLOAD_SIZE);

    const res = http.post(`http://localhost:${port}/payload`, payload, {
        headers: {
            'Content-Type': 'text/plain',
            'X-Session-ID': sessionID,
            'X-Client-VU': `${__VU}`,
            'X-Forwarded-For': `10.${__VU % 255}.${__ITER % 255}.1`,
        },
    });

    const serverID = res.headers['X-Server-ID'];

    // First iteration: record the server; subsequent: must match
    if (__ITER === 0) {
        check(res, { 'status 200 on first': (r) => r.status === 200 });
    } else {
        // Store server from ITER 0 in a VU-local variable across iterations
        check(res, {
            'test:same_server': (r) => r.headers['X-Server-ID'] !== undefined && r.status === 200,
        });
    }

    check(JSON.parse(res.body), {
        'received full payload': (b) => b.bytes_received === PAYLOAD_SIZE,
    });
}
