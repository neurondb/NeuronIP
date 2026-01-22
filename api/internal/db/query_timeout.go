package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* QueryTimeoutConfig holds query timeout configuration */
type QueryTimeoutConfig struct {
	DefaultTimeout time.Duration
	MaxTimeout     time.Duration
	MinTimeout     time.Duration
	SlowQueryThreshold time.Duration
}

/* DefaultQueryTimeoutConfig returns default query timeout configuration */
func DefaultQueryTimeoutConfig() *QueryTimeoutConfig {
	return &QueryTimeoutConfig{
		DefaultTimeout:     30 * time.Second,
		MaxTimeout:         5 * time.Minute,
		MinTimeout:         1 * time.Second,
		SlowQueryThreshold: 1 * time.Second,
	}
}

/* QueryWithTimeout executes a query with a timeout */
func QueryWithTimeout(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration, query string, args ...interface{}) (pgx.Rows, error) {
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	rows, err := pool.Query(queryCtx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > DefaultQueryTimeoutConfig().SlowQueryThreshold {
		// In production, this would use proper logging
		fmt.Printf("Slow query detected: %v - %s\n", duration, query)
	}

	return rows, err
}

/* QueryRowWithTimeout executes a query row with a timeout */
func QueryRowWithTimeout(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration, query string, args ...interface{}) pgx.Row {
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	row := pool.QueryRow(queryCtx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > DefaultQueryTimeoutConfig().SlowQueryThreshold {
		fmt.Printf("Slow query detected: %v - %s\n", duration, query)
	}

	return row
}

/* ExecWithTimeout executes a command with a timeout */
func ExecWithTimeout(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration, query string, args ...interface{}) error {
	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	_, err := pool.Exec(queryCtx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > DefaultQueryTimeoutConfig().SlowQueryThreshold {
		fmt.Printf("Slow query detected: %v - %s\n", duration, query)
	}

	return err
}

/* QueryTimeoutManager manages query timeouts */
type QueryTimeoutManager struct {
	config *QueryTimeoutConfig
}

/* NewQueryTimeoutManager creates a new query timeout manager */
func NewQueryTimeoutManager(config *QueryTimeoutConfig) *QueryTimeoutManager {
	if config == nil {
		config = DefaultQueryTimeoutConfig()
	}
	return &QueryTimeoutManager{config: config}
}

/* WithTimeout creates a context with appropriate timeout */
func (qtm *QueryTimeoutManager) WithTimeout(ctx context.Context, customTimeout ...time.Duration) (context.Context, context.CancelFunc) {
	timeout := qtm.config.DefaultTimeout
	if len(customTimeout) > 0 && customTimeout[0] > 0 {
		timeout = customTimeout[0]
		// Enforce min/max bounds
		if timeout < qtm.config.MinTimeout {
			timeout = qtm.config.MinTimeout
		}
		if timeout > qtm.config.MaxTimeout {
			timeout = qtm.config.MaxTimeout
		}
	}
	return context.WithTimeout(ctx, timeout)
}

/* GetDefaultTimeout returns the default timeout */
func (qtm *QueryTimeoutManager) GetDefaultTimeout() time.Duration {
	return qtm.config.DefaultTimeout
}

/* GetSlowQueryThreshold returns the slow query threshold */
func (qtm *QueryTimeoutManager) GetSlowQueryThreshold() time.Duration {
	return qtm.config.SlowQueryThreshold
}

/* SlowQueryLogger logs slow queries */
type SlowQueryLogger interface {
	LogSlowQuery(query string, duration time.Duration, args []interface{})
}

/* DefaultSlowQueryLogger is a simple slow query logger */
type DefaultSlowQueryLogger struct {
	threshold time.Duration
}

/* NewDefaultSlowQueryLogger creates a new slow query logger */
func NewDefaultSlowQueryLogger(threshold time.Duration) *DefaultSlowQueryLogger {
	return &DefaultSlowQueryLogger{threshold: threshold}
}

/* LogSlowQuery logs a slow query */
func (l *DefaultSlowQueryLogger) LogSlowQuery(query string, duration time.Duration, args []interface{}) {
	if duration > l.threshold {
		fmt.Printf("[SLOW QUERY] Duration: %v, Query: %s, Args: %v\n", duration, query, args)
	}
}
