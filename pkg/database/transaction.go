package database

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) *GormTransactionManager {
	return &GormTransactionManager{db: db}
}

type contextKey string

const txKey contextKey = "DB_TX"

//File này giúp gom cả 2 hành động trên vào một Transaction
//Nếu cả 2 cùng thành công -> Commit (Lưu tất cả).
//Nếu 1 trong 2 bị lỗi -> Rollback (Hoàn tác tất cả, xóa đơn hàng vừa tạo).
func (tm *GormTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		//lấy tx kết nối db trong transaction
		//nhét nó vào ctx bằng 1 cái chìa khóa txkey
		//tạo ra một context mới (txCTX) chứa cái tx này
		txCtx := context.WithValue(ctx, txKey, tx)
		//3 chạy hàm logic chính (fn)
		//nhưng truyền vào cái context mới đã chứa transaction
		return fn(txCtx)

	})
}
