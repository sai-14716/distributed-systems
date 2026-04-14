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
        'checks{test:sticky}': ['rate>0.95'],
        'http_req_failed': ['rate<0.01'],
    },
};

const LB_PORTS = ['8081', '8082', '8083'];

export default function () {
    const sessionID = `sticky-session-vu-${__VU}`;
    const role = ['user', 'admin', 'member'][__VU % 3];

    // Round-robin across LBs for each iteration (simulates multi-LB environment)
    const port = LB_PORTS[__ITER % LB_PORTS.length];
    const base = `http://localhost:${port}`;

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
        const chatServer = chatRes.headers['X-Server-ID'];
        const payloadServer = payloadRes.headers['X-Server-ID'];

        check(chatRes, { 'chat 200': (r) => r.status === 200 });
        check(payloadRes, { 'payload 200': (r) => r.status === 200 });

        // /chat must go to backend-1 or backend-2 (StickyRule subset)
        check(null, {
            'test:sticky /chat in backend-1 or backend-2': () =>
                chatServer === 'B1' || chatServer === 'B2',
        });

        // /payload must go to backend-3 (StickyRule subset)
        check(null, {
            'test:sticky /payload in backend-3': () => payloadServer === 'B3',
        });
    });
}
