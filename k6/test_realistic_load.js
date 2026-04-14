// Test 5: Realistic Mixed Load
// Simulates a GPT-like app: chat sessions, large payloads, mixed roles and IPs.
// 50% /chat user sessions, 30% /chat admin, 20% /payload uploads.

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        realistic_load: {
            executor: 'constant-arrival-rate',
            rate: 1000,
            timeUnit: '1s',
            duration: '20s',
            preAllocatedVUs: 200,
            maxVUs: 500,
        },
    },
    thresholds: {
        'http_req_failed': ['rate<0.01'],
        'http_req_duration': ['p(95)<300'],
        'checks': ['rate>0.99'],
    },
};

const LB_PORTS = ['8081', '8082', '8083'];
const ROLES = ['user', 'user', 'user', 'admin', 'admin', 'member'];
const CHAT_IDS = Array.from({ length: 20 }, (_, i) => `chat-${i + 1}`);

function randomIP() {
    return `${10 + (__VU % 10)}.${Math.floor(Math.random() * 254) + 1}.${Math.floor(Math.random() * 254) + 1}.${Math.floor(Math.random() * 254) + 1}`;
}

export default function () {
    const port = LB_PORTS[__VU % LB_PORTS.length];
    const base = `http://localhost:${port}`;
    const role = ROLES[__VU % ROLES.length];
    const sessionID = `session-vu-${__VU}`;
    const chatID = CHAT_IDS[__VU % CHAT_IDS.length];

    const headers = {
        'X-Session-ID': sessionID,
        'X-Client-VU': `${__VU}`,
        'X-Role': role,
        'X-Chat-ID': chatID,
        'X-Forwarded-For': randomIP(),
        'X-Real-Port': `${10000 + (__VU % 50000)}`,
    };

    const rand = Math.random();
    let res;

    if (rand < 0.50) {
        // 50% — /chat user
        res = http.get(`${base}/chat`, { headers });
        check(res, { 'chat 200': (r) => r.status === 200 });
    } else if (rand < 0.80) {
        // 30% — /chat admin
        res = http.get(`${base}/chat`, {
            headers: { ...headers, 'X-Role': 'admin' },
        });
        check(res, { 'admin chat 200': (r) => r.status === 200 });
    } else {
        // 20% — /payload (large-ish body)
        const body = 'X'.repeat(64 * 1024); // 64KB
        res = http.post(`${base}/payload`, body, { headers });
        check(res, { 'payload 200': (r) => r.status === 200 });
        check(JSON.parse(res.body || '{}'), {
            'bytes received correct': (b) => b.bytes_received === 64 * 1024,
        });
    }

    sleep(0.01);
}
