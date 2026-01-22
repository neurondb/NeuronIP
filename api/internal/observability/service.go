package observability

import (
	"context"
	"database/sql"
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

/* RecordUsageMetric records a usage metric */
func (s *ObservabilityService) RecordUsageMetric(ctx context.Context, userID *string, resourceType string, resourceID *string, metricName string, metricValue float64, unit *string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	
	query := `
		INSERT INTO neuronip.usage_metrics 
		(id, user_id, resource_type, resource_id, metric_name, metric_value, unit, metadata, timestamp)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW())`
	
	_, err := s.pool.Exec(ctx, query, userID, resourceType, resourceID, metricName, metricValue, unit, metadataJSON)
	return err
}

/* GetUsageMetrics retrieves usage metrics */
func (s *ObservabilityService) GetUsageMetrics(ctx context.Context, filters UsageFilters, limit int) ([]UsageMetric, error) {
	query := `
		SELECT id, user_id, resource_type, resource_id, metric_name, metric_value, unit, metadata, timestamp
		FROM neuronip.usage_metrics
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}
	
	if filters.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *filters.ResourceType)
		argIndex++
	}
	
	if filters.MetricName != nil {
		query += fmt.Sprintf(" AND metric_name = $%d", argIndex)
		args = append(args, *filters.MetricName)
		argIndex++
	}
	
	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}
	
	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}
	
	query += " ORDER BY timestamp DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %w", err)
	}
	defer rows.Close()
	
	metrics := make([]UsageMetric, 0)
	for rows.Next() {
		var m UsageMetric
		var metadataJSON json.RawMessage
		
		err := rows.Scan(&m.ID, &m.UserID, &m.ResourceType, &m.ResourceID, &m.MetricName, &m.MetricValue, &m.Unit, &metadataJSON, &m.Timestamp)
		if err != nil {
			continue
		}
		
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &m.Metadata)
		}
		
		metrics = append(metrics, m)
	}
	
	return metrics, nil
}

/* RecordCost records a cost */
func (s *ObservabilityService) RecordCost(ctx context.Context, userID *string, resourceType string, resourceID *string, costAmount float64, currency string, costCategory string, periodStart, periodEnd time.Time, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	
	query := `
		INSERT INTO neuronip.cost_tracking 
		(id, user_id, resource_type, resource_id, cost_amount, currency, cost_category,
		 billing_period_start, billing_period_end, metadata, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`
	
	_, err := s.pool.Exec(ctx, query, userID, resourceType, resourceID, costAmount, currency, costCategory, periodStart, periodEnd, metadataJSON)
	return err
}

/* GetCostSummary retrieves cost summary */
func (s *ObservabilityService) GetCostSummary(ctx context.Context, userID *string, startTime, endTime time.Time) (*CostSummary, error) {
	query := `
		SELECT 
			SUM(cost_amount) as total_cost,
			cost_category,
			COUNT(*) as record_count
		FROM neuronip.cost_tracking
		WHERE billing_period_start >= $1 AND billing_period_end <= $2`
	
	args := []interface{}{startTime, endTime}
	argIndex := 3
	
	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}
	
	query += " GROUP BY cost_category"
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}
	defer rows.Close()
	
	summary := &CostSummary{
		TotalCost: 0,
		ByCategory: make(map[string]float64),
		PeriodStart: startTime,
		PeriodEnd: endTime,
	}
	
	for rows.Next() {
		var category string
		var cost float64
		var count int
		
		err := rows.Scan(&cost, &category, &count)
		if err != nil {
			continue
		}
		
		summary.TotalCost += cost
		summary.ByCategory[category] = cost
	}
	
	return summary, nil
}

/* UsageMetric represents a usage metric */
type UsageMetric struct {
	ID           uuid.UUID              `json:"id"`
	UserID       *string                `json:"user_id,omitempty"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	MetricName   string                 `json:"metric_name"`
	MetricValue  float64                `json:"metric_value"`
	Unit         *string                `json:"unit,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

/* UsageFilters filters for usage metric queries */
type UsageFilters struct {
	UserID       *string
	ResourceType *string
	MetricName   *string
	StartTime    *time.Time
	EndTime      *time.Time
}

/* CostSummary represents a cost summary */
type CostSummary struct {
	TotalCost   float64             `json:"total_cost"`
	ByCategory  map[string]float64  `json:"by_category"`
	PeriodStart time.Time           `json:"period_start"`
	PeriodEnd   time.Time           `json:"period_end"`
}

/* GetRealTimeMetrics retrieves real-time aggregated metrics */
func (s *ObservabilityService) GetRealTimeMetrics(ctx context.Context, timeWindow string) (*RealTimeMetrics, error) {
	var interval string
	switch timeWindow {
	case "1m", "1min":
		interval = "1 minute"
	case "5m", "5min":
		interval = "5 minutes"
	case "15m", "15min":
		interval = "15 minutes"
	case "1h", "1hour":
		interval = "1 hour"
	default:
		interval = "5 minutes"
	}

	// Note: time_bucket requires TimescaleDB extension
	// For standard PostgreSQL, use date_trunc instead
	fallbackQuery := fmt.Sprintf(`
		SELECT 
			date_trunc('minute', executed_at) as time_bucket,
			COUNT(*) as query_count,
			AVG(execution_time_ms / 1000.0) as avg_latency,
			MAX(execution_time_ms / 1000.0) as max_latency,
			MIN(execution_time_ms / 1000.0) as min_latency,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as error_count
		FROM neuronip.warehouse_queries
		WHERE executed_at > NOW() - INTERVAL '%s'
		GROUP BY time_bucket
		ORDER BY time_bucket DESC
		LIMIT 100`, interval)

	rows, err := s.pool.Query(ctx, fallbackQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get real-time metrics: %w", err)
	}
	defer rows.Close()

	metrics := &RealTimeMetrics{
		TimeWindow: timeWindow,
		DataPoints: []RealTimeDataPoint{},
	}

	for rows.Next() {
		var dp RealTimeDataPoint
		var avgLatency, maxLatency, minLatency *float64

		err := rows.Scan(&dp.Timestamp, &dp.QueryCount, &avgLatency, &maxLatency, &minLatency, &dp.SuccessCount, &dp.ErrorCount)
		if err != nil {
			continue
		}

		if avgLatency != nil {
			dp.AvgLatency = *avgLatency
		}
		if maxLatency != nil {
			dp.MaxLatency = *maxLatency
		}
		if minLatency != nil {
			dp.MinLatency = *minLatency
		}

		metrics.DataPoints = append(metrics.DataPoints, dp)
	}

	return metrics, nil
}

/* RealTimeMetrics represents real-time aggregated metrics */
type RealTimeMetrics struct {
	TimeWindow string              `json:"time_window"`
	DataPoints []RealTimeDataPoint `json:"data_points"`
}

/* RealTimeDataPoint represents a single data point in real-time metrics */
type RealTimeDataPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	QueryCount   int       `json:"query_count"`
	AvgLatency   float64   `json:"avg_latency"`
	MaxLatency   float64   `json:"max_latency"`
	MinLatency   float64   `json:"min_latency"`
	SuccessCount int       `json:"success_count"`
	ErrorCount   int       `json:"error_count"`
}

/* GetLogStream retrieves a stream of logs (returns recent logs for streaming simulation) */
func (s *ObservabilityService) GetLogStream(ctx context.Context, logType, level string, since time.Time, limit int) ([]SystemLog, error) {
	if limit <= 0 {
		limit = 1000
	}

	query := `
		SELECT id, log_type, level, message, context, timestamp
		FROM neuronip.system_logs
		WHERE ($1 = '' OR log_type = $1) 
			AND ($2 = '' OR level = $2)
			AND timestamp >= $3
		ORDER BY timestamp DESC
		LIMIT $4`

	rows, err := s.pool.Query(ctx, query, logType, level, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stream: %w", err)
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

/* GetPerformanceBenchmark retrieves performance benchmarks */
func (s *ObservabilityService) GetPerformanceBenchmark(ctx context.Context, metricType string, timeRange string) (*PerformanceBenchmark, error) {
	var interval string
	switch timeRange {
	case "1h", "1hour":
		interval = "1 hour"
	case "24h", "1day":
		interval = "24 hours"
	case "7d", "1week":
		interval = "7 days"
	case "30d", "1month":
		interval = "30 days"
	default:
		interval = "24 hours"
	}

	var query string
	switch metricType {
	case "query":
		query = fmt.Sprintf(`
			SELECT 
				AVG(execution_time_ms / 1000.0) as avg_latency,
				PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY execution_time_ms / 1000.0) as p50_latency,
				PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY execution_time_ms / 1000.0) as p95_latency,
				PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY execution_time_ms / 1000.0) as p99_latency,
				MAX(execution_time_ms / 1000.0) as max_latency,
				COUNT(*) as total_queries,
				SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as success_count,
				SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as error_count
			FROM neuronip.warehouse_queries
			WHERE executed_at > NOW() - INTERVAL '%s'`, interval)
	case "agent":
		query = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_executions,
				SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) as error_count
			FROM neuronip.system_logs
			WHERE log_type = 'agent' AND timestamp > NOW() - INTERVAL '%s'`, interval)
	case "workflow":
		query = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_executions,
				SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) as error_count
			FROM neuronip.system_logs
			WHERE log_type = 'workflow' AND timestamp > NOW() - INTERVAL '%s'`, interval)
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance benchmark: %w", err)
	}
	defer rows.Close()

	benchmark := &PerformanceBenchmark{
		MetricType: metricType,
		TimeRange:  timeRange,
	}

	if rows.Next() {
		if metricType == "query" {
			var avgLatency, p50Latency, p95Latency, p99Latency, maxLatency *float64
			var totalQueries, successCount, errorCount *int

			err := rows.Scan(&avgLatency, &p50Latency, &p95Latency, &p99Latency, &maxLatency, &totalQueries, &successCount, &errorCount)
			if err == nil {
				if avgLatency != nil {
					benchmark.AvgLatency = *avgLatency
				}
				if p50Latency != nil {
					benchmark.P50Latency = *p50Latency
				}
				if p95Latency != nil {
					benchmark.P95Latency = *p95Latency
				}
				if p99Latency != nil {
					benchmark.P99Latency = *p99Latency
				}
				if maxLatency != nil {
					benchmark.MaxLatency = *maxLatency
				}
				if totalQueries != nil {
					benchmark.TotalExecutions = *totalQueries
				}
				if successCount != nil {
					benchmark.SuccessCount = *successCount
				}
				if errorCount != nil {
					benchmark.ErrorCount = *errorCount
				}
			}
		} else {
			var totalExecutions, errorCount *int
			err := rows.Scan(&totalExecutions, &errorCount)
			if err == nil {
				if totalExecutions != nil {
					benchmark.TotalExecutions = *totalExecutions
				}
				if errorCount != nil {
					benchmark.ErrorCount = *errorCount
				}
			}
		}
	}

	return benchmark, nil
}

/* PerformanceBenchmark represents performance benchmark data */
type PerformanceBenchmark struct {
	MetricType      string  `json:"metric_type"`
	TimeRange       string  `json:"time_range"`
	AvgLatency      float64 `json:"avg_latency,omitempty"`
	P50Latency      float64 `json:"p50_latency,omitempty"`
	P95Latency      float64 `json:"p95_latency,omitempty"`
	P99Latency      float64 `json:"p99_latency,omitempty"`
	MaxLatency      float64 `json:"max_latency,omitempty"`
	TotalExecutions int     `json:"total_executions"`
	SuccessCount    int     `json:"success_count"`
	ErrorCount      int     `json:"error_count"`
}

/* GetCostBreakdown retrieves detailed cost breakdown */
func (s *ObservabilityService) GetCostBreakdown(ctx context.Context, userID *string, startTime, endTime time.Time, groupBy string) ([]CostBreakdown, error) {
	var groupByClause string
	switch groupBy {
	case "category":
		groupByClause = "cost_category"
	case "resource_type":
		groupByClause = "resource_type"
	case "user":
		groupByClause = "user_id"
	default:
		groupByClause = "cost_category"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as group_key,
			SUM(cost_amount) as total_cost,
			AVG(cost_amount) as avg_cost,
			COUNT(*) as record_count
		FROM neuronip.cost_tracking
		WHERE billing_period_start >= $1 AND billing_period_end <= $2`, groupByClause)

	args := []interface{}{startTime, endTime}
	argIndex := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	query += fmt.Sprintf(" GROUP BY %s ORDER BY total_cost DESC", groupByClause)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost breakdown: %w", err)
	}
	defer rows.Close()

	var breakdown []CostBreakdown
	for rows.Next() {
		var cb CostBreakdown
		var groupKey *string
		var totalCost, avgCost *float64
		var recordCount *int

		err := rows.Scan(&groupKey, &totalCost, &avgCost, &recordCount)
		if err != nil {
			continue
		}

		if groupKey != nil {
			cb.GroupKey = *groupKey
		}
		if totalCost != nil {
			cb.TotalCost = *totalCost
		}
		if avgCost != nil {
			cb.AvgCost = *avgCost
		}
		if recordCount != nil {
			cb.RecordCount = *recordCount
		}

		breakdown = append(breakdown, cb)
	}

	return breakdown, nil
}

/* CostBreakdown represents cost breakdown data */
type CostBreakdown struct {
	GroupKey    string  `json:"group_key"`
	TotalCost   float64 `json:"total_cost"`
	AvgCost     float64 `json:"avg_cost"`
	RecordCount int     `json:"record_count"`
}

/* AgentExecutionLog represents an agent execution log entry */
type AgentExecutionLog struct {
	ID              uuid.UUID              `json:"id"`
	AgentID          string                 `json:"agent_id"`
	AgentRunID       uuid.UUID              `json:"agent_run_id"`
	StepID           *string                `json:"step_id,omitempty"`
	StepType         string                 `json:"step_type"` // tool_call, reasoning, decision, etc.
	ToolName         *string                `json:"tool_name,omitempty"`
	Input            map[string]interface{} `json:"input,omitempty"`
	Output           map[string]interface{} `json:"output,omitempty"`
	Decision         *string                `json:"decision,omitempty"`
	LatencyMs        int64                  `json:"latency_ms"`
	TokensUsed       *int64                 `json:"tokens_used,omitempty"`
	Cost             *float64               `json:"cost,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

/* GetAgentExecutionLogs retrieves agent execution logs */
func (s *ObservabilityService) GetAgentExecutionLogs(ctx context.Context, agentID *string, agentRunID *uuid.UUID, limit int) ([]AgentExecutionLog, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, agent_id, agent_run_id, step_id, step_type, tool_name, input_data, output_data,
		       decision, latency_ms, tokens_used, cost, metadata, timestamp
		FROM neuronip.agent_execution_logs
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if agentID != nil {
		query += fmt.Sprintf(" AND agent_id = $%d", argIndex)
		args = append(args, *agentID)
		argIndex++
	}

	if agentRunID != nil {
		query += fmt.Sprintf(" AND agent_run_id = $%d", argIndex)
		args = append(args, *agentRunID)
		argIndex++
	}

	query += " ORDER BY timestamp ASC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent execution logs: %w", err)
	}
	defer rows.Close()

	var logs []AgentExecutionLog
	for rows.Next() {
		var log AgentExecutionLog
		var stepID, toolName, decision sql.NullString
		var inputJSON, outputJSON, metadataJSON json.RawMessage
		var tokensUsed sql.NullInt64
		var cost sql.NullFloat64

		err := rows.Scan(
			&log.ID, &log.AgentID, &log.AgentRunID, &stepID, &log.StepType, &toolName,
			&inputJSON, &outputJSON, &decision, &log.LatencyMs, &tokensUsed, &cost,
			&metadataJSON, &log.Timestamp,
		)
		if err != nil {
			continue
		}

		if stepID.Valid {
			log.StepID = &stepID.String
		}
		if toolName.Valid {
			log.ToolName = &toolName.String
		}
		if decision.Valid {
			log.Decision = &decision.String
		}
		if tokensUsed.Valid {
			log.TokensUsed = &tokensUsed.Int64
		}
		if cost.Valid {
			log.Cost = &cost.Float64
		}
		if inputJSON != nil {
			json.Unmarshal(inputJSON, &log.Input)
		}
		if outputJSON != nil {
			json.Unmarshal(outputJSON, &log.Output)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

/* RecordAgentExecutionLog records an agent execution log */
func (s *ObservabilityService) RecordAgentExecutionLog(ctx context.Context, agentID string, agentRunID uuid.UUID, stepID *string, stepType string, toolName *string, input map[string]interface{}, output map[string]interface{}, decision *string, latencyMs int64, tokensUsed *int64, cost *float64, metadata map[string]interface{}) error {
	id := uuid.New()
	inputJSON, _ := json.Marshal(input)
	outputJSON, _ := json.Marshal(output)
	metadataJSON, _ := json.Marshal(metadata)

	var stepIDVal, toolNameVal, decisionVal sql.NullString
	if stepID != nil {
		stepIDVal = sql.NullString{String: *stepID, Valid: true}
	}
	if toolName != nil {
		toolNameVal = sql.NullString{String: *toolName, Valid: true}
	}
	if decision != nil {
		decisionVal = sql.NullString{String: *decision, Valid: true}
	}

	var tokensUsedVal sql.NullInt64
	if tokensUsed != nil {
		tokensUsedVal = sql.NullInt64{Int64: *tokensUsed, Valid: true}
	}

	var costVal sql.NullFloat64
	if cost != nil {
		costVal = sql.NullFloat64{Float64: *cost, Valid: true}
	}

	query := `
		INSERT INTO neuronip.agent_execution_logs 
		(id, agent_id, agent_run_id, step_id, step_type, tool_name, input_data, output_data,
		 decision, latency_ms, tokens_used, cost, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())`

	_, err := s.pool.Exec(ctx, query, id, agentID, agentRunID, stepIDVal, stepType, toolNameVal,
		inputJSON, outputJSON, decisionVal, latencyMs, tokensUsedVal, costVal, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to record agent execution log: %w", err)
	}

	return nil
}

/* GetQueryCost retrieves cost for a specific query */
func (s *ObservabilityService) GetQueryCost(ctx context.Context, queryID uuid.UUID) (*QueryCost, error) {
	query := `
		SELECT 
			SUM(cost_amount) as total_cost,
			COUNT(*) as cost_records,
			AVG(cost_amount) as avg_cost
		FROM neuronip.cost_tracking
		WHERE resource_type = 'query' AND resource_id = $1::text`

	var cost QueryCost
	var totalCost, avgCost sql.NullFloat64
	var recordCount sql.NullInt64

	err := s.pool.QueryRow(ctx, query, queryID.String()).Scan(&totalCost, &recordCount, &avgCost)
	if err != nil {
		return nil, fmt.Errorf("failed to get query cost: %w", err)
	}

	cost.QueryID = queryID
	if totalCost.Valid {
		cost.TotalCost = totalCost.Float64
	}
	if avgCost.Valid {
		cost.AvgCost = avgCost.Float64
	}
	if recordCount.Valid {
		cost.RecordCount = int(recordCount.Int64)
	}

	return &cost, nil
}

/* GetAgentRunCost retrieves cost for a specific agent run */
func (s *ObservabilityService) GetAgentRunCost(ctx context.Context, agentRunID uuid.UUID) (*AgentRunCost, error) {
	query := `
		SELECT 
			SUM(cost_amount) as total_cost,
			COUNT(*) as cost_records,
			AVG(cost_amount) as avg_cost,
			SUM(tokens_used) as total_tokens
		FROM neuronip.cost_tracking ct
		LEFT JOIN neuronip.agent_execution_logs ael ON ct.resource_id = ael.agent_run_id::text
		WHERE ct.resource_type = 'agent' AND ct.resource_id = $1::text`

	var cost AgentRunCost
	var totalCost, avgCost sql.NullFloat64
	var recordCount, totalTokens sql.NullInt64

	err := s.pool.QueryRow(ctx, query, agentRunID.String()).Scan(&totalCost, &recordCount, &avgCost, &totalTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run cost: %w", err)
	}

	cost.AgentRunID = agentRunID
	if totalCost.Valid {
		cost.TotalCost = totalCost.Float64
	}
	if avgCost.Valid {
		cost.AvgCost = avgCost.Float64
	}
	if recordCount.Valid {
		cost.RecordCount = int(recordCount.Int64)
	}
	if totalTokens.Valid {
		cost.TotalTokens = totalTokens.Int64
	}

	return &cost, nil
}

/* QueryCost represents cost for a query */
type QueryCost struct {
	QueryID    uuid.UUID `json:"query_id"`
	TotalCost  float64   `json:"total_cost"`
	AvgCost    float64   `json:"avg_cost"`
	RecordCount int      `json:"record_count"`
}

/* AgentRunCost represents cost for an agent run */
type AgentRunCost struct {
	AgentRunID  uuid.UUID `json:"agent_run_id"`
	TotalCost   float64   `json:"total_cost"`
	AvgCost     float64   `json:"avg_cost"`
	TotalTokens int64     `json:"total_tokens"`
	RecordCount int       `json:"record_count"`
}
