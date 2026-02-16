package redis

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestInventoryCache_DeductStock_RaceCondition(t *testing.T) {
	// 1. Setup MiniRedis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 2. Initialize Cache
	cache := NewInventoryCache(client)

	// 3. Setup Data
	productID := uuid.New()
	initialStock := 10
	totalRequests := 50

	// Set stock directly in redis (or via helper if available)
	// We need to match the key format used in implementation
	stockKey := fmt.Sprintf("product:%s:stock", productID.String())
	_ = client.Set(context.Background(), stockKey, initialStock, 0).Err()

	// 4. Run Concurrent Requests
	var successCount int32
	var failCount int32
	var wg sync.WaitGroup

	// Use a channel to synchronize start for maximum contention
	start := make(chan struct{})

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			<-start // Wait for signal

			// Unique user for each request to avoid rate limits/logic per user checks if any
			// But DeductStock signature is: DeductStock(ctx, productID, userID, quantity, limit)
			// Assuming limit 1 per user, so we need different userIDs
			uid := fmt.Sprintf("user-%d", userID)

			err := cache.DeductStock(context.Background(), productID, uid, 1, 1)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}(i)
	}

	// Release all goroutines at once
	close(start)
	wg.Wait()

	// 5. Assertions
	t.Logf("Success: %d, Failed: %d", successCount, failCount)

	assert.Equal(t, int32(initialStock), successCount, "Should successfully deduct exactly initial stock amount")
	assert.Equal(t, int32(totalRequests-initialStock), failCount, "Remaining requests should fail")

	// Verify final stock is 0
	finalStockStr, _ := client.Get(context.Background(), stockKey).Result()
	assert.Equal(t, "0", finalStockStr, "Final stock should be 0")
}
