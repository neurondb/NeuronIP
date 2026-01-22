package governance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PromptTemplateService provides prompt template versioning functionality */
type PromptTemplateService struct {
	pool *pgxpool.Pool
}

/* NewPromptTemplateService creates a new prompt template service */
func NewPromptTemplateService(pool *pgxpool.Pool) *PromptTemplateService {
	return &PromptTemplateService{pool: pool}
}

/* PromptTemplate represents a prompt template */
type PromptTemplate struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	TemplateText    string                 `json:"template_text"`
	Variables       []string               `json:"variables"`
	Description     *string                `json:"description,omitempty"`
	Status          string                 `json:"status"` // draft, pending_approval, approved, deprecated
	ApprovedBy      *string                `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty"`
	ParentTemplateID *uuid.UUID             `json:"parent_template_id,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* CreatePromptTemplate creates a new prompt template version */
func (s *PromptTemplateService) CreatePromptTemplate(ctx context.Context, name string, version string, templateText string, variables []string, description *string, createdBy string, parentTemplateID *uuid.UUID, metadata map[string]interface{}) (*PromptTemplate, error) {
	id := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)
	now := time.Now()

	query := `
		INSERT INTO neuronip.prompt_templates 
		(id, name, version, template_text, variables, description, status, parent_template_id, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7, $8, $9, $10, $11)
		ON CONFLICT (name, version) 
		DO UPDATE SET 
			template_text = EXCLUDED.template_text,
			variables = EXCLUDED.variables,
			description = EXCLUDED.description,
			parent_template_id = EXCLUDED.parent_template_id,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
		RETURNING id, name, version, template_text, variables, description, status, approved_by, approved_at, parent_template_id, metadata, created_by, created_at, updated_at`

	var prompt PromptTemplate
	var metadataJSONRaw json.RawMessage
	var descriptionVal sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	var parentID sql.NullString

	err := s.pool.QueryRow(ctx, query, id, name, version, templateText, variables, description, parentTemplateID, metadataJSON, createdBy, now, now).Scan(
		&prompt.ID, &prompt.Name, &prompt.Version, &prompt.TemplateText, &prompt.Variables,
		&descriptionVal, &prompt.Status, &approvedBy, &approvedAt, &parentID,
		&metadataJSONRaw, &prompt.CreatedBy, &prompt.CreatedAt, &prompt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create prompt template: %w", err)
	}

	if descriptionVal.Valid {
		prompt.Description = &descriptionVal.String
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &prompt.Metadata)
	}
	if approvedBy.Valid {
		prompt.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		prompt.ApprovedAt = &approvedAt.Time
	}
	if parentID.Valid {
		parentUUID, _ := uuid.Parse(parentID.String)
		prompt.ParentTemplateID = &parentUUID
	}

	return &prompt, nil
}

/* ApprovePrompt approves a prompt template version */
func (s *PromptTemplateService) ApprovePrompt(ctx context.Context, promptID uuid.UUID, approverID string) error {
	now := time.Now()

	query := `
		UPDATE neuronip.prompt_templates 
		SET status = 'approved', approved_by = $1, approved_at = $2, updated_at = $2
		WHERE id = $3`

	_, err := s.pool.Exec(ctx, query, approverID, now, promptID)
	if err != nil {
		return fmt.Errorf("failed to approve prompt: %w", err)
	}

	return nil
}

/* RequestPromptApproval requests approval for a prompt */
func (s *PromptTemplateService) RequestPromptApproval(ctx context.Context, promptID uuid.UUID, approverID string) error {
	// Update prompt status to pending_approval
	query := `
		UPDATE neuronip.prompt_templates 
		SET status = 'pending_approval', updated_at = NOW()
		WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, promptID)
	if err != nil {
		return fmt.Errorf("failed to request prompt approval: %w", err)
	}

	// Create approval record
	approvalID := uuid.New()
	approvalQuery := `
		INSERT INTO neuronip.prompt_approvals 
		(id, prompt_id, approver_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', NOW(), NOW())`

	_, err = s.pool.Exec(ctx, approvalQuery, approvalID, promptID, approverID)
	if err != nil {
		return fmt.Errorf("failed to create approval record: %w", err)
	}

	return nil
}

/* GetPromptTemplate retrieves a prompt template by ID */
func (s *PromptTemplateService) GetPromptTemplate(ctx context.Context, promptID uuid.UUID) (*PromptTemplate, error) {
	query := `
		SELECT id, name, version, template_text, variables, description, status, approved_by, approved_at, parent_template_id, metadata, created_by, created_at, updated_at
		FROM neuronip.prompt_templates
		WHERE id = $1`

	var prompt PromptTemplate
	var metadataJSONRaw json.RawMessage
	var descriptionVal sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	var parentID sql.NullString

	err := s.pool.QueryRow(ctx, query, promptID).Scan(
		&prompt.ID, &prompt.Name, &prompt.Version, &prompt.TemplateText, &prompt.Variables,
		&descriptionVal, &prompt.Status, &approvedBy, &approvedAt, &parentID,
		&metadataJSONRaw, &prompt.CreatedBy, &prompt.CreatedAt, &prompt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt template: %w", err)
	}

	if descriptionVal.Valid {
		prompt.Description = &descriptionVal.String
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &prompt.Metadata)
	}
	if approvedBy.Valid {
		prompt.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		prompt.ApprovedAt = &approvedAt.Time
	}
	if parentID.Valid {
		parentUUID, _ := uuid.Parse(parentID.String)
		prompt.ParentTemplateID = &parentUUID
	}

	return &prompt, nil
}

/* ListPrompts lists all prompt templates */
func (s *PromptTemplateService) ListPrompts(ctx context.Context, limit int) ([]PromptTemplate, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, name, version, template_text, variables, description, status,
		       approved_by, approved_at, parent_template_id, metadata, created_by, created_at, updated_at
		FROM neuronip.prompt_templates
		WHERE status != 'deprecated'
		ORDER BY name, created_at DESC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}
	defer rows.Close()

	var prompts []PromptTemplate
	for rows.Next() {
		var prompt PromptTemplate
		var desc, approvedBy sql.NullString
		var approvedAt sql.NullTime
		var parentID sql.NullString
		var metadataJSON json.RawMessage
		var variablesJSON json.RawMessage

		err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Version, &prompt.TemplateText,
			&variablesJSON, &desc, &prompt.Status, &approvedBy, &approvedAt,
			&parentID, &metadataJSON, &prompt.CreatedBy, &prompt.CreatedAt, &prompt.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if desc.Valid {
			prompt.Description = &desc.String
		}
		if approvedBy.Valid {
			prompt.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			prompt.ApprovedAt = &approvedAt.Time
		}
		if parentID.Valid {
			pID, _ := uuid.Parse(parentID.String)
			prompt.ParentTemplateID = &pID
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &prompt.Metadata)
		}
		if variablesJSON != nil {
			json.Unmarshal(variablesJSON, &prompt.Variables)
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

/* GetPromptVersions retrieves all versions of a prompt */
func (s *PromptTemplateService) GetPromptVersions(ctx context.Context, promptName string) ([]PromptTemplate, error) {
	query := `
		SELECT id, name, version, template_text, variables, description, status, approved_by, approved_at, parent_template_id, metadata, created_by, created_at, updated_at
		FROM neuronip.prompt_templates
		WHERE name = $1
		ORDER BY version DESC`

	rows, err := s.pool.Query(ctx, query, promptName)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt versions: %w", err)
	}
	defer rows.Close()

	var prompts []PromptTemplate
	for rows.Next() {
		var prompt PromptTemplate
		var metadataJSONRaw json.RawMessage
		var descriptionVal sql.NullString
		var approvedBy sql.NullString
		var approvedAt sql.NullTime
		var parentID sql.NullString

		err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Version, &prompt.TemplateText, &prompt.Variables,
			&descriptionVal, &prompt.Status, &approvedBy, &approvedAt, &parentID,
			&metadataJSONRaw, &prompt.CreatedBy, &prompt.CreatedAt, &prompt.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if descriptionVal.Valid {
			prompt.Description = &descriptionVal.String
		}
		if metadataJSONRaw != nil {
			json.Unmarshal(metadataJSONRaw, &prompt.Metadata)
		}
		if approvedBy.Valid {
			prompt.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			prompt.ApprovedAt = &approvedAt.Time
		}
		if parentID.Valid {
			parentUUID, _ := uuid.Parse(parentID.String)
			prompt.ParentTemplateID = &parentUUID
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

/* RenderPrompt renders a prompt template with variables */
func (s *PromptTemplateService) RenderPrompt(ctx context.Context, promptID uuid.UUID, variables map[string]interface{}) (string, error) {
	prompt, err := s.GetPromptTemplate(ctx, promptID)
	if err != nil {
		return "", err
	}

	if prompt.Status != "approved" {
		return "", fmt.Errorf("prompt template is not approved")
	}

	// Render template by replacing variables
	rendered := prompt.TemplateText
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		rendered = strings.ReplaceAll(rendered, placeholder, fmt.Sprintf("%v", value))
	}

	// Check for unresolved variables
	for _, varName := range prompt.Variables {
		placeholder := fmt.Sprintf("{{%s}}", varName)
		if strings.Contains(rendered, placeholder) {
			return "", fmt.Errorf("variable %s not provided", varName)
		}
	}

	return rendered, nil
}

/* ComparePromptVersions compares two prompt versions */
func (s *PromptTemplateService) ComparePromptVersions(ctx context.Context, promptName string, version1 string, version2 string) (map[string]interface{}, error) {
	v1, err := s.getPromptVersion(ctx, promptName, version1)
	if err != nil {
		return nil, fmt.Errorf("failed to get version 1: %w", err)
	}

	v2, err := s.getPromptVersion(ctx, promptName, version2)
	if err != nil {
		return nil, fmt.Errorf("failed to get version 2: %w", err)
	}

	diff := map[string]interface{}{
		"version1": map[string]interface{}{
			"version":      v1.Version,
			"template_text": v1.TemplateText,
			"variables":    v1.Variables,
			"created_at":   v1.CreatedAt,
		},
		"version2": map[string]interface{}{
			"version":      v2.Version,
			"template_text": v2.TemplateText,
			"variables":    v2.Variables,
			"created_at":   v2.CreatedAt,
		},
		"changes": map[string]interface{}{
			"template_changed": v1.TemplateText != v2.TemplateText,
			"variables_changed": !equalStringSlices(v1.Variables, v2.Variables),
		},
	}

	return diff, nil
}

/* getPromptVersion gets a specific prompt version */
func (s *PromptTemplateService) getPromptVersion(ctx context.Context, promptName string, version string) (*PromptTemplate, error) {
	query := `
		SELECT id, name, version, template_text, variables, description, status, approved_by, approved_at, parent_template_id, metadata, created_by, created_at, updated_at
		FROM neuronip.prompt_templates
		WHERE name = $1 AND version = $1`

	var prompt PromptTemplate
	var metadataJSONRaw json.RawMessage
	var descriptionVal sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	var parentID sql.NullString

	err := s.pool.QueryRow(ctx, query, promptName, version).Scan(
		&prompt.ID, &prompt.Name, &prompt.Version, &prompt.TemplateText, &prompt.Variables,
		&descriptionVal, &prompt.Status, &approvedBy, &approvedAt, &parentID,
		&metadataJSONRaw, &prompt.CreatedBy, &prompt.CreatedAt, &prompt.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt version: %w", err)
	}

	if descriptionVal.Valid {
		prompt.Description = &descriptionVal.String
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &prompt.Metadata)
	}
	if approvedBy.Valid {
		prompt.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		prompt.ApprovedAt = &approvedAt.Time
	}
	if parentID.Valid {
		parentUUID, _ := uuid.Parse(parentID.String)
		prompt.ParentTemplateID = &parentUUID
	}

	return &prompt, nil
}

/* equalStringSlices checks if two string slices are equal */
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

/* RollbackPrompt rolls back to a previous prompt version */
func (s *PromptTemplateService) RollbackPrompt(ctx context.Context, promptName string, targetVersion string) error {
	// Create a new version based on the target version
	targetPrompt, err := s.getPromptVersion(ctx, promptName, targetVersion)
	if err != nil {
		return fmt.Errorf("failed to get target version: %w", err)
	}

	// Increment version number for rollback
	newVersion := incrementVersion(targetPrompt.Version)

	// Create new version
	_, err = s.CreatePromptTemplate(
		context.Background(),
		promptName,
		newVersion,
		targetPrompt.TemplateText,
		targetPrompt.Variables,
		targetPrompt.Description,
		targetPrompt.CreatedBy,
		&targetPrompt.ID, // Set parent to target version
		targetPrompt.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create rollback version: %w", err)
	}

	return nil
}

/* incrementVersion increments a version string */
func incrementVersion(version string) string {
	// Simple version incrementing (e.g., "1.0.0" -> "1.0.1")
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		parts = append(parts, "0", "0")
	}
	
	lastPart := parts[len(parts)-1]
	// Try to increment last part
	if lastPartNum, err := parseInt(lastPart); err == nil {
		parts[len(parts)-1] = fmt.Sprintf("%d", lastPartNum+1)
	} else {
		parts[len(parts)-1] = lastPart + ".1"
	}
	
	return strings.Join(parts, ".")
}

/* parseInt tries to parse a string as int */
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
