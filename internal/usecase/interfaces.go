package usecase

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	Register(ctx context.Context, req *dtos.RegisterRequest) (*dtos.AuthResponse, error)
	Login(ctx context.Context, req *dtos.LoginRequest) (*dtos.AuthResponse, error)
}

type ProductUseCase interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	CreateProduct(ctx context.Context, req *dtos.CreateProductRequest) (*models.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req *dtos.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error)
}

type OrderUseCase interface {
	CreateFlashSaleOrder(ctx context.Context, userID uuid.UUID, req *dtos.CreateOrderRequest) (*dtos.OrderResponse, error)
	ProcessOrder(ctx context.Context, msg *dtos.OrderMessage) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, newStatus models.OrderStatus) error
	ListOrders(ctx context.Context, page, limit int) ([]*models.Order, int64, error)
}

type InventoryUseCase interface {
	SyncStockToRedis(ctx context.Context, productID uuid.UUID) error
	ReconcileStock(ctx context.Context, productID uuid.UUID) error
}
