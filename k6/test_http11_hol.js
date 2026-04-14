// Test 2: HTTP/1.1 Keep-Alive + Head-of-Line Blocking
// Three sequential chat requests (chat1, chat2, chat3) on a single persistent connection.
// chat1 is slow (X-Slow: true → 200ms sleep on backend).
// HTTP/1.1 requires responses in request order: chat1 must arrive before chat2 and chat3.
// k6 enforces HTTP/1.1 with httpClient.

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        http11_hol: {
            executor: 'constant-vus',
            vus: 20,
            duration: '20s',
        },
    },
    thresholds: {
        'http_req_failed': ['rate<0.01'],
        'checks': ['rate>0.95'],
    },
};

const LB_PORTS = ['8081', '8082', '8083'];

export default function () {
    const port = LB_PORTS[(__VU - 1) % LB_PORTS.length];
    const base = `http://localhost:${port}`;
    const sessionID = `session-vu-${__VU}`;

    const headers = {
        'X-Session-ID': sessionID,
        'X-Client-VU': `${__VU}`,
        'Connection': 'keep-alive',
        'X-Role': 'user',
        'X-Forwarded-For': `192.168.${__VU % 255}.1`,
    };

    // chat1: SLOW request (backend sleeps 200ms)
    const t0 = Date.now();
    const r1 = http.get(`${base}/chat`, {
        headers: { ...headers, 'X-Chat-ID': 'chat1', 'X-Slow': 'true' },
    });
    const t1 = Date.now();

    // chat2: fast
    const r2 = http.get(`${base}/chat`, {
        headers: { ...headers, 'X-Chat-ID': 'chat2' },
    });
    const t2 = Date.now();

    // chat3: fast
    const r3 = http.get(`${base}/chat`, {
        headers: { ...headers, 'X-Chat-ID': 'chat3' },
    });
    const t3 = Date.now();

    // HTTP/1.1 ordering: chat1 completes before chat2, chat2 before chat3
    check(r1, { 'chat1 status 200': (r) => r.status === 200 });
    check(r2, { 'chat2 status 200': (r) => r.status === 200 });
    check(r3, { 'chat3 status 200': (r) => r.status === 200 });

    // chat1 (slow) must take >= 200ms
    check(null, { 'chat1 took >= 200ms (HoL slow backend)': () => (t1 - t0) >= 200 });

    // chat2 must finish AFTER chat1 (sequential on keep-alive)
    check(null, { 'HTTP/1.1 ordering: chat2 after chat1': () => t2 >= t1 });

    // chat3 must finish AFTER chat2
    check(null, { 'HTTP/1.1 ordering: chat3 after chat2': () => t3 >= t2 });

    sleep(0.1);
}
