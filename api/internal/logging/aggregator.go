package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* LogAggregator provides centralized log aggregation */
type LogAggregator struct {
	pool *pgxpool.Pool
}

/* NewLogAggregator creates a new log aggregator */
func NewLogAggregator(pool *pgxpool.Pool) *LogAggregator {
	return &LogAggregator{pool: pool}
}

/* LogEntry represents a structured log entry */
type LogEntry struct {
	ID        uuid.UUID              `json:"id"`
	Level     string                 `json:"level"` // "debug", "info", "warn", "error", "fatal"
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Component string                 `json:"component,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

/* Log logs a structured log entry */
func (la *LogAggregator) Log(ctx context.Context, entry LogEntry) error {
	entry.ID = uuid.New()
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	fieldsJSON, _ := json.Marshal(entry.Fields)

	query := `
		INSERT INTO neuronip.aggregated_logs 
		(id, level, message, service, component, trace_id, span_id, fields, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := la.pool.Exec(ctx, query,
		entry.ID, entry.Level, entry.Message, entry.Service, entry.Component,
		entry.TraceID, entry.SpanID, fieldsJSON, entry.Timestamp,
	)
	return err
}

/* GetLogs retrieves logs with filtering */
func (la *LogAggregator) GetLogs(ctx context.Context, filters LogFilters) ([]LogEntry, error) {
	query := `
		SELECT id, level, message, service, component, trace_id, span_id, fields, timestamp
		FROM neuronip.aggregated_logs
		WHERE ($1 = '' OR level = $1)
			AND ($2 = '' OR service = $2)
			AND ($3 = '' OR component = $3)
			AND ($4 IS NULL OR timestamp >= $4)
			AND ($5 IS NULL OR timestamp <= $5)
		ORDER BY timestamp DESC
		LIMIT $6`

	rows, err := la.pool.Query(ctx, query,
		filters.Level, filters.Service, filters.Component,
		filters.StartTime, filters.EndTime, filters.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		var fieldsJSON json.RawMessage
		var traceID, spanID *string

		err := rows.Scan(
			&entry.ID, &entry.Level, &entry.Message, &entry.Service, &entry.Component,
			&traceID, &spanID, &fieldsJSON, &entry.Timestamp,
		)
		if err != nil {
			continue
		}

		if traceID != nil {
			entry.TraceID = *traceID
		}
		if spanID != nil {
			entry.SpanID = *spanID
		}
		if fieldsJSON != nil {
			json.Unmarshal(fieldsJSON, &entry.Fields)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

/* CleanupOldLogs removes logs older than retention period */
func (la *LogAggregator) CleanupOldLogs(ctx context.Context, retentionDays int) error {
	query := `
		DELETE FROM neuronip.aggregated_logs
		WHERE timestamp < NOW() - INTERVAL '%d days'`

	query = fmt.Sprintf(query, retentionDays)
	_, err := la.pool.Exec(ctx, query)
	return err
}

/* LogFilters represents log query filters */
type LogFilters struct {
	Level      string
	Service    string
	Component  string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
}
