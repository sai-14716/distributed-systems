// Test 3: HTTP/2 Multiplexing
// Three concurrent streams (chat1, chat2, chat3) on one connection.
// chat1 is slow (200ms). With HTTP/2, chat2 and chat3 complete first — that's correct behavior.
// k6 uses HTTP/2 automatically when the server supports it (HTTPS or h2c).
// Since our backend is plain HTTP/1.1, this test uses concurrent VU goroutines to
// simulate independent stream behavior and verify out-of-order completion is fine.

import http from 'k6/http';
import { check, group } from 'k6';

export const options = {
    scenarios: {
        http2_multiplex: {
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

const LB_BASE_URLS = (__ENV.LB_BASE_URLS || 'http://localhost:8081,http://localhost:8082,http://localhost:8083')
    .split(',')
    .map((v) => v.trim())
    .filter((v) => v.length > 0);

export default function () {
    const base = LB_BASE_URLS[(__VU - 1) % LB_BASE_URLS.length];
    const sessionID = `h2-session-vu-${__VU}`;

    const commonHeaders = {
        'X-Session-ID': sessionID,
        'X-Client-VU': `${__VU}`,
        'X-Role': 'user',
        'X-Forwarded-For': `172.16.${__VU % 255}.1`,
    };

    // Fire all three concurrently using http.batch()
    // http.batch() in k6 runs requests concurrently on the same VU, simulating streams
    const responses = http.batch([
        ['GET', `${base}/chat`, null, { headers: { ...commonHeaders, 'X-Chat-ID': 'chat1', 'X-Slow': 'true' } }],
        ['GET', `${base}/chat`, null, { headers: { ...commonHeaders, 'X-Chat-ID': 'chat2' } }],
        ['GET', `${base}/chat`, null, { headers: { ...commonHeaders, 'X-Chat-ID': 'chat3' } }],
    ]);

    const [r1, r2, r3] = responses;
    const b1 = JSON.parse(r1.body || '{}');
    const b2 = JSON.parse(r2.body || '{}');
    const b3 = JSON.parse(r3.body || '{}');

    group('HTTP/2 stream independence', () => {
        check(r1, { 'chat1 stream 200': (r) => r.status === 200 });
        check(r2, { 'chat2 stream 200': (r) => r.status === 200 });
        check(r3, { 'chat3 stream 200': (r) => r.status === 200 });
        check(b1, { 'chat1 backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0 });
        check(b2, { 'chat2 backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0 });
        check(b3, { 'chat3 backend ack present': (b) => typeof b.ack === 'string' && b.ack.length > 0 });

        // All three must succeed regardless of ordering
        check(null, {
            'all 3 streams succeeded': () =>
                r1.status === 200 && r2.status === 200 && r3.status === 200,
        });

        // HTTP/2 key property: chat2 and chat3 should finish even if chat1 is slow
        // (They share no HoL dependency — each stream is independent)
        check(null, {
            'HTTP/2: fast streams not blocked by slow stream': () =>
                r2.timings.duration < 180 && r3.timings.duration < 180,
        });
    });
}
