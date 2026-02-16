import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 1000 },
    { duration: '1m', target: 5000 }, 
    { duration: '30s', target: 10000 }, // Stress peak
    { duration: '1m', target: 10000 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // Allow higher latency under stress
    http_req_failed: ['rate<0.1'],     // Allow up to 10% errors under stress
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
    full_name: 'Test Stress User',
    role: 'user'
  }), { headers: { 'Content-Type': 'application/json' } });
  
  check(res, { 'registered': (r) => r.status === 201 || r.status === 200 });

  // 2. Login
  res = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: email,
    password: password
  }), { headers: { 'Content-Type': 'application/json' } });
  
  check(res, { 'logged in': (r) => r.status === 200 });
  
  let token = "";
  try {
      token = res.json('data.token');
  } catch (e) {
     // fail silently under stress
  }

  if (!token) return;

  const authHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  };

  // 3. Get Products
  res = http.get(`${BASE_URL}/products/flash-sale`, { headers: authHeaders });
  
  let productID = "";
  try {
      const products = res.json('data.products');
      if (products && products.length > 0) {
          productID = products[0].id; // Just pick the first one
      }
  } catch(e) {}

  // 4. Create Order
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
