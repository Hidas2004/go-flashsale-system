package redis

import (
	"context"
	"errors"
	"fmt"

	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed lua_scripts/deduct_stock.lua
var deductStockScript string
var (
	ErrOutOfStock    = errors.New("products out of stock/sản phẩm hết hàng")
	ErrLimitExceeded = errors.New("user purchase limit exceeded/đã vượt quá giới hạn mua của người dùng")
)

type StockRepository struct {
	client *redis.Client
}

func NewStockRepository(client *redis.Client) *StockRepository {
	return &StockRepository{client: client}
}

func (r *StockRepository) DeductStock(ctx context.Context, productID string, userID string, quantity int, limit int) error {
	//định nghĩa key
	stockKey := fmt.Sprintf("inventory:product:%s", productID)
	boughtKey := fmt.Sprintf("inventory:product:%s:bought", productID)
	//gọi lua script
	//Hành động: Gửi lệnh sang Redis Server -> Chạy script Lua -> Lấy kết quả về.
	result, err := r.client.Eval(ctx, deductStockScript, []string{stockKey, boughtKey}, quantity, limit, userID).Int()
	if err != nil {
		return fmt.Errorf("redis eval error: %w", err)
	}
	switch result {
	case 1:
		return nil //// Thành công
	case -1:
		return ErrOutOfStock
	case -2:
		return ErrLimitExceeded
	default:
		return fmt.Errorf("unknown result from redis script/mã lỗi lua không xác định: %d", result)
	}

}
