package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Email        string    `gorm:"unique;not null;index" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"` // JSON "-" để không bao giờ trả về password
	FullName     string    `gorm:"type:varchar(100)" json:"full_name"`

	Role      string    `gorm:"type:varchar(20);default:'customer';not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
