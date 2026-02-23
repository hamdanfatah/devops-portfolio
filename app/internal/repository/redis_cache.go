package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hamfa/task-manager/internal/model"
)

const (
	taskCachePrefix = "task:"
	taskCacheTTL    = 5 * time.Minute
)

// RedisCache handles Redis caching operations
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// GetTask retrieves a cached task
func (c *RedisCache) GetTask(ctx context.Context, id string) (*model.Task, error) {
	key := taskCachePrefix + id

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached task: %w", err)
	}

	var task model.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached task: %w", err)
	}

	return &task, nil
}

// SetTask caches a task
func (c *RedisCache) SetTask(ctx context.Context, task *model.Task) error {
	key := taskCachePrefix + task.ID

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return c.client.Set(ctx, key, data, taskCacheTTL).Err()
}

// InvalidateTask removes a task from cache
func (c *RedisCache) InvalidateTask(ctx context.Context, id string) error {
	key := taskCachePrefix + id
	return c.client.Del(ctx, key).Err()
}

// InvalidateAll clears all task caches
func (c *RedisCache) InvalidateAll(ctx context.Context) error {
	iter := c.client.Scan(ctx, 0, taskCachePrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Ping checks the Redis connection
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
