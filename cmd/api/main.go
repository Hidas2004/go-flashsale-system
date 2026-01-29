// main.go (File tạm để test, dùng xong xoá)
package main

import (
	"fmt"
	"log"

	"github.com/Hidas2004/go-flashsale-system/config"
)

func main() {
	// Load config: Dấu chấm (.) nghĩa là thư mục hiện tại nơi bạn chạy lệnh
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("❌ LoadConfig failed: %v", err)
	}

	fmt.Println("🚀 Config loaded successfully!")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("✅ App Mode: %s\n", cfg.Server.Mode)
	fmt.Printf("✅ DB Host: %s\n", cfg.Database.Host)
	fmt.Printf("✅ JWT Secret: %s\n", cfg.JWT.Secret)
	fmt.Println("--------------------------------------------------")
}
