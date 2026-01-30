package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"` // Index để user xem lịch sử đơn
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`

	Quantity   int             `gorm:"not null" json:"quantity"`
	TotalPrice decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_price"`

	Status OrderStatus `gorm:"type:varchar(50);default:'pending';index" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}
