package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string    `gorm:"not null;type:varchar(255)" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`

	// Sử dụng decimal để tính toán tiền chính xác
	Price    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	ImageURL *string         `gorm:"type:varchar(500)" json:"image_url"`

	// Flash Sale Logic
	IsFlashSale    bool             `gorm:"default:false;index" json:"is_flash_sale"` // Index để lọc nhanh sp flash sale
	FlashSalePrice *decimal.Decimal `gorm:"type:decimal(10,2)" json:"flash_sale_price"`
	FlashSaleStart *time.Time       `json:"flash_sale_start"`
	FlashSaleEnd   *time.Time       `json:"flash_sale_end"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Product) TableName() string {
	return "products"
}
