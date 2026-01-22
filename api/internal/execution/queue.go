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

/* JobQueueService provides job queue functionality */
type JobQueueService struct {
	pool *pgxpool.Pool
}

/* NewJobQueueService creates a new job queue service */
func NewJobQueueService(pool *pgxpool.Pool) *JobQueueService {
	return &JobQueueService{pool: pool}
}

/* Job represents a job in the queue */
type Job struct {
	ID                  uuid.UUID              `json:"id"`
	JobType             string                 `json:"job_type"`
	JobPayload          map[string]interface{} `json:"job_payload"`
	Priority            int                    `json:"priority"`
	Status              string                 `json:"status"`
	ResourceRequirements map[string]interface{} `json:"resource_requirements,omitempty"`
	WorkerID            *string                `json:"worker_id,omitempty"`
	ScheduledAt         time.Time              `json:"scheduled_at"`
	StartedAt           *time.Time             `json:"started_at,omitempty"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage        *string                `json:"error_message,omitempty"`
	RetryCount          int                    `json:"retry_count"`
	MaxRetries          int                    `json:"max_retries"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

/* EnqueueJob adds a job to the queue */
func (s *JobQueueService) EnqueueJob(ctx context.Context, jobType string, payload map[string]interface{}, priority int, resourceRequirements map[string]interface{}) (*Job, error) {
	jobID := uuid.New()
	payloadJSON, _ := json.Marshal(payload)
	requirementsJSON, _ := json.Marshal(resourceRequirements)
	now := time.Now()

	query := `
		INSERT INTO neuronip.job_queue 
		(id, job_type, job_payload, priority, status, resource_requirements, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8)
		RETURNING id, job_type, job_payload, priority, status, resource_requirements, worker_id, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, created_at, updated_at`

	var job Job
	var payloadJSONRaw, requirementsJSONRaw json.RawMessage
	var workerID sql.NullString
	var startedAt, completedAt sql.NullTime
	var errorMsg sql.NullString

	err := s.pool.QueryRow(ctx, query, jobID, jobType, payloadJSON, priority, requirementsJSON, now, now, now).Scan(
		&job.ID, &job.JobType, &payloadJSONRaw, &job.Priority, &job.Status,
		&requirementsJSONRaw, &workerID, &job.ScheduledAt, &startedAt, &completedAt,
		&errorMsg, &job.RetryCount, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	json.Unmarshal(payloadJSONRaw, &job.JobPayload)
	json.Unmarshal(requirementsJSONRaw, &job.ResourceRequirements)
	if workerID.Valid {
		job.WorkerID = &workerID.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		job.ErrorMessage = &errorMsg.String
	}

	return &job, nil
}

/* DequeueJob dequeues the next available job */
func (s *JobQueueService) DequeueJob(ctx context.Context, workerID string, jobTypes []string) (*Job, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build query for job types
	var query string
	var args []interface{}
	if len(jobTypes) > 0 {
		query = `
			SELECT id, job_type, job_payload, priority, status, resource_requirements, worker_id, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, created_at, updated_at
			FROM neuronip.job_queue
			WHERE status = 'pending' AND job_type = ANY($1::text[])
			ORDER BY priority DESC, scheduled_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED`
		args = []interface{}{jobTypes}
	} else {
		query = `
			SELECT id, job_type, job_payload, priority, status, resource_requirements, worker_id, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, created_at, updated_at
			FROM neuronip.job_queue
			WHERE status = 'pending'
			ORDER BY priority DESC, scheduled_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED`
		args = []interface{}{}
	}

	var job Job
	var payloadJSONRaw, requirementsJSONRaw json.RawMessage
	var workerIDVal sql.NullString
	var startedAt, completedAt sql.NullTime
	var errorMsg sql.NullString

	err = tx.QueryRow(ctx, query, args...).Scan(
		&job.ID, &job.JobType, &payloadJSONRaw, &job.Priority, &job.Status,
		&requirementsJSONRaw, &workerIDVal, &job.ScheduledAt, &startedAt, &completedAt,
		&errorMsg, &job.RetryCount, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	json.Unmarshal(payloadJSONRaw, &job.JobPayload)
	json.Unmarshal(requirementsJSONRaw, &job.ResourceRequirements)
	if workerIDVal.Valid {
		job.WorkerID = &workerIDVal.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		job.ErrorMessage = &errorMsg.String
	}

	// Update job status and worker
	now := time.Now()
	updateQuery := `
		UPDATE neuronip.job_queue 
		SET status = 'running', worker_id = $1, started_at = $2, updated_at = $2
		WHERE id = $3`
	_, err = tx.Exec(ctx, updateQuery, workerID, now, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	job.Status = "running"
	job.WorkerID = &workerID
	job.StartedAt = &now

	return &job, nil
}

/* CompleteJob marks a job as completed */
func (s *JobQueueService) CompleteJob(ctx context.Context, jobID uuid.UUID, result map[string]interface{}) error {
	// Store result if needed
	_ = result
	now := time.Now()

	query := `
		UPDATE neuronip.job_queue 
		SET status = 'completed', completed_at = $1, updated_at = $1
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, now, jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}

/* FailJob marks a job as failed */
func (s *JobQueueService) FailJob(ctx context.Context, jobID uuid.UUID, errorMessage string, shouldRetry bool) error {
	now := time.Now()

	var query string
	if shouldRetry {
		// Increment retry count and set back to pending if not exceeded
		query = `
			UPDATE neuronip.job_queue 
			SET retry_count = retry_count + 1, 
			    status = CASE WHEN retry_count + 1 >= max_retries THEN 'failed' ELSE 'pending' END,
			    error_message = $1, updated_at = $2
			WHERE id = $3`
	} else {
		query = `
			UPDATE neuronip.job_queue 
			SET status = 'failed', error_message = $1, completed_at = $2, updated_at = $2
			WHERE id = $3`
	}

	_, err := s.pool.Exec(ctx, query, errorMessage, now, jobID)
	if err != nil {
		return fmt.Errorf("failed to fail job: %w", err)
	}

	return nil
}

/* GetJob retrieves a job by ID */
func (s *JobQueueService) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	query := `
		SELECT id, job_type, job_payload, priority, status, resource_requirements, worker_id, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, created_at, updated_at
		FROM neuronip.job_queue
		WHERE id = $1`

	var job Job
	var payloadJSONRaw, requirementsJSONRaw json.RawMessage
	var workerID sql.NullString
	var startedAt, completedAt sql.NullTime
	var errorMsg sql.NullString

	err := s.pool.QueryRow(ctx, query, jobID).Scan(
		&job.ID, &job.JobType, &payloadJSONRaw, &job.Priority, &job.Status,
		&requirementsJSONRaw, &workerID, &job.ScheduledAt, &startedAt, &completedAt,
		&errorMsg, &job.RetryCount, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	json.Unmarshal(payloadJSONRaw, &job.JobPayload)
	json.Unmarshal(requirementsJSONRaw, &job.ResourceRequirements)
	if workerID.Valid {
		job.WorkerID = &workerID.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		job.ErrorMessage = &errorMsg.String
	}

	return &job, nil
}

/* ListJobs lists jobs in the queue */
func (s *JobQueueService) ListJobs(ctx context.Context, status *string, jobType *string, limit int) ([]Job, error) {
	query := `
		SELECT id, job_type, job_payload, priority, status, resource_requirements, worker_id, scheduled_at, started_at, completed_at, error_message, retry_count, max_retries, created_at, updated_at
		FROM neuronip.job_queue
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	if jobType != nil {
		query += fmt.Sprintf(" AND job_type = $%d", argIndex)
		args = append(args, *jobType)
		argIndex++
	}

	query += " ORDER BY priority DESC, scheduled_at ASC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		var payloadJSONRaw, requirementsJSONRaw json.RawMessage
		var workerID sql.NullString
		var startedAt, completedAt sql.NullTime
		var errorMsg sql.NullString

		err := rows.Scan(
			&job.ID, &job.JobType, &payloadJSONRaw, &job.Priority, &job.Status,
			&requirementsJSONRaw, &workerID, &job.ScheduledAt, &startedAt, &completedAt,
			&errorMsg, &job.RetryCount, &job.MaxRetries, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(payloadJSONRaw, &job.JobPayload)
		json.Unmarshal(requirementsJSONRaw, &job.ResourceRequirements)
		if workerID.Valid {
			job.WorkerID = &workerID.String
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}
		if errorMsg.Valid {
			job.ErrorMessage = &errorMsg.String
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

/* EnqueueAgentJob enqueues a long-running agent job with resource requirements */
func (s *JobQueueService) EnqueueAgentJob(ctx context.Context, agentID string, task string, tools []string, priority int, maxExecutionTime time.Duration) (*Job, error) {
	payload := map[string]interface{}{
		"agent_id": agentID,
		"task":     task,
		"tools":    tools,
		"type":     "agent_execution",
	}

	resourceRequirements := map[string]interface{}{
		"max_execution_time_ms": maxExecutionTime.Milliseconds(),
		"memory_mb":             512, // Default memory requirement
		"cpu_cores":             1,   // Default CPU requirement
	}

	return s.EnqueueJob(ctx, "agent", payload, priority, resourceRequirements)
}

/* GetAgentJobs retrieves agent jobs */
func (s *JobQueueService) GetAgentJobs(ctx context.Context, agentID *string, status *string, limit int) ([]Job, error) {
	jobType := "agent"
	return s.ListJobs(ctx, status, &jobType, limit)
}
