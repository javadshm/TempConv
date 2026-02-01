/**
 * k6 load test: many virtual users (frontends) calling the TempConv backend.
 * Run: k6 run k6-load.js
 * Target: backend base URL (default http://localhost:8080 for local; set BASE_URL for K8s/Ingress).
 *
 * Example for GKE (after Ingress has an IP):
 *   BASE_URL=http://<INGRESS_IP> k6 run k6-load.js
 *
 * Options:
 *   k6 run --vus 50 --duration 60s k6-load.js   (50 users for 60s)
 *   k6 run --vus 100 --duration 120s --rps 200 k6-load.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 20 },   // ramp up to 20 users
    { duration: '1m', target: 20 },    // stay at 20
    { duration: '30s', target: 50 },   // ramp to 50
    { duration: '1m', target: 50 },    // stay at 50
    { duration: '30s', target: 0 },    // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests under 500ms
    http_req_failed: ['rate<0.01'],    // error rate under 1%
  },
};

export default function () {
  // Health check
  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, { 'health status 200': (r) => r.status === 200 });

  // Celsius -> Fahrenheit
  const c2fRes = http.post(
    `${BASE_URL}/api/convert`,
    JSON.stringify({ value: 25, from_unit: 'CELSIUS', to_unit: 'FAHRENHEIT' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(c2fRes, { 'c2f status 200': (r) => r.status === 200 });
  check(c2fRes, { 'c2f value 77': (r) => JSON.parse(r.body).value === 77 });

  // Fahrenheit -> Celsius
  const f2cRes = http.post(
    `${BASE_URL}/api/convert`,
    JSON.stringify({ value: 32, from_unit: 'FAHRENHEIT', to_unit: 'CELSIUS' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(f2cRes, { 'f2c status 200': (r) => r.status === 200 });
  check(f2cRes, { 'f2c value 0': (r) => JSON.parse(r.body).value === 0 });

  sleep(0.5 + Math.random());
}
