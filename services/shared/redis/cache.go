package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Cache provides Redis-backed caching with deduplication support.
type Cache struct {
	client *goredis.Client
}

// NewCache creates a new Cache instance wrapping a Redis client.
func NewCache(client *goredis.Client) *Cache {
	return &Cache{client: client}
}

// Get retrieves a cached value by key and unmarshals it into dest.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return fmt.Errorf("cache miss: %s", key)
	}
	if err != nil {
		return fmt.Errorf("cache get error: %w", err)
	}
	return json.Unmarshal(val, dest)
}

// Set stores a value in cache with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// InvalidatePattern deletes all keys matching a glob pattern.
func (c *Cache) InvalidatePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan error: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

// TryDedup attempts to acquire a dedup lock. Returns true if the operation
// should proceed (lock acquired), false if a duplicate is already in progress.
func (c *Cache) TryDedup(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := c.client.SetNX(ctx, "dedup:"+key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("dedup lock error: %w", err)
	}
	return ok, nil
}
