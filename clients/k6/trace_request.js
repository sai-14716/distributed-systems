import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 1,
  iterations: 1,
};

const TRACE_ID = __ENV.TRACE_ID || `trace-${Date.now()}`;

export default function () {
  const url = 'http://api.service.com:8000/chat';
  const headers = {
    'X-Trace-ID': TRACE_ID,
    'X-Session-ID': `session-${TRACE_ID}`,
    'X-Client-VU': `${__VU}`,
    'X-Chat-ID': `chat-${TRACE_ID}`,
    'X-Role': 'user',
  };

  console.log(
    `[client-app] event=request_sent trace=${TRACE_ID} method=GET url=${url} session=${headers['X-Session-ID']} chat=${headers['X-Chat-ID']} role=${headers['X-Role']}`
  );

  const res = http.get(url, { headers });
  const backend = res.headers['X-Server-Id'] || res.headers['X-Server-ID'] || 'unknown';
  let ack = 'unavailable';
  let hasAck = false;
  try {
    const body = res.json();
    ack = body && body.ack ? body.ack : 'missing';
    hasAck = typeof body?.ack === 'string' && body.ack.length > 0;
  } catch (_err) {
    ack = 'invalid_json';
    hasAck = false;
  }

  console.log(
    `[client-app] event=response_received trace=${TRACE_ID} status=${res.status} endpoint=${res.headers['X-Discovery-Endpoint'] || 'unknown'} lb_backend=${res.headers['X-LB-Backend'] || res.headers['X-Lb-Backend'] || 'unknown'} backend=${backend} ack="${ack}"`
  );

  check(res, {
    'status is 200': (r) => r.status === 200,
    'backend id present': () => backend !== 'unknown',
    'backend ack present': () => hasAck,
  });
}
