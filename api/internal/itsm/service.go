package itsm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Incident represents an ITSM incident */
type Incident struct {
	ID                  uuid.UUID              `json:"id"`
	Number              string                 `json:"number"`
	Title               string                 `json:"title"`
	Description         *string                `json:"description,omitempty"`
	Priority            string                 `json:"priority"`
	Status              string                 `json:"status"`
	AssigneeID          *string                `json:"assignee_id,omitempty"`
	RequesterID         string                 `json:"requester_id"`
	RunbookID           *uuid.UUID             `json:"runbook_id,omitempty"`
	WorkflowExecutionID *uuid.UUID             `json:"workflow_execution_id,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	ResolvedAt          *time.Time             `json:"resolved_at,omitempty"`
	ClosedAt            *time.Time             `json:"closed_at,omitempty"`
}

/* Change represents an ITSM change */
type Change struct {
	ID                  uuid.UUID              `json:"id"`
	Number              string                 `json:"number"`
	Title               string                 `json:"title"`
	Description         *string                `json:"description,omitempty"`
	ChangeType          string                 `json:"change_type"`
	Status              string                 `json:"status"`
	RequesterID         string                 `json:"requester_id"`
	ApproverID          *string                `json:"approver_id,omitempty"`
	ApprovedAt          *time.Time             `json:"approved_at,omitempty"`
	ScheduledStart      *time.Time             `json:"scheduled_start,omitempty"`
	ScheduledEnd        *time.Time             `json:"scheduled_end,omitempty"`
	RunbookID           *uuid.UUID             `json:"runbook_id,omitempty"`
	WorkflowExecutionID *uuid.UUID             `json:"workflow_execution_id,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
}

/* Runbook represents an ITSM runbook */
type Runbook struct {
	ID                uuid.UUID                `json:"id"`
	Name              string                   `json:"name"`
	Description       *string                  `json:"description,omitempty"`
	WorkflowID        uuid.UUID                `json:"workflow_id"`
	TriggerConditions []map[string]interface{} `json:"trigger_conditions,omitempty"`
	Enabled           bool                     `json:"enabled"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

/* Service provides ITSM (incidents, changes, runbooks) */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new ITSM service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* CreateIncident creates an incident */
func (s *Service) CreateIncident(ctx context.Context, title, description, priority, requesterID string, assigneeID *string, runbookID *uuid.UUID) (*Incident, error) {
	if priority == "" {
		priority = "medium"
	}
	id := uuid.New()
	now := time.Now()
	number := fmt.Sprintf("INC-%d", now.Unix()%1000000)
	metaJSON, _ := json.Marshal(map[string]interface{}{})
	query := `
		INSERT INTO neuronip.itsm_incidents (id, number, title, description, priority, status, assignee_id, requester_id, runbook_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'new', $6, $7, $8, $9, $10, $11)
		RETURNING id, number, title, description, priority, status, assignee_id, requester_id, runbook_id, workflow_execution_id, metadata, created_at, updated_at, resolved_at, closed_at
	`
	var i Incident
	var desc *string
	var metaRaw []byte
	var resolvedAt, closedAt *time.Time
	err := s.pool.QueryRow(ctx, query, id, number, title, nullStr(description), priority, assigneeID, requesterID, runbookID, metaJSON, now, now).Scan(
		&i.ID, &i.Number, &i.Title, &desc, &i.Priority, &i.Status, &i.AssigneeID, &i.RequesterID, &i.RunbookID, &i.WorkflowExecutionID, &metaRaw, &i.CreatedAt, &i.UpdatedAt, &resolvedAt, &closedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}
	i.Description = desc
	i.ResolvedAt = resolvedAt
	i.ClosedAt = closedAt
	json.Unmarshal(metaRaw, &i.Metadata)
	return &i, nil
}

/* ListIncidents lists incidents */
func (s *Service) ListIncidents(ctx context.Context, status, assigneeID string) ([]Incident, error) {
	query := `SELECT id, number, title, description, priority, status, assignee_id, requester_id, runbook_id, workflow_execution_id, metadata, created_at, updated_at, resolved_at, closed_at FROM neuronip.itsm_incidents WHERE 1=1`
	args := []interface{}{}
	n := 1
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", n)
		args = append(args, status)
		n++
	}
	if assigneeID != "" {
		query += fmt.Sprintf(" AND assignee_id = $%d", n)
		args = append(args, assigneeID)
		n++
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Incident
	for rows.Next() {
		var i Incident
		var desc *string
		var metaRaw []byte
		var resolvedAt, closedAt *time.Time
		if err := rows.Scan(&i.ID, &i.Number, &i.Title, &desc, &i.Priority, &i.Status, &i.AssigneeID, &i.RequesterID, &i.RunbookID, &i.WorkflowExecutionID, &metaRaw, &i.CreatedAt, &i.UpdatedAt, &resolvedAt, &closedAt); err != nil {
			return nil, err
		}
		i.Description = desc
		i.ResolvedAt = resolvedAt
		i.ClosedAt = closedAt
		json.Unmarshal(metaRaw, &i.Metadata)
		list = append(list, i)
	}
	return list, rows.Err()
}

/* CreateChange creates a change */
func (s *Service) CreateChange(ctx context.Context, title, description, changeType, requesterID string, scheduledStart, scheduledEnd *time.Time, runbookID *uuid.UUID) (*Change, error) {
	if changeType == "" {
		changeType = "normal"
	}
	id := uuid.New()
	now := time.Now()
	number := fmt.Sprintf("CHG-%d", now.Unix()%1000000)
	metaJSON, _ := json.Marshal(map[string]interface{}{})
	query := `
		INSERT INTO neuronip.itsm_changes (id, number, title, description, change_type, status, requester_id, scheduled_start, scheduled_end, runbook_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, number, title, description, change_type, status, requester_id, approver_id, approved_at, scheduled_start, scheduled_end, runbook_id, workflow_execution_id, metadata, created_at, updated_at, completed_at
	`
	var c Change
	var desc *string
	var metaRaw []byte
	var approvedAt, completedAt *time.Time
	err := s.pool.QueryRow(ctx, query, id, number, title, nullStr(description), changeType, requesterID, scheduledStart, scheduledEnd, runbookID, metaJSON, now, now).Scan(
		&c.ID, &c.Number, &c.Title, &desc, &c.ChangeType, &c.Status, &c.RequesterID, &c.ApproverID, &approvedAt, &c.ScheduledStart, &c.ScheduledEnd, &c.RunbookID, &c.WorkflowExecutionID, &metaRaw, &c.CreatedAt, &c.UpdatedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create change: %w", err)
	}
	c.Description = desc
	c.ApprovedAt = approvedAt
	c.CompletedAt = completedAt
	json.Unmarshal(metaRaw, &c.Metadata)
	return &c, nil
}

/* ListChanges lists changes */
func (s *Service) ListChanges(ctx context.Context, status string) ([]Change, error) {
	query := `SELECT id, number, title, description, change_type, status, requester_id, approver_id, approved_at, scheduled_start, scheduled_end, runbook_id, workflow_execution_id, metadata, created_at, updated_at, completed_at FROM neuronip.itsm_changes WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		query += " AND status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Change
	for rows.Next() {
		var c Change
		var desc *string
		var metaRaw []byte
		var approvedAt, completedAt *time.Time
		if err := rows.Scan(&c.ID, &c.Number, &c.Title, &desc, &c.ChangeType, &c.Status, &c.RequesterID, &c.ApproverID, &approvedAt, &c.ScheduledStart, &c.ScheduledEnd, &c.RunbookID, &c.WorkflowExecutionID, &metaRaw, &c.CreatedAt, &c.UpdatedAt, &completedAt); err != nil {
			return nil, err
		}
		c.Description = desc
		c.ApprovedAt = approvedAt
		c.CompletedAt = completedAt
		json.Unmarshal(metaRaw, &c.Metadata)
		list = append(list, c)
	}
	return list, rows.Err()
}

/* CreateRunbook creates a runbook */
func (s *Service) CreateRunbook(ctx context.Context, name, description string, workflowID uuid.UUID, triggerConditions []map[string]interface{}) (*Runbook, error) {
	id := uuid.New()
	now := time.Now()
	condJSON, _ := json.Marshal(triggerConditions)
	query := `
		INSERT INTO neuronip.itsm_runbooks (id, name, description, workflow_id, trigger_conditions, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, $7)
		RETURNING id, name, description, workflow_id, trigger_conditions, enabled, created_at, updated_at
	`
	var r Runbook
	var desc *string
	var condRaw []byte
	err := s.pool.QueryRow(ctx, query, id, name, nullStr(description), workflowID, condJSON, now, now).Scan(
		&r.ID, &r.Name, &desc, &r.WorkflowID, &condRaw, &r.Enabled, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create runbook: %w", err)
	}
	r.Description = desc
	json.Unmarshal(condRaw, &r.TriggerConditions)
	return &r, nil
}

/* ListRunbooks lists runbooks */
func (s *Service) ListRunbooks(ctx context.Context) ([]Runbook, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, description, workflow_id, trigger_conditions, enabled, created_at, updated_at FROM neuronip.itsm_runbooks ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Runbook
	for rows.Next() {
		var r Runbook
		var desc *string
		var condRaw []byte
		if err := rows.Scan(&r.ID, &r.Name, &desc, &r.WorkflowID, &condRaw, &r.Enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Description = desc
		json.Unmarshal(condRaw, &r.TriggerConditions)
		list = append(list, r)
	}
	return list, rows.Err()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
