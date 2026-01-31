package cache

import (
	"context"
	"path"
	"sync"
	"time"
)

/* MemoryCache provides in-memory caching with TTL support */
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	ttl   time.Duration
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
	createdAt time.Time
}

/* NewMemoryCache creates a new in-memory cache */
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
		ttl:   defaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

/* Get retrieves a value from cache */
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.value, true
}

/* Set stores a value in cache with TTL */
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = c.ttl
	}

	c.items[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		createdAt: time.Now(),
	}

	return nil
}

/* Delete removes a value from cache */
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

/* DeleteByPattern removes all keys matching the glob pattern (e.g. "user:*") */
func (c *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		matched, err := path.Match(pattern, key)
		if err == nil && matched {
			delete(c.items, key)
		}
	}
	return nil
}

/* Clear removes all items from cache */
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	return nil
}

/* cleanup periodically removes expired items */
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

/* Stats returns cache statistics */
func (c *MemoryCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	now := time.Now()
	for _, item := range c.items {
		if now.After(item.expiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_items":   len(c.items),
		"expired_items": expired,
		"active_items":  len(c.items) - expired,
	}
}
