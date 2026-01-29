package database

import (
	"fmt"
	"time"

	"github.com/Hidas2004/go-flashsale-system/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB khởi tạo kết nối PostgreSQL với cấu hình Connection Pool tối ưu
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 1. Tạo DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode,
	)
	// 2. Mở kết nối GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 3. Cấu hình Connection Pool (Quan trọng nhất)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// MaxOpenConns: Số lượng kết nối tối đa được mở cùng lúc.
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	// MaxIdleConns: Số lượng kết nối "rảnh" được giữ lại để dùng tiếp.
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	// ConnMaxLifetime: Thời gian tối đa 1 kết nối tồn tại.
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
