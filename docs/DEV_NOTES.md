# 📘 GO-FLASHSALE-SYSTEM: DOCUMENTATION

Tài liệu này giải thích kiến trúc dự án theo chuẩn **Clean Architecture** và cách ánh xạ tư duy từ mô hình MVC cũ.

---

## 1. BẢNG ĐỐI CHIẾU TƯ DUY (MENTAL MODEL)
Dành cho người quen với MVC/N-Tier (Laravel, Node.js Express, Spring Boot):

| Layer (Clean Arch) | Tương đương (MVC/Cũ) | Nhiệm vụ chính |
| :--- | :--- | :--- |
| **Delivery** | **Route + Controller** | Cửa ngõ đón khách. Nhận request HTTP, validate JSON, chuyển việc cho Usecase. |
| **Usecase** | **Service** | Bộ não. Chứa logic nghiệp vụ (tính toán, check điều kiện, flow chính). |
| **Repository** | **DAO / Model Access** | Thủ kho. Chỉ biết `Get`, `Save`, `Update` vào DB/Redis. Không chứa logic business. |
| **Domain** | **Model / Entity** | Định nghĩa dữ liệu (Struct User, Order...). |

---

## 2. GIẢI THÍCH CHI TIẾT CẤU TRÚC THƯ MỤC

### 📂 `cmd/` (Cửa ra vào)
Nơi chứa các hàm `main` để khởi chạy ứng dụng.
- `api/main.go`: Chạy Web Server (REST API).
- `consumer/main.go`: Chạy Worker xử lý ngầm (RabbitMQ Consumer).

### 📂 `internal/` (Khu vực nội bộ - Private)
Code trong này là cốt lõi, không được phép import từ bên ngoài vào.

#### 1. `domain/` (Nguyên liệu)
Chứa các định nghĩa dữ liệu thuần túy, không phụ thuộc vào ai.
- `models/`: Cấu trúc bảng Database (User, Product...).
- `dtos/`: Cấu trúc dữ liệu JSON gửi nhận qua API.

#### 2. `delivery/` (Lễ tân & Bồi bàn)
Lớp giao tiếp với thế giới bên ngoài.
- `http/`: Xử lý REST API (Gin Framework).
    - `router.go`: Định nghĩa đường dẫn (Routes).
    - `handler`: Các hàm xử lý request (Controllers).
- `worker/`: Xử lý tin nhắn từ Queue (RabbitMQ Consumer).

#### 3. `usecase/` (Đầu bếp trưởng - Logic)
Chứa toàn bộ Business Logic.
- Ví dụ: `CreateOrder` sẽ kiểm tra tồn kho, tính tiền, gọi Repository lưu DB, bắn event Queue.

#### 4. `repository/` (Phụ bếp - Data Access)
Lớp tương tác trực tiếp với nơi lưu trữ dữ liệu.
- `postgres/`: Các câu SQL truy xuất PostgreSQL.
- `redis/`: Các lệnh thao tác Redis (bao gồm cả Lua Script).

### 📂 `pkg/` (Dụng cụ dùng chung)
Chứa các thư viện tiện ích (Library) có thể dùng lại ở nhiều dự án khác.
- `database/`: Code mở kết nối Postgres/Redis.
- `logger/`: Cấu hình log.
- `utils/`: Hàm hỗ trợ nhỏ lẻ.

### 📂 `config/` (Sổ quy định)
- Chứa file `config.yaml` và code load biến môi trường.

---

## 3. LUỒNG ĐI CỦA 1 REQUEST (WORKFLOW)
Ví dụ: **Khách mua hàng (POST /orders)**

1. **Client** gọi API.
2. **Delivery (Handler)** nhận request, check JSON hợp lệ không -> Gọi Usecase.
3. **Usecase** kiểm tra logic (còn hàng không, giờ mở bán chưa) -> Gọi Repository.
4. **Repository** chạy Lua Script trừ kho Redis -> Trả kết quả về Usecase.
5. **Usecase** thấy thành công -> Bắn tin vào RabbitMQ -> Trả kết quả OK cho Delivery.
6. **Delivery** trả JSON `200 OK` cho khách.