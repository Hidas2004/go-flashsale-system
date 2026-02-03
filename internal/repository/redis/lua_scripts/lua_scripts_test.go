package redis_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	redisDriver "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestDeductStock_Concurrency(t *testing.T) {
	// 1. Setup MiniRedis
	s := miniredis.RunT(t)
	defer s.Close()

	rdb := redisDriver.NewClient(&redisDriver.Options{Addr: s.Addr()})
	inventoryCache := redis.NewInventoryCache(rdb)

	ctx := context.Background()
	productID := uuid.New()
	
    // 2. Setup Data: Set kho = 100 sản phẩm
	initialStock := 100
	err := inventoryCache.SetInitialStock(ctx, productID, initialStock)
	assert.NoError(t, err)

	// 3. Act: 150 người cùng lao vào mua (Concurrency)
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	mu := sync.Mutex{} // Mutex để đếm successCount an toàn

	totalRequests := 150
	limitPerUser := 1

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := uuid.New().String()
			
            // Gọi hàm trừ kho
			err := inventoryCache.DeductStock(ctx, productID, userID, 1, limitPerUser)
			
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// 4. Assert & Deep Dive Result
	// Kỳ vọng: Chỉ có đúng 100 người mua được (vì kho có 100)
    // 50 người phải bị fail (OutOfStock)
	assert.Equal(t, 100, successCount, "Chỉ được phép bán đúng 100 đơn")
	assert.Equal(t, 50, failCount, "Phải có 50 đơn bị từ chối")

	// Check lại Redis: Stock phải bằng 0 (không được âm)
	finalStock, _ := inventoryCache.GetStock(ctx, productID)
	assert.Equal(t, 0, finalStock, "Kho phải về 0")
}