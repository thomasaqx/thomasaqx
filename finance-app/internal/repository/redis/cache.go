package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a Redis-backed cache implementation.
type Cache struct {
	client *redis.Client
}

func NewCache(addr, password string, db int) *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Cache{client: rdb}
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value: %w", err)
	}
	return c.client.Set(ctx, key, data, time.Duration(ttlSeconds)*time.Second).Err()
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Ping checks the Redis connection.
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close releases the Redis client.
func (c *Cache) Close() error {
	return c.client.Close()
}
