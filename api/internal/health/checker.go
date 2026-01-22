package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/* HealthStatus represents health status */
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

/* HealthCheck represents a health check */
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
}

/* HealthResult represents the result of a health check */
type HealthResult struct {
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Duration  time.Duration
	Details   map[string]interface{}
}

/* Checker performs health checks */
type Checker struct {
	checks []HealthCheck
	mu     sync.RWMutex
}

/* NewChecker creates a new health checker */
func NewChecker() *Checker {
	return &Checker{
		checks: make([]HealthCheck, 0),
	}
}

/* RegisterCheck registers a health check */
func (c *Checker) RegisterCheck(check HealthCheck) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks = append(c.checks, check)
}

/* CheckAll performs all registered health checks */
func (c *Checker) CheckAll(ctx context.Context) map[string]HealthResult {
	c.mu.RLock()
	checks := make([]HealthCheck, len(c.checks))
	copy(checks, c.checks)
	c.mu.RUnlock()

	results := make(map[string]HealthResult)

	for _, check := range checks {
		start := time.Now()
		result := check.Check(ctx)
		result.Duration = time.Since(start)
		result.Timestamp = time.Now()
		results[check.Name()] = result
	}

	return results
}

/* AggregateStatus aggregates health check results into overall status */
func (c *Checker) AggregateStatus(results map[string]HealthResult) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

/* DependencyCheck checks a dependency */
type DependencyCheck struct {
	name     string
	checkFunc func(ctx context.Context) error
	required bool
}

/* NewDependencyCheck creates a new dependency check */
func NewDependencyCheck(name string, checkFunc func(ctx context.Context) error, required bool) *DependencyCheck {
	return &DependencyCheck{
		name:      name,
		checkFunc: checkFunc,
		required:  required,
	}
}

/* Name implements HealthCheck */
func (c *DependencyCheck) Name() string {
	return c.name
}

/* Check implements HealthCheck */
func (c *DependencyCheck) Check(ctx context.Context) HealthResult {
	err := c.checkFunc(ctx)
	if err != nil {
		if c.required {
			return HealthResult{
				Status:  StatusUnhealthy,
				Message: fmt.Sprintf("Required dependency %s is unhealthy: %v", c.name, err),
			}
		}
		return HealthResult{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("Optional dependency %s is unhealthy: %v", c.name, err),
		}
	}

	return HealthResult{
		Status:  StatusHealthy,
		Message: fmt.Sprintf("Dependency %s is healthy", c.name),
	}
}

/* ReadinessCheck checks if the service is ready */
type ReadinessCheck struct {
	readyFunc func(ctx context.Context) bool
}

/* NewReadinessCheck creates a new readiness check */
func NewReadinessCheck(readyFunc func(ctx context.Context) bool) *ReadinessCheck {
	return &ReadinessCheck{readyFunc: readyFunc}
}

/* Name implements HealthCheck */
func (c *ReadinessCheck) Name() string {
	return "readiness"
}

/* Check implements HealthCheck */
func (c *ReadinessCheck) Check(ctx context.Context) HealthResult {
	if c.readyFunc == nil {
		return HealthResult{
			Status:  StatusUnhealthy,
			Message: "Readiness function not set",
		}
	}

	if c.readyFunc(ctx) {
		return HealthResult{
			Status:  StatusHealthy,
			Message: "Service is ready",
		}
	}

	return HealthResult{
		Status:  StatusUnhealthy,
		Message: "Service is not ready",
	}
}

/* LivenessCheck checks if the service is alive */
type LivenessCheck struct {
	aliveFunc func(ctx context.Context) bool
}

/* NewLivenessCheck creates a new liveness check */
func NewLivenessCheck(aliveFunc func(ctx context.Context) bool) *LivenessCheck {
	return &LivenessCheck{aliveFunc: aliveFunc}
}

/* Name implements HealthCheck */
func (c *LivenessCheck) Name() string {
	return "liveness"
}

/* Check implements HealthCheck */
func (c *LivenessCheck) Check(ctx context.Context) HealthResult {
	if c.aliveFunc == nil {
		return HealthResult{
			Status:  StatusUnhealthy,
			Message: "Liveness function not set",
		}
	}

	if c.aliveFunc(ctx) {
		return HealthResult{
			Status:  StatusHealthy,
			Message: "Service is alive",
		}
	}

	return HealthResult{
		Status:  StatusUnhealthy,
		Message: "Service is not alive",
	}
}
