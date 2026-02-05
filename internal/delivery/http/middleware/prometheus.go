package middleware

import (
	"strconv"
	"time"

	"github.com/Hidas2004/go-flashsale-system/pkg/metrics"
	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // Xử lý request

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}

		// Ghi lại số liệu
		metrics.HttpRequestDuration.WithLabelValues(
			method,
			path,
			strconv.Itoa(status),
		).Observe(duration)
	}
}
