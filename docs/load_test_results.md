# Kết Quả K6 Load & Stress Test

## Tổng Quan

Tài liệu này tóm tắt kết quả của các bài kiểm tra tải (load test) và sức chịu đựng (stress test) được thực hiện trên `go-flashsale-system` API.

### Môi Trường

- **API URL**: `http://localhost:8081/api/v1`
- **Database**: PostgreSQL (Local Docker)
- **Cache**: Redis (Local Docker)
- **Message Queue**: RabbitMQ (Local Docker)

## 1. Load Test (100 VUs)

- **Script**: `k6/load_test.js`
- **Cấu hình**:
    - Thời gian: 1 phút 40 giây
    - VUs (Virtual Users): 100 (Ramp up trong 30s)
- **Kết Quả**:
    - **Trạng thái**: KHÔNG ĐẠT (Vượt ngưỡng cho phép)
    - **Độ trễ p(95)**: ~883ms (Mục tiêu: <500ms)
    - **Tỷ lệ lỗi**: Thấp (<1%)
- **Nhận xét**:
    - Hệ thống xử lý được 100 người dùng đồng thời nhưng độ trễ hơi cao so với mục tiêu 500ms. Cần tối ưu hóa (ví dụ: đánh index database, chiến lược caching).

## 2. Stress Test (Lên tới 10000 VUs)

- **Script**: `k6/stress_test.js`
- **Cấu hình**:
    - Thời gian: 5 phút
    - Các giai đoạn (Stages): 100 -> 1000 -> 5000 -> 10000 VUs
- **Kết Quả**:
    - **Trạng thái**: KHÔNG ĐẠT (Vượt ngưỡng nghiêm trọng)
    - **Độ trễ p(95)**: ~44.53s (Mục tiêu: <2000ms)
    - **Tỷ lệ lỗi**: Cao (Vượt ngưỡng)
- **Quan sát**:
    - Ở mức tải cao (gần 5000-10000 VUs), hệ thống bị suy giảm hiệu năng nghiêm trọng.
    - Độ trễ tăng vọt lên hơn 44 giây.
    - Điều này cho thấy các điểm nghẽn (bottlenecks) có thể nằm ở:
        - Kết nối Database (kiểm tra `max_connections` và cấu hình pool).
        - Tốc độ xử lý của Worker (RabbitMQ consumer bị lag/chậm).
        - Giới hạn CPU/Memory của môi trường Docker local.

## Khuyến Nghị

1.  **Tối ưu Database**: Phân tích các câu truy vấn chậm (slow queries) và đảm bảo đánh index đầy đủ cho các bảng `products`, `orders`, và `inventory`.
2.  **Scale Workers**: Tăng số lượng instance của worker để xử lý hàng đợi RabbitMQ nhanh hơn.
3.  **Caching**: Tận dụng triệt để Redis để cache thông tin sản phẩm và số lượng tồn kho (inventory), giảm tải cho DB.
4.  **Rate Limiting**: Điều chỉnh rate limiter để từ chối bớt request khi vượt quá khả năng chịu đựng, tránh để hệ thống bị treo.
5.  **Hạ tầng**: Với mức 10000 VUs, môi trường Docker local là không đủ. Cần môi trường test phân tán và hạ tầng server mạnh mẽ hơn (ví dụ: Kubernetes cluster).

## Kết Quả Sau Tối Ưu (Phase 2)

Sau khi thực hiện các tối ưu hóa (Index DB, Caching Product, Scale Worker x5, Rate Limiting), kết quả test lại với **100 VUs** như sau:

- **Độ trễ p(95)**: ~635ms (Cải thiện so với 883ms trước đó).
- **Độ ổn định**: Hệ thống hoạt động ổn định hơn, worker xử lý nhanh hơn nhờ scale out.
- **Rate Limiting**: Đã hoạt động để bảo vệ hệ thống khỏi overload (Stress test cho thấy hệ thống không bị crash ngay lập tức).
- **Caching**: Redis keys (`product:id`, `product:id:stock`) được ghi nhận đầy đủ, giảm tải cho DB.

> [!NOTE]
> Mặc dù độ trễ vẫn chưa đạt mức lý tưởng (<500ms), nhưng đã có sự cải thiện đáng kể logic xử lý. Việc chạy trên local Docker (resource limited) là nguyên nhân chính khiến độ trễ chưa giảm sâu hơn.
