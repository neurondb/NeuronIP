package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ApprovalWorkflowService provides multi-stage approval workflow functionality */
type ApprovalWorkflowService struct {
	pool *pgxpool.Pool
}

/* NewApprovalWorkflowService creates a new approval workflow service */
func NewApprovalWorkflowService(pool *pgxpool.Pool) *ApprovalWorkflowService {
	return &ApprovalWorkflowService{pool: pool}
}

/* ApprovalWorkflow represents a multi-stage approval workflow */
type ApprovalWorkflow struct {
	ID          uuid.UUID              `json:"id"`
	ResourceType string                `json:"resource_type"` // "model", "prompt", "metric"
	ResourceID   uuid.UUID             `json:"resource_id"`
	Stages      []ApprovalStage        `json:"stages"`
	Status      string                 `json:"status"` // "pending", "in_progress", "approved", "rejected"
	CurrentStage int                   `json:"current_stage"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* ApprovalStage represents a stage in the approval workflow */
type ApprovalStage struct {
	StageNumber int                    `json:"stage_number"`
	StageName   string                 `json:"stage_name"`
	Approvers   []string               `json:"approvers"`
	RequiredApprovals int               `json:"required_approvals"`
	Approvals   []Approval             `json:"approvals,omitempty"`
	Status      string                 `json:"status"` // "pending", "approved", "rejected"
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

/* Approval represents an individual approval */
type Approval struct {
	ID          uuid.UUID              `json:"id"`
	StageID     uuid.UUID              `json:"stage_id"`
	ApproverID  string                 `json:"approver_id"`
	Decision    string                 `json:"decision"` // "approve", "reject", "request_changes"
	Comments    string                 `json:"comments,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

/* CreateWorkflow creates a new approval workflow */
func (aws *ApprovalWorkflowService) CreateWorkflow(ctx context.Context, workflow ApprovalWorkflow) (*ApprovalWorkflow, error) {
	workflow.ID = uuid.New()
	workflow.Status = "pending"
	workflow.CurrentStage = 0
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()
	
	stagesJSON, _ := json.Marshal(workflow.Stages)
	metadataJSON, _ := json.Marshal(workflow.Metadata)
	
	query := `
		INSERT INTO neuronip.approval_workflows 
		(id, resource_type, resource_id, stages, status, current_stage, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	
	err := aws.pool.QueryRow(ctx, query,
		workflow.ID, workflow.ResourceType, workflow.ResourceID, stagesJSON,
		workflow.Status, workflow.CurrentStage, metadataJSON, workflow.CreatedAt, workflow.UpdatedAt,
	).Scan(&workflow.ID, &workflow.CreatedAt, &workflow.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}
	
	return &workflow, nil
}

/* SubmitApproval submits an approval for a workflow stage */
func (aws *ApprovalWorkflowService) SubmitApproval(ctx context.Context, workflowID uuid.UUID, stageNumber int, approverID, decision, comments string) error {
	approvalID := uuid.New()
	now := time.Now()
	
	// Create approval record
	query := `
		INSERT INTO neuronip.approval_workflow_approvals 
		(id, workflow_id, stage_number, approver_id, decision, comments, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := aws.pool.Exec(ctx, query, approvalID, workflowID, stageNumber, approverID, decision, comments, now)
	if err != nil {
		return fmt.Errorf("failed to submit approval: %w", err)
	}
	
	// Check if stage is complete
	if err := aws.checkStageCompletion(ctx, workflowID, stageNumber); err != nil {
		return fmt.Errorf("failed to check stage completion: %w", err)
	}
	
	return nil
}

/* checkStageCompletion checks if a stage has enough approvals */
func (aws *ApprovalWorkflowService) checkStageCompletion(ctx context.Context, workflowID uuid.UUID, stageNumber int) error {
	// Get workflow
	var stagesJSON json.RawMessage
	var currentStage int
	var status string
	
	err := aws.pool.QueryRow(ctx, `SELECT stages, current_stage, status FROM neuronip.approval_workflows WHERE id = $1`, workflowID).Scan(
		&stagesJSON, &currentStage, &status,
	)
	if err != nil {
		return err
	}
	
	var stages []ApprovalStage
	json.Unmarshal(stagesJSON, &stages)
	
	if stageNumber >= len(stages) {
		return fmt.Errorf("invalid stage number")
	}
	
	stage := stages[stageNumber]
	
	// Count approvals for this stage
	var approvalCount int
	err = aws.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM neuronip.approval_workflow_approvals 
		WHERE workflow_id = $1 AND stage_number = $2 AND decision = 'approve'
	`, workflowID, stageNumber).Scan(&approvalCount)
	if err != nil {
		return err
	}
	
	// Check if required approvals met
	if approvalCount >= stage.RequiredApprovals {
		// Mark stage as approved
		stages[stageNumber].Status = "approved"
		now := time.Now()
		stages[stageNumber].CompletedAt = &now
		
		// Update workflow
		stagesJSON, _ = json.Marshal(stages)
		newCurrentStage := stageNumber + 1
		newStatus := "in_progress"
		
		// Check if all stages complete
		if newCurrentStage >= len(stages) {
			newStatus = "approved"
		}
		
		_, err = aws.pool.Exec(ctx, `
			UPDATE neuronip.approval_workflows 
			SET stages = $1, current_stage = $2, status = $3, updated_at = NOW()
			WHERE id = $4
		`, stagesJSON, newCurrentStage, newStatus, workflowID)
		
		return err
	}
	
	return nil
}

/* GetWorkflow retrieves an approval workflow */
func (aws *ApprovalWorkflowService) GetWorkflow(ctx context.Context, workflowID uuid.UUID) (*ApprovalWorkflow, error) {
	query := `
		SELECT id, resource_type, resource_id, stages, status, current_stage, metadata, created_at, updated_at
		FROM neuronip.approval_workflows
		WHERE id = $1
	`
	
	var workflow ApprovalWorkflow
	var stagesJSON, metadataJSON json.RawMessage
	
	err := aws.pool.QueryRow(ctx, query, workflowID).Scan(
		&workflow.ID, &workflow.ResourceType, &workflow.ResourceID, &stagesJSON,
		&workflow.Status, &workflow.CurrentStage, &metadataJSON, &workflow.CreatedAt, &workflow.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	
	if stagesJSON != nil {
		json.Unmarshal(stagesJSON, &workflow.Stages)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &workflow.Metadata)
	}
	
	// Get approvals for each stage
	for i := range workflow.Stages {
		approvals, _ := aws.getStageApprovals(ctx, workflowID, i)
		workflow.Stages[i].Approvals = approvals
	}
	
	return &workflow, nil
}

/* getStageApprovals retrieves approvals for a stage */
func (aws *ApprovalWorkflowService) getStageApprovals(ctx context.Context, workflowID uuid.UUID, stageNumber int) ([]Approval, error) {
	query := `
		SELECT id, workflow_id, stage_number, approver_id, decision, comments, created_at
		FROM neuronip.approval_workflow_approvals
		WHERE workflow_id = $1 AND stage_number = $2
		ORDER BY created_at ASC
	`
	
	rows, err := aws.pool.Query(ctx, query, workflowID, stageNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var approvals []Approval
	for rows.Next() {
		var approval Approval
		var stageNum int
		
		err := rows.Scan(
			&approval.ID, &approval.StageID, &stageNum, &approval.ApproverID,
			&approval.Decision, &approval.Comments, &approval.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		approvals = append(approvals, approval)
	}
	
	return approvals, nil
}
