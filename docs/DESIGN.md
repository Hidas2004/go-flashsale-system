# 🏗️ SYSTEM DESIGN DOCUMENT

## 1. API Endpoints (Dự kiến)

| Method | Endpoint | Mô tả | Input |
| :--- | :--- | :--- | :--- |
| **POST** | `/api/v1/auth/login` | Đăng nhập lấy Token | `{email, password}` |
| **GET** | `/api/v1/products/flash-sale` | Lấy danh sách SP đang giảm giá | None |
| **POST** | `/api/v1/orders` | **(Quan trọng)** Mua hàng | `{product_id, quantity}` |
| **GET** | `/api/v1/orders/:id` | Kiểm tra trạng thái đơn | None |

---

## 2. Message Queue Format (RabbitMQ Payload)
Đây là cấu trúc JSON mà API sẽ đẩy xuống Queue cho Worker:

```json
{
  "order_id": "UUID-v4",
  "user_id": "UUID-v4",
  "product_id": "UUID-v4",
  "quantity": 1,
  "created_at": "2024-01-30T10:00:00Z"
}


3. Bottlenecks & Solutions (Vấn đề & Giải pháp)
🔴 Vấn đề 1: Race Condition (Tranh chấp dữ liệu)
Mô tả: Khi 10.000 người cùng bấm mua, Database không kịp cập nhật số lượng tồn kho (Stock), dẫn đến bán quá số lượng (Overselling).

Giải pháp: Sử dụng Redis Lua Script. Redis xử lý đơn luồng (single-threaded), đảm bảo lệnh kiểm tra và trừ kho diễn ra tuần tự, không ai chen ngang được.

🔴 Vấn đề 2: Database Overload (Sập Database)
Mô tả: Database (PostgreSQL) chỉ chịu được khoảng 500-1000 kết nối đồng thời. Nếu 10.000 người chọc thẳng vào DB cùng lúc, hệ thống sẽ sập.

Giải pháp: Sử dụng RabbitMQ (Message Queue).

API nhận đơn cực nhanh rồi đẩy vào hàng đợi.

Worker lấy đơn từ hàng đợi ra xử lý từ từ (tùy chỉnh tốc độ Worker).

Giúp "san phẳng" đỉnh tải (Load Levelling).