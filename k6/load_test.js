import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 100 }, // Ramp up to 100 users
    { duration: '1m', target: 100 },  // Stay at 100
    { duration: '10s', target: 0 },   // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate should be less than 1%
  },
};

const BASE_URL = 'http://localhost:8080/api/v1';

export default function () {
  // 1. Register
  const email = `user_${randomString(10)}@example.com`;
  const password = 'password123';
  
  let res = http.post(`${BASE_URL}/auth/register`, JSON.stringify({
    email: email,
    password: password,
    full_name: 'Test Load User',
    role: 'user'
  }), { headers: { 'Content-Type': 'application/json' } });
  
  // Accept 201 Created or 200 OK
  check(res, { 'registered': (r) => r.status === 201 || r.status === 200 });

  // 2. Login
  res = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: email,
    password: password
  }), { headers: { 'Content-Type': 'application/json' } });
  
  check(res, { 'logged in': (r) => r.status === 200 });
  
  // Extract token
  let token = "";
  try {
      token = res.json('data.token');
  } catch (e) {
      console.log("Login failed or invalid response:", res.body);
  }

  if (!token) return;

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };

  // 3. Get Products
  res = http.get(`${BASE_URL}/products/flash-sale`, { headers: authHeaders });
  check(res, { 'got products': (r) => r.status === 200 });
  
  let productID = "";
  try {
      const products = res.json('data.products');
      if (products && products.length > 0) {
          // Pick a random product or the first one
          productID = products[0].id;
      }
  } catch(e) {
      console.log("Failed to parse products:", res.body);
  }

  // 4. Create Order (if product found)
  if (productID) {
      const orderPayload = JSON.stringify({
          product_id: productID,
          quantity: 1
      });
      
      res = http.post(`${BASE_URL}/orders`, orderPayload, { headers: authHeaders });
      check(res, { 'order placed': (r) => r.status === 202 || r.status === 200 });
  }

  sleep(1);
}
