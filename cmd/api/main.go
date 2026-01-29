// main.go (File tạm để test, dùng xong xoá)
package main

import (
	"log"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
)

func main() {
	//1 load config
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("❌ LoadConfig failed: %v", err)
	}
	log.Println("🚀 Config loaded successfully")

	//2 connect database
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	log.Println("🚀 Connected to PostgreSQL successfully")

	//3 connect redis
	rdb, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	log.Println("🚀 Connected to Redis successfully")

}
