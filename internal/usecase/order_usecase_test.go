package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFlashSaleOrder_Success(t *testing.T) {
	// Setup Mocks
	mockProductRepo := new(MockProductRepo)
	mockInventoryRepo := new(MockInventoryRepo)
	mockInventoryCache := new(MockInventoryCache)
	mockMQ := new(MockMQ)
	mockTxManager := new(MockTxManager)
	mockOrderRepo := new(MockOrderRepo) // Not used in CreateFlashSaleOrder but needed for constructor

	uc := NewOrderUseCase(mockOrderRepo, mockProductRepo, mockInventoryRepo, mockInventoryCache, mockMQ, mockTxManager)

	// Data
	userID := uuid.New()
	productID := uuid.New()
	req := &dtos.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	price := decimal.NewFromInt(100)
	flashPrice := decimal.NewFromInt(80)

	product := &models.Product{
		ID:             productID,
		IsFlashSale:    true,
		FlashSaleStart: &startTime,
		FlashSaleEnd:   &endTime,
		Price:          price,
		FlashSalePrice: &flashPrice,
	}

	// Expectations
	mockProductRepo.On("FindByID", mock.Anything, productID).Return(product, nil)
	mockInventoryCache.On("GetStock", mock.Anything, productID).Return(10, nil) // In stock in cache
	mockInventoryCache.On("DeductStock", mock.Anything, productID, userID.String(), 1, 1).Return(nil)
	mockMQ.On("Publish", mock.Anything, mock.MatchedBy(func(msg dtos.OrderMessage) bool {
		return msg.UserID == userID && msg.ProductID == productID && msg.Quantity == 1
	})).Return(nil)

	// Execute
	resp, err := uc.CreateFlashSaleOrder(context.Background(), userID, req)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "processing", resp.Status)

	mockProductRepo.AssertExpectations(t)
	mockInventoryCache.AssertExpectations(t)
	mockMQ.AssertExpectations(t)
}

func TestCreateFlashSaleOrder_OutOfStock(t *testing.T) {
	// Setup
	mockProductRepo := new(MockProductRepo)
	mockInventoryCache := new(MockInventoryCache)
	mockMQ := new(MockMQ)
	// Other mocks nil is fine if not called, but NewOrderUseCase requires them
	uc := NewOrderUseCase(new(MockOrderRepo), mockProductRepo, new(MockInventoryRepo), mockInventoryCache, mockMQ, new(MockTxManager))

	userID := uuid.New()
	productID := uuid.New()
	req := &dtos.CreateOrderRequest{ProductID: productID, Quantity: 1}

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)

	product := &models.Product{
		ID:             productID,
		IsFlashSale:    true,
		FlashSaleStart: &startTime,
		FlashSaleEnd:   &endTime,
		Price:          decimal.NewFromInt(100),
	}

	// Expectations
	mockProductRepo.On("FindByID", mock.Anything, productID).Return(product, nil)
	mockInventoryCache.On("GetStock", mock.Anything, productID).Return(10, nil)
	mockInventoryCache.On("DeductStock", mock.Anything, productID, userID.String(), 1, 1).Return(errors.New("out of stock"))

	// Execute
	resp, err := uc.CreateFlashSaleOrder(context.Background(), userID, req)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to deduct stock")
}

func TestCreateFlashSaleOrder_NotStarted(t *testing.T) {
	mockProductRepo := new(MockProductRepo)
	uc := NewOrderUseCase(new(MockOrderRepo), mockProductRepo, new(MockInventoryRepo), new(MockInventoryCache), new(MockMQ), new(MockTxManager))

	productID := uuid.New()
	req := &dtos.CreateOrderRequest{ProductID: productID, Quantity: 1}

	now := time.Now()
	startTime := now.Add(1 * time.Hour) // Future
	endTime := now.Add(2 * time.Hour)

	product := &models.Product{
		ID:             productID,
		IsFlashSale:    true,
		FlashSaleStart: &startTime,
		FlashSaleEnd:   &endTime,
	}

	mockProductRepo.On("FindByID", mock.Anything, productID).Return(product, nil)

	_, err := uc.CreateFlashSaleOrder(context.Background(), uuid.New(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flash sale has not started yet")
}
