package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/web3-frozen/demo-api/internal/model"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(redisURL string) (*RedisCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisCache{client: client, ttl: 5 * time.Minute}, nil
}

func (c *RedisCache) Get(ctx context.Context, id string) (*model.Task, error) {
	data, err := c.client.Get(ctx, taskKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var task model.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *RedisCache) Set(ctx context.Context, task *model.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, taskKey(task.ID), data, c.ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, taskKey(id)).Err()
}

func (c *RedisCache) InvalidateList(ctx context.Context) error {
	return c.client.Del(ctx, "tasks:list").Err()
}

func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func taskKey(id string) string {
	return "task:" + id
}
