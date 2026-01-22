package errors

import (
	"context"
	"time"
)

/* RecoveryStrategy defines how to recover from an error */
type RecoveryStrategy interface {
	ShouldRetry(err *APIError) bool
	GetRetryDelay(attempt int) time.Duration
	GetMaxRetries() int
}

/* DefaultRecoveryStrategy provides default recovery behavior */
type DefaultRecoveryStrategy struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

/* NewDefaultRecoveryStrategy creates a new default recovery strategy */
func NewDefaultRecoveryStrategy() *DefaultRecoveryStrategy {
	return &DefaultRecoveryStrategy{
		MaxRetries:       3,
		InitialDelay:     1 * time.Second,
		MaxDelay:         30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

/* ShouldRetry determines if an error should be retried */
func (s *DefaultRecoveryStrategy) ShouldRetry(err *APIError) bool {
	if err == nil {
		return false
	}
	
	// Only retry transient errors
	return err.IsTransient()
}

/* GetRetryDelay calculates the delay before retrying */
func (s *DefaultRecoveryStrategy) GetRetryDelay(attempt int) time.Duration {
	delay := s.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * s.BackoffMultiplier)
		if delay > s.MaxDelay {
			delay = s.MaxDelay
			break
		}
	}
	return delay
}

/* GetMaxRetries returns the maximum number of retries */
func (s *DefaultRecoveryStrategy) GetMaxRetries() int {
	return s.MaxRetries
}

/* RetryWithRecovery executes a function with retry logic based on recovery strategy */
func RetryWithRecovery(ctx context.Context, strategy RecoveryStrategy, fn func() error) error {
	var lastErr error
	
	for attempt := 1; attempt <= strategy.GetMaxRetries(); attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error should be retried
		apiErr := AsAPIError(err)
		if apiErr == nil {
			// Not an API error, wrap it
			apiErr = Wrap(err, ErrCodeInternalServer, "Operation failed")
		}
		
		if !strategy.ShouldRetry(apiErr) {
			return err
		}
		
		// Don't retry on last attempt
		if attempt >= strategy.GetMaxRetries() {
			break
		}
		
		// Calculate delay
		delay := strategy.GetRetryDelay(attempt)
		
		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return lastErr
}

/* GetRecoveryStrategyForError returns an appropriate recovery strategy for an error */
func GetRecoveryStrategyForError(err *APIError) RecoveryStrategy {
	if err == nil {
		return NewDefaultRecoveryStrategy()
	}
	
	// Customize strategy based on error type
	switch err.Code {
	case ErrCodeTimeout:
		// Timeouts should retry quickly
		return &DefaultRecoveryStrategy{
			MaxRetries:       5,
			InitialDelay:     500 * time.Millisecond,
			MaxDelay:         5 * time.Second,
			BackoffMultiplier: 1.5,
		}
	case ErrCodeServiceUnavailable:
		// Service unavailable should retry with longer delays
		return &DefaultRecoveryStrategy{
			MaxRetries:       3,
			InitialDelay:     2 * time.Second,
			MaxDelay:         60 * time.Second,
			BackoffMultiplier: 2.0,
		}
	default:
		return NewDefaultRecoveryStrategy()
	}
}
