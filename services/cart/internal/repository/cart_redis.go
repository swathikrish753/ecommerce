package repository

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/swathikrish753/ecommerce/services/cart/internal/domain"
)

// CartRedis stores each cart as a Redis hash: key cart:{userID}, field = SKU, value = quantity.
// Every write refreshes the key's TTL, so abandoned carts expire on their own.
type CartRedis struct {
	rdb *goredis.Client
	ttl time.Duration
}

func NewCartRedis(rdb *goredis.Client, ttl time.Duration) *CartRedis {
	return &CartRedis{rdb: rdb, ttl: ttl}
}

func key(userID string) string { return "cart:" + userID }

func (r *CartRedis) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	m, err := r.rdb.HGetAll(ctx, key(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall: %w", err)
	}
	cart := &domain.Cart{UserID: userID, Items: make([]*domain.CartItem, 0, len(m))}
	for sku, v := range m {
		qty, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse qty for %s: %w", sku, err)
		}
		cart.Items = append(cart.Items, &domain.CartItem{SKU: sku, Quantity: qty})
	}
	// Hash order is random; sort for stable responses.
	sort.Slice(cart.Items, func(i, j int) bool { return cart.Items[i].SKU < cart.Items[j].SKU })
	return cart, nil
}

func (r *CartRedis) AddItem(ctx context.Context, userID, sku string, delta int64) (int64, error) {
	k := key(userID)
	pipe := r.rdb.TxPipeline()
	incr := pipe.HIncrBy(ctx, k, sku, delta)
	pipe.Expire(ctx, k, r.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("hincrby: %w", err)
	}
	return incr.Val(), nil
}

func (r *CartRedis) SetQuantity(ctx context.Context, userID, sku string, qty int64) error {
	k := key(userID)
	exists, err := r.rdb.HExists(ctx, k, sku).Result()
	if err != nil {
		return fmt.Errorf("hexists: %w", err)
	}
	if !exists {
		return domain.ErrItemNotFound
	}
	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, k, sku, qty)
	pipe.Expire(ctx, k, r.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("hset: %w", err)
	}
	return nil
}

func (r *CartRedis) RemoveItem(ctx context.Context, userID, sku string) error {
	n, err := r.rdb.HDel(ctx, key(userID), sku).Result()
	if err != nil {
		return fmt.Errorf("hdel: %w", err)
	}
	if n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

func (r *CartRedis) Clear(ctx context.Context, userID string) error {
	if err := r.rdb.Del(ctx, key(userID)).Err(); err != nil {
		return fmt.Errorf("del: %w", err)
	}
	return nil
}
