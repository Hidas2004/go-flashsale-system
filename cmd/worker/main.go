package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/worker"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
)

func main() {
	//1 load cofig
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}

	//// 2. Setup Infrastructure (Database, RabbitMQ)
	//postgreSQL
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✅ PostgreSQL Connected")

	//rabbitMQ connect
	rabbitClient, err := rabbitmq.NewRabbitMQClient(rabbitmq.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		Exchange:      cfg.RabbitMQ.Exchange,
		Queue:         cfg.RabbitMQ.Queue,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()
	log.Println("✅ RabbitMQ Connected")
	// 3. Setup Layers (Repository -> Usecase -> Consumer)
	// Init Redis
	rdb, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	log.Println("✅ Redis Connected")

	// Repository & Transaction Manager
	txManager := database.NewGormTransactionManager(db)
	orderRepo := postgres.NewOrderRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryCache := redis.NewInventoryCache(rdb)

	// Usecase
	orderUC := usecase.NewOrderUseCase(orderRepo, inventoryRepo, inventoryCache, rabbitClient, txManager)
	//Consumer
	orderConsumer := worker.NewOrderConsumer(rabbitClient, orderUC)
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	go func() {
		if err := orderConsumer.Start(ctx, wg); err != nil {
			log.Printf("❌ Consumer error: %v", err)
			cancel()
		}
	}()
	quit := make(chan os.Signal, 1) //Tạo một đường ống (channel) để nhận tín hiệu.
	//nếu người dùng bấm Ctr + C hoặc lệnh docker stop
	//đừng giết tôi ngay ,mà gửi vào biến quit cho tôi xử lý trươc
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("🚀 Worker is running. Press Ctrl+C to stop.")
	<-quit
	log.Println("🛑 Shutting down worker...")
	cancel()
	log.Println("⏳ Waiting for active messages to finish...")
	wg.Wait()
	log.Println("👋 Worker exited properly")
}
