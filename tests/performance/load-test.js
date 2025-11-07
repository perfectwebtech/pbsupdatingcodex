// ============================================================================
// IPTV Platform - Load Testing Script (k6)
// ============================================================================
// Usage: k6 run tests/performance/load-test.js
// ============================================================================

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const apiDuration = new Trend('api_duration');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp-up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Ramp-up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp-up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '5m', target: 0 },    // Ramp-down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'],  // Error rate < 1%
    errors: ['rate<0.05'],           // Custom error rate < 5%
  },
};

// Configuration
const BASE_URL = __ENV.API_URL || 'http://localhost:8080';
const API_BASE = BASE_URL + '/api/v1';

// Test data
const testUsers = [
  { email: 'user1@example.com', password: 'Admin@123' },
  { email: 'user2@example.com', password: 'Admin@123' },
  { email: 'user3@example.com', password: 'Admin@123' },
];

// Helper function to get random user
function getRandomUser() {
  return testUsers[Math.floor(Math.random() * testUsers.length)];
}

// Setup function - runs once
export function setup() {
  console.log('Starting load test...');
  console.log('Target: ' + BASE_URL);
  return { startTime: new Date().toISOString() };
}

// Main test function
export default function (data) {
  let token;

  // 1. Login Test
  group('Authentication', function () {
    const user = getRandomUser();
    const loginRes = http.post(API_BASE + '/auth/login', JSON.stringify({
      email: user.email,
      password: user.password,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });

    requestCount.add(1);
    loginDuration.add(loginRes.timings.duration);

    const loginSuccess = check(loginRes, {
      'login status is 200': (r) => r.status === 200,
      'login returns token': (r) => r.json('data.access_token') !== undefined,
    });

    if (!loginSuccess) {
      errorRate.add(1);
      return; // Skip rest of test if login fails
    }

    token = loginRes.json('data.access_token');
  });

  sleep(1);

  // 2. Get User Profile
  group('User Profile', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
      'Content-Type': 'application/json',
    };

    const profileRes = http.get(API_BASE + '/auth/me', { headers });
    requestCount.add(1);
    apiDuration.add(profileRes.timings.duration);

    const profileSuccess = check(profileRes, {
      'profile status is 200': (r) => r.status === 200,
      'profile returns user data': (r) => r.json('data.email') !== undefined,
    });

    if (!profileSuccess) errorRate.add(1);
  });

  sleep(1);

  // 3. Get Streams List
  group('Get Streams', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
    };

    const streamsRes = http.get(API_BASE + '/streams?limit=20', { headers });
    requestCount.add(1);
    apiDuration.add(streamsRes.timings.duration);

    const streamsSuccess = check(streamsRes, {
      'streams status is 200': (r) => r.status === 200,
      'streams returns data': (r) => r.json('data') !== undefined,
      'response time < 500ms': (r) => r.timings.duration < 500,
    });

    if (!streamsSuccess) errorRate.add(1);
  });

  sleep(2);

  // 4. Get Categories
  group('Get Categories', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
    };

    const categoriesRes = http.get(API_BASE + '/categories', { headers });
    requestCount.add(1);
    apiDuration.add(categoriesRes.timings.duration);

    check(categoriesRes, {
      'categories status is 200': (r) => r.status === 200,
    });
  });

  sleep(1);

  // 5. Get VOD Movies
  group('Get VOD', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
    };

    const vodRes = http.get(API_BASE + '/mobile/vod?limit=20', { headers });
    requestCount.add(1);
    apiDuration.add(vodRes.timings.duration);

    check(vodRes, {
      'vod status is 200 or 404': (r) => [200, 404].includes(r.status),
    });
  });

  sleep(2);

  // 6. Get Recommendations (if available)
  group('Get Recommendations', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
    };

    const recRes = http.get(API_BASE + '/recommendations/personalized?limit=10', { headers });
    requestCount.add(1);

    check(recRes, {
      'recommendations returned': (r) => [200, 404, 500].includes(r.status),
    });
  });

  sleep(1);

  // 7. Get Analytics (Admin only - may fail for regular users)
  group('Get Analytics', function () {
    const headers = {
      'Authorization': 'Bearer ' + token,
    };

    http.get(API_BASE + '/analytics/dashboard', { headers });
    requestCount.add(1);
  });

  sleep(3);
}

// Teardown function - runs once
export function teardown(data) {
  console.log('Load test completed!');
  console.log('Started at: ' + data.startTime);
  console.log('Ended at: ' + new Date().toISOString());
}
