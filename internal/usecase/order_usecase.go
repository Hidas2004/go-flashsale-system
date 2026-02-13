package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type orderUseCase struct {
	orderRepo      domain.OrderRepository
	productRepo    domain.ProductRepository
	inventoryRepo  domain.InventoryRepository
	inventoryCache domain.InventoryCache
	mq             domain.MessageQueue
	txManager      database.TransactionManager
}

func NewOrderUseCase(
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	inventoryRepo domain.InventoryRepository,
	inventoryCache domain.InventoryCache,
	mq domain.MessageQueue,
	txManager database.TransactionManager,
) OrderUseCase {
	return &orderUseCase{
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		inventoryRepo:  inventoryRepo,
		inventoryCache: inventoryCache,
		mq:             mq,
		txManager:      txManager,
	}
}

// CreateFlashSaleOrder - Xử lý tạo đơn hàng Flash Sale (High Concurrency)
func (u *orderUseCase) CreateFlashSaleOrder(ctx context.Context, userID uuid.UUID, req *dtos.CreateOrderRequest) (*dtos.OrderResponse, error) {
	// 1. Validate Product & Flash Sale Logic
	product, err := u.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	if !product.IsFlashSale {
		return nil, fmt.Errorf("product is not in flash sale")
	}

	// Check thời gian Flash Sale
	now := time.Now()
	if product.FlashSaleStart != nil && now.Before(*product.FlashSaleStart) {
		return nil, fmt.Errorf("flash sale has not started yet")
	}
	if product.FlashSaleEnd != nil && now.After(*product.FlashSaleEnd) {
		return nil, fmt.Errorf("flash sale has ended")
	}

	// 2. Deduct Redis Stock (Lua Script - Atomic)
	// limit = 1 (mỗi user chỉ được mua 1 cái trong đợt sale này để tránh đầu cơ)
	if err := u.inventoryCache.DeductStock(ctx, req.ProductID, userID.String(), req.Quantity, 1); err != nil {
		return nil, fmt.Errorf("failed to deduct stock: %w", err)
	}

	// 3. Tính toán giá tiền
	// Ưu tiên lấy giá Flash Sale nếu có
	price := product.Price
	if product.FlashSalePrice != nil {
		price = *product.FlashSalePrice
	}
	totalPrice := price.Mul(decimal.NewFromInt(int64(req.Quantity)))
	// Lưu ý: totalPrice hiện tại chỉ dùng để log hoặc bắn message,
	// việc tính toán chính xác cuối cùng nên override ở Consumer để bảo mật hơn.

	// 4. Tạo Message đẩy vào Queue
	orderID := uuid.New()
	msg := dtos.OrderMessage{
		OrderID:    orderID,
		UserID:     userID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice.InexactFloat64(), // Convert sang float để json transport dễ dàng
		CreatedAt:  time.Now(),
	}

	// 5. Publish to RabbitMQ
	if err := u.mq.Publish(ctx, msg); err != nil {
		// [CRITICAL] Rollback Redis Stock nếu publish lỗi
		// Đây là cơ chế "Best Effort" để tránh mất hàng (Stock Leak)
		log.Printf("❌ Publish RabbitMQ failed for order %s. Rolling back stock...", orderID)
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

// ProcessOrder - Worker sẽ gọi hàm này để xử lý message từ Queue
func (u *orderUseCase) ProcessOrder(ctx context.Context, msg *dtos.OrderMessage) error {
	// 1. Idempotency Check (Tránh duplicate đơn hàng)
	exists, err := u.orderRepo.CheckOrderExists(ctx, msg.OrderID)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("⚠️ Order %s processed. Skipping.", msg.OrderID)
		return nil
	}

	// 2. Transaction DB (Tạo Order + Trừ Inventory DB)
	return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		order := &models.Order{
			ID:         msg.OrderID,
			UserID:     msg.UserID,
			ProductID:  msg.ProductID,
			Quantity:   msg.Quantity,
			TotalPrice: decimal.NewFromFloat(msg.TotalPrice),
			Status:     models.OrderStatusPending,
		}

		// 2.1 Lưu đơn hàng
		if err := u.orderRepo.CreateOrder(txCtx, order); err != nil {
			return err
		}

		// 2.2 Trừ kho DB (InventoryRepository - Postgres) để đồng bộ với Redis
		if err := u.inventoryRepo.DeductStock(txCtx, msg.ProductID, msg.Quantity); err != nil {
			return err
		}

		// 2.3 Update trạng thái thành công
		return u.orderRepo.UpdateStatus(txCtx, order.ID, models.OrderStatusConfirmed)
	})
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	return u.orderRepo.FindByID(ctx, orderID)
}

func (u *orderUseCase) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	return u.orderRepo.FindByUserID(ctx, userID)
}
