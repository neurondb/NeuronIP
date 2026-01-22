package cache

import "time"

/* CacheStrategy defines caching strategies */
type CacheStrategy string

const (
	StrategyTTL      CacheStrategy = "ttl"
	StrategyLRU      CacheStrategy = "lru"
	StrategyLFU      CacheStrategy = "lfu"
	StrategyFIFO    CacheStrategy = "fifo"
)

/* CacheConfig holds cache configuration */
type CacheConfig struct {
	Strategy        CacheStrategy
	DefaultTTL      time.Duration
	MaxSize         int
	CleanupInterval time.Duration
}

/* DefaultCacheConfig returns default cache configuration */
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Strategy:        StrategyLRU,
		DefaultTTL:      5 * time.Minute,
		MaxSize:         1000,
		CleanupInterval: 1 * time.Minute,
	}
}
