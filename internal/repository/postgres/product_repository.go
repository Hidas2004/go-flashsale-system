package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// 1. FindByID - Tìm sản phẩm theo ID
func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	if err := GetDB(ctx, r.db).WithContext(ctx).First(&product, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
	}
	return &product, nil
}

// 2. Create - Tạo sản phẩm mới
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	return GetDB(ctx, r.db).Create(product).Error
}

// 3. Update - Cập nhật sản phẩm
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	return GetDB(ctx, r.db).Save(product).Error
}

// 4. FindFlashSaleProducts - Lấy danh sách đang Flash Sale
func (r *ProductRepository) FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	var products []*models.Product
	now := time.Now()
	err := GetDB(ctx, r.db).
		Where("is_flash_sale = ?", true).
		Where("flash_sale_start <= ?", now).
		Where("flash_sale_end >= ?", now).
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}
