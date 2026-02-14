package dtos

import "github.com/shopspring/decimal"

type CreateProductRequest struct {
	Name           string          `json:"name" binding:"required"`
	Description    string          `json:"description"`
	Price          decimal.Decimal `json:"price" binding:"required"` // Tạm bỏ gt=0 vì validator chưa support decimal
	ImageURL       string          `json:"image_url"`
	Inventory      int             `json:"inventory" binding:"required,gte=0"` // Tồn kho ban đầu
	IsFlashSale    bool            `json:"is_flash_sale"`
	FlashSalePrice decimal.Decimal `json:"flash_sale_price"`
}

type UpdateProductRequest struct {
	Name           *string          `json:"name"`
	Description    *string          `json:"description"`
	Price          *decimal.Decimal `json:"price"` // Tạm bỏ binding check
	ImageURL       *string          `json:"image_url"`
	IsFlashSale    *bool            `json:"is_flash_sale"`
	FlashSalePrice *decimal.Decimal `json:"flash_sale_price"`
}
