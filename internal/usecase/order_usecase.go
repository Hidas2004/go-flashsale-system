package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error)
	CreateOrder(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error
}

type InventoryRepository interface {
	DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

type InventoryCache interface {
	DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error
	IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error
}
type MessageQueue interface {
	Publish(ctx context.Context, msg interface{}) error
}

type OrderUseCase struct {
	orderRepo      OrderRepository
	inventoryRepo  InventoryRepository
	inventoryCache InventoryCache
	mq             MessageQueue
	txManager      database.TransactionManager
}

func NewOrderUseCase(
	orderRepo OrderRepository,
	inventoryRepo InventoryRepository,
	inventoryCache InventoryCache,
	mq MessageQueue,
	txManager database.TransactionManager,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:      orderRepo,
		inventoryRepo:  inventoryRepo,
		inventoryCache: inventoryCache,
		mq:             mq,
		txManager:      txManager,
	}
}

func (u *OrderUseCase) CreateFlashSaleOrder(ctx context.Context, userID uuid.UUID, req *dtos.CreateOrderRequest) (*dtos.OrderResponse, error) {
	// 1.1 Trừ kho Redis
	if err := u.inventoryCache.DeductStock(ctx, req.ProductID, userID.String(), req.Quantity, 1); err != nil {
		return nil, fmt.Errorf("deduct redis failed: %w", err)
	}
	// 1.2 Tạo Message
	orderID := uuid.New()
	msg := dtos.OrderMessage{
		OrderID:    orderID,
		UserID:     userID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: 0,
		CreatedAt:  time.Now(),
	}
	//1.3 bắn vào queue
	if err := u.mq.Publish(ctx, msg); err != nil {
		_ = u.inventoryCache.IncrStock(ctx, req.ProductID, req.Quantity)
		return nil, fmt.Errorf("publish message failed: %w", err)
	}
	return &dtos.OrderResponse{
		OrderID:  orderID,
		Status:   "processing",
		Message:  "Order accepted",
		QueuedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (u *OrderUseCase) ProcessOrder(ctx context.Context, msg *dtos.OrderMessage) error {
	exists, err := u.orderRepo.CheckOrderExists(ctx, msg.OrderID)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("⚠️ Order %s processed. Skipping.", msg.OrderID)
		return nil
	}
	// Dùng txManager của pkg/database
	return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		order := &models.Order{
			ID:         msg.OrderID,
			UserID:     msg.UserID,
			ProductID:  msg.ProductID,
			Quantity:   msg.Quantity,
			TotalPrice: decimal.NewFromFloat(msg.TotalPrice),
			Status:     models.OrderStatusPending,
		}
		if err := u.orderRepo.CreateOrder(txCtx, order); err != nil {
			return err
		}

		// Trừ kho DB (InventoryRepository - Postgres)
		if err := u.inventoryRepo.DeductStock(txCtx, msg.ProductID, msg.Quantity); err != nil {
			return err
		}

		return u.orderRepo.UpdateStatus(txCtx, order.ID, models.OrderStatusConfirmed)
	})
}
