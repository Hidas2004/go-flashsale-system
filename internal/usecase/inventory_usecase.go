package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type InventoryUseCase struct {
	inventoryRepo  InventoryRepository
	inventoryCache InventoryCache
}

func NewInventoryUseCase(
	inventoryRepo InventoryRepository,
	inventoryCache InventoryCache,
) *InventoryUseCase {
	return &InventoryUseCase{
		inventoryRepo:  inventoryRepo,
		inventoryCache: inventoryCache,
	}
}

// SyncStockToRedis - Đồng bộ tồn kho từ DB lên Redis (Cache Warming)
func (u *InventoryUseCase) SyncStockToRedis(ctx context.Context, productID uuid.UUID) error {
	//1 lấy tồn kho thực tế từ DB
	inventory, err := u.inventoryRepo.FindByProductID(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get inventory from db: %w", err)
	}
	//2 set vào redis
	if err := u.inventoryCache.SetInitialStock(ctx, productID, inventory.Stock); err != nil {
		return fmt.Errorf("failed to set stock to redis: %w", err)
	}
	log.Printf("✅ Synced stock for product %s: %d", productID, inventory.Stock)
	return nil
}
