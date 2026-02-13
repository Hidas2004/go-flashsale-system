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
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error)
}

type OrderUseCase interface {
	CreateFlashSaleOrder(ctx context.Context, userID uuid.UUID, req *dtos.CreateOrderRequest) (*dtos.OrderResponse, error)
	ProcessOrder(ctx context.Context, msg *dtos.OrderMessage) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
}

type InventoryUseCase interface {
	SyncStockToRedis(ctx context.Context, productID uuid.UUID) error
}
