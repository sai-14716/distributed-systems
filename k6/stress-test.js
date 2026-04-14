import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 1000,          // 1000 requests per second
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 200,
      maxVUs: 1000,
    },
  },
};

export default function () {
  // Hit the single active LB in Phase 1
  const port = '8081';

  const res = http.get(`http://localhost:${port}/`, {
    headers: { 'X-Client-VU': __VU.toString() },
  });

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}