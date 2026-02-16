package models

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;unique;not null;index" json:"product_id"` // 1 Product chỉ có 1 dòng Inventory

	Stock         int `gorm:"not null;default:0;check:stock >= 0" json:"stock"`
	ReservedStock int `gorm:"not null;default:0" json:"reserved_stock"`
	Sold          int `gorm:"not null;default:0" json:"sold"`
	Version       int `gorm:"default:1" json:"version"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Inventory) TableName() string {
	return "inventory"
}
