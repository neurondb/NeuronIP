package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DLQService provides dead letter queue functionality */
type DLQService struct {
	pool *pgxpool.Pool
}

/* NewDLQService creates a new DLQ service */
func NewDLQService(pool *pgxpool.Pool) *DLQService {
	return &DLQService{pool: pool}
}

/* DLQEntry represents a dead letter queue entry */
type DLQEntry struct {
	ID            uuid.UUID              `json:"id"`
	JobID         uuid.UUID              `json:"job_id"`
	DataSourceID  uuid.UUID              `json:"data_source_id"`
	ErrorType     string                 `json:"error_type"`
	ErrorMessage  string                 `json:"error_message"`
	FailedData    map[string]interface{} `json:"failed_data,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	LastAttemptAt time.Time              `json:"last_attempt_at"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* AddToDLQ adds a failed job to the dead letter queue */
func (s *DLQService) AddToDLQ(ctx context.Context, jobID, dataSourceID uuid.UUID, errorType, errorMessage string, failedData map[string]interface{}, retryCount int) (*DLQEntry, error) {
	entryID := uuid.New()
	failedDataJSON, _ := json.Marshal(failedData)

	query := `
		INSERT INTO neuronip.dlq_entries (
			id, job_id, data_source_id, error_type, error_message,
			failed_data, retry_count, last_attempt_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err := s.pool.QueryRow(ctx, query,
		entryID, jobID, dataSourceID, errorType, errorMessage,
		failedDataJSON, retryCount,
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add to DLQ: %w", err)
	}

	return &DLQEntry{
		ID:            entryID,
		JobID:         jobID,
		DataSourceID:  dataSourceID,
		ErrorType:     errorType,
		ErrorMessage:  errorMessage,
		FailedData:    failedData,
		RetryCount:    retryCount,
		LastAttemptAt: time.Now(),
		CreatedAt:     createdAt,
	}, nil
}

/* ListDLQEntries lists entries in the dead letter queue */
func (s *DLQService) ListDLQEntries(ctx context.Context, limit int, offset int) ([]DLQEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, job_id, data_source_id, error_type, error_message,
		       failed_data, retry_count, last_attempt_at, created_at
		FROM neuronip.dlq_entries
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list DLQ entries: %w", err)
	}
	defer rows.Close()

	var entries []DLQEntry
	for rows.Next() {
		var entry DLQEntry
		var failedDataJSON []byte
		err := rows.Scan(
			&entry.ID, &entry.JobID, &entry.DataSourceID, &entry.ErrorType,
			&entry.ErrorMessage, &failedDataJSON, &entry.RetryCount,
			&entry.LastAttemptAt, &entry.CreatedAt,
		)
		if err != nil {
			continue
		}
		json.Unmarshal(failedDataJSON, &entry.FailedData)
		entries = append(entries, entry)
	}

	return entries, nil
}

/* RetryDLQEntry retries a DLQ entry */
func (s *DLQService) RetryDLQEntry(ctx context.Context, entryID uuid.UUID) error {
	query := `DELETE FROM neuronip.dlq_entries WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, entryID)
	return err
}

/* GetDLQStats gets statistics about the dead letter queue */
func (s *DLQService) GetDLQStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_entries,
			COUNT(DISTINCT data_source_id) as affected_sources,
			COUNT(DISTINCT error_type) as error_types,
			MAX(created_at) as latest_entry
		FROM neuronip.dlq_entries
	`
	var totalEntries, affectedSources, errorTypes int
	var latestEntry *time.Time

	err := s.pool.QueryRow(ctx, query).Scan(&totalEntries, &affectedSources, &errorTypes, &latestEntry)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_entries":    totalEntries,
		"affected_sources": affectedSources,
		"error_types":      errorTypes,
		"latest_entry":     latestEntry,
	}, nil
}
