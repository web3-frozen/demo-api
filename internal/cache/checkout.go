package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CheckoutCache struct {
	client *redis.Client
}

func NewCheckoutCache(redisURL string) (*CheckoutCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &CheckoutCache{client: client}, nil
}

func (c *CheckoutCache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *CheckoutCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *CheckoutCache) Close() error {
	return c.client.Close()
}
