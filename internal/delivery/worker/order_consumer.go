package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
)

type OrderConsumer struct {
	client       *rabbitmq.RabbitMQClient
	orderUseCase usecase.OrderUseCase
}

func NewOrderConsumer(client *rabbitmq.RabbitMQClient, uc usecase.OrderUseCase) *OrderConsumer {
	return &OrderConsumer{client: client, orderUseCase: uc}
}

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
		processCtx := context.Background()
		return c.orderUseCase.ProcessOrder(processCtx, &msg)
	}
	return c.client.Consume(ctx, handler)
}
