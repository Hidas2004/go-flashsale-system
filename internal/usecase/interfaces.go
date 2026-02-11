package usecase

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
)

// OrderRepository: Định nghĩa các hành vi liên quan đến lưu trữ đơn hàng
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
}

type ProductRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
    Create(ctx context.Context, product *models.Product) error
    Update(ctx context.Context, product *models.Product) error
    FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error)
}

// InventoryRepository: Làm việc với DB (Postgres)
type InventoryRepository interface {
	FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error)
	DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

type InventoryCache interface {
	// DeductStock: Trừ kho trên Redis (có xử lý Race Condition bằng Lua Script)
	DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error
	//cộng kho lại dùng rollback khi sử lý đơn hàng thất bại
	IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error
	GetStock(ctx context.Context, productID uuid.UUID) (int, error)
	SetInitialStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

// MessageQueue:  việc gửi message (RabbitMQ)
type MessageQueue interface {
	Publish(ctx context.Context, msg interface{}) error
}
