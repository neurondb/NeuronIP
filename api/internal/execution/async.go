package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AsyncQueryService provides async query execution functionality */
type AsyncQueryService struct {
	pool *pgxpool.Pool
}

/* NewAsyncQueryService creates a new async query service */
func NewAsyncQueryService(pool *pgxpool.Pool) *AsyncQueryService {
	return &AsyncQueryService{pool: pool}
}

/* AsyncQuery represents an async query */
type AsyncQuery struct {
	ID             uuid.UUID              `json:"id"`
	QueryText      string                 `json:"query_text"`
	QueryParams    map[string]interface{} `json:"query_params,omitempty"`
	Status         string                 `json:"status"` // pending, running, completed, failed, cancelled
	ResultLocation *string                `json:"result_location,omitempty"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	RowsReturned   *int64                 `json:"rows_returned,omitempty"`
	ExecutionTimeMs *int                  `json:"execution_time_ms,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	CreatedBy      *string                `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* SubmitAsyncQuery submits a query for async execution */
func (s *AsyncQueryService) SubmitAsyncQuery(ctx context.Context, queryText string, queryParams map[string]interface{}, createdBy *string, expiresInHours int) (*AsyncQuery, error) {
	id := uuid.New()
	paramsJSON, _ := json.Marshal(queryParams)
	now := time.Now()
	
	var expiresAt *time.Time
	if expiresInHours > 0 {
		exp := now.Add(time.Duration(expiresInHours) * time.Hour)
		expiresAt = &exp
	}

	query := `
		INSERT INTO neuronip.async_queries 
		(id, query_text, query_params, status, created_by, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, $7)
		RETURNING id, query_text, query_params, status, result_location, error_message, rows_returned, execution_time_ms, started_at, completed_at, expires_at, created_by, created_at, updated_at`

	var asyncQuery AsyncQuery
	var paramsJSONRaw json.RawMessage
	var resultLocation, errorMsg, createdByVal sql.NullString
	var rowsReturned sql.NullInt64
	var executionTimeMs sql.NullInt32
	var startedAt, completedAt, expiresAtVal sql.NullTime

	err := s.pool.QueryRow(ctx, query, id, queryText, paramsJSON, createdBy, expiresAt, now, now).Scan(
		&asyncQuery.ID, &asyncQuery.QueryText, &paramsJSONRaw, &asyncQuery.Status,
		&resultLocation, &errorMsg, &rowsReturned, &executionTimeMs,
		&startedAt, &completedAt, &expiresAtVal, &createdByVal,
		&asyncQuery.CreatedAt, &asyncQuery.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit async query: %w", err)
	}

	if paramsJSONRaw != nil {
		json.Unmarshal(paramsJSONRaw, &asyncQuery.QueryParams)
	}
	if resultLocation.Valid {
		asyncQuery.ResultLocation = &resultLocation.String
	}
	if errorMsg.Valid {
		asyncQuery.ErrorMessage = &errorMsg.String
	}
	if rowsReturned.Valid {
		asyncQuery.RowsReturned = &rowsReturned.Int64
	}
	if executionTimeMs.Valid {
		execTime := int(executionTimeMs.Int32)
		asyncQuery.ExecutionTimeMs = &execTime
	}
	if startedAt.Valid {
		asyncQuery.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		asyncQuery.CompletedAt = &completedAt.Time
	}
	if expiresAtVal.Valid {
		asyncQuery.ExpiresAt = &expiresAtVal.Time
	}
	if createdByVal.Valid {
		asyncQuery.CreatedBy = &createdByVal.String
	}

	// Enqueue for execution
	jobService := NewJobQueueService(s.pool)
	_, err = jobService.EnqueueJob(ctx, "query", map[string]interface{}{
		"query_id": id.String(),
		"query_text": queryText,
		"query_params": queryParams,
	}, 100, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue query: %w", err)
	}

	return &asyncQuery, nil
}

/* GetAsyncQuery retrieves an async query by ID */
func (s *AsyncQueryService) GetAsyncQuery(ctx context.Context, queryID uuid.UUID) (*AsyncQuery, error) {
	query := `
		SELECT id, query_text, query_params, status, result_location, error_message, rows_returned, execution_time_ms, started_at, completed_at, expires_at, created_by, created_at, updated_at
		FROM neuronip.async_queries
		WHERE id = $1`

	var asyncQuery AsyncQuery
	var paramsJSONRaw json.RawMessage
	var resultLocation, errorMsg, createdByVal sql.NullString
	var rowsReturned sql.NullInt64
	var executionTimeMs sql.NullInt32
	var startedAt, completedAt, expiresAtVal sql.NullTime

	err := s.pool.QueryRow(ctx, query, queryID).Scan(
		&asyncQuery.ID, &asyncQuery.QueryText, &paramsJSONRaw, &asyncQuery.Status,
		&resultLocation, &errorMsg, &rowsReturned, &executionTimeMs,
		&startedAt, &completedAt, &expiresAtVal, &createdByVal,
		&asyncQuery.CreatedAt, &asyncQuery.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get async query: %w", err)
	}

	if paramsJSONRaw != nil {
		json.Unmarshal(paramsJSONRaw, &asyncQuery.QueryParams)
	}
	if resultLocation.Valid {
		asyncQuery.ResultLocation = &resultLocation.String
	}
	if errorMsg.Valid {
		asyncQuery.ErrorMessage = &errorMsg.String
	}
	if rowsReturned.Valid {
		asyncQuery.RowsReturned = &rowsReturned.Int64
	}
	if executionTimeMs.Valid {
		execTime := int(executionTimeMs.Int32)
		asyncQuery.ExecutionTimeMs = &execTime
	}
	if startedAt.Valid {
		asyncQuery.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		asyncQuery.CompletedAt = &completedAt.Time
	}
	if expiresAtVal.Valid {
		asyncQuery.ExpiresAt = &expiresAtVal.Time
	}
	if createdByVal.Valid {
		asyncQuery.CreatedBy = &createdByVal.String
	}

	return &asyncQuery, nil
}

/* ListAsyncQueries lists async queries */
func (s *AsyncQueryService) ListAsyncQueries(ctx context.Context, createdBy *string, status *string, limit int) ([]AsyncQuery, error) {
	query := `
		SELECT id, query_text, query_params, status, result_location, error_message, rows_returned, execution_time_ms, started_at, completed_at, expires_at, created_by, created_at, updated_at
		FROM neuronip.async_queries
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if createdBy != nil {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, *createdBy)
		argIndex++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list async queries: %w", err)
	}
	defer rows.Close()

	var queries []AsyncQuery
	for rows.Next() {
		var asyncQuery AsyncQuery
		var paramsJSONRaw json.RawMessage
		var resultLocation, errorMsg, createdByVal sql.NullString
		var rowsReturned sql.NullInt64
		var executionTimeMs sql.NullInt32
		var startedAt, completedAt, expiresAtVal sql.NullTime

		err := rows.Scan(
			&asyncQuery.ID, &asyncQuery.QueryText, &paramsJSONRaw, &asyncQuery.Status,
			&resultLocation, &errorMsg, &rowsReturned, &executionTimeMs,
			&startedAt, &completedAt, &expiresAtVal, &createdByVal,
			&asyncQuery.CreatedAt, &asyncQuery.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if paramsJSONRaw != nil {
			json.Unmarshal(paramsJSONRaw, &asyncQuery.QueryParams)
		}
		if resultLocation.Valid {
			asyncQuery.ResultLocation = &resultLocation.String
		}
		if errorMsg.Valid {
			asyncQuery.ErrorMessage = &errorMsg.String
		}
		if rowsReturned.Valid {
			asyncQuery.RowsReturned = &rowsReturned.Int64
		}
		if executionTimeMs.Valid {
			execTime := int(executionTimeMs.Int32)
			asyncQuery.ExecutionTimeMs = &execTime
		}
		if startedAt.Valid {
			asyncQuery.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			asyncQuery.CompletedAt = &completedAt.Time
		}
		if expiresAtVal.Valid {
			asyncQuery.ExpiresAt = &expiresAtVal.Time
		}
		if createdByVal.Valid {
			asyncQuery.CreatedBy = &createdByVal.String
		}

		queries = append(queries, asyncQuery)
	}

	return queries, nil
}

/* CancelAsyncQuery cancels an async query */
func (s *AsyncQueryService) CancelAsyncQuery(ctx context.Context, queryID uuid.UUID) error {
	query := `
		UPDATE neuronip.async_queries 
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'running')`

	result, err := s.pool.Exec(ctx, query, queryID)
	if err != nil {
		return fmt.Errorf("failed to cancel async query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("query not found or cannot be cancelled")
	}

	return nil
}
