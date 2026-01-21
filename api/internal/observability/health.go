package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* EnhancedHealthService provides comprehensive health check functionality */
type EnhancedHealthService struct {
	pool *pgxpool.Pool
}

/* NewEnhancedHealthService creates a new enhanced health service */
func NewEnhancedHealthService(pool *pgxpool.Pool) *EnhancedHealthService {
	return &EnhancedHealthService{pool: pool}
}

/* HealthCheckResult represents comprehensive health check result */
type HealthCheckResult struct {
	Status       string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Timestamp    time.Time              `json:"timestamp"`
	Uptime       time.Duration          `json:"uptime,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Components   map[string]ComponentHealth `json:"components,omitempty"`
	Metrics      HealthMetrics          `json:"metrics,omitempty"`
}

/* ComponentHealth represents health of a component */
type ComponentHealth struct {
	Status      string    `json:"status"` // "healthy", "degraded", "unhealthy"
	Message     string    `json:"message,omitempty"`
	LastCheck   time.Time `json:"last_check"`
	ResponseTime *time.Duration `json:"response_time_ms,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

/* HealthMetrics represents system health metrics */
type HealthMetrics struct {
	DatabaseConnections HealthConnectionMetrics `json:"database_connections"`
	SystemResources     HealthResourceMetrics   `json:"system_resources"`
}

/* HealthConnectionMetrics represents database connection metrics */
type HealthConnectionMetrics struct {
	Active      int     `json:"active"`
	Idle        int     `json:"idle"`
	Max         int     `json:"max"`
	WaitCount   int     `json:"wait_count"` // Reserved for future implementation
	Utilization float64 `json:"utilization_percent"`
}

/* HealthResourceMetrics represents system resource metrics */
type HealthResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage_percent,omitempty"`
	MemoryUsage float64 `json:"memory_usage_percent,omitempty"`
	DiskUsage   float64 `json:"disk_usage_percent,omitempty"`
}

/* PerformHealthCheck performs comprehensive health check */
func (s *EnhancedHealthService) PerformHealthCheck(ctx context.Context) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Status:     "healthy",
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentHealth),
	}

	// Check database
	dbHealth := s.checkDatabase(ctx)
	result.Components["database"] = dbHealth

	// Check connection pool
	poolHealth := s.checkConnectionPool(ctx)
	result.Components["connection_pool"] = poolHealth

	// Check query performance
	queryHealth := s.checkQueryPerformance(ctx)
	result.Components["query_performance"] = queryHealth

	// Determine overall status
	result.Status = s.determineOverallStatus(result.Components)

	// Get metrics
	result.Metrics = s.getHealthMetrics(ctx)

	return result, nil
}

/* checkDatabase checks database health */
func (s *EnhancedHealthService) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastCheck: time.Now(),
	}

	err := s.pool.Ping(ctx)
	responseTime := time.Since(start)

	if err != nil {
		health.Status = "unhealthy"
		health.Message = fmt.Sprintf("Database ping failed: %v", err)
		health.ResponseTime = &responseTime
		return health
	}

	// Check database read
	var version string
	err = s.pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		health.Status = "degraded"
		health.Message = fmt.Sprintf("Database read failed: %v", err)
		health.ResponseTime = &responseTime
		return health
	}

	health.Status = "healthy"
	health.Message = "Database is healthy"
	health.ResponseTime = &responseTime
	health.Details = map[string]interface{}{
		"version": version,
	}

	return health
}

/* checkConnectionPool checks connection pool health */
func (s *EnhancedHealthService) checkConnectionPool(ctx context.Context) ComponentHealth {
	health := ComponentHealth{
		LastCheck: time.Now(),
	}

	stats := s.pool.Stat()
	active := int(stats.AcquiredConns())
	idle := int(stats.IdleConns())
	max := int(stats.MaxConns())

	utilization := float64(active+idle) / float64(max) * 100.0

	health.Details = map[string]interface{}{
		"active":      active,
		"idle":        idle,
		"max":         max,
		"utilization": utilization,
	}

	if utilization > 90.0 {
		health.Status = "degraded"
		health.Message = "Connection pool utilization is high"
	} else {
		health.Status = "healthy"
		health.Message = "Connection pool is healthy"
	}

	return health
}

/* checkQueryPerformance checks query performance */
func (s *EnhancedHealthService) checkQueryPerformance(ctx context.Context) ComponentHealth {
	health := ComponentHealth{
		LastCheck: time.Now(),
	}

	// Check recent query performance
	query := `
		SELECT AVG(execution_time_ms) as avg_time,
		       MAX(execution_time_ms) as max_time,
		       COUNT(*) as query_count
		FROM neuronip.warehouse_queries
		WHERE executed_at > NOW() - INTERVAL '5 minutes'
		AND execution_time_ms IS NOT NULL`

	var avgTime, maxTime *float64
	var queryCount int

	err := s.pool.QueryRow(ctx, query).Scan(&avgTime, &maxTime, &queryCount)
	if err != nil {
		health.Status = "degraded"
		health.Message = "Unable to check query performance"
		return health
	}

	health.Details = map[string]interface{}{
		"avg_time_ms":    avgTime,
		"max_time_ms":    maxTime,
		"query_count":    queryCount,
	}

	if avgTime != nil && *avgTime > 5000 {
		health.Status = "degraded"
		health.Message = "Average query time is high"
	} else if maxTime != nil && *maxTime > 30000 {
		health.Status = "degraded"
		health.Message = "Some queries are taking too long"
	} else {
		health.Status = "healthy"
		health.Message = "Query performance is acceptable"
	}

	return health
}

/* determineOverallStatus determines overall health status */
func (s *EnhancedHealthService) determineOverallStatus(components map[string]ComponentHealth) string {
	hasUnhealthy := false
	hasDegraded := false

	for _, component := range components {
		switch component.Status {
		case "unhealthy":
			hasUnhealthy = true
		case "degraded":
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return "unhealthy"
	} else if hasDegraded {
		return "degraded"
	}

	return "healthy"
}

/* getHealthMetrics retrieves health metrics */
func (s *EnhancedHealthService) getHealthMetrics(ctx context.Context) HealthMetrics {
	metrics := HealthMetrics{}

	// Get connection pool metrics
	stats := s.pool.Stat()
	active := int(stats.AcquiredConns())
	idle := int(stats.IdleConns())
	max := int(stats.MaxConns())

	metrics.DatabaseConnections = HealthConnectionMetrics{
		Active:      active,
		Idle:        idle,
		Max:         max,
		WaitCount:   0, // Not available in pgxpool.Stat
		Utilization: float64(active+idle) / float64(max) * 100.0,
	}

	// System resources would be retrieved from monitoring system
	// For now, leave empty as they would require system-level access

	return metrics
}

/* GetReadinessCheck performs readiness check */
func (s *EnhancedHealthService) GetReadinessCheck(ctx context.Context) bool {
	// Check if database is accessible
	err := s.pool.Ping(ctx)
	if err != nil {
		return false
	}

	// Check if connection pool is not at capacity
	stats := s.pool.Stat()
	utilization := float64(stats.AcquiredConns()+stats.IdleConns()) / float64(stats.MaxConns())
	if utilization > 0.95 {
		return false
	}

	return true
}

/* GetLivenessCheck performs liveness check */
func (s *EnhancedHealthService) GetLivenessCheck(ctx context.Context) bool {
	// Simple liveness check - can we ping the database?
	err := s.pool.Ping(ctx)
	return err == nil
}
