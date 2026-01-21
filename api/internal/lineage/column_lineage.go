package lineage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ColumnLineageService provides column-level lineage functionality */
type ColumnLineageService struct {
	pool *pgxpool.Pool
}

/* NewColumnLineageService creates a new column lineage service */
func NewColumnLineageService(pool *pgxpool.Pool) *ColumnLineageService {
	return &ColumnLineageService{pool: pool}
}

/* ColumnLineageNode represents a column lineage node */
type ColumnLineageNode struct {
	ID          uuid.UUID              `json:"id"`
	ConnectorID *uuid.UUID              `json:"connector_id,omitempty"`
	SchemaName  string                 `json:"schema_name"`
	TableName   string                 `json:"table_name"`
	ColumnName  string                 `json:"column_name"`
	NodeType    string                 `json:"node_type"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* CreateColumnNode creates a column lineage node */
func (s *ColumnLineageService) CreateColumnNode(ctx context.Context, node ColumnLineageNode) (*ColumnLineageNode, error) {
	node.ID = uuid.New()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(node.Metadata)
	var connectorID sql.NullString
	if node.ConnectorID != nil {
		connectorID = sql.NullString{String: node.ConnectorID.String(), Valid: true}
	}

	query := `
		INSERT INTO neuronip.column_lineage_nodes
		(id, connector_id, schema_name, table_name, column_name, node_type, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (connector_id, schema_name, table_name, column_name)
		DO UPDATE SET
			node_type = EXCLUDED.node_type,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		node.ID, connectorID, node.SchemaName, node.TableName, node.ColumnName,
		node.NodeType, metadataJSON, node.CreatedAt, node.UpdatedAt,
	).Scan(&node.ID, &node.CreatedAt, &node.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create column node: %w", err)
	}

	return &node, nil
}

/* GetColumnLineage gets column lineage (upstream and downstream) */
func (s *ColumnLineageService) GetColumnLineage(ctx context.Context, connectorID uuid.UUID, schemaName, tableName, columnName string) (*ColumnLineageGraph, error) {
	// Get node
	var nodeID uuid.UUID
	query := `
		SELECT id FROM neuronip.column_lineage_nodes
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3 AND column_name = $4`

	err := s.pool.QueryRow(ctx, query, connectorID, schemaName, tableName, columnName).Scan(&nodeID)
	if err != nil {
		return nil, fmt.Errorf("column node not found: %w", err)
	}

	// Get upstream (sources)
	upstreamNodes, upstreamEdges := s.getUpstreamLineage(ctx, nodeID)

	// Get downstream (dependents)
	downstreamNodes, downstreamEdges := s.getDownstreamLineage(ctx, nodeID)

	// Combine
	allNodes := append(upstreamNodes, downstreamNodes...)
	allEdges := append(upstreamEdges, downstreamEdges...)

	return &ColumnLineageGraph{
		Nodes: allNodes,
		Edges: allEdges,
	}, nil
}

/* getUpstreamLineage gets upstream lineage */
func (s *ColumnLineageService) getUpstreamLineage(ctx context.Context, nodeID uuid.UUID) ([]ColumnLineageNode, []ColumnLineageEdge) {
	// Recursive query to get all upstream nodes
	query := `
		WITH RECURSIVE upstream AS (
			SELECT source_node_id, target_node_id, edge_type, transformation
			FROM neuronip.column_lineage_edges
			WHERE target_node_id = $1
			
			UNION
			
			SELECT e.source_node_id, e.target_node_id, e.edge_type, e.transformation
			FROM neuronip.column_lineage_edges e
			INNER JOIN upstream u ON e.target_node_id = u.source_node_id
		)
		SELECT DISTINCT n.id, n.connector_id, n.schema_name, n.table_name, n.column_name,
		       n.node_type, n.metadata, n.created_at, n.updated_at
		FROM upstream u
		JOIN neuronip.column_lineage_nodes n ON n.id = u.source_node_id`

	rows, err := s.pool.Query(ctx, query, nodeID)
	if err != nil {
		return []ColumnLineageNode{}, []ColumnLineageEdge{}
	}
	defer rows.Close()

	nodes := []ColumnLineageNode{}
	for rows.Next() {
		var node ColumnLineageNode
		var connectorID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&node.ID, &connectorID, &node.SchemaName, &node.TableName, &node.ColumnName,
			&node.NodeType, &metadataJSON, &node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			connUUID, _ := uuid.Parse(connectorID.String)
			node.ConnectorID = &connUUID
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodes = append(nodes, node)
	}

	// Get edges
	edgesQuery := `
		SELECT id, source_node_id, target_node_id, edge_type, transformation, created_at
		FROM neuronip.column_lineage_edges
		WHERE target_node_id = $1 OR source_node_id IN (
			SELECT source_node_id FROM neuronip.column_lineage_edges WHERE target_node_id = $1
		)`

	edgesRows, _ := s.pool.Query(ctx, edgesQuery, nodeID)
	defer edgesRows.Close()

	edges := []ColumnLineageEdge{}
	for edgesRows.Next() {
		var edge ColumnLineageEdge
		var transformationJSON []byte

		err := edgesRows.Scan(
			&edge.ID, &edge.SourceNodeID, &edge.TargetNodeID, &edge.EdgeType,
			&transformationJSON, &edge.CreatedAt,
		)
		if err != nil {
			continue
		}

		if transformationJSON != nil {
			json.Unmarshal(transformationJSON, &edge.Transformation)
		}

		edges = append(edges, edge)
	}

	return nodes, edges
}

/* getDownstreamLineage gets downstream lineage */
func (s *ColumnLineageService) getDownstreamLineage(ctx context.Context, nodeID uuid.UUID) ([]ColumnLineageNode, []ColumnLineageEdge) {
	// Recursive query to get all downstream nodes
	query := `
		WITH RECURSIVE downstream AS (
			SELECT source_node_id, target_node_id, edge_type, transformation
			FROM neuronip.column_lineage_edges
			WHERE source_node_id = $1
			
			UNION
			
			SELECT e.source_node_id, e.target_node_id, e.edge_type, e.transformation
			FROM neuronip.column_lineage_edges e
			INNER JOIN downstream d ON e.source_node_id = d.target_node_id
		)
		SELECT DISTINCT n.id, n.connector_id, n.schema_name, n.table_name, n.column_name,
		       n.node_type, n.metadata, n.created_at, n.updated_at
		FROM downstream d
		JOIN neuronip.column_lineage_nodes n ON n.id = d.target_node_id`

	rows, err := s.pool.Query(ctx, query, nodeID)
	if err != nil {
		return []ColumnLineageNode{}, []ColumnLineageEdge{}
	}
	defer rows.Close()

	nodes := []ColumnLineageNode{}
	for rows.Next() {
		var node ColumnLineageNode
		var connectorID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&node.ID, &connectorID, &node.SchemaName, &node.TableName, &node.ColumnName,
			&node.NodeType, &metadataJSON, &node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			connUUID, _ := uuid.Parse(connectorID.String)
			node.ConnectorID = &connUUID
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodes = append(nodes, node)
	}

	// Get edges
	edgesQuery := `
		SELECT id, source_node_id, target_node_id, edge_type, transformation, created_at
		FROM neuronip.column_lineage_edges
		WHERE source_node_id = $1 OR target_node_id IN (
			SELECT target_node_id FROM neuronip.column_lineage_edges WHERE source_node_id = $1
		)`

	edgesRows, _ := s.pool.Query(ctx, edgesQuery, nodeID)
	defer edgesRows.Close()

	edges := []ColumnLineageEdge{}
	for edgesRows.Next() {
		var edge ColumnLineageEdge
		var transformationJSON []byte

		err := edgesRows.Scan(
			&edge.ID, &edge.SourceNodeID, &edge.TargetNodeID, &edge.EdgeType,
			&transformationJSON, &edge.CreatedAt,
		)
		if err != nil {
			continue
		}

		if transformationJSON != nil {
			json.Unmarshal(transformationJSON, &edge.Transformation)
		}

		edges = append(edges, edge)
	}

	return nodes, edges
}

/* CreateColumnEdge creates a column lineage edge */
func (s *ColumnLineageService) CreateColumnEdge(ctx context.Context, edge ColumnLineageEdge) (*ColumnLineageEdge, error) {
	edge.ID = uuid.New()
	edge.CreatedAt = time.Now()

	transformationJSON, _ := json.Marshal(edge.Transformation)

	query := `
		INSERT INTO neuronip.column_lineage_edges
		(id, source_node_id, target_node_id, edge_type, transformation, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (source_node_id, target_node_id, edge_type)
		DO UPDATE SET
			transformation = EXCLUDED.transformation
		RETURNING id, created_at`

	err := s.pool.QueryRow(ctx, query,
		edge.SourceNodeID, edge.TargetNodeID, edge.EdgeType,
		transformationJSON, edge.CreatedAt,
	).Scan(&edge.ID, &edge.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create column edge: %w", err)
	}

	return &edge, nil
}

/* TrackColumnLineage tracks column-level lineage between two columns */
func (s *ColumnLineageService) TrackColumnLineage(ctx context.Context,
	sourceConnectorID *uuid.UUID, sourceSchemaName, sourceTableName, sourceColumnName string,
	targetConnectorID *uuid.UUID, targetSchemaName, targetTableName, targetColumnName string,
	edgeType string, transformation map[string]interface{}) (*ColumnLineageEdge, error) {
	// Get or create source node
	sourceNode := ColumnLineageNode{
		ConnectorID: sourceConnectorID,
		SchemaName:  sourceSchemaName,
		TableName:   sourceTableName,
		ColumnName:  sourceColumnName,
		NodeType:    "source",
	}
	createdSourceNode, err := s.CreateColumnNode(ctx, sourceNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create source node: %w", err)
	}

	// Get or create target node
	targetNode := ColumnLineageNode{
		ConnectorID: targetConnectorID,
		SchemaName:  targetSchemaName,
		TableName:   targetTableName,
		ColumnName:  targetColumnName,
		NodeType:    "derived",
	}
	createdTargetNode, err := s.CreateColumnNode(ctx, targetNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create target node: %w", err)
	}

	// Create edge
	edge := ColumnLineageEdge{
		SourceNodeID:   createdSourceNode.ID,
		TargetNodeID:   createdTargetNode.ID,
		EdgeType:       edgeType,
		Transformation: transformation,
	}

	return s.CreateColumnEdge(ctx, edge)
}

/* ColumnLineageGraph represents column lineage graph */
type ColumnLineageGraph struct {
	Nodes []ColumnLineageNode `json:"nodes"`
	Edges []ColumnLineageEdge `json:"edges"`
}

/* ColumnLineageEdge represents a column lineage edge */
type ColumnLineageEdge struct {
	ID             uuid.UUID              `json:"id"`
	SourceNodeID   uuid.UUID              `json:"source_node_id"`
	TargetNodeID   uuid.UUID              `json:"target_node_id"`
	EdgeType       string                 `json:"edge_type"`
	Transformation map[string]interface{} `json:"transformation,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}
