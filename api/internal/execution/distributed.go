package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DistributedExecutor provides distributed job execution functionality */
type DistributedExecutor struct {
	pool           *pgxpool.Pool
	queueManager   *PriorityQueueManager
	nodeID         string
	workerPoolSize int
}

/* NewDistributedExecutor creates a new distributed executor */
func NewDistributedExecutor(pool *pgxpool.Pool, nodeID string, workerPoolSize int) *DistributedExecutor {
	return &DistributedExecutor{
		pool:           pool,
		queueManager:   NewPriorityQueueManager(),
		nodeID:         nodeID,
		workerPoolSize: workerPoolSize,
	}
}

/* DistributedJob represents a job for distributed execution */
type DistributedJob struct {
	ID            uuid.UUID              `json:"id"`
	JobType       string                 `json:"job_type"`
	Priority      int                    `json:"priority"`
	TenantID      *string                `json:"tenant_id,omitempty"`
	ResourceQuota map[string]interface{} `json:"resource_quota,omitempty"`
	JobData       map[string]interface{} `json:"job_data"`
	Status        string                 `json:"status"`
	AssignedNode  *string                `json:"assigned_node,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

/* SubmitJob submits a job for distributed execution */
func (de *DistributedExecutor) SubmitJob(ctx context.Context, job *DistributedJob) error {
	job.ID = uuid.New()
	job.Status = "pending"
	job.CreatedAt = time.Now()
	
	// Store job in database
	jobDataJSON, _ := json.Marshal(job.JobData)
	quotaJSON, _ := json.Marshal(job.ResourceQuota)
	
	query := `
		INSERT INTO neuronip.distributed_jobs 
		(id, job_type, priority, tenant_id, resource_quota, job_data, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := de.pool.Exec(ctx, query,
		job.ID, job.JobType, job.Priority, job.TenantID, quotaJSON, jobDataJSON, job.Status, job.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to submit job: %w", err)
	}
	
	// Add to priority queue
	queueName := "default"
	if job.TenantID != nil {
		queueName = *job.TenantID
	}
	
	queue := de.queueManager.GetQueue(queueName)
	queue.Push(&JobItem{
		ID:        job.ID,
		Priority:  job.Priority,
		CreatedAt: job.CreatedAt,
		JobType:   job.JobType,
		JobData:   job,
	})
	
	return nil
}

/* GetJob retrieves a distributed job */
func (de *DistributedExecutor) GetJob(ctx context.Context, jobID uuid.UUID) (*DistributedJob, error) {
	query := `
		SELECT id, job_type, priority, tenant_id, resource_quota, job_data, status, assigned_node, created_at, started_at, completed_at
		FROM neuronip.distributed_jobs
		WHERE id = $1
	`
	
	var job DistributedJob
	var tenantID *string
	var quotaJSON, jobDataJSON json.RawMessage
	var assignedNode *string
	var startedAt, completedAt *time.Time
	
	err := de.pool.QueryRow(ctx, query, jobID).Scan(
		&job.ID, &job.JobType, &job.Priority, &tenantID, &quotaJSON, &jobDataJSON,
		&job.Status, &assignedNode, &job.CreatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	
	job.TenantID = tenantID
	job.AssignedNode = assignedNode
	job.StartedAt = startedAt
	job.CompletedAt = completedAt
	
	if quotaJSON != nil {
		json.Unmarshal(quotaJSON, &job.ResourceQuota)
	}
	if jobDataJSON != nil {
		json.Unmarshal(jobDataJSON, &job.JobData)
	}
	
	return &job, nil
}

/* AssignJob assigns a job to this node */
func (de *DistributedExecutor) AssignJob(ctx context.Context, jobID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE neuronip.distributed_jobs
		SET status = 'running', assigned_node = $1, started_at = $2
		WHERE id = $3 AND status = 'pending'
	`
	
	result, err := de.pool.Exec(ctx, query, de.nodeID, now, jobID)
	if err != nil {
		return fmt.Errorf("failed to assign job: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found or already assigned")
	}
	
	return nil
}

/* CompleteJob marks a job as completed */
func (de *DistributedExecutor) CompleteJob(ctx context.Context, jobID uuid.UUID, result map[string]interface{}) error {
	now := time.Now()
	resultJSON, _ := json.Marshal(result)
	
	query := `
		UPDATE neuronip.distributed_jobs
		SET status = 'completed', completed_at = $1, result_data = $2
		WHERE id = $3
	`
	
	_, err := de.pool.Exec(ctx, query, now, resultJSON, jobID)
	return err
}

/* FailJob marks a job as failed */
func (de *DistributedExecutor) FailJob(ctx context.Context, jobID uuid.UUID, errorMsg string) error {
	now := time.Now()
	
	query := `
		UPDATE neuronip.distributed_jobs
		SET status = 'failed', completed_at = $1, error_message = $2
		WHERE id = $3
	`
	
	_, err := de.pool.Exec(ctx, query, now, errorMsg, jobID)
	return err
}

/* GetPendingJobs retrieves pending jobs for this node */
func (de *DistributedExecutor) GetPendingJobs(ctx context.Context, limit int) ([]*DistributedJob, error) {
	if limit <= 0 {
		limit = 10
	}
	
	query := `
		SELECT id, job_type, priority, tenant_id, resource_quota, job_data, status, created_at
		FROM neuronip.distributed_jobs
		WHERE status = 'pending'
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
	`
	
	rows, err := de.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()
	
	var jobs []*DistributedJob
	for rows.Next() {
		var job DistributedJob
		var tenantID *string
		var quotaJSON, jobDataJSON json.RawMessage
		
		err := rows.Scan(
			&job.ID, &job.JobType, &job.Priority, &tenantID, &quotaJSON, &jobDataJSON,
			&job.Status, &job.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		job.TenantID = tenantID
		if quotaJSON != nil {
			json.Unmarshal(quotaJSON, &job.ResourceQuota)
		}
		if jobDataJSON != nil {
			json.Unmarshal(jobDataJSON, &job.JobData)
		}
		
		jobs = append(jobs, &job)
	}
	
	return jobs, nil
}
