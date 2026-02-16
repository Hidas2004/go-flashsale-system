package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hidas2004/go-flashsale-system/config"
	httpDelivery "github.com/Hidas2004/go-flashsale-system/internal/delivery/http"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/http/middleware"
	v1 "github.com/Hidas2004/go-flashsale-system/internal/delivery/http/v1"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect Infrastructure
	// Postgres
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// --- AUTO MIGRATE (Thêm đoạn này) ---
	// Tự động tạo bảng dựa trên struct Model
	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.Inventory{},
		&models.InventoryAudit{}, // Nếu có
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Database Migrated Successfully")

	// Redis
	rdb, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// RabbitMQ
	rabbitClient, err := rabbitmq.NewRabbitMQClient(rabbitmq.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		Exchange:      cfg.RabbitMQ.Exchange,
		Queue:         cfg.RabbitMQ.Queue,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()
	log.Println("✅ Connected to RabbitMQ")

	// 3. Initialize Repositories (Data Layer)
	userRepo := postgres.NewUserRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	productRepo := postgres.NewProductRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryCache := redis.NewInventoryCache(rdb)
	productCache := redis.NewProductCache(rdb)
	txManager := database.NewGormTransactionManager(db)

	// 4. Initialize UseCases (Domain Layer)
	// Auth UseCase: 24h expiration (hoặc config)
	authUseCase := usecase.NewAuthUseCase(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours*60)
	productUseCase := usecase.NewProductUseCase(
		productRepo,
		inventoryRepo,
		txManager,
		productCache,
	)
	// Order UseCase cần tất cả power: DB, Redis, RabbitMQ
	orderUseCase := usecase.NewOrderUseCase(orderRepo, productRepo, inventoryRepo, inventoryCache, rabbitClient, txManager)

	// 5. Initialize Handlers (Delivery Layer)
	authHandler := v1.NewAuthHandler(authUseCase)
	productHandler := v1.NewProductHandler(productUseCase)
	orderHandler := v1.NewOrderHandler(orderUseCase)

	// 6. Init Middlewares (Những trạm kiểm soát)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	rateLimiter := middleware.NewRateLimiterMiddleware(rdb)
	adminMiddleware := middleware.NewAdminMiddleware(db)

	// 7. Setup Router (Wiring Handlers & Middlewares)
	router := httpDelivery.NewRouter(&httpDelivery.RouterConfig{
		AuthHandler:     authHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		AuthMiddleware:  authMiddleware,
		RateLimiter:     rateLimiter,
		AdminMiddleware: adminMiddleware,
		RateLimit:       cfg.RateLimit.Limit,
		RateDuration:    cfg.RateLimit.Duration,
	})

	// 8. Setup HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// 9. Start Server in Goroutine
	// Tại sao phải chạy trong Goroutine? Để thread chính (main) không bị block,
	// nó đi xuống dưới chờ tín hiệu shutdown.
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 10. Graceful Shutdown
	// Tạo channel để lắng nghe tín hiệu từ hệ điều hành (Ctrl+C, Kill command)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block ở đây cho đến khi nhận được tín hiệu
	<-quit
	log.Println("🛑 Shutting down server...")

	// Tạo context với timeout 5 giây để server kịp hoàn thành nốt các request đang dở
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited properly")
}
