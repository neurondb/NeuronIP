package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TransformationService provides transformation logic capture functionality */
type TransformationService struct {
	pool *pgxpool.Pool
}

/* NewTransformationService creates a new transformation service */
func NewTransformationService(pool *pgxpool.Pool) *TransformationService {
	return &TransformationService{pool: pool}
}

/* TransformationLogic represents captured transformation logic */
type TransformationLogic struct {
	ID             uuid.UUID              `json:"id"`
	SourceNodeID   uuid.UUID              `json:"source_node_id"`
	TargetNodeID   uuid.UUID              `json:"target_node_id"`
	TransformationType string             `json:"transformation_type"` // "sql", "python", "dbt", "airflow"
	Logic          string                 `json:"logic"` // Transformation code/logic
	Language       string                 `json:"language"` // "sql", "python", "yaml"
	SourceSystem   string                 `json:"source_system"` // "warehouse", "dbt", "airflow", "custom"
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CapturedAt     time.Time              `json:"captured_at"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* CaptureTransformation captures transformation logic from various sources */
func (s *TransformationService) CaptureTransformation(ctx context.Context,
	transformation TransformationLogic) (*TransformationLogic, error) {

	transformation.ID = uuid.New()
	transformation.CapturedAt = time.Now()
	transformation.CreatedAt = time.Now()
	transformation.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(transformation.Metadata)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.transformation_logic
		(id, source_node_id, target_node_id, transformation_type, logic,
		 language, source_system, metadata, captured_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		transformation.ID, transformation.SourceNodeID, transformation.TargetNodeID,
		transformation.TransformationType, transformation.Logic,
		transformation.Language, transformation.SourceSystem, metadataJSON,
		transformation.CapturedAt, transformation.CreatedAt, transformation.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to capture transformation: %w", err)
	}

	return &transformation, nil
}

/* CaptureFromSQL captures transformation from SQL query */
func (s *TransformationService) CaptureFromSQL(ctx context.Context,
	sourceNodeID, targetNodeID uuid.UUID, sqlQuery string) (*TransformationLogic, error) {

	return s.CaptureTransformation(ctx, TransformationLogic{
		SourceNodeID:     sourceNodeID,
		TargetNodeID:     targetNodeID,
		TransformationType: "sql",
		Logic:            sqlQuery,
		Language:         "sql",
		SourceSystem:     "warehouse",
		Metadata: map[string]interface{}{
			"query_type": "transformation",
		},
	})
}

/* CaptureFromDBT captures transformation from dbt model */
func (s *TransformationService) CaptureFromDBT(ctx context.Context,
	sourceNodeID, targetNodeID uuid.UUID, dbtSQL string, modelName string) (*TransformationLogic, error) {

	return s.CaptureTransformation(ctx, TransformationLogic{
		SourceNodeID:     sourceNodeID,
		TargetNodeID:     targetNodeID,
		TransformationType: "dbt",
		Logic:            dbtSQL,
		Language:         "sql",
		SourceSystem:     "dbt",
		Metadata: map[string]interface{}{
			"model_name": modelName,
			"framework":  "dbt",
		},
	})
}

/* CaptureFromAirflow captures transformation from Airflow task */
func (s *TransformationService) CaptureFromAirflow(ctx context.Context,
	sourceNodeID, targetNodeID uuid.UUID, taskCode string, dagID, taskID string) (*TransformationLogic, error) {

	return s.CaptureTransformation(ctx, TransformationLogic{
		SourceNodeID:     sourceNodeID,
		TargetNodeID:     targetNodeID,
		TransformationType: "airflow",
		Logic:            taskCode,
		Language:         "python",
		SourceSystem:     "airflow",
		Metadata: map[string]interface{}{
			"dag_id":   dagID,
			"task_id":  taskID,
			"framework": "airflow",
		},
	})
}

/* GetTransformation retrieves transformation logic */
func (s *TransformationService) GetTransformation(ctx context.Context,
	sourceNodeID, targetNodeID uuid.UUID) (*TransformationLogic, error) {

	var transformation TransformationLogic
	var metadataJSON []byte

	err := s.pool.QueryRow(ctx, `
		SELECT id, source_node_id, target_node_id, transformation_type, logic,
		       language, source_system, metadata, captured_at, created_at, updated_at
		FROM neuronip.transformation_logic
		WHERE source_node_id = $1 AND target_node_id = $2
		ORDER BY captured_at DESC
		LIMIT 1`, sourceNodeID, targetNodeID,
	).Scan(&transformation.ID, &transformation.SourceNodeID, &transformation.TargetNodeID,
		&transformation.TransformationType, &transformation.Logic,
		&transformation.Language, &transformation.SourceSystem, &metadataJSON,
		&transformation.CapturedAt, &transformation.CreatedAt, &transformation.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get transformation: %w", err)
	}

	json.Unmarshal(metadataJSON, &transformation.Metadata)

	return &transformation, nil
}

/* ListTransformations lists all transformations for a node */
func (s *TransformationService) ListTransformations(ctx context.Context,
	nodeID uuid.UUID, direction string) ([]TransformationLogic, error) {

	var query string
	if direction == "source" || direction == "" {
		query = `
			SELECT id, source_node_id, target_node_id, transformation_type, logic,
			       language, source_system, metadata, captured_at, created_at, updated_at
			FROM neuronip.transformation_logic
			WHERE source_node_id = $1
			ORDER BY captured_at DESC`
	} else {
		query = `
			SELECT id, source_node_id, target_node_id, transformation_type, logic,
			       language, source_system, metadata, captured_at, created_at, updated_at
			FROM neuronip.transformation_logic
			WHERE target_node_id = $1
			ORDER BY captured_at DESC`
	}

	rows, err := s.pool.Query(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list transformations: %w", err)
	}
	defer rows.Close()

	var transformations []TransformationLogic
	for rows.Next() {
		var transformation TransformationLogic
		var metadataJSON []byte

		err := rows.Scan(&transformation.ID, &transformation.SourceNodeID, &transformation.TargetNodeID,
			&transformation.TransformationType, &transformation.Logic,
			&transformation.Language, &transformation.SourceSystem, &metadataJSON,
			&transformation.CapturedAt, &transformation.CreatedAt, &transformation.UpdatedAt)
		if err != nil {
			continue
		}

		json.Unmarshal(metadataJSON, &transformation.Metadata)
		transformations = append(transformations, transformation)
	}

	return transformations, nil
}
