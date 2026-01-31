package onboarding

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* OnboardingService provides enterprise onboarding functionality */
type OnboardingService struct {
	pool *pgxpool.Pool
}

/* NewOnboardingService creates a new onboarding service */
func NewOnboardingService(pool *pgxpool.Pool) *OnboardingService {
	return &OnboardingService{pool: pool}
}

/* StartOnboarding starts an onboarding workflow */
func (os *OnboardingService) StartOnboarding(ctx context.Context, workspaceID uuid.UUID, workflowID uuid.UUID, userID string) (*OnboardingProgress, error) {
	progress := &OnboardingProgress{
		ID:                uuid.New(),
		WorkspaceID:       &workspaceID,
		UserID:            &userID,
		WorkflowID:        workflowID,
		Status:            "in_progress",
		ProgressPercentage: 0,
		StartedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	completedStepsJSON, _ := json.Marshal([]string{})

	query := `
		INSERT INTO neuronip.onboarding_progress 
		(id, workspace_id, user_id, workflow_id, current_step, completed_steps, progress_percentage,
		 status, started_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, workspace_id, user_id, workflow_id, current_step, completed_steps,
		          progress_percentage, status, started_at, updated_at`

	var completedStepsJSONRaw json.RawMessage
	err := os.pool.QueryRow(ctx, query,
		progress.ID, progress.WorkspaceID, progress.UserID, progress.WorkflowID,
		progress.CurrentStep, completedStepsJSON, progress.ProgressPercentage,
		progress.Status, progress.StartedAt, progress.UpdatedAt,
	).Scan(
		&progress.ID, &progress.WorkspaceID, &progress.UserID, &progress.WorkflowID,
		&progress.CurrentStep, &completedStepsJSONRaw, &progress.ProgressPercentage,
		&progress.Status, &progress.StartedAt, &progress.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start onboarding: %w", err)
	}

	if completedStepsJSONRaw != nil {
		json.Unmarshal(completedStepsJSONRaw, &progress.CompletedSteps)
	}

	return progress, nil
}

/* CompleteStep completes an onboarding step */
func (os *OnboardingService) CompleteStep(ctx context.Context, progressID uuid.UUID, stepID string) error {
	// Get current progress
	progress, err := os.GetProgress(ctx, progressID)
	if err != nil {
		return err
	}

	// Add step to completed
	completedSteps := progress.CompletedSteps
	if completedSteps == nil {
		completedSteps = []string{}
	}
	completedSteps = append(completedSteps, stepID)

	completedStepsJSON, _ := json.Marshal(completedSteps)

	// Update progress
	query := `
		UPDATE neuronip.onboarding_progress
		SET completed_steps = $1, updated_at = NOW()
		WHERE id = $2`

	_, err = os.pool.Exec(ctx, query, completedStepsJSON, progressID)
	return err
}

/* GetProgress retrieves onboarding progress */
func (os *OnboardingService) GetProgress(ctx context.Context, progressID uuid.UUID) (*OnboardingProgress, error) {
	query := `
		SELECT id, workspace_id, user_id, workflow_id, current_step, completed_steps,
		       progress_percentage, status, started_at, completed_at, updated_at
		FROM neuronip.onboarding_progress
		WHERE id = $1`

	var progress OnboardingProgress
	var completedStepsJSONRaw json.RawMessage
	var completedAt *time.Time

	err := os.pool.QueryRow(ctx, query, progressID).Scan(
		&progress.ID, &progress.WorkspaceID, &progress.UserID, &progress.WorkflowID,
		&progress.CurrentStep, &completedStepsJSONRaw, &progress.ProgressPercentage,
		&progress.Status, &progress.StartedAt, &completedAt, &progress.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("progress not found: %w", err)
	}

	if completedStepsJSONRaw != nil {
		json.Unmarshal(completedStepsJSONRaw, &progress.CompletedSteps)
	}
	if completedAt != nil {
		progress.CompletedAt = completedAt
	}

	return &progress, nil
}

/* OnboardingProgress represents onboarding progress */
type OnboardingProgress struct {
	ID                uuid.UUID   `json:"id"`
	WorkspaceID       *uuid.UUID  `json:"workspace_id,omitempty"`
	UserID            *string     `json:"user_id,omitempty"`
	WorkflowID        uuid.UUID   `json:"workflow_id"`
	CurrentStep       *string     `json:"current_step,omitempty"`
	CompletedSteps    []string    `json:"completed_steps"`
	ProgressPercentage int        `json:"progress_percentage"`
	Status            string      `json:"status"`
	StartedAt         time.Time   `json:"started_at"`
	CompletedAt       *time.Time  `json:"completed_at,omitempty"`
	UpdatedAt         time.Time   `json:"updated_at"`
}
