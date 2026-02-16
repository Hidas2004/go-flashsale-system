package domain

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
)

// UserRepository defines methods for user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

// ProductRepository defines methods for product data access
type ProductRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, product *models.Product) error
	FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error)
}

// OrderRepository defines methods for order data access
type OrderRepository interface {
	CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
	ListAll(ctx context.Context, page, limit int) ([]*models.Order, int64, error)
}

// InventoryRepository defines methods for inventory data access (DB)
type InventoryRepository interface {
	DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error
	FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
	Create(ctx context.Context, inv *models.Inventory) error
	// [NEW] Hoàn trả kho (Dùng cho Cancel Order)
	RestoreStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

// InventoryCache defines methods for inventory caching (Redis)
type InventoryCache interface {
	DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error
	IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error
	SetStock(ctx context.Context, productID uuid.UUID, stock int) error
	GetStock(ctx context.Context, productID uuid.UUID) (int, error)
}

// MessageQueue defines methods for messaging (RabbitMQ)
type MessageQueue interface {
	Publish(ctx context.Context, msg interface{}) error
}

// ProductCache defines methods for product caching (Redis)
type ProductCache interface {
	GetFlashSaleProducts(ctx context.Context) ([]*models.Product, error)
	SetFlashSaleProducts(ctx context.Context, products []*models.Product) error
	InvalidateFlashSaleProducts(ctx context.Context) error
	GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error)
	SetProduct(ctx context.Context, product *models.Product) error
}
