package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/* State represents the circuit breaker state */
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

/* String returns the string representation of the state */
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

/* Config holds circuit breaker configuration */
type Config struct {
	FailureThreshold    int           // Number of failures before opening
	SuccessThreshold    int           // Number of successes in half-open before closing
	Timeout             time.Duration // Time to wait before attempting half-open
	MaxRequests         int           // Max requests in half-open state
	ResetInterval       time.Duration // Interval to reset failure count
}

/* DefaultConfig returns default circuit breaker configuration */
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          60 * time.Second,
		MaxRequests:      3,
		ResetInterval:    60 * time.Second,
	}
}

/* CircuitBreaker implements the circuit breaker pattern */
type CircuitBreaker struct {
	config          *Config
	state           State
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	halfOpenRequests int
	mu              sync.RWMutex
	onStateChange   func(from, to State)
}

/* NewCircuitBreaker creates a new circuit breaker */
func NewCircuitBreaker(config *Config) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}
	return &CircuitBreaker{
		config:        config,
		state:         StateClosed,
		lastFailureTime: time.Now(),
	}
}

/* OnStateChange sets a callback for state changes */
func (cb *CircuitBreaker) OnStateChange(fn func(from, to State)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

/* Execute executes a function with circuit breaker protection */
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we should allow the request
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker is open")
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.recordResult(err)

	return err
}

/* allowRequest checks if a request should be allowed */
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.config.Timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.config.Timeout {
				cb.transitionTo(StateHalfOpen)
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		// Allow limited requests in half-open state
		return cb.halfOpenRequests < cb.config.MaxRequests
	default:
		return false
	}
}

/* recordResult records the result of an operation */
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

/* recordFailure records a failure */
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen)
		}
	case StateHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.transitionTo(StateOpen)
		cb.halfOpenRequests = 0
	}
}

/* recordSuccess records a success */
func (cb *CircuitBreaker) recordSuccess() {
	cb.successCount++

	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case StateHalfOpen:
		cb.halfOpenRequests++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
			cb.failureCount = 0
			cb.successCount = 0
			cb.halfOpenRequests = 0
		}
	}
}

/* transitionTo transitions to a new state */
func (cb *CircuitBreaker) transitionTo(newState State) {
	oldState := cb.state
	cb.state = newState

	if cb.onStateChange != nil {
		cb.onStateChange(oldState, newState)
	}
}

/* GetState returns the current state */
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

/* GetStats returns circuit breaker statistics */
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":            cb.state.String(),
		"failure_count":    cb.failureCount,
		"success_count":    cb.successCount,
		"last_failure":     cb.lastFailureTime,
		"half_open_requests": cb.halfOpenRequests,
	}
}

/* Reset resets the circuit breaker to closed state */
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenRequests = 0
	cb.lastFailureTime = time.Now()

	if cb.onStateChange != nil {
		cb.onStateChange(oldState, StateClosed)
	}
}
