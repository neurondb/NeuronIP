package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* GetPoolFromContext gets the database pool from context, falling back to default */
func (q *Queries) GetPoolFromContext(ctx context.Context) *pgxpool.Pool {
	// Try to get pool from context (set by database middleware)
	if pool, ok := GetDatabaseFromContext(ctx); ok {
		return pool
	}
	// Fallback to default pool
	return q.DB
}

/* QueryRowContext executes a query and returns a single row, using context-aware pool */
func (q *Queries) QueryRowContext(ctx context.Context, query string, args ...interface{}) pgx.Row {
	pool := q.GetPoolFromContext(ctx)
	return pool.QueryRow(ctx, query, args...)
}

/* ExecContext executes a query, using context-aware pool */
func (q *Queries) ExecContext(ctx context.Context, query string, args ...interface{}) error {
	pool := q.GetPoolFromContext(ctx)
	_, err := pool.Exec(ctx, query, args...)
	return err
}

/* QueryContext executes a query and returns rows, using context-aware pool */
func (q *Queries) QueryContext(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	pool := q.GetPoolFromContext(ctx)
	return pool.Query(ctx, query, args...)
}
