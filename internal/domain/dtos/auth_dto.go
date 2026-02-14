package dtos

import (
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role"` // Optional: admin, user (default)
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// AuthClaims dùng cho JWT
type AuthClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}
