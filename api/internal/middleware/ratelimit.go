package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
)

/* RateLimiter provides in-memory rate limiting */
type RateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.RWMutex
	maxRequests int
	window      time.Duration
	cleanupInterval time.Duration
}

type clientLimiter struct {
	count      int
	windowStart time.Time
}

/* NewRateLimiter creates a new rate limiter */
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string]*clientLimiter),
		maxRequests: maxRequests,
		window:      window,
		cleanupInterval: window * 2,
	}

	// Start cleanup goroutine
	go rl.startCleanup()

	return rl
}

/* startCleanup removes old entries periodically */
func (rl *RateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, cl := range rl.clients {
			if now.Sub(cl.windowStart) > rl.cleanupInterval {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

/* Allow checks if a request is allowed for the given client */
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cl, exists := rl.clients[clientID]

	if !exists || now.Sub(cl.windowStart) > rl.window {
		// New client or window expired, start new window
		rl.clients[clientID] = &clientLimiter{
			count:      1,
			windowStart: now,
		}
		return true
	}

	// Check if limit exceeded
	if cl.count >= rl.maxRequests {
		return false
	}

	cl.count++
	return true
}

/* GetRemaining returns the number of remaining requests in the current window */
func (rl *RateLimiter) GetRemaining(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	cl, exists := rl.clients[clientID]
	if !exists {
		return rl.maxRequests
	}

	now := time.Now()
	if now.Sub(cl.windowStart) > rl.window {
		return rl.maxRequests
	}

	return rl.maxRequests - cl.count
}

/* RateLimit is a middleware that enforces rate limiting */
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (API key from context or IP address)
			clientID := getClientID(r)

			if !limiter.Allow(clientID) {
				remaining := limiter.GetRemaining(clientID)
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.maxRequests))
				w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.window).Unix()))
				handlers.WriteErrorResponse(w, errors.TooManyRequests("Rate limit exceeded"))
				return
			}

			// Add rate limit headers to response
			remaining := limiter.GetRemaining(clientID)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.maxRequests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.window).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

/* getClientID extracts client identifier from request */
func getClientID(r *http.Request) string {
	// Try to extract user ID from auth context if available
	if userID, ok := auth.GetUserIDFromContext(r.Context()); ok {
		return "user:" + userID
	}
	
	// Try to extract API key ID from context
	if apiKey, ok := auth.GetAPIKeyFromContext(r.Context()); ok && apiKey.ID.String() != "" {
		return "apikey:" + apiKey.ID.String()
	}
	
	// Fallback to IP address
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return "ip:" + ip
}
