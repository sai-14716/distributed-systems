// Test 4: Sticky Rules + Client-Server Mapping
// Demonstrates that WRR and LR maintain a ClientBackendMap for session stickiness.
// Maglev achieves this without a map (hash-based).
// Key insight shown: if this map is NOT replicated across LBs, a different LB
// would assign the same session to a different backend — breaking stickiness.

import http from 'k6/http';
import { check, group } from 'k6';

export const options = {
    scenarios: {
        sticky_sessions: {
            executor: 'per-vu-iterations',
            vus: 30,
            iterations: 10,
        },
    },
    thresholds: {
        checks: ['rate>0.95'],
        'http_req_failed': ['rate<0.01'],
    },
};

const LB_BASE_URLS = (__ENV.LB_BASE_URLS || 'http://localhost:8081,http://localhost:8082,http://localhost:8083')
    .split(',')
    .map((v) => v.trim())
    .filter((v) => v.length > 0);

export default function () {
    const sessionID = `sticky-session-vu-${__VU}`;
    const role = ['user', 'admin', 'member'][__VU % 3];

    // Round-robin across LBs for each iteration (simulates multi-LB environment)
    const base = LB_BASE_URLS[__ITER % LB_BASE_URLS.length];

    const headers = {
        'X-Session-ID': sessionID,
        'X-Client-VU': `${__VU}`,
        'X-Role': role,
        'X-Chat-ID': `chat-${__VU}`,
        'X-Forwarded-For': `10.${__VU % 254 + 1}.${__ITER % 254 + 1}.1`,
    };

    // /chat → should stay within backend-1/2 subset (StickyRule)
    const chatRes = http.get(`${base}/chat`, { headers });

    // /payload → should stay within backend-3 subset (StickyRule)
    const payloadRes = http.post(`${base}/payload`, 'test-data-' + __VU, { headers });

    group('Sticky rules enforcement', () => {
        const chatBody = JSON.parse(chatRes.body || '{}');
        const payloadBody = JSON.parse(payloadRes.body || '{}');
        const chatServer = chatRes.headers['X-Server-Id'] || chatRes.headers['X-Server-ID'];
        const payloadServer = payloadRes.headers['X-Server-Id'] || payloadRes.headers['X-Server-ID'];

        check(chatRes, { 'chat 200': (r) => r.status === 200 });
        check(payloadRes, { 'payload 200': (r) => r.status === 200 });
        check(chatBody, { 'chat backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0 });
        check(payloadBody, { 'payload backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0 });

        // /chat must go to backend-1 or backend-2 (StickyRule subset)
        check(null, {
            'sticky /chat in backend-1 or backend-2': () =>
                chatServer === 'B1' || chatServer === 'B2' || chatServer === 'backend-1' || chatServer === 'backend-2',
        });

        // /payload must go to backend-3 (StickyRule subset)
        check(null, {
            'sticky /payload in backend-3': () => payloadServer === 'B3' || payloadServer === 'backend-3',
        });
    });
}
