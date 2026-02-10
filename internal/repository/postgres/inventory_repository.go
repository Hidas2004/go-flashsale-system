package postgres

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/Hidas2004/go-flashsale-system/internal/domain/models"
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