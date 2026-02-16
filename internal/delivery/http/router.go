package http

import (
	"fmt"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/delivery/http/middleware"
	v1 "github.com/Hidas2004/go-flashsale-system/internal/delivery/http/v1"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterConfig struct {
	AuthHandler     *v1.AuthHandler
	ProductHandler  *v1.ProductHandler
	OrderHandler    *v1.OrderHandler
	AdminMiddleware *middleware.AdminMiddleware
	AuthMiddleware  *middleware.AuthMiddleware
	RateLimiter     *middleware.RateLimiterMiddleware
	RateLimit       int
	RateDuration    int
}

func NewRouter(config *RouterConfig) *gin.Engine {
	r := gin.New()
	//loger ghi log request/response
	r.Use(gin.Logger())
	//recovery phục hồi nếu panic xảy ra
	r.Use(gin.Recovery())
	//Cors cấu hình cho phép fontend goi API
	r.Use(middleware.CORS())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "System is running 🚀"})
	})
	// [NEW] Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	fmt.Println("✅ Registered /metrics endpoint")

	v1Group := r.Group("/api/v1")
	{
		// --- Public Routes ---
		auth := v1Group.Group("/auth")
		{
			auth.POST("/register", config.AuthHandler.Register)
			auth.POST("/login", config.AuthHandler.Login)
		}

		products := v1Group.Group("/products")
		{
			products.GET("/flash-sale", config.ProductHandler.GetFlashSaleProducts)
			products.GET("/:id", config.ProductHandler.GetProduct)
		}
		// --- Protected Routes (Yêu cầu đăng nhập & Rate Limit) ---
		protected := v1Group.Group("")
		//1 kiểm tra token
		protected.Use(config.AuthMiddleware.Handle())
		{
			orders := protected.Group("/orders")
			{
				// Áp dụng Rate Limiter cho tạo đơn hàng (tránh spam)
				orders.POST("", config.RateLimiter.Limit(config.RateLimit, time.Duration(config.RateDuration)*time.Second), config.OrderHandler.CreateOrder)
				orders.GET("/:id", config.OrderHandler.GetOrder)
				orders.GET("", config.OrderHandler.GetUserOrders)
			}
			//---chỉ admin mới được vào ---
			admin := protected.Group("/admin")
			admin.Use(config.AdminMiddleware.Handle())
			{

				products := admin.Group("/products")
				{
					products.POST("", config.ProductHandler.CreateProduct)
					products.PUT("/:id", config.ProductHandler.UpdateProduct)
					products.DELETE("/:id", config.ProductHandler.DeleteProduct)
				}
				orders := admin.Group("/orders")
				{
					orders.GET("", config.OrderHandler.AdminListOrders)                // Xem danh sách
					orders.PATCH("/:id/status", config.OrderHandler.AdminUpdateStatus) // Update trạng thái
				}
			}
		}

	}
	return r

}
