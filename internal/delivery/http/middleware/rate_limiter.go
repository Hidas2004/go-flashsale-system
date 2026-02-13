package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiterMiddleware struct {
	rdb *redis.Client
}

func NewRateLimiterMiddleware(rdb *redis.Client) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{rdb: rdb}
}

// Limit: limit request / duration (Ví dụ: 5 request / 60 giây)
func (m *RateLimiterMiddleware) Limit(limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		//b1 xác định key
		key := "ratelimit" + c.ClientIP()
		if userID, exists := c.Get("x-user-id"); exists {
			key = fmt.Sprintf("ratelimit:user:%v", userID)
		}
		ctx := c.Request.Context()
		// 2. Tương tác với Redis (Dùng Pipeline để gom lệnh -> Tối ưu Network)
		pipe := m.rdb.Pipeline()
		//lệnh INCR tăng biến đếm lên 1, nếu chưa có -> tạo mới = 1
		incr := pipe.Incr(ctx, key)
		// Lệnh EXPIRE: Đặt thời gian hết hạn cho key
		pipe.Expire(ctx, key, duration)
		// Thực thi pipeline
		_, err := pipe.Exec(ctx)
		if err != nil {
			//Nếu kết nối Redis bị lỗi (Redis sập, mạng lag...), ta in lỗi ra Log và gọi c.Next() (cho phép request đi qua)
			fmt.Printf("Redis Rate Limit Error: %v\n", err)
			c.Next()
			return
		}
		//3 kiếm tra kết quả
		count := incr.Val()
		if count > int64(limit) {
			// Quá giới hạn -> Chặn
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Slow down! Too many requests.",
				"retry_after": duration.Seconds(),
			})
			return
		}
		c.Next()
	}
}
