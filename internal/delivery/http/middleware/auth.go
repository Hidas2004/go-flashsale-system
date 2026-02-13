package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	secretKey string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: secret,
	}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		//A lấy Header
		authHeader := c.GetHeader("Authorization")
		//nếu ko có gì
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		//B parse Forrmat "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		//nếu ko có chữ "bearer hoặc không đủ 2 phần
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}
		tokenString := parts[1] //lấy phần mã thẻ phía sau
		// C. Parse & Verify Token
		token, err := jwt.ParseWithClaims(tokenString, &dtos.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			//c1 kiểm tra thuật toán
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			//c2 đưa chìa khóa ra để so khớp
			return []byte(m.secretKey), nil
		})
		// D. Handle Error
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		//E Trích xuất thông tin
		if claims, ok := token.Claims.(*dtos.AuthClaims); ok {
			c.Set("x-user-id", claims.UserID)
			c.Set("x-user-email", claims.Email)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		// F. Cho đi tiếp
		c.Next()
	}
}
