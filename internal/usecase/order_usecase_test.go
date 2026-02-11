package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK DEFINITIONS ---

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepository) CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

type MockInventoryRepository struct {
	mock.Mock
}

func (m *MockInventoryRepository) FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepository) DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepository) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}



type MockInventoryCache struct {
	mock.Mock
}

func (m *MockInventoryCache) DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error {
	args := m.Called(ctx, productID, userID, quantity, limit)
	return args.Error(0)
}

func (m *MockInventoryCache) IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryCache) GetStock(ctx context.Context, productID uuid.UUID) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockInventoryCache) SetInitialStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

type MockMessageQueue struct {
	mock.Mock
}

func (m *MockMessageQueue) Publish(ctx context.Context, msg interface{}) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Mock implementation: just execute the function
	return fn(ctx)
}

// --- TEST CASES ---

func TestCreateFlashSaleOrder_Success(t *testing.T) {
	// 1. Setup
	mockOrderRepo := new(MockOrderRepository)
	mockProductRepo := new(MockProductRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	mockInventoryCache := new(MockInventoryCache)
	mockMQ := new(MockMessageQueue)
	mockTx := new(MockTransactionManager)

	uc := usecase.NewOrderUseCase(mockOrderRepo, mockProductRepo, mockInventoryRepo, mockInventoryCache, mockMQ, mockTx)

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()
	req := &dtos.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}

	// 2. Expectation
	flashSalePrice := decimal.NewFromFloat(9.99)
	mockProductRepo.On("FindByID", ctx, productID).Return(&models.Product{
		ID:             productID,
		IsFlashSale:    true,
		FlashSalePrice: &flashSalePrice,
		Price:          decimal.NewFromFloat(20.00),
	}, nil)

	mockInventoryCache.On("DeductStock", ctx, productID, userID.String(), 1, 1).Return(nil)
	mockMQ.On("Publish", ctx, mock.AnythingOfType("dtos.OrderMessage")).Return(nil)

	// 3. Execute
	resp, err := uc.CreateFlashSaleOrder(ctx, userID, req)

	// 4. Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "processing", resp.Status)
	mockMQ.AssertExpectations(t)
	mockInventoryCache.AssertExpectations(t)
}

func TestCreateFlashSaleOrder_Fail_RollbackRedis(t *testing.T) {
	// 1. Setup
	mockOrderRepo := new(MockOrderRepository)
	mockProductRepo := new(MockProductRepository)
	mockInventoryRepo := new(MockInventoryRepository)
	mockInventoryCache := new(MockInventoryCache)
	mockMQ := new(MockMessageQueue) // Simulate RabbitMQ connect failure
	mockTx := new(MockTransactionManager)

	uc := usecase.NewOrderUseCase(mockOrderRepo, mockProductRepo, mockInventoryRepo, mockInventoryCache, mockMQ, mockTx)

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()
	req := &dtos.CreateOrderRequest{
		ProductID: productID,
		Quantity:  1,
	}

	// 2. Expectation
	mockProductRepo.On("FindByID", ctx, productID).Return(&models.Product{
		ID:          productID,
		IsFlashSale: true,
	}, nil)

	// Redis Deduct Success
	mockInventoryCache.On("DeductStock", ctx, productID, userID.String(), 1, 1).Return(nil)

	// RabbitMQ Publish Failed
	mockMQ.On("Publish", ctx, mock.Anything).Return(errors.New("rabbitmq down"))

	// Expect Rollback (IncrStock)
	mockInventoryCache.On("IncrStock", ctx, productID, 1).Return(nil)

	// 3. Execute
	resp, err := uc.CreateFlashSaleOrder(ctx, userID, req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "publish message failed")

	// Verify Rollback was called
	mockInventoryCache.AssertCalled(t, "IncrStock", ctx, productID, 1)
}

func TestCreateFlashSaleOrder_Fail_ProductNotFound(t *testing.T) {
	// 1. Setup
	mockProductRepo := new(MockProductRepository)
	uc := usecase.NewOrderUseCase(nil, mockProductRepo, nil, nil, nil, nil)
	ctx := context.Background()
	req := &dtos.CreateOrderRequest{ProductID: uuid.New(), Quantity: 1}

	// 2. Expectation
	mockProductRepo.On("FindByID", ctx, req.ProductID).Return(nil, errors.New("db error"))

	// 3. Execute
	resp, err := uc.CreateFlashSaleOrder(ctx, uuid.New(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "product not found")
}

func TestCreateFlashSaleOrder_Fail_FlashSaleEnded(t *testing.T) {
	// 1. Setup
	mockProductRepo := new(MockProductRepository)
	uc := usecase.NewOrderUseCase(nil, mockProductRepo, nil, nil, nil, nil)
	ctx := context.Background()
	
	productID := uuid.New()
	req := &dtos.CreateOrderRequest{ProductID: productID, Quantity: 1}
	
	yesterday := time.Now().Add(-24 * time.Hour)

	// 2. Expectation
	mockProductRepo.On("FindByID", ctx, productID).Return(&models.Product{
		ID:             productID,
		IsFlashSale:    true,
		FlashSaleEnd:   &yesterday, // Ended
	}, nil)

	// 3. Execute
	resp, err := uc.CreateFlashSaleOrder(ctx, uuid.New(), req)

	// 4. Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "flash sale has ended")
}
