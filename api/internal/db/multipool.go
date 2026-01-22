package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/config"
)

/* MultiPool manages multiple database connection pools */
type MultiPool struct {
	pools map[string]*pgxpool.Pool
	mu    sync.RWMutex
	cfg   config.Config
}

/* NewMultiPool creates a new multi-database pool manager */
func NewMultiPool(ctx context.Context, cfg config.Config) (*MultiPool, error) {
	mp := &MultiPool{
		pools: make(map[string]*pgxpool.Pool),
		cfg:   cfg,
	}

	// Initialize default neuronip pool
	neuronipPool, err := createPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create neuronip pool: %w", err)
	}
	mp.pools["neuronip"] = neuronipPool

	// Initialize neuronai-demo pool if configured
	if cfg.Auth.NeuronAIDemo.Host != "" {
		demoPool, err := createPool(ctx, cfg.Auth.NeuronAIDemo)
		if err != nil {
			// Log error but don't fail - demo database is optional
			fmt.Printf("Warning: Failed to create neuronai-demo pool: %v\n", err)
		} else {
			mp.pools["neuronai-demo"] = demoPool
		}
	}

	return mp, nil
}

/* GetPool returns the connection pool for the specified database */
func (mp *MultiPool) GetPool(database string) (*pgxpool.Pool, error) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	// Default to neuronip if not specified
	if database == "" {
		database = "neuronip"
	}

	pool, exists := mp.pools[database]
	if !exists {
		return nil, fmt.Errorf("database pool not found: %s", database)
	}

	return pool, nil
}

/* GetPoolFromContext gets the pool based on database in context */
func (mp *MultiPool) GetPoolFromContext(ctx context.Context) (*pgxpool.Pool, error) {
	// Try to get database from context (set by session middleware)
	database := "neuronip" // default
	
	// Check if database is in context (from session)
	if dbVal := ctx.Value("database"); dbVal != nil {
		if dbStr, ok := dbVal.(string); ok && dbStr != "" {
			database = dbStr
		}
	}

	return mp.GetPool(database)
}

/* EnsurePool ensures a pool exists for the given database, creating it if needed */
func (mp *MultiPool) EnsurePool(ctx context.Context, database string) (*pgxpool.Pool, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Check if pool already exists
	if pool, exists := mp.pools[database]; exists {
		return pool, nil
	}

	// Create new pool based on database name
	var dbCfg config.DatabaseConfig
	switch database {
	case "neuronai-demo":
		dbCfg = mp.cfg.Auth.NeuronAIDemo
		// If not configured, use same as neuronip but different database name
		if dbCfg.Host == "" {
			dbCfg = mp.cfg.Database
			dbCfg.Name = "neuronai-demo"
		}
	case "neuronip":
		dbCfg = mp.cfg.Database
	default:
		return nil, fmt.Errorf("unknown database: %s", database)
	}

	pool, err := createPool(ctx, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool for %s: %w", database, err)
	}

	mp.pools[database] = pool
	return pool, nil
}

/* Close closes all connection pools */
func (mp *MultiPool) Close() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for name, pool := range mp.pools {
		pool.Close()
		delete(mp.pools, name)
	}
}

/* Health checks the health of all pools */
func (mp *MultiPool) Health(ctx context.Context) map[string]error {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	health := make(map[string]error)
	for name, pool := range mp.pools {
		if err := pool.Ping(ctx); err != nil {
			health[name] = err
		} else {
			health[name] = nil
		}
	}
	return health
}

/* Stats returns statistics for all pools */
func (mp *MultiPool) Stats() map[string]*pgxpool.Stat {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	stats := make(map[string]*pgxpool.Stat)
	for name, pool := range mp.pools {
		stats[name] = pool.Stat()
	}
	return stats
}

/* createPool creates a new connection pool */
func createPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
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

	return pool, nil
}
