package startup

import (
	"context"
	"fmt"
	"time"
)

/* Check represents a startup health check */
type Check interface {
	Name() string
	Execute(ctx context.Context) error
}

/* CheckResult represents the result of a check */
type CheckResult struct {
	Name      string
	Status    CheckStatus
	Message   string
	Duration  time.Duration
	Error     error
}

/* CheckStatus represents the status of a check */
type CheckStatus string

const (
	CheckStatusPass CheckStatus = "pass"
	CheckStatusFail CheckStatus = "fail"
	CheckStatusWarn CheckStatus = "warn"
)

/* Checker performs startup health checks */
type Checker struct {
	checks []Check
	timeout time.Duration
}

/* NewChecker creates a new startup checker */
func NewChecker(timeout time.Duration) *Checker {
	return &Checker{
		checks:  make([]Check, 0),
		timeout: timeout,
	}
}

/* AddCheck adds a check to the checker */
func (c *Checker) AddCheck(check Check) {
	c.checks = append(c.checks, check)
}

/* RunChecks runs all checks */
func (c *Checker) RunChecks(ctx context.Context) ([]CheckResult, error) {
	checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	results := make([]CheckResult, 0, len(c.checks))

	for _, check := range c.checks {
		start := time.Now()
		err := check.Execute(checkCtx)
		duration := time.Since(start)

		status := CheckStatusPass
		message := "Check passed"
		if err != nil {
			status = CheckStatusFail
			message = err.Error()
		}

		results = append(results, CheckResult{
			Name:     check.Name(),
			Status:   status,
			Message:  message,
			Duration: duration,
			Error:    err,
		})
	}

	// Check if any checks failed
	for _, result := range results {
		if result.Status == CheckStatusFail {
			return results, fmt.Errorf("startup check failed: %s", result.Name)
		}
	}

	return results, nil
}

/* DatabaseCheck checks database connectivity */
type DatabaseCheck struct {
	pingFunc func(ctx context.Context) error
}

/* NewDatabaseCheck creates a new database check */
func NewDatabaseCheck(pingFunc func(ctx context.Context) error) *DatabaseCheck {
	return &DatabaseCheck{pingFunc: pingFunc}
}

/* Name implements Check */
func (c *DatabaseCheck) Name() string {
	return "database"
}

/* Execute implements Check */
func (c *DatabaseCheck) Execute(ctx context.Context) error {
	if c.pingFunc == nil {
		return fmt.Errorf("ping function not set")
	}
	return c.pingFunc(ctx)
}

/* ReadinessCheck checks if the service is ready */
type ReadinessCheck struct {
	readyFunc func(ctx context.Context) error
}

/* NewReadinessCheck creates a new readiness check */
func NewReadinessCheck(readyFunc func(ctx context.Context) error) *ReadinessCheck {
	return &ReadinessCheck{readyFunc: readyFunc}
}

/* Name implements Check */
func (c *ReadinessCheck) Name() string {
	return "readiness"
}

/* Execute implements Check */
func (c *ReadinessCheck) Execute(ctx context.Context) error {
	if c.readyFunc == nil {
		return fmt.Errorf("ready function not set")
	}
	return c.readyFunc(ctx)
}
