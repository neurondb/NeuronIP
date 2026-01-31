package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DecisionDashboard represents a decision intelligence dashboard */
type DecisionDashboard struct {
	ID                  uuid.UUID                `json:"id"`
	Name                string                   `json:"name"`
	Description         *string                  `json:"description,omitempty"`
	WorkspaceID         *uuid.UUID               `json:"workspace_id,omitempty"`
	OwnerID             string                   `json:"owner_id"`
	Layout              []map[string]interface{} `json:"layout"`
	MetricIDs           []uuid.UUID              `json:"metric_ids,omitempty"`
	EvidenceSources     []map[string]interface{} `json:"evidence_sources,omitempty"`
	LineageResourceType *string                  `json:"lineage_resource_type,omitempty"`
	LineageResourceID   *uuid.UUID               `json:"lineage_resource_id,omitempty"`
	WorkflowIDs         []uuid.UUID              `json:"workflow_ids,omitempty"`
	ApprovalWorkflowID  *uuid.UUID               `json:"approval_workflow_id,omitempty"`
	Visibility          string                   `json:"visibility"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
}

/* DecisionDashboardRun represents a dashboard run snapshot */
type DecisionDashboardRun struct {
	ID          uuid.UUID              `json:"id"`
	DashboardID uuid.UUID              `json:"dashboard_id"`
	TriggeredBy string                 `json:"triggered_by"`
	Snapshot    map[string]interface{} `json:"snapshot"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
}

/* DecisionDashboardService provides decision dashboard CRUD and runs */
type DecisionDashboardService struct {
	pool *pgxpool.Pool
}

/* NewDecisionDashboardService creates a new decision dashboard service */
func NewDecisionDashboardService(pool *pgxpool.Pool) *DecisionDashboardService {
	return &DecisionDashboardService{pool: pool}
}

/* Create creates a decision dashboard */
func (s *DecisionDashboardService) Create(ctx context.Context, name, description, ownerID string, workspaceID *uuid.UUID, layout []map[string]interface{}, metricIDs []uuid.UUID, evidenceSources []map[string]interface{}, lineageResourceType *string, lineageResourceID *uuid.UUID, workflowIDs []uuid.UUID, approvalWorkflowID *uuid.UUID, visibility string) (*DecisionDashboard, error) {
	if visibility == "" {
		visibility = "private"
	}
	id := uuid.New()
	now := time.Now()
	layoutJSON, _ := json.Marshal(layout)
	metricJSON, _ := json.Marshal(metricIDs)
	evidenceJSON, _ := json.Marshal(evidenceSources)
	workflowJSON, _ := json.Marshal(workflowIDs)
	var desc interface{}
	if description != "" {
		desc = description
	}
	query := `
		INSERT INTO neuronip.decision_dashboards (id, name, description, workspace_id, owner_id, layout, metric_ids, evidence_sources, lineage_resource_type, lineage_resource_id, workflow_ids, approval_workflow_id, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, name, description, workspace_id, owner_id, layout, metric_ids, evidence_sources, lineage_resource_type, lineage_resource_id, workflow_ids, approval_workflow_id, visibility, created_at, updated_at
	`
	var d DecisionDashboard
	var descScanned *string
	var layoutRaw, metricRaw, evidenceRaw, workflowRaw []byte
	var lineageType *string
	var lineageID *uuid.UUID
	var approvalID *uuid.UUID
	err := s.pool.QueryRow(ctx, query, id, name, desc, workspaceID, ownerID, layoutJSON, metricJSON, evidenceJSON, lineageResourceType, lineageResourceID, workflowJSON, approvalWorkflowID, visibility, now, now).Scan(
		&d.ID, &d.Name, &descScanned, &d.WorkspaceID, &d.OwnerID, &layoutRaw, &metricRaw, &evidenceRaw, &lineageType, &lineageID, &workflowRaw, &approvalID, &d.Visibility, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create decision dashboard: %w", err)
	}
	d.Description = descScanned
	d.LineageResourceType = lineageType
	d.LineageResourceID = lineageID
	d.ApprovalWorkflowID = approvalID
	json.Unmarshal(layoutRaw, &d.Layout)
	json.Unmarshal(metricRaw, &d.MetricIDs)
	json.Unmarshal(evidenceRaw, &d.EvidenceSources)
	json.Unmarshal(workflowRaw, &d.WorkflowIDs)
	return &d, nil
}

/* Get returns a decision dashboard by ID */
func (s *DecisionDashboardService) Get(ctx context.Context, id uuid.UUID) (*DecisionDashboard, error) {
	query := `SELECT id, name, description, workspace_id, owner_id, layout, metric_ids, evidence_sources, lineage_resource_type, lineage_resource_id, workflow_ids, approval_workflow_id, visibility, created_at, updated_at FROM neuronip.decision_dashboards WHERE id = $1`
	var d DecisionDashboard
	var descScanned *string
	var layoutRaw, metricRaw, evidenceRaw, workflowRaw []byte
	var lineageType *string
	var lineageID *uuid.UUID
	var approvalID *uuid.UUID
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.Name, &descScanned, &d.WorkspaceID, &d.OwnerID, &layoutRaw, &metricRaw, &evidenceRaw, &lineageType, &lineageID, &workflowRaw, &approvalID, &d.Visibility, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get decision dashboard: %w", err)
	}
	d.Description = descScanned
	d.LineageResourceType = lineageType
	d.LineageResourceID = lineageID
	d.ApprovalWorkflowID = approvalID
	json.Unmarshal(layoutRaw, &d.Layout)
	json.Unmarshal(metricRaw, &d.MetricIDs)
	json.Unmarshal(evidenceRaw, &d.EvidenceSources)
	json.Unmarshal(workflowRaw, &d.WorkflowIDs)
	return &d, nil
}

/* List lists decision dashboards for workspace or owner */
func (s *DecisionDashboardService) List(ctx context.Context, workspaceID *uuid.UUID, ownerID string) ([]DecisionDashboard, error) {
	query := `SELECT id, name, description, workspace_id, owner_id, layout, metric_ids, evidence_sources, lineage_resource_type, lineage_resource_id, workflow_ids, approval_workflow_id, visibility, created_at, updated_at FROM neuronip.decision_dashboards WHERE 1=1`
	args := []interface{}{}
	n := 1
	if workspaceID != nil {
		query += fmt.Sprintf(" AND (workspace_id = $%d OR visibility = 'public')", n)
		args = append(args, workspaceID)
		n++
	}
	if ownerID != "" {
		query += fmt.Sprintf(" AND owner_id = $%d", n)
		args = append(args, ownerID)
		n++
	}
	query += " ORDER BY updated_at DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []DecisionDashboard
	for rows.Next() {
		var d DecisionDashboard
		var descScanned *string
		var layoutRaw, metricRaw, evidenceRaw, workflowRaw []byte
		var lineageType *string
		var lineageID *uuid.UUID
		var approvalID *uuid.UUID
		if err := rows.Scan(&d.ID, &d.Name, &descScanned, &d.WorkspaceID, &d.OwnerID, &layoutRaw, &metricRaw, &evidenceRaw, &lineageType, &lineageID, &workflowRaw, &approvalID, &d.Visibility, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		d.Description = descScanned
		d.LineageResourceType = lineageType
		d.LineageResourceID = lineageID
		d.ApprovalWorkflowID = approvalID
		json.Unmarshal(layoutRaw, &d.Layout)
		json.Unmarshal(metricRaw, &d.MetricIDs)
		json.Unmarshal(evidenceRaw, &d.EvidenceSources)
		json.Unmarshal(workflowRaw, &d.WorkflowIDs)
		list = append(list, d)
	}
	return list, rows.Err()
}

/* RecordRun records a decision dashboard run snapshot */
func (s *DecisionDashboardService) RecordRun(ctx context.Context, dashboardID uuid.UUID, triggeredBy string, snapshot map[string]interface{}, status string) (*DecisionDashboardRun, error) {
	if status == "" {
		status = "completed"
	}
	id := uuid.New()
	now := time.Now()
	snapshotJSON, _ := json.Marshal(snapshot)
	query := `INSERT INTO neuronip.decision_dashboard_runs (id, dashboard_id, triggered_by, snapshot, status, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, dashboard_id, triggered_by, snapshot, status, created_at`
	var r DecisionDashboardRun
	var snapshotRaw []byte
	err := s.pool.QueryRow(ctx, query, id, dashboardID, triggeredBy, snapshotJSON, status, now).Scan(&r.ID, &r.DashboardID, &r.TriggeredBy, &snapshotRaw, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("record run: %w", err)
	}
	json.Unmarshal(snapshotRaw, &r.Snapshot)
	return &r, nil
}
