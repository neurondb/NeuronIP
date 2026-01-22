package governance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ModelRegistryService provides model registry functionality */
type ModelRegistryService struct {
	pool *pgxpool.Pool
}

/* NewModelRegistryService creates a new model registry service */
func NewModelRegistryService(pool *pgxpool.Pool) *ModelRegistryService {
	return &ModelRegistryService{pool: pool}
}

/* Model represents a model in the registry */
type Model struct {
	ID          uuid.UUID              `json:"id"`
	ModelName   string                 `json:"model_name"`
	Version     string                 `json:"version"`
	Provider    string                 `json:"provider"`
	ModelID     string                 `json:"model_id"`
	Status      string                 `json:"status"` // draft, pending_approval, approved, deprecated
	ApprovedBy  *string                `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time             `json:"approved_at,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* RegisterModel registers a new model version */
func (s *ModelRegistryService) RegisterModel(ctx context.Context, modelName string, version string, provider string, modelID string, createdBy string, config map[string]interface{}, metadata map[string]interface{}) (*Model, error) {
	id := uuid.New()
	configJSON, _ := json.Marshal(config)
	metadataJSON, _ := json.Marshal(metadata)
	now := time.Now()

	query := `
		INSERT INTO neuronip.model_registry 
		(id, model_name, version, provider, model_id, status, config, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'draft', $6, $7, $8, $9, $10)
		ON CONFLICT (model_name, version) 
		DO UPDATE SET 
			provider = EXCLUDED.provider,
			model_id = EXCLUDED.model_id,
			config = EXCLUDED.config,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
		RETURNING id, model_name, version, provider, model_id, status, approved_by, approved_at, config, metadata, created_by, created_at, updated_at`

	var model Model
	var configJSONRaw, metadataJSONRaw json.RawMessage
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, id, modelName, version, provider, modelID, configJSON, metadataJSON, createdBy, now, now).Scan(
		&model.ID, &model.ModelName, &model.Version, &model.Provider, &model.ModelID,
		&model.Status, &approvedBy, &approvedAt, &configJSONRaw, &metadataJSONRaw,
		&model.CreatedBy, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register model: %w", err)
	}

	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &model.Config)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &model.Metadata)
	}
	if approvedBy.Valid {
		model.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		model.ApprovedAt = &approvedAt.Time
	}

	return &model, nil
}

/* ApproveModel approves a model version */
func (s *ModelRegistryService) ApproveModel(ctx context.Context, modelID uuid.UUID, approverID string) error {
	now := time.Now()

	query := `
		UPDATE neuronip.model_registry 
		SET status = 'approved', approved_by = $1, approved_at = $2, updated_at = $2
		WHERE id = $3`

	_, err := s.pool.Exec(ctx, query, approverID, now, modelID)
	if err != nil {
		return fmt.Errorf("failed to approve model: %w", err)
	}

	return nil
}

/* RequestModelApproval requests approval for a model */
func (s *ModelRegistryService) RequestModelApproval(ctx context.Context, modelID uuid.UUID, approverID string) error {
	// Update model status to pending_approval
	query := `
		UPDATE neuronip.model_registry 
		SET status = 'pending_approval', updated_at = NOW()
		WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, modelID)
	if err != nil {
		return fmt.Errorf("failed to request model approval: %w", err)
	}

	// Create approval record
	approvalID := uuid.New()
	approvalQuery := `
		INSERT INTO neuronip.model_approvals 
		(id, model_id, approver_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', NOW(), NOW())`

	_, err = s.pool.Exec(ctx, approvalQuery, approvalID, modelID, approverID)
	if err != nil {
		return fmt.Errorf("failed to create approval record: %w", err)
	}

	return nil
}

/* GetModel retrieves a model by ID */
func (s *ModelRegistryService) GetModel(ctx context.Context, modelID uuid.UUID) (*Model, error) {
	query := `
		SELECT id, model_name, version, provider, model_id, status, approved_by, approved_at, config, metadata, created_by, created_at, updated_at
		FROM neuronip.model_registry
		WHERE id = $1`

	var model Model
	var configJSONRaw, metadataJSONRaw json.RawMessage
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, modelID).Scan(
		&model.ID, &model.ModelName, &model.Version, &model.Provider, &model.ModelID,
		&model.Status, &approvedBy, &approvedAt, &configJSONRaw, &metadataJSONRaw,
		&model.CreatedBy, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &model.Config)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &model.Metadata)
	}
	if approvedBy.Valid {
		model.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		model.ApprovedAt = &approvedAt.Time
	}

	return &model, nil
}

/* ListModels lists all models */
func (s *ModelRegistryService) ListModels(ctx context.Context, provider *string, status *string) ([]Model, error) {
	query := `
		SELECT id, model_name, version, provider, model_id, status, approved_by, approved_at, config, metadata, created_by, created_at, updated_at
		FROM neuronip.model_registry
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if provider != nil {
		query += fmt.Sprintf(" AND provider = $%d", argIndex)
		args = append(args, *provider)
		argIndex++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	query += " ORDER BY model_name, version DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var model Model
		var configJSONRaw, metadataJSONRaw json.RawMessage
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&model.ID, &model.ModelName, &model.Version, &model.Provider, &model.ModelID,
			&model.Status, &approvedBy, &approvedAt, &configJSONRaw, &metadataJSONRaw,
			&model.CreatedBy, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configJSONRaw != nil {
			json.Unmarshal(configJSONRaw, &model.Config)
		}
		if metadataJSONRaw != nil {
			json.Unmarshal(metadataJSONRaw, &model.Metadata)
		}
		if approvedBy.Valid {
			model.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			model.ApprovedAt = &approvedAt.Time
		}

		models = append(models, model)
	}

	return models, nil
}

/* SetWorkspaceModel sets the default model for a workspace */
func (s *ModelRegistryService) SetWorkspaceModel(ctx context.Context, workspaceID uuid.UUID, modelID uuid.UUID, isDefault bool) error {
	// If setting as default, unset other defaults
	if isDefault {
		unsetQuery := `
			UPDATE neuronip.workspace_model_selection 
			SET is_default = false, updated_at = NOW()
			WHERE workspace_id = $1`
		s.pool.Exec(ctx, unsetQuery, workspaceID)
	}

	// Insert or update workspace model selection
	query := `
		INSERT INTO neuronip.workspace_model_selection 
		(id, workspace_id, model_id, is_default, enabled, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (workspace_id, model_id) 
		DO UPDATE SET 
			is_default = EXCLUDED.is_default,
			enabled = true,
			updated_at = NOW()`

	_, err := s.pool.Exec(ctx, query, workspaceID, modelID, isDefault)
	if err != nil {
		return fmt.Errorf("failed to set workspace model: %w", err)
	}

	return nil
}

/* GetWorkspaceDefaultModel retrieves the default model for a workspace */
func (s *ModelRegistryService) GetWorkspaceDefaultModel(ctx context.Context, workspaceID uuid.UUID) (*Model, error) {
	query := `
		SELECT m.id, m.model_name, m.version, m.provider, m.model_id, m.status, m.approved_by, m.approved_at, m.config, m.metadata, m.created_by, m.created_at, m.updated_at
		FROM neuronip.model_registry m
		JOIN neuronip.workspace_model_selection wms ON m.id = wms.model_id
		WHERE wms.workspace_id = $1 AND wms.is_default = true AND wms.enabled = true AND m.status = 'approved'
		LIMIT 1`

	var model Model
	var configJSONRaw, metadataJSONRaw json.RawMessage
	var approvedBy sql.NullString
	var approvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, workspaceID).Scan(
		&model.ID, &model.ModelName, &model.Version, &model.Provider, &model.ModelID,
		&model.Status, &approvedBy, &approvedAt, &configJSONRaw, &metadataJSONRaw,
		&model.CreatedBy, &model.CreatedAt, &model.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace default model: %w", err)
	}

	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &model.Config)
	}
	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &model.Metadata)
	}
	if approvedBy.Valid {
		model.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		model.ApprovedAt = &approvedAt.Time
	}

	return &model, nil
}

/* ListWorkspaceModels lists all models available for a workspace */
func (s *ModelRegistryService) ListWorkspaceModels(ctx context.Context, workspaceID uuid.UUID) ([]Model, error) {
	query := `
		SELECT m.id, m.model_name, m.version, m.provider, m.model_id, m.status, m.approved_by, m.approved_at, m.config, m.metadata, m.created_by, m.created_at, m.updated_at
		FROM neuronip.model_registry m
		JOIN neuronip.workspace_model_selection wms ON m.id = wms.model_id
		WHERE wms.workspace_id = $1 AND wms.enabled = true AND m.status = 'approved'
		ORDER BY wms.is_default DESC, m.model_name`

	rows, err := s.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspace models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var model Model
		var configJSONRaw, metadataJSONRaw json.RawMessage
		var approvedBy sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&model.ID, &model.ModelName, &model.Version, &model.Provider, &model.ModelID,
			&model.Status, &approvedBy, &approvedAt, &configJSONRaw, &metadataJSONRaw,
			&model.CreatedBy, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configJSONRaw != nil {
			json.Unmarshal(configJSONRaw, &model.Config)
		}
		if metadataJSONRaw != nil {
			json.Unmarshal(metadataJSONRaw, &model.Metadata)
		}
		if approvedBy.Valid {
			model.ApprovedBy = &approvedBy.String
		}
		if approvedAt.Valid {
			model.ApprovedAt = &approvedAt.Time
		}

		models = append(models, model)
	}

	return models, nil
}

/* RollbackModel rolls back to a previous model version */
func (s *ModelRegistryService) RollbackModel(ctx context.Context, modelName string, targetVersion string) error {
	// Update current approved version to deprecated
	deprecateQuery := `
		UPDATE neuronip.model_registry 
		SET status = 'deprecated', updated_at = NOW()
		WHERE model_name = $1 AND status = 'approved'`

	s.pool.Exec(ctx, deprecateQuery, modelName)

	// Set target version to approved
	approveQuery := `
		UPDATE neuronip.model_registry 
		SET status = 'approved', updated_at = NOW()
		WHERE model_name = $1 AND version = $2`

	_, err := s.pool.Exec(ctx, approveQuery, modelName, targetVersion)
	if err != nil {
		return fmt.Errorf("failed to rollback model: %w", err)
	}

	return nil
}
