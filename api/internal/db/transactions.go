package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TransactionOptions holds options for transaction execution */
type TransactionOptions struct {
	Timeout      time.Duration
	Isolation    pgx.TxIsoLevel
	ReadOnly     bool
	Deferrable   bool
	RetryOnError bool
	MaxRetries   int
}

/* DefaultTransactionOptions returns default transaction options */
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		Timeout:      30 * time.Second,
		Isolation:    pgx.ReadCommitted,
		ReadOnly:     false,
		Deferrable:   false,
		RetryOnError: false,
		MaxRetries:   0,
	}
}

/* WithTimeout sets the transaction timeout */
func (o *TransactionOptions) WithTimeout(timeout time.Duration) *TransactionOptions {
	o.Timeout = timeout
	return o
}

/* WithIsolation sets the transaction isolation level */
func (o *TransactionOptions) WithIsolation(level pgx.TxIsoLevel) *TransactionOptions {
	o.Isolation = level
	return o
}

/* WithReadOnly sets the transaction as read-only */
func (o *TransactionOptions) WithReadOnly(readOnly bool) *TransactionOptions {
	o.ReadOnly = readOnly
	return o
}

/* WithRetry enables retry on error */
func (o *TransactionOptions) WithRetry(maxRetries int) *TransactionOptions {
	o.RetryOnError = true
	o.MaxRetries = maxRetries
	return o
}

/* TransactionFunc is a function that executes within a transaction */
type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

/* WithTransaction executes a function within a transaction with automatic rollback on error */
func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn TransactionFunc) error {
	return WithTransactionOptions(ctx, pool, DefaultTransactionOptions(), fn)
}

/* WithTransactionOptions executes a function within a transaction with custom options */
func WithTransactionOptions(ctx context.Context, pool *pgxpool.Pool, opts *TransactionOptions, fn TransactionFunc) error {
	// Create context with timeout if specified
	txCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		txCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Retry logic
	var lastErr error
	maxAttempts := 1
	if opts.RetryOnError {
		maxAttempts = opts.MaxRetries + 1
		if maxAttempts < 1 {
			maxAttempts = 1
		}
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Begin transaction
		tx, err := pool.BeginTx(txCtx, pgx.TxOptions{
			IsoLevel:   opts.Isolation,
			AccessMode: pgx.ReadWrite,
		})
		if err != nil {
			if attempt < maxAttempts {
				time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Set read-only mode if specified
		if opts.ReadOnly {
			_, err = tx.Exec(txCtx, "SET TRANSACTION READ ONLY")
			if err != nil {
				tx.Rollback(txCtx)
				if attempt < maxAttempts {
					time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
					continue
				}
				return fmt.Errorf("failed to set read-only mode: %w", err)
			}
		}

		// Execute the function
		err = fn(txCtx, tx)
		if err != nil {
			// Rollback on error
			rollbackErr := tx.Rollback(txCtx)
			if rollbackErr != nil {
				return fmt.Errorf("transaction error: %w, rollback error: %v", err, rollbackErr)
			}

			// Retry if enabled and not last attempt
			if opts.RetryOnError && attempt < maxAttempts {
				lastErr = err
				time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
				continue
			}

			return err
		}

		// Commit transaction
		err = tx.Commit(txCtx)
		if err != nil {
			// Try to rollback if commit fails
			tx.Rollback(txCtx)
			if attempt < maxAttempts {
				lastErr = err
				time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Success
		return nil
	}

	// All retries exhausted
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("transaction failed after %d attempts", maxAttempts)
}

/* WithReadTransaction executes a read-only transaction */
func WithReadTransaction(ctx context.Context, pool *pgxpool.Pool, fn TransactionFunc) error {
	opts := DefaultTransactionOptions().WithReadOnly(true)
	return WithTransactionOptions(ctx, pool, opts, fn)
}

/* WithTimeoutTransaction executes a transaction with a timeout */
func WithTimeoutTransaction(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration, fn TransactionFunc) error {
	opts := DefaultTransactionOptions().WithTimeout(timeout)
	return WithTransactionOptions(ctx, pool, opts, fn)
}

/* TransactionManager manages transactions with context propagation */
type TransactionManager struct {
	pool *pgxpool.Pool
}

/* NewTransactionManager creates a new transaction manager */
func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

/* Execute executes a function within a transaction */
func (tm *TransactionManager) Execute(ctx context.Context, fn TransactionFunc) error {
	return WithTransaction(ctx, tm.pool, fn)
}

/* ExecuteWithOptions executes a function within a transaction with options */
func (tm *TransactionManager) ExecuteWithOptions(ctx context.Context, opts *TransactionOptions, fn TransactionFunc) error {
	return WithTransactionOptions(ctx, tm.pool, opts, fn)
}

/* ExecuteReadOnly executes a read-only transaction */
func (tm *TransactionManager) ExecuteReadOnly(ctx context.Context, fn TransactionFunc) error {
	return WithReadTransaction(ctx, tm.pool, fn)
}
