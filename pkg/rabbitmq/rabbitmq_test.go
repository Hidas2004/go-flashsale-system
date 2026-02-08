package rabbitmq_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRabbitMQ_Integration test connecting, publishing and consuming from RabbitMQ.
// Requirement: RabbitMQ must be running on localhost:5672
func TestRabbitMQ_Integration(t *testing.T) {
	// 1. Setup Config
	cfg := rabbitmq.RabbitMQConfig{
		URL:           "amqp://guest:guest@127.0.0.1:5672/",
		Exchange:      "test_exchange",
		Queue:         "test_queue",
		RoutingKey:    "test_routing_key",
		PrefetchCount: 1,
	}

	// 2. Connect
	client, err := rabbitmq.NewRabbitMQClient(cfg)
	if err != nil {
		t.Skipf("Skipping RabbitMQ integration test: %v", err)
	}
	defer client.Close()
	assert.NotNil(t, client)

	// 3. Define Message
	type TestMessage struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	messageID := uuid.New().String()
	payload := TestMessage{
		ID:      messageID,
		Content: "Hello RabbitMQ",
	}

	// 4. Publish Message
	ctx := context.Background()
	err = client.Publish(ctx, payload)
	assert.NoError(t, err)

	// 5. Consume Message
	// Channel to signal completion
	done := make(chan bool)
	
	// Start consumer in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Consume(ctx, func(body []byte) error {
			var msg TestMessage
			if err := json.Unmarshal(body, &msg); err != nil {
				return err
			}
			
			// Verify content
			if msg.ID == messageID {
				assert.Equal(t, "Hello RabbitMQ", msg.Content)
				done <- true
				return nil
			}
			return nil
		})
		if err != nil {
			// Handle consume error (if client is closed usually)
		}
	}()

	// 6. Wait for result with timeout
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message consumption")
	}
}
