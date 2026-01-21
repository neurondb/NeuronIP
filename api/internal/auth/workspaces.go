package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* WorkspaceService provides workspace management functionality */
type WorkspaceService struct {
	queries *db.Queries
}

/* NewWorkspaceService creates a new workspace service */
func NewWorkspaceService(queries *db.Queries) *WorkspaceService {
	return &WorkspaceService{
		queries: queries,
	}
}

/* Workspace represents a workspace */
type Workspace struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

/* CreateWorkspace creates a new workspace */
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, orgID uuid.UUID, name string, description *string) (*Workspace, error) {
	query := `
		INSERT INTO neuronip.workspaces (organization_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, name, description, created_at, updated_at
	`
	var workspace Workspace
	err := s.queries.DB.QueryRow(ctx, query, orgID, name, description).Scan(
		&workspace.ID, &workspace.OrganizationID, &workspace.Name,
		&workspace.Description, &workspace.CreatedAt, &workspace.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	return &workspace, nil
}

/* GetWorkspace retrieves a workspace by ID */
func (s *WorkspaceService) GetWorkspace(ctx context.Context, workspaceID uuid.UUID) (*Workspace, error) {
	query := `
		SELECT id, organization_id, name, description, created_at, updated_at
		FROM neuronip.workspaces
		WHERE id = $1
	`
	var workspace Workspace
	err := s.queries.DB.QueryRow(ctx, query, workspaceID).Scan(
		&workspace.ID, &workspace.OrganizationID, &workspace.Name,
		&workspace.Description, &workspace.CreatedAt, &workspace.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}
	return &workspace, nil
}

/* ListWorkspaces lists workspaces for an organization */
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, orgID uuid.UUID) ([]Workspace, error) {
	query := `
		SELECT id, organization_id, name, description, created_at, updated_at
		FROM neuronip.workspaces
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.queries.DB.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var workspace Workspace
		err := rows.Scan(
			&workspace.ID, &workspace.OrganizationID, &workspace.Name,
			&workspace.Description, &workspace.CreatedAt, &workspace.UpdatedAt,
		)
		if err != nil {
			continue
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

/* AddUserToWorkspace adds a user to a workspace with a role */
func (s *WorkspaceService) AddUserToWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO neuronip.workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, user_id)
		DO UPDATE SET role = $3, updated_at = NOW()
	`
	_, err := s.queries.DB.Exec(ctx, query, workspaceID, userID, role)
	return err
}

/* RemoveUserFromWorkspace removes a user from a workspace */
func (s *WorkspaceService) RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `DELETE FROM neuronip.workspace_members WHERE workspace_id = $1 AND user_id = $2`
	_, err := s.queries.DB.Exec(ctx, query, workspaceID, userID)
	return err
}

/* GetWorkspaceRole gets a user's role in a workspace */
func (s *WorkspaceService) GetWorkspaceRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	query := `SELECT role FROM neuronip.workspace_members WHERE workspace_id = $1 AND user_id = $2`
	var role string
	err := s.queries.DB.QueryRow(ctx, query, workspaceID, userID).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("user not in workspace: %w", err)
	}
	return role, nil
}
