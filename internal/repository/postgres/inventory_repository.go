package postgres

import (
	"context"
	"fmt"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	db := GetDB(ctx, r.db)

	result := db.Model(&models.Inventory{}).
		Where("product_id = ? AND stock >= ?", productID, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient stock or product not found")
	}

	return nil
}

// FindByProductID - Lấy thông tin tồn kho của sản phẩm
func (r *InventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error) {
	var inventory models.Inventory
	err := GetDB(ctx, r.db).
		Where("product_id = ?", productID).
		First(&inventory).Error

	if err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *InventoryRepository) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return GetDB(ctx, r.db).Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Update("stock", quantity).Error
}

func (r *InventoryRepository) Create(ctx context.Context, inv *models.Inventory) error {
	return GetDB(ctx, r.db).Create(inv).Error
}

// RestoreStock - Cộng lại số lượng tồn kho (Atomic Update)
func (r *InventoryRepository) RestoreStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return GetDB(ctx, r.db).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}
