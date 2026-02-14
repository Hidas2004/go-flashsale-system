package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	// 2. Các phương thức được phép (HTTP Methods)
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	// 3. Các Header được phép gửi lên (Headers)
	// Quan trọng nhất là "Authorization" để gửi Token và "Content-Type" để gửi JSON
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
	}
	// 4. Các Header mà Frontend được phép đọc từ Response
	// Đôi khi em trả về custom header, frontend cần đọc được nó
	config.ExposeHeaders = []string{
		"Content-Length",
	}
	config.MaxAge = 12 * time.Hour

	// 6. Cho phép gửi Cookies/Credentials (nếu cần thiết)
	config.AllowCredentials = true

	return cors.New(config)
}
