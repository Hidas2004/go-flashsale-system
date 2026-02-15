package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	redisRepo "github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/google/uuid"
)

type inventoryUseCase struct {
	inventoryRepo  domain.InventoryRepository
	inventoryCache domain.InventoryCache
}

func NewInventoryUseCase(
	inventoryRepo domain.InventoryRepository,
	inventoryCache domain.InventoryCache,
) InventoryUseCase {
	return &inventoryUseCase{
		inventoryRepo:  inventoryRepo,
		inventoryCache: inventoryCache,
	}
}

// SyncStockToRedis - Đồng bộ tồn kho từ DB lên Redis (Cache Warming)
func (u *inventoryUseCase) SyncStockToRedis(ctx context.Context, productID uuid.UUID) error {
	//1 lấy tồn kho thực tế từ DB
	inventory, err := u.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get inventory from db: %w", err)
	}
	//2 set vào redis
	if err := u.inventoryCache.SetStock(ctx, productID, inventory.Stock); err != nil {
		return fmt.Errorf("failed to set stock to redis: %w", err)
	}
	log.Printf("✅ Synced stock for product %s: %d", productID, inventory.Stock)
	return nil
}

// ReconcileStock thực hiện đối soát và đồng bộ dữ liệu tồn kho giữa Database và Redis.
func (u *inventoryUseCase) ReconcileStock(ctx context.Context, productID uuid.UUID) error {
	//1 lấy stock từ DB
	inventory, err := u.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get inventory from db: %w", err)
	}
	dbStock := inventory.Stock

	//2 lấy stock từ redis
	redisStock, err := u.inventoryCache.GetStock(ctx, productID)
	// Case 1: Redis Missing Key -> Phục hồi từ DB
	if errors.Is(err, redisRepo.ErrKeyNotFound) {
		log.Printf("[RECOVERY] Redis key missing for %s. Restoring from DB (Stock: %d)", productID, dbStock)
		return u.inventoryCache.SetStock(ctx, productID, dbStock)
	}
	if err != nil {
		return fmt.Errorf("failed to get redis stock: %w", err)
	}
	// 3. Logic Đối soát
	if redisStock == dbStock {
		return nil
	}

	// Case 2: Redis > DB
	if redisStock > dbStock {
		log.Printf("[SAFETY_FIX] Redis (%d) > DB (%d) for %s. Updating Redis to match DB.", redisStock, dbStock, productID)
		return u.inventoryCache.SetStock(ctx, productID, dbStock)
	}

	// Case 3: Redis < DB
	if redisStock < dbStock {
		log.Printf("⚠️ [LEAK_WARNING] Redis (%d) < DB (%d) for %s. No action taken (Conservative Mode). Check Queue consumers.", redisStock, dbStock, productID)
	}
	return nil
}
