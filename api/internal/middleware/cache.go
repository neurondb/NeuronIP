package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/cache"
)

/* CacheConfig configures caching behavior */
type CacheConfig struct {
	Cache          cache.CacheInterface
	DefaultTTL     time.Duration
	CacheablePaths []string // Paths that should be cached
	SkipPaths      []string // Paths that should not be cached
	VaryHeaders    []string // Headers that affect cache key
}

/* DefaultCacheConfig returns default cache configuration */
func DefaultCacheConfig(cacheService cache.CacheInterface) CacheConfig {
	return CacheConfig{
		Cache:          cacheService,
		DefaultTTL:     5 * time.Minute,
		CacheablePaths: []string{"/api/v1/catalog", "/api/v1/metrics", "/api/v1/schemas"},
		SkipPaths:      []string{"/api/v1/auth", "/api/v1/warehouse/query"},
		VaryHeaders:    []string{"Authorization"},
	}
}

/* CacheMiddleware provides HTTP response caching */
func CacheMiddleware(config CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip caching for non-GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Check if path should be skipped
			for _, skipPath := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, skipPath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check if path should be cached
			shouldCache := false
			for _, cacheablePath := range config.CacheablePaths {
				if strings.HasPrefix(r.URL.Path, cacheablePath) {
					shouldCache = true
					break
				}
			}

			if !shouldCache {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := generateCacheKey(r, config.VaryHeaders)

			// Try to get from cache
			if val, found := config.Cache.Get(r.Context(), cacheKey); found {
				if cachedResponse, ok := val.(CachedResponse); ok {
					// Set cache headers
					w.Header().Set("X-Cache", "HIT")
					w.Header().Set("Content-Type", cachedResponse.ContentType)

					// Set ETag if available
					if cachedResponse.ETag != "" {
						w.Header().Set("ETag", cachedResponse.ETag)
					}

					// Check If-None-Match header
					if match := r.Header.Get("If-None-Match"); match != "" && match == cachedResponse.ETag {
						w.WriteHeader(http.StatusNotModified)
						return
					}

					w.WriteHeader(cachedResponse.StatusCode)
					w.Write(cachedResponse.Body)
					return
				}
			}

			// Cache miss - use response writer that captures response
			rw := &cacheResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           []byte{},
			}

			next.ServeHTTP(rw, r)

			// Only cache successful responses
			if rw.statusCode >= 200 && rw.statusCode < 300 {
				// Generate ETag
				etag := generateETag(rw.body)

				// Store in cache
				cachedResponse := CachedResponse{
					StatusCode:  rw.statusCode,
					Body:        rw.body,
					ContentType: rw.Header().Get("Content-Type"),
					ETag:        etag,
					CachedAt:    time.Now(),
				}

				config.Cache.Set(r.Context(), cacheKey, cachedResponse, config.DefaultTTL)

				// Set cache headers
				w.Header().Set("X-Cache", "MISS")
				w.Header().Set("ETag", etag)
			}
		})
	}
}

/* CachedResponse represents a cached HTTP response */
type CachedResponse struct {
	StatusCode  int       `json:"status_code"`
	Body        []byte    `json:"body"`
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag"`
	CachedAt    time.Time `json:"cached_at"`
}

/* cacheResponseWriter captures response for caching */
type cacheResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *cacheResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *cacheResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

/* generateCacheKey generates a cache key from request */
func generateCacheKey(r *http.Request, varyHeaders []string) string {
	key := r.Method + ":" + r.URL.Path + "?" + r.URL.RawQuery

	// Include varying headers in cache key
	for _, header := range varyHeaders {
		if val := r.Header.Get(header); val != "" {
			key += ":" + header + "=" + val
		}
	}

	// Hash the key to keep it reasonable length
	hash := sha256.Sum256([]byte(key))
	return "cache:" + hex.EncodeToString(hash[:])
}

/* generateETag generates an ETag from response body */
func generateETag(body []byte) string {
	hash := sha256.Sum256(body)
	return `"` + hex.EncodeToString(hash[:16]) + `"`
}

/* InvalidateCache invalidates cache for a given pattern */
func InvalidateCache(cacheService cache.CacheInterface, pattern string) error {
	if cacheService == nil || pattern == "" {
		return nil
	}
	return cacheService.DeleteByPattern(context.Background(), pattern)
}
