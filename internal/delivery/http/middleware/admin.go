package middleware

import (
	"net/http"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminMiddleware struct {
	db *gorm.DB
}

func NewAdminMiddleware(db *gorm.DB) *AdminMiddleware {
	return &AdminMiddleware{db: db}
}

func (m *AdminMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		//1 lấy userid từ context
		userIDVal, exists := c.Get("x-user-id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(uuid.UUID)
		//2 tra cứu role
		var user models.User
		if err := m.db.Select("role").First(&user, "id = ?", userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		//3 kiểm tra role
		if user.Role != "admin" {
			// 403 Forbidden: Bạn là user hợp lệ, nhưng không đủ quyền
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin access required"})
			return
		}
		c.Next()
	}
}
