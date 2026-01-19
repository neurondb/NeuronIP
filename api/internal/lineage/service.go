package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* LineageService provides data lineage functionality */
type LineageService struct {
	pool *pgxpool.Pool
}

/* NewLineageService creates a new lineage service */
func NewLineageService(pool *pgxpool.Pool) *LineageService {
	return &LineageService{pool: pool}
}

/* LineageNode represents a lineage node */
type LineageNode struct {
	ID         uuid.UUID              `json:"id"`
	NodeType   string                 `json:"node_type"`
	NodeName   string                 `json:"node_name"`
	SchemaInfo map[string]interface{} `json:"schema_info,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

/* LineageEdge represents a lineage edge */
type LineageEdge struct {
	ID             uuid.UUID              `json:"id"`
	SourceNodeID   uuid.UUID              `json:"source_node_id"`
	TargetNodeID   uuid.UUID              `json:"target_node_id"`
	EdgeType       string                 `json:"edge_type"`
	Transformation map[string]interface{} `json:"transformation,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

/* LineageGraph represents the full lineage graph */
type LineageGraph struct {
	Nodes []LineageNode `json:"nodes"`
	Edges []LineageEdge `json:"edges"`
}

/* GetLineage retrieves lineage for a resource */
func (s *LineageService) GetLineage(ctx context.Context, resourceType, resourceID string) (*LineageGraph, error) {
	// Find nodes related to this resource
	query := `
		SELECT ln.id, ln.node_type, ln.node_name, ln.schema_info, ln.metadata, ln.created_at
		FROM neuronip.lineage_nodes ln
		WHERE ln.metadata->>'resource_type' = $1 AND ln.metadata->>'resource_id' = $2`

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lineage: %w", err)
	}
	defer rows.Close()

	var nodeIDs []uuid.UUID
	var nodes []LineageNode

	for rows.Next() {
		var node LineageNode
		var schemaJSON, metadataJSON json.RawMessage

		err := rows.Scan(&node.ID, &node.NodeType, &node.NodeName, &schemaJSON, &metadataJSON, &node.CreatedAt)
		if err != nil {
			continue
		}

		if schemaJSON != nil {
			json.Unmarshal(schemaJSON, &node.SchemaInfo)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodeIDs = append(nodeIDs, node.ID)
		nodes = append(nodes, node)
	}

	// Get edges for these nodes
	var edges []LineageEdge
	if len(nodeIDs) > 0 {
		edgeQuery := `
			SELECT id, source_node_id, target_node_id, edge_type, transformation, created_at
			FROM neuronip.lineage_edges
			WHERE source_node_id = ANY($1) OR target_node_id = ANY($1)`

		edgeRows, err := s.pool.Query(ctx, edgeQuery, nodeIDs)
		if err == nil {
			defer edgeRows.Close()

			for edgeRows.Next() {
				var edge LineageEdge
				var transJSON json.RawMessage

				err := edgeRows.Scan(&edge.ID, &edge.SourceNodeID, &edge.TargetNodeID, &edge.EdgeType, &transJSON, &edge.CreatedAt)
				if err != nil {
					continue
				}

				if transJSON != nil {
					json.Unmarshal(transJSON, &edge.Transformation)
				}

				edges = append(edges, edge)
			}
		}
	}

	return &LineageGraph{Nodes: nodes, Edges: edges}, nil
}

/* TrackTransformation tracks a data transformation */
func (s *LineageService) TrackTransformation(ctx context.Context, sourceID, targetID uuid.UUID, edgeType string, transformation map[string]interface{}) error {
	transJSON, _ := json.Marshal(transformation)

	query := `
		INSERT INTO neuronip.lineage_edges (id, source_node_id, target_node_id, edge_type, transformation, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		ON CONFLICT (source_node_id, target_node_id, edge_type) DO NOTHING`

	_, err := s.pool.Exec(ctx, query, sourceID, targetID, edgeType, transJSON)
	return err
}

/* GetImpactAnalysis performs impact analysis for a resource */
func (s *LineageService) GetImpactAnalysis(ctx context.Context, resourceID string) ([]LineageNode, error) {
	// Find all downstream nodes (nodes that depend on this resource)
	query := `
		SELECT DISTINCT ln.id, ln.node_type, ln.node_name, ln.schema_info, ln.metadata, ln.created_at
		FROM neuronip.lineage_nodes ln
		JOIN neuronip.lineage_edges le ON le.target_node_id = ln.id
		WHERE le.source_node_id = (SELECT id FROM neuronip.lineage_nodes WHERE metadata->>'resource_id' = $1 LIMIT 1)`

	rows, err := s.pool.Query(ctx, query, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get impact analysis: %w", err)
	}
	defer rows.Close()

	var nodes []LineageNode
	for rows.Next() {
		var node LineageNode
		var schemaJSON, metadataJSON json.RawMessage

		err := rows.Scan(&node.ID, &node.NodeType, &node.NodeName, &schemaJSON, &metadataJSON, &node.CreatedAt)
		if err != nil {
			continue
		}

		if schemaJSON != nil {
			json.Unmarshal(schemaJSON, &node.SchemaInfo)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

/* GetFullGraph retrieves the complete lineage graph */
func (s *LineageService) GetFullGraph(ctx context.Context) (*LineageGraph, error) {
	// Get all nodes
	nodeQuery := `SELECT id, node_type, node_name, schema_info, metadata, created_at FROM neuronip.lineage_nodes`
	rows, err := s.pool.Query(ctx, nodeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	defer rows.Close()

	var nodes []LineageNode
	for rows.Next() {
		var node LineageNode
		var schemaJSON, metadataJSON json.RawMessage

		err := rows.Scan(&node.ID, &node.NodeType, &node.NodeName, &schemaJSON, &metadataJSON, &node.CreatedAt)
		if err != nil {
			continue
		}

		if schemaJSON != nil {
			json.Unmarshal(schemaJSON, &node.SchemaInfo)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodes = append(nodes, node)
	}

	// Get all edges
	edgeQuery := `SELECT id, source_node_id, target_node_id, edge_type, transformation, created_at FROM neuronip.lineage_edges`
	edgeRows, err := s.pool.Query(ctx, edgeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}
	defer edgeRows.Close()

	var edges []LineageEdge
	for edgeRows.Next() {
		var edge LineageEdge
		var transJSON json.RawMessage

		err := edgeRows.Scan(&edge.ID, &edge.SourceNodeID, &edge.TargetNodeID, &edge.EdgeType, &transJSON, &edge.CreatedAt)
		if err != nil {
			continue
		}

		if transJSON != nil {
			json.Unmarshal(transJSON, &edge.Transformation)
		}

		edges = append(edges, edge)
	}

	return &LineageGraph{Nodes: nodes, Edges: edges}, nil
}

/* CreateNode creates a new lineage node */
func (s *LineageService) CreateNode(ctx context.Context, node LineageNode) (*LineageNode, error) {
	id := uuid.New()
	schemaJSON, _ := json.Marshal(node.SchemaInfo)
	metadataJSON, _ := json.Marshal(node.Metadata)

	query := `
		INSERT INTO neuronip.lineage_nodes (id, node_type, node_name, schema_info, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, node_type, node_name, schema_info, metadata, created_at`

	var result LineageNode
	var resultSchemaJSON, resultMetadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, id, node.NodeType, node.NodeName, schemaJSON, metadataJSON).Scan(
		&result.ID, &result.NodeType, &result.NodeName, &resultSchemaJSON, &resultMetadataJSON, &result.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	if resultSchemaJSON != nil {
		json.Unmarshal(resultSchemaJSON, &result.SchemaInfo)
	}
	if resultMetadataJSON != nil {
		json.Unmarshal(resultMetadataJSON, &result.Metadata)
	}

	return &result, nil
}
