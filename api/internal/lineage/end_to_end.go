package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* EndToEndService provides cross-system lineage functionality */
type EndToEndService struct {
	pool *pgxpool.Pool
}

/* NewEndToEndService creates a new end-to-end lineage service */
func NewEndToEndService(pool *pgxpool.Pool) *EndToEndService {
	return &EndToEndService{pool: pool}
}

/* CrossSystemLineage represents lineage across multiple systems */
type CrossSystemLineage struct {
	ID          uuid.UUID              `json:"id"`
	SourceSystem string                `json:"source_system"` // e.g., "snowflake", "bigquery", "airflow"
	TargetSystem string                `json:"target_system"`
	SourceNodeID uuid.UUID             `json:"source_node_id"`
	TargetNodeID uuid.UUID             `json:"target_node_id"`
	EdgeType     string                `json:"edge_type"` // "sync", "transform", "copy", "replicate"
	SyncFrequency string               `json:"sync_frequency,omitempty"` // e.g., "hourly", "daily"
	Transformation string              `json:"transformation,omitempty"` // Description of transformation
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

/* LineagePath represents a path through multiple systems */
type LineagePath struct {
	ID          uuid.UUID              `json:"id"`
	Path        []LineageStep          `json:"path"`
	SourceSystem string                `json:"source_system"`
	TargetSystem string                `json:"target_system"`
	TotalHops   int                    `json:"total_hops"`
	Confidence  float64                `json:"confidence"`
	CreatedAt   time.Time              `json:"created_at"`
}

/* LineageStep represents a step in a lineage path */
type LineageStep struct {
	NodeID       uuid.UUID             `json:"node_id"`
	System       string                `json:"system"`
	ResourceType string                `json:"resource_type"`
	ResourceName string                `json:"resource_name"`
	StepOrder    int                   `json:"step_order"`
	Transformation map[string]interface{} `json:"transformation,omitempty"`
}

/* GetEndToEndLineage retrieves lineage across multiple systems */
func (s *EndToEndService) GetEndToEndLineage(ctx context.Context,
	sourceSystem, targetSystem string,
	sourceResourceID string) (*LineagePath, error) {

	// Find all paths from source to target across systems
	query := `
		WITH RECURSIVE lineage_path AS (
			-- Start with source node
			SELECT 
				ln.id as node_id,
				ln.metadata->>'system' as system,
				ln.metadata->>'resource_type' as resource_type,
				ln.node_name as resource_name,
				1 as step_order,
				ARRAY[ln.id] as visited_nodes,
				jsonb_build_object('system', ln.metadata->>'system', 'name', ln.node_name) as path
			FROM neuronip.lineage_nodes ln
			WHERE ln.metadata->>'system' = $1 
			AND (ln.metadata->>'resource_id' = $3 OR ln.id::text = $3)
			
			UNION ALL
			
			-- Recursively find connected nodes
			SELECT 
				target_ln.id as node_id,
				target_ln.metadata->>'system' as system,
				target_ln.metadata->>'resource_type' as resource_type,
				target_ln.node_name as resource_name,
				lp.step_order + 1,
				lp.visited_nodes || target_ln.id,
				lp.path || jsonb_build_object('system', target_ln.metadata->>'system', 'name', target_ln.node_name)
			FROM lineage_path lp
			JOIN neuronip.lineage_edges le ON le.source_node_id = lp.node_id
			JOIN neuronip.lineage_nodes target_ln ON target_ln.id = le.target_node_id
			WHERE NOT (target_ln.id = ANY(lp.visited_nodes))
			AND lp.step_order < 10  -- Limit depth
			AND (target_ln.metadata->>'system' = $2 OR lp.step_order < 5)
		)
		SELECT node_id, system, resource_type, resource_name, step_order, path
		FROM lineage_path
		WHERE system = $2
		ORDER BY step_order ASC
		LIMIT 1`

	var path LineagePath
	path.ID = uuid.New()
	path.SourceSystem = sourceSystem
	path.TargetSystem = targetSystem
	path.CreatedAt = time.Now()

	rows, err := s.pool.Query(ctx, query, sourceSystem, targetSystem, sourceResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get end-to-end lineage: %w", err)
	}
	defer rows.Close()

	path.Path = []LineageStep{}
	for rows.Next() {
		var step LineageStep
		var pathJSON json.RawMessage

		err := rows.Scan(&step.NodeID, &step.System, &step.ResourceType,
			&step.ResourceName, &step.StepOrder, &pathJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(pathJSON, &step.Transformation)
		path.Path = append(path.Path, step)
	}

	if len(path.Path) == 0 {
		return nil, fmt.Errorf("no path found from %s to %s", sourceSystem, targetSystem)
	}

	path.TotalHops = len(path.Path)
	path.Confidence = s.calculatePathConfidence(path.Path)

	return &path, nil
}

/* calculatePathConfidence calculates confidence score for a lineage path */
func (s *EndToEndService) calculatePathConfidence(path []LineageStep) float64 {
	// Simple confidence calculation based on path length
	// Longer paths = lower confidence
	baseConfidence := 1.0
	confidenceDecay := 0.1

	for i := range path {
		if i > 0 {
			baseConfidence -= confidenceDecay
		}
	}

	if baseConfidence < 0.3 {
		return 0.3
	}

	return baseConfidence
}

/* TrackCrossSystemLineage tracks lineage across different systems */
func (s *EndToEndService) TrackCrossSystemLineage(ctx context.Context,
	lineage CrossSystemLineage) error {

	lineage.ID = uuid.New()
	lineage.CreatedAt = time.Now()
	lineage.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(lineage.Metadata)

	query := `
		INSERT INTO neuronip.cross_system_lineage
		(id, source_system, target_system, source_node_id, target_node_id,
		 edge_type, sync_frequency, transformation, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (source_node_id, target_node_id, edge_type) 
		DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			sync_frequency = EXCLUDED.sync_frequency,
			transformation = EXCLUDED.transformation,
			metadata = EXCLUDED.metadata`

	_, err := s.pool.Exec(ctx, query,
		lineage.ID, lineage.SourceSystem, lineage.TargetSystem,
		lineage.SourceNodeID, lineage.TargetNodeID, lineage.EdgeType,
		lineage.SyncFrequency, lineage.Transformation, metadataJSON,
		lineage.CreatedAt, lineage.UpdatedAt)

	return err
}

/* GetCrossSystemLineage retrieves cross-system lineage for a system */
func (s *EndToEndService) GetCrossSystemLineage(ctx context.Context,
	system string) ([]CrossSystemLineage, error) {

	query := `
		SELECT id, source_system, target_system, source_node_id, target_node_id,
		       edge_type, sync_frequency, transformation, metadata, created_at, updated_at
		FROM neuronip.cross_system_lineage
		WHERE source_system = $1 OR target_system = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, system)
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-system lineage: %w", err)
	}
	defer rows.Close()

	var lineage []CrossSystemLineage
	for rows.Next() {
		var l CrossSystemLineage
		var metadataJSON []byte

		err := rows.Scan(&l.ID, &l.SourceSystem, &l.TargetSystem,
			&l.SourceNodeID, &l.TargetNodeID, &l.EdgeType,
			&l.SyncFrequency, &l.Transformation, &metadataJSON,
			&l.CreatedAt, &l.UpdatedAt)

		if err != nil {
			continue
		}

		json.Unmarshal(metadataJSON, &l.Metadata)
		lineage = append(lineage, l)
	}

	return lineage, nil
}

/* FindLineagePaths finds all paths between two resources across systems */
func (s *EndToEndService) FindLineagePaths(ctx context.Context,
	sourceResourceID, targetResourceID string,
	maxDepth int) ([]LineagePath, error) {

	if maxDepth == 0 {
		maxDepth = 5
	}

	// Find all paths using recursive CTE
	query := `
		WITH RECURSIVE lineage_paths AS (
			-- Start with source
			SELECT 
				ln.id as node_id,
				ln.metadata->>'system' as system,
				ln.node_name as resource_name,
				1 as depth,
				ARRAY[ln.id] as path_nodes,
				jsonb_build_array(jsonb_build_object('node_id', ln.id::text, 'system', ln.metadata->>'system', 'name', ln.node_name)) as path
			FROM neuronip.lineage_nodes ln
			WHERE ln.metadata->>'resource_id' = $1 OR ln.id::text = $1
			
			UNION ALL
			
			-- Recursively expand
			SELECT 
				target_ln.id as node_id,
				target_ln.metadata->>'system' as system,
				target_ln.node_name as resource_name,
				lp.depth + 1,
				lp.path_nodes || target_ln.id,
				lp.path || jsonb_build_object('node_id', target_ln.id::text, 'system', target_ln.metadata->>'system', 'name', target_ln.node_name)
			FROM lineage_paths lp
			JOIN neuronip.lineage_edges le ON le.source_node_id = lp.node_id
			JOIN neuronip.lineage_nodes target_ln ON target_ln.id = le.target_node_id
			WHERE NOT (target_ln.id = ANY(lp.path_nodes))
			AND lp.depth < $3
			AND (target_ln.metadata->>'resource_id' != $2 AND target_ln.id::text != $2 OR lp.depth >= $3)
		)
		SELECT path, depth
		FROM lineage_paths
		WHERE (node_id IN (SELECT id FROM neuronip.lineage_nodes WHERE metadata->>'resource_id' = $2 OR id::text = $2))
		ORDER BY depth ASC`

	rows, err := s.pool.Query(ctx, query, sourceResourceID, targetResourceID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to find paths: %w", err)
	}
	defer rows.Close()

	var paths []LineagePath
	for rows.Next() {
		var pathJSON json.RawMessage
		var depth int

		err := rows.Scan(&pathJSON, &depth)
		if err != nil {
			continue
		}

		var pathData []map[string]interface{}
		json.Unmarshal(pathJSON, &pathData)

		path := LineagePath{
			ID:        uuid.New(),
			TotalHops: depth,
			CreatedAt: time.Now(),
		}

		path.Path = []LineageStep{}
		for i, stepData := range pathData {
			step := LineageStep{
				StepOrder: i + 1,
			}
			if nodeID, ok := stepData["node_id"].(string); ok {
				id, _ := uuid.Parse(nodeID)
				step.NodeID = id
			}
			if system, ok := stepData["system"].(string); ok {
				step.System = system
			}
			if name, ok := stepData["name"].(string); ok {
				step.ResourceName = name
			}
			path.Path = append(path.Path, step)
		}

		if len(path.Path) > 0 {
			path.SourceSystem = path.Path[0].System
			path.TargetSystem = path.Path[len(path.Path)-1].System
			path.Confidence = s.calculatePathConfidence(path.Path)
			paths = append(paths, path)
		}
	}

	return paths, nil
}
