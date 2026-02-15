package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/metrics"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
)

type OrderConsumer struct {
	client       *rabbitmq.RabbitMQClient
	orderUseCase usecase.OrderUseCase
}

func NewOrderConsumer(client *rabbitmq.RabbitMQClient, uc usecase.OrderUseCase) *OrderConsumer {
	return &OrderConsumer{client: client, orderUseCase: uc}
}

// Start khởi chạy worker tiêu thụ đơn hàng từ RabbitMQ.
//
// Nhiệm vụ:
// 1. Lắng nghe tin nhắn và giải mã (Unmarshal) thành struct.
// 2. Đo lường độ trễ (Lag) và ghi nhận metrics vào Prometheus.
// 3. Đảm bảo Graceful Shutdown: Xử lý nốt đơn đang dở trước khi tắt (dùng WaitGroup).
func (c *OrderConsumer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	log.Println("🔄 Order Consumer started...")
	// Callback function xử lý message
	handler := func(body []byte) error {
		wg.Add(1)
		defer wg.Done()
		select {
		case <-ctx.Done():
			log.Println("⚠️ System shutting down, skipping message processing")
			return nil
		default:
			// Tiếp tục xử lý
		}
		var msg dtos.OrderMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		lag := time.Since(msg.CreatedAt).Seconds()

		processCtx := context.Background()
		err := c.orderUseCase.ProcessOrder(processCtx, &msg)

		status := "processed"
		if err != nil {
			status = "failed"
			log.Printf("❌ Failed to process order %s: %v", msg.OrderID, err)
		}
		metrics.WorkerLag.WithLabelValues(status).Observe(lag)
		return nil

	}
	return c.client.Consume(ctx, handler)
}
