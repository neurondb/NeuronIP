package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* HealthCheckResult represents the result of a health check */
type HealthCheckResult struct {
	Healthy     bool
	Message     string
	Latency     time.Duration
	Connections *PoolStats
}

/* PoolStats represents connection pool statistics */
type PoolStats struct {
	TotalConns    int32
	AcquiredConns int32
	IdleConns     int32
	MaxConns      int32
}

/* HealthChecker performs comprehensive database health checks */
type HealthChecker struct {
	pool *pgxpool.Pool
}

/* NewHealthChecker creates a new health checker */
func NewHealthChecker(pool *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{pool: pool}
}

/* CheckConnectivity checks basic database connectivity */
func (hc *HealthChecker) CheckConnectivity(ctx context.Context) HealthCheckResult {
	start := time.Now()
	err := hc.pool.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Database ping failed: %v", err),
			Latency: latency,
		}
	}

	return HealthCheckResult{
		Healthy: true,
		Message: "Database is reachable",
		Latency: latency,
	}
}

/* CheckReadHealth checks database read health */
func (hc *HealthChecker) CheckReadHealth(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	// Perform a simple read query
	var result int
	err := hc.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	latency := time.Since(start)

	if err != nil {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Read check failed: %v", err),
			Latency: latency,
		}
	}

	return HealthCheckResult{
		Healthy: true,
		Message: "Read operations are healthy",
		Latency: latency,
	}
}

/* CheckWriteHealth checks database write health */
func (hc *HealthChecker) CheckWriteHealth(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	// Perform a simple write query (using a transaction that rolls back)
	tx, err := hc.pool.Begin(ctx)
	if err != nil {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Failed to begin transaction: %v", err),
			Latency: time.Since(start),
		}
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SELECT 1")
	latency := time.Since(start)

	if err != nil {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Write check failed: %v", err),
			Latency: latency,
		}
	}

	return HealthCheckResult{
		Healthy: true,
		Message: "Write operations are healthy",
		Latency: latency,
	}
}

/* CheckPoolHealth checks connection pool health */
func (hc *HealthChecker) CheckPoolHealth(ctx context.Context) HealthCheckResult {
	stats := hc.pool.Stat()
	
	poolStats := &PoolStats{
		TotalConns:    stats.TotalConns(),
		AcquiredConns: stats.AcquiredConns(),
		IdleConns:     stats.IdleConns(),
		MaxConns:      stats.MaxConns(),
	}

	// Check if pool is exhausted
	exhaustionThreshold := float64(poolStats.MaxConns) * 0.9
	if float64(poolStats.AcquiredConns) >= exhaustionThreshold {
		return HealthCheckResult{
			Healthy:     false,
			Message:     fmt.Sprintf("Connection pool near exhaustion: %d/%d connections", poolStats.AcquiredConns, poolStats.MaxConns),
			Connections: poolStats,
		}
	}

	// Check if pool has minimum connections
	if poolStats.TotalConns < 1 {
		return HealthCheckResult{
			Healthy:     false,
			Message:     "No connections in pool",
			Connections: poolStats,
		}
	}

	return HealthCheckResult{
		Healthy:     true,
		Message:     fmt.Sprintf("Pool healthy: %d/%d connections", poolStats.AcquiredConns, poolStats.MaxConns),
		Connections: poolStats,
	}
}

/* CheckLatency checks database query latency */
func (hc *HealthChecker) CheckLatency(ctx context.Context, timeout time.Duration) HealthCheckResult {
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	var result int
	err := hc.pool.QueryRow(checkCtx, "SELECT 1").Scan(&result)
	latency := time.Since(start)

	if err != nil {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("Latency check failed: %v", err),
			Latency: latency,
		}
	}

	// Consider latency unhealthy if it exceeds threshold
	latencyThreshold := 100 * time.Millisecond
	if latency > latencyThreshold {
		return HealthCheckResult{
			Healthy: false,
			Message: fmt.Sprintf("High latency detected: %v (threshold: %v)", latency, latencyThreshold),
			Latency: latency,
		}
	}

	return HealthCheckResult{
		Healthy: true,
		Message: fmt.Sprintf("Latency acceptable: %v", latency),
		Latency: latency,
	}
}

/* ComprehensiveHealthCheck performs all health checks */
func (hc *HealthChecker) ComprehensiveHealthCheck(ctx context.Context) map[string]HealthCheckResult {
	results := make(map[string]HealthCheckResult)

	// Connectivity check
	results["connectivity"] = hc.CheckConnectivity(ctx)

	// Read health check
	results["read"] = hc.CheckReadHealth(ctx)

	// Write health check
	results["write"] = hc.CheckWriteHealth(ctx)

	// Pool health check
	results["pool"] = hc.CheckPoolHealth(ctx)

	// Latency check
	results["latency"] = hc.CheckLatency(ctx, 5*time.Second)

	return results
}

/* IsHealthy determines overall health based on all checks */
func (hc *HealthChecker) IsHealthy(ctx context.Context) (bool, map[string]HealthCheckResult) {
	results := hc.ComprehensiveHealthCheck(ctx)
	
	// All checks must be healthy
	for _, result := range results {
		if !result.Healthy {
			return false, results
		}
	}

	return true, results
}
