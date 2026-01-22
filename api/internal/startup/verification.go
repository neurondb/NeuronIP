package startup

import (
	"context"
	"fmt"
	"time"
)

/* DependencyVerifier verifies dependencies are available */
type DependencyVerifier struct {
	dependencies map[string]Dependency
	timeout      time.Duration
}

/* Dependency represents a dependency to verify */
type Dependency struct {
	Name     string
	Verify   func(ctx context.Context) error
	Required bool
	Retries  int
	Delay    time.Duration
}

/* NewDependencyVerifier creates a new dependency verifier */
func NewDependencyVerifier(timeout time.Duration) *DependencyVerifier {
	return &DependencyVerifier{
		dependencies: make(map[string]Dependency),
		timeout:      timeout,
	}
}

/* AddDependency adds a dependency to verify */
func (dv *DependencyVerifier) AddDependency(dep Dependency) {
	dv.dependencies[dep.Name] = dep
}

/* VerifyAll verifies all dependencies */
func (dv *DependencyVerifier) VerifyAll(ctx context.Context) error {
	verifyCtx, cancel := context.WithTimeout(ctx, dv.timeout)
	defer cancel()

	var errors []error

	for name, dep := range dv.dependencies {
		err := dv.verifyDependency(verifyCtx, dep)
		if err != nil {
			if dep.Required {
				errors = append(errors, fmt.Errorf("required dependency %s failed: %w", name, err))
			} else {
				// Log warning for optional dependencies
				fmt.Printf("Warning: optional dependency %s failed: %v\n", name, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("dependency verification failed: %v", errors)
	}

	return nil
}

/* verifyDependency verifies a single dependency with retries */
func (dv *DependencyVerifier) verifyDependency(ctx context.Context, dep Dependency) error {
	var lastErr error

	for attempt := 0; attempt <= dep.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dep.Delay):
				// Continue to next attempt
			}
		}

		err := dep.Verify(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return lastErr
}

/* StartupFailureRecovery handles startup failure recovery */
type StartupFailureRecovery struct {
	maxRetries int
	retryDelay time.Duration
}

/* NewStartupFailureRecovery creates a new startup failure recovery */
func NewStartupFailureRecovery(maxRetries int, retryDelay time.Duration) *StartupFailureRecovery {
	return &StartupFailureRecovery{
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

/* Recover attempts to recover from startup failure */
func (sfr *StartupFailureRecovery) Recover(ctx context.Context, startupFunc func(ctx context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= sfr.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sfr.retryDelay):
				// Continue to next attempt
			}
		}

		err := startupFunc(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("startup failed after %d attempts: %w", sfr.maxRetries+1, lastErr)
}
