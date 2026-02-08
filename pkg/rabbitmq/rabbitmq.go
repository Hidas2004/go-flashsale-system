package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	URL           string
	Exchange      string
	Queue         string
	RoutingKey    string
	PrefetchCount int
}

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     RabbitMQConfig
}

func NewRabbitMQClient(cfg RabbitMQConfig) (*RabbitMQClient, error) {
	// 1. Kết nối RabbitMQ
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 2. Tạo Channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// -------------------------------------------------------
	// CONFIG DEAD LETTER QUEUE (DLQ)
	// -------------------------------------------------------
	dlxName := cfg.Exchange + ".dlx"
	dlqName := cfg.Queue + ".dlq"

	// 3. Khai báo DLX
	err = ch.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	// 4. Khai báo DLQ
	_, err = ch.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	// 5. Bind DLQ
	err = ch.QueueBind(dlqName, "dead_letter", dlxName, false, nil)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------
	// CONFIG MAIN QUEUE
	// -------------------------------------------------------

	// 6. Khai báo Main Exchange
	err = ch.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	// 7. Khai báo Main Queue VỚI CẤU HÌNH DLQ
	args := amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": "dead_letter",
	}

	q, err := ch.QueueDeclare(
		cfg.Queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,  // args DLQ
	)
	if err != nil {
		return nil, err
	}

	// 8. Bind Main Queue
	err = ch.QueueBind(q.Name, cfg.RoutingKey, cfg.Exchange, false, nil)
	if err != nil {
		return nil, err
	}

	// 9. QoS
	err = ch.Qos(cfg.PrefetchCount, 0, false)
	if err != nil {
		return nil, err
	}

	log.Println("✅ RabbitMQ setup complete with DLQ enable")

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

// Publish - Gửi message
func (r *RabbitMQClient) Publish(ctx context.Context, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(
		ctx,
		r.cfg.Exchange,
		r.cfg.RoutingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

// Consume - Lắng nghe message từ Queue
func (r *RabbitMQClient) Consume(ctx context.Context, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		r.cfg.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume message: %w", err)
	}

	go func() {
		defer log.Println("Consumer stopped")
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				// Gọi handler xử lý
				if err := handler(msg.Body); err != nil {
					log.Printf("Error handling message: %v. Moving to DLQ...", err)
					msg.Nack(false, false)
				} else {
					msg.Ack(false)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Close connection
func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
