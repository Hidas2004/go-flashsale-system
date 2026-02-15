package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/worker"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/redis"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/Hidas2004/go-flashsale-system/pkg/rabbitmq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	//1 load cofig
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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
	productRepo := postgres.NewProductRepository(db)
	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryCache := redis.NewInventoryCache(rdb)

	// Usecase
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo, inventoryRepo, inventoryCache, rabbitClient, txManager)
	inventoryUC := usecase.NewInventoryUseCase(inventoryRepo, inventoryCache)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case <-ticker.C:
				log.Println("🔄 Starting periodic stock reconciliation...")
				products, err := productRepo.FindFlashSaleProducts(context.Background())
				if err != nil {
					log.Printf("❌ Failed to get flash sale products: %v", err)
					continue
				}
				// 2. Duyệt qua từng sản phẩm để đối soát
				for _, p := range products {
					if err := inventoryUC.ReconcileStock(context.Background(), p.ID); err != nil {
						log.Printf("❌ Failed to reconcile stock for product %s: %v", p.ID, err)
					}
				}
				log.Println("✅ Reconciliation finished.")
			case <-ctx.Done():
				log.Println("🛑 Stopping reconciliation ticker...")
				return
			}
		}
	}()

	// Log that InventoryUC is ready (or use it if needed)
	_ = inventoryUC
	//Consumer
	orderConsumer := worker.NewOrderConsumer(rabbitClient, orderUC)
	go func() {
		if err := orderConsumer.Start(ctx, wg); err != nil {
			log.Printf("❌ Consumer error: %v", err)
			cancel()
		}
	}()
	// [NEW] Metrics Server for Worker
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Println("📊 Worker Metrics running on :9091")
		if err := http.ListenAndServe(":9091", mux); err != nil {
			log.Printf("❌ Failed to start metrics server: %v", err)
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
