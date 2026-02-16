package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/http/middleware"
	v1 "github.com/Hidas2004/go-flashsale-system/internal/delivery/http/v1"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setupRouter(t *testing.T) (*gin.Engine, *models.Product) {
	// Load Config (Hardcoded for test environment or load from file)
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5433", // Docker exposed port
			User:     "postgres",
			Password: "postgres",
			DBName:   "flashdeal_db",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:           "amqp://guest:guest@localhost:5672/",
			Exchange:      "flashsale_exchange",
			Queue:         "orders_queue",
			RoutingKey:    "order.created",
			PrefetchCount: 10,
		},
		JWT: config.JWTConfig{
			Secret:      "test_secret",
			ExpireHours: 1,
		},
		RateLimit: config.RateLimitConfig{
			Limit:    100, // High limit for tests
			Duration: 1,
		},
	}

	// 1. Infrastructure
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	rdb, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	mq, err := rabbitmq.NewRabbitMQClient(rabbitmq.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		Exchange:      cfg.RabbitMQ.Exchange,
		Queue:         cfg.RabbitMQ.Queue,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	})
	if err != nil {
		t.Logf("Warning: RabbitMQ connection failed: %v. Integration test might fail if MQ is required.", err)
		// We can mock MQ here if we want, but let's try real connection
	}

	// 2. Repos
	userRepo := postgres.NewUserRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	productRepo := postgres.NewProductRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryCache := redis.NewInventoryCache(rdb)
	productCache := redis.NewProductCache(rdb) // Added
	txManager := database.NewGormTransactionManager(db)

	// 3. UseCases
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWT.Secret, 60)
	productUC := usecase.NewProductUseCase(productRepo, inventoryRepo, txManager, productCache)
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo, inventoryRepo, inventoryCache, mq, txManager)

	// 4. Handlers
	authHandler := v1.NewAuthHandler(authUC)
	productHandler := v1.NewProductHandler(productUC)
	orderHandler := v1.NewOrderHandler(orderUC)

	// 5. Middlewares
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	rateLimiter := middleware.NewRateLimiterMiddleware(rdb)
	adminMiddleware := middleware.NewAdminMiddleware(db)

	// 6. Router
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

	// Setup Test Data (Product)
	product := &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Price:       decimal.NewFromInt(100),
		IsFlashSale: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	// Set flash sale time valid
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	product.FlashSaleStart = &start
	product.FlashSaleEnd = &end
	val := decimal.NewFromInt(50)
	product.FlashSalePrice = &val

	_ = productRepo.Create(context.Background(), product)
	// Create Inventory
	inv := &models.Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Stock:     100,
	}
	_ = inventoryRepo.Create(context.Background(), inv)
	// Warmup cache
	_ = inventoryCache.SetStock(context.Background(), product.ID, 100)

	return router, product
}

func TestAPI_FullFlow(t *testing.T) {
	router, product := setupRouter(t)

	// 1. Register
	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	registerBody := map[string]interface{}{
		"email":     email,
		"password":  "password123",
		"full_name": "Test User",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Login
	loginBody := map[string]interface{}{
		"email":    email,
		"password": "password123",
	}
	body, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	data := loginResp["data"].(map[string]interface{})
	token := data["token"].(string)
	assert.NotEmpty(t, token)

	// 3. Create Order
	orderBody := map[string]interface{}{
		"product_id": product.ID.String(),
		"quantity":   1,
	}
	body, _ = json.Marshal(orderBody)
	req = httptest.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Expect 200 OK because we process async but return success immediately, structure might be used
	// Or 202 Accepted depending on handler implementation.
	// Checked order_handler.go: c.JSON(http.StatusOK, response.SuccessResponse("Order flash sale created successfully", orderResp))
	assert.Equal(t, http.StatusOK, w.Code)
}
