package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Hidas2004/go-flashsale-system/internal/domain"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type productCache struct {
	client *redis.Client
}

func NewProductCache(client *redis.Client) domain.ProductCache {
	return &productCache{client: client}
}

const flashSaleKey = "products:flash_sale"

// (Đọc Cache)
func (c *productCache) GetFlashSaleProducts(ctx context.Context) ([]*models.Product, error) {
	val, err := c.client.Get(ctx, flashSaleKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var products []*models.Product
	if err := json.Unmarshal([]byte(val), &products); err != nil {
		return nil, err
	}
	return products, nil
}

// ghi cache
func (c *productCache) SetFlashSaleProducts(ctx context.Context, products []*models.Product) error {
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}
	//cache 5 phút
	return c.client.Set(ctx, flashSaleKey, data, 5*time.Minute).Err()

}

// (Xóa Cache)
func (c *productCache) InvalidateFlashSaleProducts(ctx context.Context) error {
	return c.client.Del(ctx, flashSaleKey).Err()
}

func (c *productCache) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	key := "product:" + id.String()
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var product models.Product
	if err := json.Unmarshal([]byte(val), &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (c *productCache) SetProduct(ctx context.Context, product *models.Product) error {
	key := "product:" + product.ID.String()
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	// Cache 10 mins
	return c.client.Set(ctx, key, data, 10*time.Minute).Err()
}
