package cache

import (
	"context"
	"time"
)

/* CacheInterface defines the interface for cache implementations with context support */
type CacheInterface interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Clear(ctx context.Context) error
}

/* Adapter adapts the existing Cache struct to CacheInterface */
type Adapter struct {
	cache *Cache
}

/* NewAdapter creates a new cache adapter */
func NewAdapter(cache *Cache) *Adapter {
	return &Adapter{cache: cache}
}

/* Get retrieves a value from cache */
func (a *Adapter) Get(ctx context.Context, key string) (interface{}, bool) {
	return a.cache.Get(key)
}

/* Set stores a value in cache with TTL */
func (a *Adapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return a.cache.Set(key, value, ttl)
}

/* Delete removes a value from cache */
func (a *Adapter) Delete(ctx context.Context, key string) error {
	a.cache.Delete(key)
	return nil
}

/* DeleteByPattern removes all keys matching the glob pattern (e.g. "user:*") */
func (a *Adapter) DeleteByPattern(ctx context.Context, pattern string) error {
	return a.cache.DeleteByPattern(pattern)
}

/* Clear removes all items from cache */
func (a *Adapter) Clear(ctx context.Context) error {
	a.cache.Clear()
	return nil
}
