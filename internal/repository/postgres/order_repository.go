package postgres

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CheckOrderExists kiểm tra trùng lặp (Idempotency)
func (r *OrderRepository) CheckOrderExists(ctx context.Context, orderID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", orderID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateOrder lưu đơn hàng
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	return GetDB(ctx, r.db).Create(order).Error
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status models.OrderStatus) error {
	return GetDB(ctx, r.db).Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}
