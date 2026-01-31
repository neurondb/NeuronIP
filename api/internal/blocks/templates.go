package blocks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PageTemplate represents a page template */
type PageTemplate struct {
	ID            uuid.UUID                `json:"id"`
	Name          string                   `json:"name"`
	Description   *string                  `json:"description,omitempty"`
	Icon          *string                  `json:"icon,omitempty"`
	CoverURL      *string                  `json:"cover_url,omitempty"`
	DefaultBlocks []map[string]interface{} `json:"default_blocks"`
	WorkspaceID   *uuid.UUID               `json:"workspace_id,omitempty"`
	Visibility    string                   `json:"visibility"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

/* DatabaseTemplate represents a database template */
type DatabaseTemplate struct {
	ID               uuid.UUID                `json:"id"`
	Name             string                   `json:"name"`
	Description      *string                  `json:"description,omitempty"`
	SchemaDefinition map[string]interface{}   `json:"schema_definition"`
	DefaultViews     []map[string]interface{} `json:"default_views,omitempty"`
	WorkspaceID      *uuid.UUID               `json:"workspace_id,omitempty"`
	Visibility       string                   `json:"visibility"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

/* TemplatesService provides page and database templates */
type TemplatesService struct {
	db *pgxpool.Pool
}

/* NewTemplatesService creates a new templates service */
func NewTemplatesService(db *pgxpool.Pool) *TemplatesService {
	return &TemplatesService{db: db}
}

/* ListPageTemplates lists page templates */
func (s *TemplatesService) ListPageTemplates(ctx context.Context, workspaceID *uuid.UUID) ([]PageTemplate, error) {
	var query string
	args := []interface{}{}
	if workspaceID != nil {
		query = `SELECT id, name, description, icon, cover_url, default_blocks, workspace_id, visibility, created_at, updated_at FROM neuronip.page_templates WHERE (workspace_id = $1 OR visibility = 'public') ORDER BY name`
		args = append(args, workspaceID)
	} else {
		query = `SELECT id, name, description, icon, cover_url, default_blocks, workspace_id, visibility, created_at, updated_at FROM neuronip.page_templates WHERE visibility IN ('public', 'workspace') ORDER BY name`
	}
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list page templates: %w", err)
	}
	defer rows.Close()
	var list []PageTemplate
	for rows.Next() {
		var t PageTemplate
		var desc, icon, cover *string
		var blocksRaw []byte
		if err := rows.Scan(&t.ID, &t.Name, &desc, &icon, &cover, &blocksRaw, &t.WorkspaceID, &t.Visibility, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Description = desc
		t.Icon = icon
		t.CoverURL = cover
		json.Unmarshal(blocksRaw, &t.DefaultBlocks)
		list = append(list, t)
	}
	return list, rows.Err()
}

/* ListDatabaseTemplates lists database templates */
func (s *TemplatesService) ListDatabaseTemplates(ctx context.Context, workspaceID *uuid.UUID) ([]DatabaseTemplate, error) {
	var query string
	args := []interface{}{}
	if workspaceID != nil {
		query = `SELECT id, name, description, schema_definition, default_views, workspace_id, visibility, created_at, updated_at FROM neuronip.database_templates WHERE (workspace_id = $1 OR visibility = 'public') ORDER BY name`
		args = append(args, workspaceID)
	} else {
		query = `SELECT id, name, description, schema_definition, default_views, workspace_id, visibility, created_at, updated_at FROM neuronip.database_templates WHERE visibility IN ('public', 'workspace') ORDER BY name`
	}
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list database templates: %w", err)
	}
	defer rows.Close()
	var list []DatabaseTemplate
	for rows.Next() {
		var t DatabaseTemplate
		var desc *string
		var schemaRaw, viewsRaw []byte
		if err := rows.Scan(&t.ID, &t.Name, &desc, &schemaRaw, &viewsRaw, &t.WorkspaceID, &t.Visibility, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Description = desc
		json.Unmarshal(schemaRaw, &t.SchemaDefinition)
		json.Unmarshal(viewsRaw, &t.DefaultViews)
		list = append(list, t)
	}
	return list, rows.Err()
}
