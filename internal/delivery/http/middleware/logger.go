package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// khỏi tạo logger
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		//lưu lại đường dẫn khách muốn vào
		path := c.Request.URL.Path
		//lưu lại tham số
		query := c.Request.URL.RawQuery
		c.Next()
		//thu thập kết quả
		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		//phân loại mức độ nghiêm trọng
		msg := "Request processed"
		level := slog.LevelInfo

		if status >= 500 {
			level = slog.LevelError
			msg = "Server Error"
		} else if status >= 400 {
			level = slog.LevelWarn
			msg = "Client Error"
		}
		//Ghi Log ra giấy (Final Output)
		Logger.Log(c.Request.Context(), level, msg,
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", status),
			slog.String("ip", clientIP),
			slog.Duration("duration", duration),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}
