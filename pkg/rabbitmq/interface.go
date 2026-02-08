package rabbitmq

import "context"

type RabbitMQService interface {
	Publish(ctx context.Context, message any) error
	Consume(ctx context.Context, handler func([]byte) error) error
	Close()
}
