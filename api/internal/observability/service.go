package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ObservabilityService provides observability functionality */
type ObservabilityService struct {
	pool *pgxpool.Pool
}

/* NewObservabilityService creates a new observability service */
func NewObservabilityService(pool *pgxpool.Pool) *ObservabilityService {
	return &ObservabilityService{pool: pool}
}

/* QueryPerformance represents query performance metrics */
type QueryPerformance struct {
	QueryID      uuid.UUID  `json:"query_id"`
	Duration     float64    `json:"duration"`
	RowCount     int        `json:"row_count"`
	Status       string     `json:"status"`
	ExecutedAt   time.Time  `json:"executed_at"`
}

/* SystemLog represents a system log entry */
type SystemLog struct {
	ID        uuid.UUID              `json:"id"`
	LogType   string                 `json:"log_type"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

/* SystemMetrics represents system-wide metrics */
type SystemMetrics struct {
	Latency    float64 `json:"latency"`
	Throughput float64 `json:"throughput"`
	Cost       float64 `json:"cost"`
	Timestamp  time.Time `json:"timestamp"`
}

/* GetQueryPerformance retrieves query performance metrics */
func (s *ObservabilityService) GetQueryPerformance(ctx context.Context, limit int) ([]QueryPerformance, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, executed_at, execution_time_ms, row_count, status
		FROM neuronip.warehouse_queries
		WHERE executed_at IS NOT NULL
		ORDER BY executed_at DESC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get query performance: %w", err)
	}
	defer rows.Close()

	var perf []QueryPerformance
	for rows.Next() {
		var p QueryPerformance
		var durationMs *float64

		err := rows.Scan(&p.QueryID, &p.ExecutedAt, &durationMs, &p.RowCount, &p.Status)
		if err != nil {
			continue
		}

		if durationMs != nil {
			p.Duration = *durationMs / 1000.0 // Convert ms to seconds
		}

		perf = append(perf, p)
	}

	return perf, nil
}

/* GetSystemLogs retrieves system logs with filtering */
func (s *ObservabilityService) GetSystemLogs(ctx context.Context, logType, level string, limit int) ([]SystemLog, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, log_type, level, message, context, timestamp
		FROM neuronip.system_logs
		WHERE ($1 = '' OR log_type = $1) AND ($2 = '' OR level = $2)
		ORDER BY timestamp DESC
		LIMIT $3`

	rows, err := s.pool.Query(ctx, query, logType, level, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get system logs: %w", err)
	}
	defer rows.Close()

	var logs []SystemLog
	for rows.Next() {
		var log SystemLog
		var contextJSON json.RawMessage

		err := rows.Scan(&log.ID, &log.LogType, &log.Level, &log.Message, &contextJSON, &log.Timestamp)
		if err != nil {
			continue
		}

		if contextJSON != nil {
			json.Unmarshal(contextJSON, &log.Context)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

/* LogSystemEvent logs a system event */
func (s *ObservabilityService) LogSystemEvent(ctx context.Context, logType, level, message string, context map[string]interface{}) error {
	contextJSON, _ := json.Marshal(context)

	query := `
		INSERT INTO neuronip.system_logs (id, log_type, level, message, context, timestamp)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`

	_, err := s.pool.Exec(ctx, query, logType, level, message, contextJSON)
	return err
}

/* GetSystemMetrics retrieves system metrics */
func (s *ObservabilityService) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	// Calculate average latency from recent queries
	latencyQuery := `
		SELECT AVG(execution_time_ms / 1000.0) as avg_latency
		FROM neuronip.warehouse_queries
		WHERE executed_at > NOW() - INTERVAL '1 hour' AND execution_time_ms IS NOT NULL`

	var latency *float64
	s.pool.QueryRow(ctx, latencyQuery).Scan(&latency)

	// Calculate throughput (queries per minute)
	throughputQuery := `
		SELECT COUNT(*) / 60.0 as throughput
		FROM neuronip.warehouse_queries
		WHERE executed_at > NOW() - INTERVAL '1 hour'`

	var throughput *float64
	s.pool.QueryRow(ctx, throughputQuery).Scan(&throughput)

	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	if latency != nil {
		metrics.Latency = *latency
	}
	if throughput != nil {
		metrics.Throughput = *throughput
	}

	// Cost would typically be calculated from usage metrics
	metrics.Cost = 0.0

	return metrics, nil
}

/* GetAgentLogs retrieves agent-specific logs */
func (s *ObservabilityService) GetAgentLogs(ctx context.Context, limit int) ([]SystemLog, error) {
	return s.GetSystemLogs(ctx, "agent", "", limit)
}

/* GetWorkflowLogs retrieves workflow-specific logs */
func (s *ObservabilityService) GetWorkflowLogs(ctx context.Context, limit int) ([]SystemLog, error) {
	return s.GetSystemLogs(ctx, "workflow", "", limit)
}
