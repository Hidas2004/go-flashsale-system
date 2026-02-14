package postgres

import (
	"context"
	"fmt"

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

func (r *OrderRepository) CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := GetDB(ctx, r.db).
		Model(&models.Order{}).
		Where("id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	return GetDB(ctx, r.db).Create(order).Error
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	return GetDB(ctx, r.db).
		Model(&models.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := GetDB(ctx, r.db).
		Preload("User").
		Preload("Product").
		First(&order, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	var orders []*models.Order
	err := GetDB(ctx, r.db).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) ListAll(ctx context.Context, page, limit int) ([]*models.Order, int64, error) {
	var orders []*models.Order
	var total int64
	offset := (page - 1) * limit
	//1 đếm tổng số lượng đơn hàng (để làm phân trang fontend)
	if err := GetDB(ctx, r.db).Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 2. Query lấy dữ liệu (preload User để biết ai mua)
	err := GetDB(ctx, r.db).
		Preload("User").    // Ai mua?
		Preload("Product"). // Mua cái gì?
		Limit(limit).
		Offset(offset).
		Order("created_at DESC"). // Mới nhất lên đầu
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
