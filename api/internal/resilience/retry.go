package resilience

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

/* RetryConfig holds retry configuration */
type RetryConfig struct {
	MaxAttempts      int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	Multiplier       float64
	Jitter           bool
	RetryableErrors  []error // Specific errors to retry
	NonRetryableErrors []error // Errors that should not be retried
}

/* DefaultRetryConfig returns default retry configuration */
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

/* ExponentialBackoffRetryConfig returns config with exponential backoff */
func ExponentialBackoffRetryConfig(maxAttempts int) *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

/* RetryBudget tracks retry budget to prevent cascading failures */
type RetryBudget struct {
	MaxRetriesPerMinute int
	retriesThisMinute   int
	windowStart         time.Time
	mu                  sync.Mutex
}

/* NewRetryBudget creates a new retry budget */
func NewRetryBudget(maxRetriesPerMinute int) *RetryBudget {
	return &RetryBudget{
		MaxRetriesPerMinute: maxRetriesPerMinute,
		windowStart:         time.Now(),
	}
}

/* CanRetry checks if a retry is allowed within the budget */
func (rb *RetryBudget) CanRetry() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	now := time.Now()
	// Reset window if a minute has passed
	if now.Sub(rb.windowStart) >= time.Minute {
		rb.retriesThisMinute = 0
		rb.windowStart = now
	}

	if rb.retriesThisMinute >= rb.MaxRetriesPerMinute {
		return false
	}

	rb.retriesThisMinute++
	return true
}

/* Retry executes a function with retry logic */
func Retry(ctx context.Context, config *RetryConfig, fn func() error) error {
	return RetryWithBudget(ctx, config, nil, fn)
}

/* RetryWithBudget executes a function with retry logic and budget */
func RetryWithBudget(ctx context.Context, config *RetryConfig, budget *RetryBudget, fn func() error) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check budget if provided
		if budget != nil && !budget.CanRetry() {
			if lastErr != nil {
				return fmt.Errorf("retry budget exhausted: %w", lastErr)
			}
			return fmt.Errorf("retry budget exhausted")
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err, config) {
			return err
		}

		// Don't retry on last attempt
		if attempt >= config.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay = calculateDelay(delay, attempt, config)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

/* calculateDelay calculates the delay for the next retry */
func calculateDelay(currentDelay time.Duration, attempt int, config *RetryConfig) time.Duration {
	// Exponential backoff
	delay := time.Duration(float64(currentDelay) * config.Multiplier)

	// Apply jitter if enabled
	if config.Jitter {
		// Add ±20% jitter
		jitter := time.Duration(float64(delay) * 0.2 * (rand.Float64()*2 - 1))
		delay = delay + jitter
	}

	// Cap at max delay
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}

	return delay
}

/* isRetryableError checks if an error should be retried */
func isRetryableError(err error, config *RetryConfig) bool {
	if err == nil {
		return false
	}

	// Check non-retryable errors first
	for _, nonRetryable := range config.NonRetryableErrors {
		if err == nonRetryable || fmt.Sprintf("%v", err) == fmt.Sprintf("%v", nonRetryable) {
			return false
		}
	}

	// Check retryable errors
	if len(config.RetryableErrors) > 0 {
		for _, retryable := range config.RetryableErrors {
			if err == retryable || fmt.Sprintf("%v", err) == fmt.Sprintf("%v", retryable) {
				return true
			}
		}
		// If specific retryable errors are defined and this isn't one, don't retry
		return false
	}

	// Default: retry all errors
	return true
}

/* RetryMetrics tracks retry statistics */
type RetryMetrics struct {
	TotalAttempts   int64
	TotalRetries    int64
	TotalFailures   int64
	TotalSuccesses  int64
	BudgetExhausted int64
	mu              sync.Mutex
}

/* NewRetryMetrics creates new retry metrics */
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{}
}

/* RecordAttempt records a retry attempt */
func (rm *RetryMetrics) RecordAttempt() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.TotalAttempts++
}

/* RecordRetry records a retry */
func (rm *RetryMetrics) RecordRetry() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.TotalRetries++
}

/* RecordFailure records a failure */
func (rm *RetryMetrics) RecordFailure() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.TotalFailures++
}

/* RecordSuccess records a success */
func (rm *RetryMetrics) RecordSuccess() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.TotalSuccesses++
}

/* RecordBudgetExhausted records budget exhaustion */
func (rm *RetryMetrics) RecordBudgetExhausted() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.BudgetExhausted++
}

/* GetStats returns current metrics */
func (rm *RetryMetrics) GetStats() map[string]int64 {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	return map[string]int64{
		"total_attempts":    rm.TotalAttempts,
		"total_retries":     rm.TotalRetries,
		"total_failures":    rm.TotalFailures,
		"total_successes":   rm.TotalSuccesses,
		"budget_exhausted":  rm.BudgetExhausted,
	}
}
