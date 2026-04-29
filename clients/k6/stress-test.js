import http from 'k6/http';
import { check } from 'k6';

export const options = {
  noConnectionReuse: true,
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 1,          // 1000 requests per second
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 200,
      maxVUs: 1000,
    },
  },
};

export default function () {
  // Hit the single active LB in Phase 1
  // const port = '8081';

  // Service discovery + connection routing is handled by the local forward proxy
  // (service_discovery_client.py). Run k6 with HTTP_PROXY=http://127.0.0.1:6700.
  const res = http.get(`http://api.service.com:19093/chat`, {
    headers: {
      'X-Client-VU': __VU.toString(),
      'X-Chat-ID': `chat-${__VU}-${__ITER}`,
    },
  });

  let body = {};
  try {
    body = JSON.parse(res.body || '{}');
  } catch (_e) {
    body = {};
  }
  const serverID = res.headers['X-Server-Id'] || res.headers['X-Server-ID'] || body.server || '';

  console.log(
    `[VU ${__VU} ITER ${__ITER}] status=${res.status} upstream=${res.headers['X-Discovery-Endpoint'] || 'unknown'} backend=${serverID || 'unknown'} ack="${body.ack || 'missing'}" body=${res.body}`
  );

  check(res, {
    'is status 200': (r) => r.status === 200,
    'backend id present': () => !!serverID,
    'backend ack present': () => typeof body.ack === 'string' && body.ack.length > 0,
  });
}
