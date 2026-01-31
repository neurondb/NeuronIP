package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TaskQueue manages distributed task queue */
type TaskQueue struct {
	pool *pgxpool.Pool
}

/* NewTaskQueue creates a new task queue */
func NewTaskQueue(pool *pgxpool.Pool) *TaskQueue {
	return &TaskQueue{pool: pool}
}

/* EnqueueTask enqueues a task for processing */
func (tq *TaskQueue) EnqueueTask(ctx context.Context, task Task) (*Task, error) {
	task.ID = uuid.New()
	task.Status = "pending"
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	payloadJSON, _ := json.Marshal(task.Payload)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
		INSERT INTO neuronip.cluster_tasks 
		(id, task_type, priority, payload, metadata, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, task_type, priority, payload, metadata, status, assigned_node_id, created_at, updated_at, started_at, completed_at`

	var payloadJSONRaw, metadataJSONRaw json.RawMessage
	var assignedNodeID *string
	var startedAt, completedAt *time.Time

	err := tq.pool.QueryRow(ctx, query,
		task.ID, task.TaskType, task.Priority, payloadJSON, metadataJSON,
		task.Status, task.CreatedAt, task.UpdatedAt,
	).Scan(
		&task.ID, &task.TaskType, &task.Priority, &payloadJSONRaw, &metadataJSONRaw,
		&task.Status, &assignedNodeID, &task.CreatedAt, &task.UpdatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	if payloadJSONRaw != nil {
		json.Unmarshal(payloadJSONRaw, &task.Payload)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &task.Metadata)
	}
	if assignedNodeID != nil {
		task.AssignedNodeID = assignedNodeID
	}
	if startedAt != nil {
		task.StartedAt = startedAt
	}
	if completedAt != nil {
		task.CompletedAt = completedAt
	}

	return &task, nil
}

/* DequeueTask dequeues a task for processing */
func (tq *TaskQueue) DequeueTask(ctx context.Context, nodeID string, taskTypes []string) (*Task, error) {
	var query string
	var args []interface{}

	if len(taskTypes) > 0 {
		query = `
			UPDATE neuronip.cluster_tasks
			SET status = 'processing', assigned_node_id = $1, started_at = NOW(), updated_at = NOW()
			WHERE id = (
				SELECT id FROM neuronip.cluster_tasks
				WHERE status = 'pending'
					AND task_type = ANY($2)
				ORDER BY priority DESC, created_at ASC
				LIMIT 1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, task_type, priority, payload, metadata, status, assigned_node_id, created_at, updated_at, started_at, completed_at`
		args = []interface{}{nodeID, taskTypes}
	} else {
		query = `
			UPDATE neuronip.cluster_tasks
			SET status = 'processing', assigned_node_id = $1, started_at = NOW(), updated_at = NOW()
			WHERE id = (
				SELECT id FROM neuronip.cluster_tasks
				WHERE status = 'pending'
				ORDER BY priority DESC, created_at ASC
				LIMIT 1
				FOR UPDATE SKIP LOCKED
			)
			RETURNING id, task_type, priority, payload, metadata, status, assigned_node_id, created_at, updated_at, started_at, completed_at`
		args = []interface{}{nodeID}
	}

	var task Task
	var payloadJSONRaw, metadataJSONRaw json.RawMessage
	var assignedNodeID *string
	var startedAt, completedAt *time.Time

	err := tq.pool.QueryRow(ctx, query, args...).Scan(
		&task.ID, &task.TaskType, &task.Priority, &payloadJSONRaw, &metadataJSONRaw,
		&task.Status, &assignedNodeID, &task.CreatedAt, &task.UpdatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no tasks available: %w", err)
	}

	if payloadJSONRaw != nil {
		json.Unmarshal(payloadJSONRaw, &task.Payload)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &task.Metadata)
	}
	if assignedNodeID != nil {
		task.AssignedNodeID = assignedNodeID
	}
	if startedAt != nil {
		task.StartedAt = startedAt
	}
	if completedAt != nil {
		task.CompletedAt = completedAt
	}

	return &task, nil
}

/* CompleteTask marks a task as completed */
func (tq *TaskQueue) CompleteTask(ctx context.Context, taskID uuid.UUID, result map[string]interface{}) error {
	resultJSON, _ := json.Marshal(result)
	query := `
		UPDATE neuronip.cluster_tasks
		SET status = 'completed', result = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := tq.pool.Exec(ctx, query, resultJSON, taskID)
	return err
}

/* FailTask marks a task as failed */
func (tq *TaskQueue) FailTask(ctx context.Context, taskID uuid.UUID, errorMsg string) error {
	query := `
		UPDATE neuronip.cluster_tasks
		SET status = 'failed', error_message = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := tq.pool.Exec(ctx, query, errorMsg, taskID)
	return err
}

/* GetTask retrieves a task by ID */
func (tq *TaskQueue) GetTask(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	query := `
		SELECT id, task_type, priority, payload, metadata, status, assigned_node_id, 
		       result, error_message, created_at, updated_at, started_at, completed_at
		FROM neuronip.cluster_tasks
		WHERE id = $1`

	var task Task
	var payloadJSONRaw, metadataJSONRaw, resultJSONRaw json.RawMessage
	var assignedNodeID *string
	var errorMsg *string
	var startedAt, completedAt *time.Time

	err := tq.pool.QueryRow(ctx, query, taskID).Scan(
		&task.ID, &task.TaskType, &task.Priority, &payloadJSONRaw, &metadataJSONRaw,
		&task.Status, &assignedNodeID, &resultJSONRaw, &errorMsg,
		&task.CreatedAt, &task.UpdatedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if payloadJSONRaw != nil {
		json.Unmarshal(payloadJSONRaw, &task.Payload)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &task.Metadata)
	}
	if resultJSONRaw != nil {
		json.Unmarshal(resultJSONRaw, &task.Result)
	}
	if assignedNodeID != nil {
		task.AssignedNodeID = assignedNodeID
	}
	if errorMsg != nil {
		task.ErrorMessage = errorMsg
	}
	if startedAt != nil {
		task.StartedAt = startedAt
	}
	if completedAt != nil {
		task.CompletedAt = completedAt
	}

	return &task, nil
}

/* GetPendingTasksCount returns the count of pending tasks */
func (tq *TaskQueue) GetPendingTasksCount(ctx context.Context, taskType string) (int, error) {
	var query string
	var args []interface{}

	if taskType != "" {
		query = `SELECT COUNT(*) FROM neuronip.cluster_tasks WHERE status = 'pending' AND task_type = $1`
		args = []interface{}{taskType}
	} else {
		query = `SELECT COUNT(*) FROM neuronip.cluster_tasks WHERE status = 'pending'`
		args = []interface{}{}
	}

	var count int
	err := tq.pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

/* Task represents a distributed task */
type Task struct {
	ID            uuid.UUID              `json:"id"`
	TaskType      string                 `json:"task_type"`
	Priority      int                    `json:"priority"` // Higher = more priority
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Status        string                 `json:"status"` // "pending", "processing", "completed", "failed"
	AssignedNodeID *string               `json:"assigned_node_id,omitempty"`
	Result        map[string]interface{} `json:"result,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}
