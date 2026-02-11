package main

import (
	"context"
	"log"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/repository/postgres"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func main() {
	// 1. Load Config & Connect DB
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("❌ LoadConfig failed: %v", err)
	}
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Connect DB failed: %v", err)
	}

	// 2. Setup Repositories
	txManager := database.NewGormTransactionManager(db)
	userRepo := postgres.NewUserRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	invRepo := postgres.NewInventoryRepository(db)
	productRepo := postgres.NewProductRepository(db)

	ctx := context.Background()

	// 3. Chuẩn bị Dữ Liệu Test (Seed Data)
	userID := uuid.New()
	productID := uuid.New()

	// Clean up old data if needed (optional)

	// Tạo User giả
	log.Println("Creating User...")
	err = userRepo.Create(ctx, &models.User{
		ID: userID, Email: "test_" + userID.String()[:8] + "@gmail.com", FullName: "Test User",
	})
	if err != nil {
		log.Printf("⚠️ Create User warning (might exist): %v", err)
	}

	// Tạo Product giả
	log.Println("Creating Product...")
	err = productRepo.Create(ctx, &models.Product{
		ID: productID, Name: "Test Product", Price: decimal.NewFromInt(1000), IsFlashSale: true,
	})
	if err != nil {
		log.Printf("⚠️ Create Product warning: %v", err)
	}

	// Tạo Inventory (Kho: 10 cái)
	log.Println("Creating Inventory...")
	inv := &models.Inventory{
		ProductID: productID, Stock: 10,
	}
	err = invRepo.Create(ctx, inv)
	if err != nil {
		log.Printf("⚠️ Create Inventory warning: %v", err)
	}

	log.Println("✅ Seed Data created. Starting Transaction Test...")

	// 4. TEST CASE: Mua Hàng Thành Công
	// Mua 2 cái
	log.Println("Attempting to purchase 2 items...")
	err = txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// a. Tạo đơn hàng
		order := &models.Order{
			ID: uuid.New(), UserID: userID, ProductID: productID,
			Quantity: 2, TotalPrice: decimal.NewFromInt(2000),
			Status: models.OrderStatusPending, // Explicitly set status
		}
		if err := orderRepo.CreateOrder(txCtx, order); err != nil {
			return err
		}
		log.Println("   -> Order Created")

		// b. Trừ kho (Trong cùng transaction)
		if err := invRepo.DeductStock(txCtx, productID, 2); err != nil {
			return err
		}
		log.Println("   -> Stock Deducted")

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Transaction Failed: %v", err)
	}
	log.Println("🎉 Transaction committed success!")

	// 5. Verify Data
	invData, err := invRepo.FindByProductID(ctx, productID)
	if err != nil {
		log.Fatalf("❌ Cannot find inventory: %v", err)
	}

	log.Printf("🔍 Current Stock: %d (Expected: 8)", invData.Stock)

	if invData.Stock != 8 {
		log.Fatalf("❌ TEST FAILED: Stock mismatch. Expected 8, got %d", invData.Stock)
	}

	log.Println("✅ TEST PASSED: Setup & Transaction Logic OK")
}
