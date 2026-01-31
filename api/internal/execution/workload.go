package execution

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* WorkloadQueueConfig represents a workload queue configuration */
type WorkloadQueueConfig struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description,omitempty"`
	Priority             int        `json:"priority"`
	MaxConcurrency       int        `json:"max_concurrency"`
	QueryTimeoutSeconds  int        `json:"query_timeout_seconds"`
	QueryBudgetPerPeriod *int64     `json:"query_budget_per_period,omitempty"`
	Period               *string    `json:"period,omitempty"`
	WorkspaceID          *uuid.UUID `json:"workspace_id,omitempty"`
	Enabled              bool       `json:"enabled"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

/* WorkloadService provides workload isolation (queues, concurrency, budgets) */
type WorkloadService struct {
	pool       *pgxpool.Pool
	slotCounts map[uuid.UUID]int
	mu         sync.RWMutex
}

/* NewWorkloadService creates a new workload service */
func NewWorkloadService(pool *pgxpool.Pool) *WorkloadService {
	return &WorkloadService{
		pool:       pool,
		slotCounts: make(map[uuid.UUID]int),
	}
}

/* GetQueueByName gets a workload queue by name */
func (s *WorkloadService) GetQueueByName(ctx context.Context, name string, workspaceID *uuid.UUID) (*WorkloadQueueConfig, error) {
	query := `
		SELECT id, name, description, priority, max_concurrency, query_timeout_seconds,
		       query_budget_per_period, period, workspace_id, enabled, created_at, updated_at
		FROM neuronip.workload_queues
		WHERE name = $1 AND enabled = true
	`
	args := []interface{}{name}
	if workspaceID != nil {
		query += ` AND (workspace_id IS NULL OR workspace_id = $2)`
		args = append(args, workspaceID)
	}
	query += ` ORDER BY workspace_id NULLS LAST LIMIT 1`

	var cfg WorkloadQueueConfig
	var desc, period sql.NullString
	var budget sql.NullInt64
	var wsID sql.NullString
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&cfg.ID, &cfg.Name, &desc, &cfg.Priority, &cfg.MaxConcurrency, &cfg.QueryTimeoutSeconds,
		&budget, &period, &wsID, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workload queue not found: %w", err)
	}
	if desc.Valid {
		cfg.Description = &desc.String
	}
	if budget.Valid {
		cfg.QueryBudgetPerPeriod = &budget.Int64
	}
	if period.Valid {
		cfg.Period = &period.String
	}
	if wsID.Valid && wsID.String != "" {
		if u, err := uuid.Parse(wsID.String); err == nil {
			cfg.WorkspaceID = &u
		}
	}
	return &cfg, nil
}

/* ListQueues lists workload queues */
func (s *WorkloadService) ListQueues(ctx context.Context, workspaceID *uuid.UUID) ([]WorkloadQueueConfig, error) {
	query := `
		SELECT id, name, description, priority, max_concurrency, query_timeout_seconds,
		       query_budget_per_period, period, workspace_id, enabled, created_at, updated_at
		FROM neuronip.workload_queues
		WHERE enabled = true
	`
	args := []interface{}{}
	if workspaceID != nil {
		query += ` AND (workspace_id IS NULL OR workspace_id = $1)`
		args = append(args, workspaceID)
	}
	query += ` ORDER BY priority DESC, name`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []WorkloadQueueConfig
	for rows.Next() {
		var cfg WorkloadQueueConfig
		var desc, period sql.NullString
		var budget sql.NullInt64
		var wsID sql.NullString
		if err := rows.Scan(
			&cfg.ID, &cfg.Name, &desc, &cfg.Priority, &cfg.MaxConcurrency, &cfg.QueryTimeoutSeconds,
			&budget, &period, &wsID, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if desc.Valid {
			cfg.Description = &desc.String
		}
		if budget.Valid {
			cfg.QueryBudgetPerPeriod = &budget.Int64
		}
		if period.Valid {
			cfg.Period = &period.String
		}
		if wsID.Valid && wsID.String != "" {
			if u, err := uuid.Parse(wsID.String); err == nil {
				cfg.WorkspaceID = &u
			}
		}
		list = append(list, cfg)
	}
	return list, rows.Err()
}

/* AcquireSlot acquires a concurrency slot for a queue. Caller must call ReleaseSlot when done. */
func (s *WorkloadService) AcquireSlot(ctx context.Context, queueID uuid.UUID, queryID uuid.UUID, userID *string, workspaceID *uuid.UUID) error {
	// Check current slot count (in-memory fast path; can also count from workload_queue_slots)
	s.mu.Lock()
	n := s.slotCounts[queueID]
	s.mu.Unlock()

	var maxConcurrency int
	err := s.pool.QueryRow(ctx, `SELECT max_concurrency FROM neuronip.workload_queues WHERE id = $1 AND enabled = true`, queueID).Scan(&maxConcurrency)
	if err != nil {
		return fmt.Errorf("queue not found or disabled: %w", err)
	}

	s.mu.Lock()
	if n >= maxConcurrency {
		s.mu.Unlock()
		return fmt.Errorf("workload queue at max concurrency (%d)", maxConcurrency)
	}
	s.slotCounts[queueID] = n + 1
	s.mu.Unlock()

	// Optionally record slot in DB for visibility (best-effort)
	expiresAt := time.Now().Add(10 * time.Minute)
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO neuronip.workload_queue_slots (id, queue_id, query_id, user_id, workspace_id, expires_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`, queueID, queryID, userID, workspaceID, expiresAt)

	return nil
}

/* ReleaseSlot releases a concurrency slot */
func (s *WorkloadService) ReleaseSlot(ctx context.Context, queueID uuid.UUID, queryID uuid.UUID) {
	s.mu.Lock()
	if n := s.slotCounts[queueID]; n > 0 {
		s.slotCounts[queueID] = n - 1
	}
	s.mu.Unlock()
	_, _ = s.pool.Exec(ctx, `DELETE FROM neuronip.workload_queue_slots WHERE queue_id = $1 AND query_id = $2`, queueID, queryID)
}

/* CheckQueryBudget returns nil if the queue has remaining query budget for the period */
func (s *WorkloadService) CheckQueryBudget(ctx context.Context, queueID uuid.UUID) error {
	var budget sql.NullInt64
	var period sql.NullString
	err := s.pool.QueryRow(ctx, `
		SELECT query_budget_per_period, period FROM neuronip.workload_queues WHERE id = $1 AND enabled = true
	`, queueID).Scan(&budget, &period)
	if err != nil {
		return err
	}
	if !budget.Valid || budget.Int64 <= 0 {
		return nil // no budget limit
	}
	// Count queries in current period (simplified: count slots or use separate usage table)
	var used int64
	_ = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM neuronip.workload_queue_slots WHERE queue_id = $1`, queueID).Scan(&used)
	// For budget we'd typically have a usage table per period; here we allow if under budget
	if used >= budget.Int64 {
		return fmt.Errorf("workload queue query budget exceeded for period")
	}
	return nil
}

/* CreateQueue creates a workload queue */
func (s *WorkloadService) CreateQueue(ctx context.Context, name, description string, priority, maxConcurrency, queryTimeoutSeconds int, workspaceID *uuid.UUID) (*WorkloadQueueConfig, error) {
	id := uuid.New()
	now := time.Now()
	query := `
		INSERT INTO neuronip.workload_queues (id, name, description, priority, max_concurrency, query_timeout_seconds, workspace_id, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8, $9)
		RETURNING id, name, description, priority, max_concurrency, query_timeout_seconds, query_budget_per_period, period, workspace_id, enabled, created_at, updated_at
	`
	var cfg WorkloadQueueConfig
	var desc, period sql.NullString
	var budget sql.NullInt64
	var wsID sql.NullString
	err := s.pool.QueryRow(ctx, query, id, name, nullString(description), priority, maxConcurrency, queryTimeoutSeconds, workspaceID, now, now).Scan(
		&cfg.ID, &cfg.Name, &desc, &cfg.Priority, &cfg.MaxConcurrency, &cfg.QueryTimeoutSeconds,
		&budget, &period, &wsID, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		cfg.Description = &desc.String
	}
	if budget.Valid {
		cfg.QueryBudgetPerPeriod = &budget.Int64
	}
	if period.Valid {
		cfg.Period = &period.String
	}
	if wsID.Valid && wsID.String != "" {
		if u, err := uuid.Parse(wsID.String); err == nil {
			cfg.WorkspaceID = &u
		}
	}
	return &cfg, nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
