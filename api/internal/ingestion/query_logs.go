package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* QueryLogService provides query log collection and analysis */
type QueryLogService struct {
	pool *pgxpool.Pool
}

/* NewQueryLogService creates a new query log service */
func NewQueryLogService(pool *pgxpool.Pool) *QueryLogService {
	return &QueryLogService{pool: pool}
}

/* QueryLog represents a logged query */
type QueryLog struct {
	ID            uuid.UUID              `json:"id"`
	ConnectorID   uuid.UUID              `json:"connector_id"`
	SchemaName    string                 `json:"schema_name"`
	QueryText     string                 `json:"query_text"`
	QueryHash     string                 `json:"query_hash"`
	ExecutionTime int64                  `json:"execution_time_ms"`
	RowCount      int64                  `json:"row_count,omitempty"`
	Status        string                 `json:"status"` // "success", "error"
	Error         string                 `json:"error,omitempty"`
	ExecutedAt    time.Time              `json:"executed_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* LogQuery logs a query execution */
func (s *QueryLogService) LogQuery(ctx context.Context, log QueryLog) error {
	log.ID = uuid.New()
	if log.ExecutedAt.IsZero() {
		log.ExecutedAt = time.Now()
	}

	if log.QueryHash == "" {
		log.QueryHash = s.hashQuery(log.QueryText)
	}

	metadataJSON, _ := json.Marshal(log.Metadata)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.query_logs
		(id, connector_id, schema_name, query_text, query_hash,
		 execution_time_ms, row_count, status, error, executed_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		log.ID, log.ConnectorID, log.SchemaName, log.QueryText, log.QueryHash,
		log.ExecutionTime, log.RowCount, log.Status, log.Error, log.ExecutedAt, metadataJSON,
	)

	return err
}

/* hashQuery creates a hash of query text */
func (s *QueryLogService) hashQuery(queryText string) string {
	// Simplified hash - normalize query first
	normalized := normalizeQuery(queryText)
	return fmt.Sprintf("%x", normalized)
}

/* normalizeQuery normalizes SQL query for hashing */
func normalizeQuery(query string) string {
	// Remove comments, normalize whitespace
	// Simplified implementation
	return query
}

/* GetQueryLogs retrieves query logs with filtering */
func (s *QueryLogService) GetQueryLogs(ctx context.Context,
	connectorID *uuid.UUID, startTime, endTime *time.Time, limit int) ([]QueryLog, error) {

	if limit == 0 {
		limit = 100
	}

	query := `
		SELECT id, connector_id, schema_name, query_text, query_hash,
		       execution_time_ms, row_count, status, error, executed_at, metadata
		FROM neuronip.query_logs
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if connectorID != nil {
		query += fmt.Sprintf(" AND connector_id = $%d", argIdx)
		args = append(args, *connectorID)
		argIdx++
	}

	if startTime != nil {
		query += fmt.Sprintf(" AND executed_at >= $%d", argIdx)
		args = append(args, *startTime)
		argIdx++
	}

	if endTime != nil {
		query += fmt.Sprintf(" AND executed_at <= $%d", argIdx)
		args = append(args, *endTime)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY executed_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get query logs: %w", err)
	}
	defer rows.Close()

	var logs []QueryLog
	for rows.Next() {
		var log QueryLog
		var metadataJSON []byte

		err := rows.Scan(&log.ID, &log.ConnectorID, &log.SchemaName, &log.QueryText,
			&log.QueryHash, &log.ExecutionTime, &log.RowCount, &log.Status,
			&log.Error, &log.ExecutedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(metadataJSON, &log.Metadata)
		logs = append(logs, log)
	}

	return logs, nil
}

/* AnalyzeQueryPatterns analyzes query patterns from logs */
func (s *QueryLogService) AnalyzeQueryPatterns(ctx context.Context,
	connectorID *uuid.UUID, startTime, endTime *time.Time) (map[string]interface{}, error) {

	query := `
		SELECT 
			query_hash,
			COUNT(*) as frequency,
			AVG(execution_time_ms) as avg_execution_time,
			MIN(execution_time_ms) as min_execution_time,
			MAX(execution_time_ms) as max_execution_time,
			SUM(row_count) as total_rows
		FROM neuronip.query_logs
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if connectorID != nil {
		query += fmt.Sprintf(" AND connector_id = $%d", argIdx)
		args = append(args, *connectorID)
		argIdx++
	}

	if startTime != nil {
		query += fmt.Sprintf(" AND executed_at >= $%d", argIdx)
		args = append(args, *startTime)
		argIdx++
	}

	if endTime != nil {
		query += fmt.Sprintf(" AND executed_at <= $%d", argIdx)
		args = append(args, *endTime)
		argIdx++
	}

	query += " GROUP BY query_hash ORDER BY frequency DESC LIMIT 20"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze patterns: %w", err)
	}
	defer rows.Close()

	patterns := []map[string]interface{}{}
	for rows.Next() {
		var hash string
		var freq int64
		var avgTime, minTime, maxTime *float64
		var totalRows *int64

		err := rows.Scan(&hash, &freq, &avgTime, &minTime, &maxTime, &totalRows)
		if err != nil {
			continue
		}

		pattern := map[string]interface{}{
			"query_hash": hash,
			"frequency":  freq,
		}

		if avgTime != nil {
			pattern["avg_execution_time_ms"] = *avgTime
		}
		if minTime != nil {
			pattern["min_execution_time_ms"] = *minTime
		}
		if maxTime != nil {
			pattern["max_execution_time_ms"] = *maxTime
		}
		if totalRows != nil {
			pattern["total_rows"] = *totalRows
		}

		patterns = append(patterns, pattern)
	}

	result := map[string]interface{}{
		"patterns":    patterns,
		"total_queries": len(patterns),
	}

	return result, nil
}
