package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- MOCK DEFINITIONS (Duplicated from order_usecase_test.go) ---
type MockInventoryRepositoryForInventoryTest struct {
	mock.Mock
}

func (m *MockInventoryRepositoryForInventoryTest) FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepositoryForInventoryTest) DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepositoryForInventoryTest) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepositoryForInventoryTest) Create(ctx context.Context, inv *models.Inventory) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

type MockInventoryCacheForInventoryTest struct {
	mock.Mock
}

func (m *MockInventoryCacheForInventoryTest) DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error {
	args := m.Called(ctx, productID, userID, quantity, limit)
	return args.Error(0)
}

func (m *MockInventoryCacheForInventoryTest) IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryCacheForInventoryTest) GetStock(ctx context.Context, productID uuid.UUID) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockInventoryCacheForInventoryTest) SetInitialStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

// --- TEST CASES ---

func TestInventoryUseCase_SyncStockToRedis_Success(t *testing.T) {
	mockRepo := new(MockInventoryRepositoryForInventoryTest)
	mockCache := new(MockInventoryCacheForInventoryTest)
	uc := usecase.NewInventoryUseCase(mockRepo, mockCache)
	ctx := context.Background()
	productID := uuid.New()
	stock := 100

	// Expect inventory to be found in DB
	mockRepo.On("FindByProductID", ctx, productID).Return(&models.Inventory{
		ProductID: productID,
		Stock:     stock,
	}, nil)

	// Expect stock to be set in Redis
	mockCache.On("SetInitialStock", ctx, productID, stock).Return(nil)

	err := uc.SyncStockToRedis(ctx, productID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestInventoryUseCase_SyncStockToRedis_DBError(t *testing.T) {
	mockRepo := new(MockInventoryRepositoryForInventoryTest)
	mockCache := new(MockInventoryCacheForInventoryTest)
	uc := usecase.NewInventoryUseCase(mockRepo, mockCache)
	ctx := context.Background()
	productID := uuid.New()

	mockRepo.On("FindByProductID", ctx, productID).Return(nil, errors.New("db error"))

	err := uc.SyncStockToRedis(ctx, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get inventory from db")
	mockRepo.AssertExpectations(t)
	// Cache should not be called
	mockCache.AssertNotCalled(t, "SetInitialStock", mock.Anything, mock.Anything, mock.Anything)
}

func TestInventoryUseCase_SyncStockToRedis_RedisError(t *testing.T) {
	mockRepo := new(MockInventoryRepositoryForInventoryTest)
	mockCache := new(MockInventoryCacheForInventoryTest)
	uc := usecase.NewInventoryUseCase(mockRepo, mockCache)
	ctx := context.Background()
	productID := uuid.New()
	stock := 100

	mockRepo.On("FindByProductID", ctx, productID).Return(&models.Inventory{
		ProductID: productID,
		Stock:     stock,
	}, nil)

	mockCache.On("SetInitialStock", ctx, productID, stock).Return(errors.New("redis error"))

	err := uc.SyncStockToRedis(ctx, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set stock to redis")
	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
