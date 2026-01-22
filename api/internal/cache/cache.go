package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/* Cache provides in-memory caching */
type Cache struct {
	items      map[string]*CacheItem
	mu         sync.RWMutex
	defaultTTL time.Duration
	maxSize    int
	cleanupInterval time.Duration
}

/* CacheItem represents a cached item */
type CacheItem struct {
	Value      interface{}
	ExpiresAt  time.Time
	CreatedAt  time.Time
	AccessCount int64
	LastAccessed time.Time
}

/* NewCache creates a new cache */
func NewCache(defaultTTL time.Duration, maxSize int) *Cache {
	c := &Cache{
		items:           make(map[string]*CacheItem),
		defaultTTL:      defaultTTL,
		maxSize:         maxSize,
		cleanupInterval: defaultTTL / 2,
	}

	// Start cleanup goroutine
	go c.startCleanup()

	return c
}

/* Get retrieves a value from cache */
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	// Update access statistics
	item.AccessCount++
	item.LastAccessed = time.Now()

	return item.Value, true
}

/* Set stores a value in cache */
func (c *Cache) Set(key string, value interface{}, ttl ...time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check size limit
	if len(c.items) >= c.maxSize && c.items[key] == nil {
		// Evict least recently used item
		c.evictLRU()
	}

	ttlDuration := c.defaultTTL
	if len(ttl) > 0 && ttl[0] > 0 {
		ttlDuration = ttl[0]
	}

	c.items[key] = &CacheItem{
		Value:       value,
		ExpiresAt:   time.Now().Add(ttlDuration),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccessed: time.Now(),
	}

	return nil
}

/* Delete removes a value from cache */
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

/* Clear clears all cache items */
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

/* evictLRU evicts the least recently used item */
func (c *Cache) evictLRU() {
	if len(c.items) == 0 {
		return
	}

	var lruKey string
	var lruTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.LastAccessed.Before(lruTime) {
			lruKey = key
			lruTime = item.LastAccessed
			first = false
		}
	}

	if lruKey != "" {
		delete(c.items, lruKey)
	}
}

/* startCleanup periodically removes expired items */
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

/* GetStats returns cache statistics */
func (c *Cache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalItems := len(c.items)
	expiredItems := 0
	totalAccesses := int64(0)

	now := time.Now()
	for _, item := range c.items {
		if now.After(item.ExpiresAt) {
			expiredItems++
		}
		totalAccesses += item.AccessCount
	}

	return map[string]interface{}{
		"total_items":    totalItems,
		"expired_items": expiredItems,
		"active_items":  totalItems - expiredItems,
		"max_size":      c.maxSize,
		"usage_percent": float64(totalItems) / float64(c.maxSize) * 100,
		"total_accesses": totalAccesses,
	}
}

/* InvalidationStrategy defines cache invalidation strategies */
type InvalidationStrategy interface {
	ShouldInvalidate(key string, value interface{}) bool
}

/* TTLInvalidationStrategy invalidates based on TTL */
type TTLInvalidationStrategy struct {
	TTL time.Duration
}

/* ShouldInvalidate implements InvalidationStrategy */
func (s *TTLInvalidationStrategy) ShouldInvalidate(key string, value interface{}) bool {
	// TTL is handled by the cache itself
	return false
}

/* PatternInvalidationStrategy invalidates keys matching a pattern */
type PatternInvalidationStrategy struct {
	Pattern string
}

/* ShouldInvalidate implements InvalidationStrategy */
func (s *PatternInvalidationStrategy) ShouldInvalidate(key string, value interface{}) bool {
	return key == s.Pattern || contains(key, s.Pattern)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

/* InvalidateByPattern invalidates all keys matching a pattern */
func (c *Cache) InvalidateByPattern(pattern string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key := range c.items {
		if contains(key, pattern) {
			delete(c.items, key)
			count++
		}
	}

	return count
}

/* WarmCache warms the cache with provided key-value pairs */
func (c *Cache) WarmCache(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := c.Set(key, value, ttl); err != nil {
				return fmt.Errorf("failed to warm cache for key %s: %w", key, err)
			}
		}
	}
	return nil
}
