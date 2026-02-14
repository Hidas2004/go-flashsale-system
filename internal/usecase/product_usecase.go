package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/google/uuid"
)

type productUseCase struct {
	productRepo   domain.ProductRepository
	inventoryRepo domain.InventoryRepository
	txManager     database.TransactionManager
	productCache  domain.ProductCache
}

func NewProductUseCase(
	productRepo domain.ProductRepository,
	inventoryRepo domain.InventoryRepository,
	txManager database.TransactionManager,
	productCache domain.ProductCache,
) ProductUseCase {
	return &productUseCase{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		txManager:     txManager,
		productCache:  productCache,
	}
}

func (uc *productUseCase) CreateProduct(ctx context.Context, req *dtos.CreateProductRequest) (*models.Product, error) {
	// 1. Chuẩn bị dữ liệu
	product := &models.Product{
		ID:             uuid.New(),
		Name:           req.Name,
		Description:    &req.Description,
		Price:          req.Price,
		ImageURL:       &req.ImageURL,
		IsFlashSale:    req.IsFlashSale,
		FlashSalePrice: &req.FlashSalePrice,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// [FIX] Nếu là Flash Sale, tự động set thời gian bắt đầu (ngay bây giờ) và kết thúc (sau 24h)
	// Để query FindFlashSaleProducts (Start <= Now <= End) tìm thấy được.
	if req.IsFlashSale {
		now := time.Now()
		endTime := now.Add(24 * time.Hour)
		product.FlashSaleStart = &now
		product.FlashSaleEnd = &endTime
	}
	inventory := &models.Inventory{
		ID:            uuid.New(), // [FIX] Tạo ID mới, không để mặc định (nil)
		ProductID:     product.ID,
		Stock:         req.Inventory,
		ReservedStock: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	// 2. Thực thi Transaction
	err := uc.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		// A. Tạo Product
		if err := uc.productRepo.Create(ctx, product); err != nil {
			return err
		}
		// B. Tạo Inventory
		if err := uc.inventoryRepo.Create(ctx, inventory); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (pu *productUseCase) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	return pu.productRepo.FindByID(ctx, id)
}

func (pu *productUseCase) UpdateProduct(ctx context.Context, id uuid.UUID, req *dtos.UpdateProductRequest) (*models.Product, error) {
	//1 check tồn tại
	product, err := pu.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found")
	}
	//2 update field nếu có
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.ImageURL != nil {
		product.ImageURL = req.ImageURL
	}
	if req.FlashSalePrice != nil {
		product.FlashSalePrice = req.FlashSalePrice
	}

	product.UpdatedAt = time.Now()

	//3 save
	if err := pu.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil

}

func (pu *productUseCase) FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	//1 kiểm tra redis
	products, err := pu.productCache.GetFlashSaleProducts(ctx)
	if err == nil && len(products) > 0 {
		return products, nil
	}
	//2 nếu không có trong redis -> lấy từ db
	products, err = pu.productRepo.FindFlashSaleProducts(ctx)
	if err != nil {
		return nil, err
	}
	// 3. Lưu lại vào Cache cho người sau dùng
	// (Dùng goroutine để không block user)
	go pu.productCache.SetFlashSaleProducts(ctx, products)
	return products, nil
}

func (pu *productUseCase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	product, err := pu.productRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return pu.productRepo.Delete(ctx, product)
}
