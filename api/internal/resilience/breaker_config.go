package resilience

import "time"

/* BreakerConfig holds configuration for circuit breakers */
type BreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening
	FailureThreshold int
	// SuccessThreshold is the number of consecutive successes in half-open before closing
	SuccessThreshold int
	// Timeout is the duration to wait before attempting half-open
	Timeout time.Duration
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests int
	// ResetInterval is the interval to reset failure count
	ResetInterval time.Duration
}

/* DefaultBreakerConfig returns default breaker configuration */
func DefaultBreakerConfig() *BreakerConfig {
	return &BreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          60 * time.Second,
		MaxRequests:      3,
		ResetInterval:    60 * time.Second,
	}
}

/* ForNeuronAgent returns configuration optimized for NeuronAgent */
func ForNeuronAgent() *BreakerConfig {
	return &BreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		MaxRequests:      2,
		ResetInterval:    30 * time.Second,
	}
}

/* ForNeuronMCP returns configuration optimized for NeuronMCP */
func ForNeuronMCP() *BreakerConfig {
	return &BreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          45 * time.Second,
		MaxRequests:      3,
		ResetInterval:    45 * time.Second,
	}
}

/* ToConfig converts BreakerConfig to Config */
func (bc *BreakerConfig) ToConfig() *Config {
	return &Config{
		FailureThreshold: bc.FailureThreshold,
		SuccessThreshold: bc.SuccessThreshold,
		Timeout:          bc.Timeout,
		MaxRequests:      bc.MaxRequests,
		ResetInterval:    bc.ResetInterval,
	}
}
