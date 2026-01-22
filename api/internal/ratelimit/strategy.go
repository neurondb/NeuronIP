package ratelimit

import (
	"sync"
	"time"
)

/* Strategy defines a rate limiting strategy */
type Strategy interface {
	Allow(key string) bool
	GetRemaining(key string) int
	GetResetTime(key string) time.Time
}

/* SlidingWindowStrategy implements sliding window rate limiting */
type SlidingWindowStrategy struct {
	requests map[string][]time.Time
	maxRequests int
	window      time.Duration
	mu          sync.RWMutex
}

/* NewSlidingWindowStrategy creates a new sliding window strategy */
func NewSlidingWindowStrategy(maxRequests int, window time.Duration) *SlidingWindowStrategy {
	return &SlidingWindowStrategy{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}
}

/* Allow checks if a request is allowed */
func (s *SlidingWindowStrategy) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-s.window)

	// Get or create request list for this key
	requests, exists := s.requests[key]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// Remove requests outside the window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= s.maxRequests {
		s.requests[key] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	s.requests[key] = validRequests
	return true
}

/* GetRemaining returns remaining requests */
func (s *SlidingWindowStrategy) GetRemaining(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-s.window)

	requests, exists := s.requests[key]
	if !exists {
		return s.maxRequests
	}

	// Count valid requests
	validCount := 0
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validCount++
		}
	}

	return s.maxRequests - validCount
}

/* GetResetTime returns the reset time */
func (s *SlidingWindowStrategy) GetResetTime(key string) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests, exists := s.requests[key]
	if !exists || len(requests) == 0 {
		return time.Now().Add(s.window)
	}

	// Find oldest request in window
	now := time.Now()
	windowStart := now.Add(-s.window)
	oldest := now

	for _, reqTime := range requests {
		if reqTime.After(windowStart) && reqTime.Before(oldest) {
			oldest = reqTime
		}
	}

	return oldest.Add(s.window)
}

/* TokenBucketStrategy implements token bucket rate limiting */
type TokenBucketStrategy struct {
	buckets map[string]*TokenBucket
	maxTokens int
	refillRate time.Duration
	mu         sync.RWMutex
}

/* TokenBucket represents a token bucket */
type TokenBucket struct {
	tokens     int
	lastRefill time.Time
}

/* NewTokenBucketStrategy creates a new token bucket strategy */
func NewTokenBucketStrategy(maxTokens int, refillRate time.Duration) *TokenBucketStrategy {
	return &TokenBucketStrategy{
		buckets:    make(map[string]*TokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

/* Allow checks if a request is allowed */
func (s *TokenBucketStrategy) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	bucket, exists := s.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			tokens:     s.maxTokens,
			lastRefill: time.Now(),
		}
		s.buckets[key] = bucket
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= s.refillRate {
		tokensToAdd := int(elapsed / s.refillRate)
		bucket.tokens = min(bucket.tokens+tokensToAdd, s.maxTokens)
		bucket.lastRefill = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

/* GetRemaining returns remaining tokens */
func (s *TokenBucketStrategy) GetRemaining(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, exists := s.buckets[key]
	if !exists {
		return s.maxTokens
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= s.refillRate {
		tokensToAdd := int(elapsed / s.refillRate)
		bucket.tokens = min(bucket.tokens+tokensToAdd, s.maxTokens)
	}

	return bucket.tokens
}

/* GetResetTime returns the reset time */
func (s *TokenBucketStrategy) GetResetTime(key string) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bucket, exists := s.buckets[key]
	if !exists {
		return time.Now().Add(s.refillRate)
	}

	// Calculate when bucket will be full
	if bucket.tokens >= s.maxTokens {
		return time.Now()
	}

	tokensNeeded := s.maxTokens - bucket.tokens
	return bucket.lastRefill.Add(time.Duration(tokensNeeded) * s.refillRate)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
