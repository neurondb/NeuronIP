package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/config"
)

/* Pool provides database connection pool management */
type Pool struct {
	*pgxpool.Pool
}

/* NewPool creates a new database connection pool */
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

/* Close closes the connection pool */
func (p *Pool) Close() {
	if p != nil && p.Pool != nil {
		p.Pool.Close()
	}
}

/* Health checks the health of the connection pool */
func (p *Pool) Health(ctx context.Context) error {
	if p == nil || p.Pool == nil {
		return fmt.Errorf("pool is nil")
	}
	return p.Pool.Ping(ctx)
}

/* Stats returns pool statistics */
func (p *Pool) Stats() *pgxpool.Stat {
	if p == nil || p.Pool == nil {
		return nil
	}
	return p.Pool.Stat()
}
