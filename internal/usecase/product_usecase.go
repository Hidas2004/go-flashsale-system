package usecase

import (
	"context"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
)

type productUseCase struct {
	productRepository domain.ProductRepository
}

func NewProductUseCase(productRepository domain.ProductRepository) ProductUseCase {
	return &productUseCase{productRepository: productRepository}
}

func (pu *productUseCase) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	return pu.productRepository.FindByID(ctx, id)
}
func (pu *productUseCase) Create(ctx context.Context, product *models.Product) error {
	return pu.productRepository.Create(ctx, product)
}
func (pu *productUseCase) Update(ctx context.Context, product *models.Product) error {
	return pu.productRepository.Update(ctx, product)
}
func (pu *productUseCase) FindFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	return pu.productRepository.FindFlashSaleProducts(ctx)
}
