package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

/* RedisCache provides Redis-based caching with fallback to memory */
type RedisCache struct {
	client      *redis.Client
	memoryCache *MemoryCache
	enabled     bool
}

/* NewRedisCache creates a new Redis cache with memory fallback */
func NewRedisCache(redisURL string, defaultTTL time.Duration) (*RedisCache, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// If Redis URL is invalid, use memory cache only
		return &RedisCache{
			client:      nil,
			memoryCache: NewMemoryCache(defaultTTL),
			enabled:     false,
		}, nil
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// Redis not available, use memory cache only
		client.Close()
		return &RedisCache{
			client:      nil,
			memoryCache: NewMemoryCache(defaultTTL),
			enabled:     false,
		}, nil
	}

	return &RedisCache{
		client:      client,
		memoryCache: NewMemoryCache(defaultTTL),
		enabled:     true,
	}, nil
}

/* Get retrieves a value from cache (Redis first, then memory) */
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	// Try Redis first if enabled
	if c.enabled && c.client != nil {
		val, err := c.client.Get(ctx, key).Result()
		if err == nil {
			// Deserialize JSON
			var result interface{}
			if err := json.Unmarshal([]byte(val), &result); err == nil {
				return result, true
			}
		}
	}

	// Fallback to memory cache
	return c.memoryCache.Get(ctx, key)
}

/* Set stores a value in cache (both Redis and memory) */
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	// Store in Redis if enabled
	if c.enabled && c.client != nil {
		if ttl <= 0 {
			ttl = 5 * time.Minute // Default TTL
		}
		if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
			// If Redis fails, continue to memory cache
		}
	}

	// Always store in memory cache as fallback
	return c.memoryCache.Set(ctx, key, value, ttl)
}

/* Delete removes a value from cache */
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if c.enabled && c.client != nil {
		c.client.Del(ctx, key)
	}
	return c.memoryCache.Delete(ctx, key)
}

/* DeleteByPattern removes all keys matching the glob pattern (e.g. "user:*") */
func (c *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c.enabled && c.client != nil {
		iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			c.client.Del(ctx, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}
	}
	return c.memoryCache.DeleteByPattern(ctx, pattern)
}

/* Clear removes all items from cache */
func (c *RedisCache) Clear(ctx context.Context) error {
	if c.enabled && c.client != nil {
		c.client.FlushDB(ctx)
	}
	return c.memoryCache.Clear(ctx)
}

/* Close closes the Redis connection */
func (c *RedisCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

/* Stats returns cache statistics */
func (c *RedisCache) Stats() map[string]interface{} {
	stats := c.memoryCache.Stats()
	stats["redis_enabled"] = c.enabled

	if c.enabled && c.client != nil {
		ctx := context.Background()
		dbSize, _ := c.client.DBSize(ctx).Result()
		stats["redis_db_size"] = dbSize
	}

	return stats
}
