package monitoring

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* QueryPerformanceMonitor tracks and logs slow queries */
type QueryPerformanceMonitor struct {
	pool          *pgxpool.Pool
	slowThreshold time.Duration
	logger        Logger
}

/* Logger interface for logging slow queries */
type Logger interface {
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

/* NewQueryPerformanceMonitor creates a new query performance monitor */
func NewQueryPerformanceMonitor(pool *pgxpool.Pool, slowThreshold time.Duration, logger Logger) *QueryPerformanceMonitor {
	return &QueryPerformanceMonitor{
		pool:          pool,
		slowThreshold: slowThreshold,
		logger:        logger,
	}
}

/* TrackQuery tracks query execution time and logs if slow */
func (m *QueryPerformanceMonitor) TrackQuery(ctx context.Context, queryName string, query string, duration time.Duration, err error) {
	if duration > m.slowThreshold {
		m.logger.Warn("Slow query detected",
			"query_name", queryName,
			"duration_ms", duration.Milliseconds(),
			"threshold_ms", m.slowThreshold.Milliseconds(),
			"query_preview", truncateQuery(query, 200),
			"error", err,
		)
	}

	// Record in database if needed
	if err == nil {
		m.recordQueryMetric(ctx, queryName, duration, false)
	} else {
		m.recordQueryMetric(ctx, queryName, duration, true)
	}
}

/* recordQueryMetric records query performance metrics */
func (m *QueryPerformanceMonitor) recordQueryMetric(ctx context.Context, queryName string, duration time.Duration, hasError bool) {
	// Insert into query performance tracking table if it exists
	query := `
		INSERT INTO neuronip.query_performance_metrics 
		(query_name, execution_time_ms, has_error, executed_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT DO NOTHING`

	// Use background context to avoid blocking
	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.pool.Exec(bgCtx, query, queryName, duration.Milliseconds(), hasError)
}

/* GetSlowQueries retrieves slow queries from the database */
func (m *QueryPerformanceMonitor) GetSlowQueries(ctx context.Context, limit int) ([]SlowQuery, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT query_name, execution_time_ms, has_error, executed_at
		FROM neuronip.query_performance_metrics
		WHERE execution_time_ms > $1
		ORDER BY execution_time_ms DESC
		LIMIT $2`

	rows, err := m.pool.Query(ctx, query, m.slowThreshold.Milliseconds(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get slow queries: %w", err)
	}
	defer rows.Close()

	var slowQueries []SlowQuery
	for rows.Next() {
		var sq SlowQuery
		err := rows.Scan(&sq.QueryName, &sq.ExecutionTimeMs, &sq.HasError, &sq.ExecutedAt)
		if err != nil {
			continue
		}
		slowQueries = append(slowQueries, sq)
	}

	return slowQueries, nil
}

/* SlowQuery represents a slow query record */
type SlowQuery struct {
	QueryName       string    `json:"query_name"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	HasError        bool      `json:"has_error"`
	ExecutedAt      time.Time `json:"executed_at"`
}

/* truncateQuery truncates a query string to specified length */
func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

/* QueryWrapper wraps a query execution with performance tracking */
func (m *QueryPerformanceMonitor) QueryWrapper(ctx context.Context, queryName string, query string, fn func(context.Context) error) error {
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	m.TrackQuery(ctx, queryName, query, duration, err)

	return err
}
