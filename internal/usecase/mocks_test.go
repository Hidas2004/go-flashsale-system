package usecase

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockProductRepo
type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepo) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) Delete(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

// MockInventoryRepo
type MockInventoryRepo struct {
	mock.Mock
}

func (m *MockInventoryRepo) DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepo) FindByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockInventoryRepo) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryRepo) Create(ctx context.Context, inv *models.Inventory) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *MockInventoryRepo) RestoreStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

// MockInventoryCache
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

func (m *MockInventoryCache) SetStock(ctx context.Context, productID uuid.UUID, stock int) error {
	args := m.Called(ctx, productID, stock)
	return args.Error(0)
}

func (m *MockInventoryCache) GetStock(ctx context.Context, productID uuid.UUID) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

// MockMQ
type MockMQ struct {
	mock.Mock
}

func (m *MockMQ) Publish(ctx context.Context, msg interface{}) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// MockTxManager
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Simply execute the function
	return fn(ctx)
}

// MockOrderRepo (Need this for NewOrderUseCase)
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CheckOrderExists(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.OrderStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}
func (m *MockOrderRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)

}
func (m *MockOrderRepo) ListAll(ctx context.Context, page, limit int) ([]*models.Order, int64, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Order), args.Get(1).(int64), args.Error(2)
}
