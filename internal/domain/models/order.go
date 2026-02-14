package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   //đang chờ sử lý
	OrderStatusConfirmed OrderStatus = "confirmed" // Đã xác nhận, chuẩn bị hàng
	OrderStatusShipping  OrderStatus = "shipping"  // Đang giao
	OrderStatusCompleted OrderStatus = "completed" // Giao thành công
	OrderStatusFailed    OrderStatus = "failed"    // Thanh toán lỗi, v.v.
	OrderStatusCancelled OrderStatus = "cancelled" // User/Admin hủy
)

type Order struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"` // Index để user xem lịch sử đơn
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`

	Quantity   int             `gorm:"not null" json:"quantity"`
	TotalPrice decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_price"`

	Status OrderStatus `gorm:"type:varchar(50);default:'pending';index" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Associations
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// IsValidTransition kiểm tra xem có được phép chuyển từ trạng thái cũ sang trạng thái mới không
func (o *Order) IsValidTransition(target OrderStatus) bool {
	switch o.Status {
	// 1. Từ Pending chỉ được sang: Confirmed, Cancelled
	case OrderStatusPending:
		return target == OrderStatusConfirmed || target == OrderStatusCancelled
	// 2. Từ Confirmed chỉ được sang: Shipping, Cancelled (trước khi giao)
	case OrderStatusConfirmed:
		return target == OrderStatusShipping || target == OrderStatusCancelled
	// 3. Từ Shipping chỉ được sang: Completed, Failed, Cancelled (Admin hủy/Hoàn trả)
	case OrderStatusShipping:
		return target == OrderStatusCompleted || target == OrderStatusFailed || target == OrderStatusCancelled
	// 4. Các trạng thái kết thúc: Không được chuyển đi đâu nữa
	case OrderStatusCompleted, OrderStatusFailed, OrderStatusCancelled:
		return false

	default:
		return false
	}
}

func (Order) TableName() string {
	return "orders"
}
