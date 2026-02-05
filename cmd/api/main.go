// main.go (File tạm để test, dùng xong xoá)
package main

import (
	"log"

	"github.com/Hidas2004/go-flashsale-system/config"
	"github.com/Hidas2004/go-flashsale-system/internal/delivery/http/middleware"
	"github.com/Hidas2004/go-flashsale-system/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	r := gin.Default()
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

	r.Use(middleware.PrometheusMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
