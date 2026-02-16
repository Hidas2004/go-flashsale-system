# ⚡ Go Flash Sale System (High Concurrency)

> **Project Status:** Active Development 🚀
>
> Một hệ thống Backend thương mại điện tử được thiết kế để chịu tải cao (High Concurrency), tập trung vào xử lý bài toán **Flash Sale** (bán hàng chớp nhoáng) nơi hàng triệu request có thể đổ về cùng một lúc.

## 🎯 Giới thiệu (Overview)

Dự án này là một PoC (Proof of Concept) cho hệ thống xử lý đơn hàng quy mô lớn, giải quyết các thách thức kỹ thuật cốt lõi trong Backend:

- **Chống bán quá số lượng kho (Overselling):** Sử dụng Redis + Lua Script để đảm bảo tính nguyên tử (Atomicity) khi trừ kho.
- **Xử lý bất đồng bộ (Asynchronous Processing):** Dùng RabbitMQ để cắt đỉnh tải (Traffic Peak Clipping), giúp API phản hồi tức thì mà không bị nghẽn DB.
- **Kiến trúc phân tán:** Tách biệt API Server (nhận request) và Worker (xử lý đơn hàng) để dễ dàng scale.
- **Data Consistency:** Cơ chế Retry và Dead Letter Queue (DLQ) để đảm bảo không mất đơn hàng.

## 🛠️ Tech Stack

Dự án được xây dựng dựa trên Clean Architecture:

- **Core:** Golang (Gin Framework).
- **Database:** PostgreSQL (Lưu trữ bền vững).
- **Cache & Locking:** Redis (Cache hàng tồn kho, Distributed Lock).
- **Message Queue:** RabbitMQ (Xử lý đơn hàng bất đồng bộ).
- **Config:** Viper (Quản lý cấu hình đa môi trường).
- **Infrastructure:** Docker, Docker Compose.

---

## ⚙️ Cài đặt & Chạy v (Installation)

### Prerequisites

- Docker & Docker Compose
- Go 1.22+

### Bước 1: Khởi động Infrastructure

Chạy Database, Redis, RabbitMQ bằng Docker:

```bash
docker-compose up -d
```

### Bước 2: Khởi động API Server

Mở terminal #1:

```bash
go run cmd/api/main.go
# Chờ thấy log: 🚀 Server starting on port 8081
```

### Bước 3: Khởi động Worker

Mở terminal #2:

```bash
go run cmd/worker/main.go
# Chờ thấy log: ✅ Worker started
```

---

## 🧪 Hướng dẫn Test (API Testing Guide)

Dưới đây là kịch bản test chi tiết (đã verify) sử dụng **Postman**.

**Cấu hình Environment trong Postman:**

- `base_url`: `http://localhost:8081`

### 🔵 GIAI ĐOẠN 1: SETUP ADMIN (Admin Flow)

_Yêu cầu: Role "admin" để quản lý sản phẩm._

**2. Đăng ký Admin**

- **URL:** `{{base_url}}/api/v1/auth/register`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "email": "admin_vip@example.com",
    "password": "adminpassword123",
    "full_name": "Super Admin",
    "role": "admin"
  }
  ```
- **Kỳ vọng:** `201 Created`

**3. Đăng nhập Admin**

- **URL:** `{{base_url}}/api/v1/auth/login`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "email": "admin_vip@example.com",
    "password": "adminpassword123"
  }
  ```
- **Hành động:** Copy `token` -> Lưu vào biến `{{admin_token}}`.

**4. Tạo Sản Phẩm Flash Sale**

- **Method:** `POST`
- **URL:** `{{base_url}}/api/v1/admin/products`
- **Header:** `Authorization: Bearer {{admin_token}}`
- **Body:**
  ```json
  {
    "name": "Samsung S24 Ultra",
    "description": "Điện thoại AI mới nhất",
    "price": 30000000,
    "image_url": "http://img.com/s24.jpg",
    "inventory": 50,
    "is_flash_sale": true,
    "flash_sale_price": 20000000
  }
  ```
- **Hành động:** Copy `id` -> Lưu vào biến `{{product_id}}`.

### 🟠 GIAI ĐOẠN 2: USER MUA HÀNG (User Flow)

**5. Đăng ký User**

- **URL:** `{{base_url}}/api/v1/auth/register`
- **Method:** `POST`
- **Body:**
  ```json
  {
    "email": "khachhang01@gmail.com",
    "password": "password123",
    "full_name": "Khach Hang Mua Le"
  }
  ```

**6. Đăng nhập User**

- **URL:** `{{base_url}}/api/v1/auth/login`
- **Method:** `POST`
- **Body:** `{"email": "khachhang01@gmail.com", "password": "password123"}`
- **Hành động:** Copy `token` -> Lưu vào biến `{{user_token}}`.

**7. Đặt hàng (Place Order)**

- **URL:** `{{base_url}}/api/v1/orders`
- **Method:** `POST`
- **Header:** `Authorization: Bearer {{user_token}}`
- **Body:**
  ```json
  {
    "product_id": "{{product_id}}",
    "quantity": 1
  }
  ```
- **Kỳ vọng:** `202 Accepted` (Order được đẩy vào Queue xử lý).

### 🟣 GIAI ĐOẠN 3: XỬ LÝ & KIỂM TRA (Verification)

**8. User xem lịch sử đơn hàng**

- **URL:** `{{base_url}}/api/v1/orders`
- **Method:** `GET`
- **Header:** `Authorization: Bearer {{user_token}}`
- **Kỳ vọng:** Thấy đơn hàng vừa đặt có trạng thái `pending` -> `confirmed` (Sau khi Worker xử lý xong).

**9. Admin cập nhật trạng thái (Giao hàng)**

- **Check List:** Admin lấy danh sách đơn `GET {{base_url}}/api/v1/admin/orders`. Lấy `id` đơn hàng -> `{{order_id}}`.
- **Update Status:**
  - **URL:** `{{base_url}}/api/v1/admin/orders/{{order_id}}/status`
  - **Method:** `PATCH`
  - **Header:** `Authorization: Bearer {{admin_token}}`
  - **Body:**
    ```json
    {
      "status": "shipping"
    }
    ```
  - **Lưu ý:** Status phải là `shipping` (đang giao), không phải `shipped`.

---
