package redis

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//go:embed lua_scripts/*.lua
var luaScripts embed.FS // [FIXED] Đổi thành số nhiều (luaScripts) cho khớp code dưới

var (
	ErrOutOfStock    = errors.New("sản phẩm đã hết hàng")
	ErrLimitExceeded = errors.New("bạn đã vượt quá giới hạn mua cho phép")
	ErrSystem        = errors.New("lỗi hệ thống")
)

// [Refactor] Đổi tên thành InventoryCache cho đúng trách nhiệm
type InventoryCache struct {
	client        *redis.Client
	deductScript  *redis.Script
	reserveScript *redis.Script
}

func NewInventoryCache(client *redis.Client) *InventoryCache {
	// Helper function để load script cho gọn code
	load := func(filename string) *redis.Script {
		content, err := luaScripts.ReadFile("lua_scripts/" + filename)
		if err != nil {
			panic(fmt.Sprintf("CRITICAL: Failed to load Lua script %s: %v", filename, err))
		}
		return redis.NewScript(string(content))
	}

	return &InventoryCache{
		client:        client,
		deductScript:  load("deduct_stock.lua"),
		reserveScript: load("check_and_reserve.lua"),
	}
}

func (c *InventoryCache) DeductStock(ctx context.Context, productID uuid.UUID, userID string, quantity int, limit int) error {

	stockKey := fmt.Sprintf("product:%s:stock", productID.String())
	historyKey := fmt.Sprintf("product:%s:bought_history", productID.String())

	result, err := c.deductScript.Run(ctx, c.client, []string{stockKey, historyKey}, quantity, limit, userID).Int()

	if err != nil {
		fmt.Printf("Redis Script Error: %v\n", err)
		return ErrSystem
	}
	switch result {
	case 1:
		return nil
	case -1:
		return ErrOutOfStock
	case -2:
		return ErrLimitExceeded
	default:
		return fmt.Errorf("unknown lua result: %d", result)
	}
}

func (c *InventoryCache) SetInitialStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	key := fmt.Sprintf("product:%s:stock", productID.String())
	return c.client.Set(ctx, key, quantity, 0).Err()
}

// GetStock - Lấy số lượng tồn kho hiện tại để hiển thị lên UI
func (c *InventoryCache) GetStock(ctx context.Context, productID uuid.UUID) (int, error) {
	key := fmt.Sprintf("product:%s:stock", productID.String())
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

// CheckAndReserve - Giữ chỗ sản phẩm cho user
func (c *InventoryCache) CheckAndReserve(ctx context.Context, productID uuid.UUID, userID string, quantity int, ttl int) (string, error) {
	stockKey := fmt.Sprintf("product:%s:stock", productID.String())
	reservedKey := fmt.Sprintf("product:%s:reserved", productID.String())
	result, err := c.reserveScript.Run(ctx, c.client, []string{stockKey, reservedKey}, quantity, userID, ttl).Text()
	if err != nil {
		return "", err
	}

	// Nếu lua trả về "0" nghĩa là hết hàng (theo logic trong script của em)
	if result == "0" {
		return "", ErrOutOfStock
	}

	return result, nil
}


//rollback stock
func (c *InventoryCache) IncrStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	key := fmt.Sprintf("product:%s:stock", productID.String())
	//lệnh IncryBy 	hãy tìm cái key này ,và cộng thêm vào giá trị hiện tại 1 lượng là quantity
	return c.client.IncrBy(ctx, key, int64(quantity)).Err()
}