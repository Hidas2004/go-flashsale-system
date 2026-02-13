package main

import (
	"log"
	"time"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/http/middleware"
	v1 "github.com/Hidas2004/go-flashsale-system/internal/delivery/http/v1"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("❌ LoadConfig failed: %v", err)
	}
	log.Println("🚀 Config loaded successfully")

	// 2. Connect Database (Postgres)
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	// Verify connection
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	log.Println("✅ PostgreSQL Connected")

	// 3. Connect Redis
	rdb, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	log.Println("✅ Redis Connected")

	// 4. Connect RabbitMQ (Producer)
	rabbitClient, err := rabbitmq.NewRabbitMQClient(rabbitmq.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		Exchange:      cfg.RabbitMQ.Exchange,
		Queue:         cfg.RabbitMQ.Queue,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	})
	if err != nil {
		log.Fatalf("❌ RabbitMQ failed: %v", err)
	}
	defer rabbitClient.Close()
	log.Println("✅ RabbitMQ Connected")

	// 5. Init Repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryCache := redis.NewInventoryCache(rdb)

	// Transaction Manager
	txManager := database.NewGormTransactionManager(db)

	// 6. Init UseCases
	// Auth UseCase: expired time in minutes
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours*60)
	productUC := usecase.NewProductUseCase(productRepo)
	// Order UseCase needs inventory logic + rabbitmq + tx
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo, inventoryRepo, inventoryCache, rabbitClient, txManager)
	// Inventory UseCase (Optional for API, mostly used by scripts or admin)
	_ = usecase.NewInventoryUseCase(inventoryRepo, inventoryCache)

	// 7. Init Handlers
	authHandler := v1.NewAuthHandler(authUC)
	productHandler := v1.NewProductHandler(productUC)
	orderHandler := v1.NewOrderHandler(orderUC)

	// 8. Setup Router
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Middlewares
	r.Use(middleware.PrometheusMiddleware()) // Prometheus
	// Rate Limiter: 60 requests / minute / IP
	rateLimiter := middleware.NewRateLimiterMiddleware(rdb)
	r.Use(rateLimiter.Limit(60, 60*time.Second))

	// Base Routes
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// API V1 Group
	apiV1 := r.Group("/api/v1")
	{
		// --- Auth Routes ---
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// --- Product Routes ---
		products := apiV1.Group("/products")
		{
			products.GET("", productHandler.GetFlashSaleProducts)
			products.GET("/:id", productHandler.GetProduct)
		}

		// --- Protected Routes (Require Login) ---
		protected := apiV1.Group("/")
		authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
		protected.Use(authMiddleware.Handle())
		{
			// Orders
			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetUserOrders)
				orders.GET("/:id", orderHandler.GetOrder) // Get detail
			}
		}
	}

	// 9. Run Server
	log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
