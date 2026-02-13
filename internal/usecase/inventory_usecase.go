package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
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
	if err := u.inventoryCache.SetInitialStock(ctx, productID, inventory.Stock); err != nil {
		return fmt.Errorf("failed to set stock to redis: %w", err)
	}
	log.Printf("✅ Synced stock for product %s: %d", productID, inventory.Stock)
	return nil
}
