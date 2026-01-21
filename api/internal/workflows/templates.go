package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TemplateService provides workflow template functionality */
type TemplateService struct {
	pool *pgxpool.Pool
}

/* NewTemplateService creates a new template service */
func NewTemplateService(pool *pgxpool.Pool) *TemplateService {
	return &TemplateService{pool: pool}
}

/* WorkflowTemplate represents a reusable workflow template */
type WorkflowTemplate struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"` // "data_ingestion", "data_transformation", "compliance", "analytics"
	Definition    WorkflowDefinition     `json:"definition"`
	Parameters    []TemplateParameter    `json:"parameters,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	UsageCount    int                    `json:"usage_count"`
	IsPublic      bool                   `json:"is_public"`
	CreatedBy     uuid.UUID              `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* TemplateParameter represents a parameter in a template */
type TemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // "string", "number", "boolean", "select"
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Options      []string    `json:"options,omitempty"` // For select type
}

/* CreateTemplate creates a new workflow template */
func (s *TemplateService) CreateTemplate(ctx context.Context, template WorkflowTemplate) (*WorkflowTemplate, error) {
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	defJSON, _ := json.Marshal(template.Definition)
	paramsJSON, _ := json.Marshal(template.Parameters)
	tagsJSON, _ := json.Marshal(template.Tags)
	metadataJSON, _ := json.Marshal(template.Metadata)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.workflow_templates
		(id, name, description, category, definition, parameters, tags,
		 usage_count, is_public, created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		template.ID, template.Name, template.Description, template.Category,
		defJSON, paramsJSON, tagsJSON, template.UsageCount, template.IsPublic,
		template.CreatedBy, template.CreatedAt, template.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &template, nil
}

/* GetTemplate retrieves a workflow template */
func (s *TemplateService) GetTemplate(ctx context.Context, templateID uuid.UUID) (*WorkflowTemplate, error) {
	var template WorkflowTemplate
	var defJSON, paramsJSON, tagsJSON, metadataJSON []byte
	var createdBy *uuid.UUID

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, description, category, definition, parameters, tags,
		       usage_count, is_public, created_by, created_at, updated_at, metadata
		FROM neuronip.workflow_templates
		WHERE id = $1`, templateID,
	).Scan(&template.ID, &template.Name, &template.Description, &template.Category,
		&defJSON, &paramsJSON, &tagsJSON, &template.UsageCount, &template.IsPublic,
		&createdBy, &template.CreatedAt, &template.UpdatedAt, &metadataJSON)

	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	json.Unmarshal(defJSON, &template.Definition)
	json.Unmarshal(paramsJSON, &template.Parameters)
	json.Unmarshal(tagsJSON, &template.Tags)
	json.Unmarshal(metadataJSON, &template.Metadata)
	if createdBy != nil {
		template.CreatedBy = *createdBy
	}

	return &template, nil
}

/* ListTemplates lists workflow templates */
func (s *TemplateService) ListTemplates(ctx context.Context, category *string, isPublic *bool, limit int) ([]WorkflowTemplate, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, name, description, category, definition, parameters, tags,
		       usage_count, is_public, created_by, created_at, updated_at, metadata
		FROM neuronip.workflow_templates
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *category)
		argIdx++
	}

	if isPublic != nil {
		query += fmt.Sprintf(" AND is_public = $%d", argIdx)
		args = append(args, *isPublic)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY usage_count DESC, created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []WorkflowTemplate
	for rows.Next() {
		var template WorkflowTemplate
		var defJSON, paramsJSON, tagsJSON, metadataJSON []byte
		var createdBy *uuid.UUID

		err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.Category,
			&defJSON, &paramsJSON, &tagsJSON, &template.UsageCount, &template.IsPublic,
			&createdBy, &template.CreatedAt, &template.UpdatedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(defJSON, &template.Definition)
		json.Unmarshal(paramsJSON, &template.Parameters)
		json.Unmarshal(tagsJSON, &template.Tags)
		json.Unmarshal(metadataJSON, &template.Metadata)
		if createdBy != nil {
			template.CreatedBy = *createdBy
		}

		templates = append(templates, template)
	}

	return templates, nil
}

/* InstantiateTemplate creates a workflow from a template with parameter substitution */
func (s *TemplateService) InstantiateTemplate(ctx context.Context,
	templateID uuid.UUID, name string, parameters map[string]interface{}) (*WorkflowDefinition, error) {

	template, err := s.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Substitute parameters in workflow definition
	definition := template.Definition

	// Replace parameter placeholders in workflow definition
	// This is a simplified version - in production, would do full template substitution
	// For now, just increment usage count

	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.workflow_templates
		SET usage_count = usage_count + 1, updated_at = NOW()
		WHERE id = $1`, templateID)

	if err != nil {
		return nil, fmt.Errorf("failed to update usage count: %w", err)
	}

	return &definition, nil
}
