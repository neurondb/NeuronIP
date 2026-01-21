package lineage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* QueryAnalysisService provides query log analysis for lineage discovery */
type QueryAnalysisService struct {
	pool *pgxpool.Pool
}

/* NewQueryAnalysisService creates a new query analysis service */
func NewQueryAnalysisService(pool *pgxpool.Pool) *QueryAnalysisService {
	return &QueryAnalysisService{pool: pool}
}

/* QueryPattern represents a pattern extracted from query logs */
type QueryPattern struct {
	ID           uuid.UUID              `json:"id"`
	PatternType  string                 `json:"pattern_type"` // "select_into", "create_table_as", "insert_select", "update", "delete"
	SourceTables []string               `json:"source_tables"`
	TargetTable  string                 `json:"target_table"`
	QueryHash    string                 `json:"query_hash"`
	Frequency    int                    `json:"frequency"`
	FirstSeen    time.Time              `json:"first_seen"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* AnalyzeQueryLogs analyzes query logs to discover lineage patterns */
func (s *QueryAnalysisService) AnalyzeQueryLogs(ctx context.Context,
	startTime, endTime time.Time) ([]QueryPattern, error) {

	if endTime.IsZero() {
		endTime = time.Now()
	}
	if startTime.IsZero() {
		startTime = endTime.AddDate(0, 0, -7) // Last 7 days by default
	}

	// Analyze warehouse query history
	query := `
		SELECT id, query_text, schema_id, created_at
		FROM neuronip.warehouse_queries
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT 1000`

	rows, err := s.pool.Query(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze query logs: %w", err)
	}
	defer rows.Close()

	patterns := make(map[string]*QueryPattern)

	for rows.Next() {
		var queryID uuid.UUID
		var queryText, schemaID string
		var createdAt time.Time

		err := rows.Scan(&queryID, &queryText, &schemaID, &createdAt)
		if err != nil {
			continue
		}

		// Extract lineage patterns from query
		pattern := s.extractPattern(queryText, schemaID)
		if pattern == nil {
			continue
		}

		// Aggregate patterns
		if existing, exists := patterns[pattern.QueryHash]; exists {
			existing.Frequency++
			if createdAt.Before(existing.FirstSeen) {
				existing.FirstSeen = createdAt
			}
			if createdAt.After(existing.LastSeen) {
				existing.LastSeen = createdAt
			}
		} else {
			pattern.ID = uuid.New()
			pattern.FirstSeen = createdAt
			pattern.LastSeen = createdAt
			pattern.Frequency = 1
			patterns[pattern.QueryHash] = pattern
		}
	}

	// Convert map to slice
	result := make([]QueryPattern, 0, len(patterns))
	for _, pattern := range patterns {
		result = append(result, *pattern)
	}

	return result, nil
}

/* extractPattern extracts lineage pattern from SQL query */
func (s *QueryAnalysisService) extractPattern(queryText, schemaID string) *QueryPattern {
	// Convert to lowercase for pattern matching
	lowerQuery := strings.ToLower(queryText)

	pattern := &QueryPattern{
		Metadata: map[string]interface{}{
			"schema_id": schemaID,
		},
	}

	// CREATE TABLE AS SELECT pattern
	if strings.Contains(lowerQuery, "create table") && strings.Contains(lowerQuery, "as select") {
		pattern.PatternType = "create_table_as"
		pattern.QueryHash = s.hashQuery(queryText)
		// Extract source and target tables (simplified)
		pattern.SourceTables = s.extractTableNames(queryText, "from")
		pattern.TargetTable = s.extractTableName(queryText, "create table")
		return pattern
	}

	// SELECT INTO pattern
	if strings.Contains(lowerQuery, "select") && strings.Contains(lowerQuery, "into") {
		pattern.PatternType = "select_into"
		pattern.QueryHash = s.hashQuery(queryText)
		pattern.SourceTables = s.extractTableNames(queryText, "from")
		pattern.TargetTable = s.extractTableName(queryText, "into")
		return pattern
	}

	// INSERT SELECT pattern
	if strings.Contains(lowerQuery, "insert") && strings.Contains(lowerQuery, "select") {
		pattern.PatternType = "insert_select"
		pattern.QueryHash = s.hashQuery(queryText)
		pattern.SourceTables = s.extractTableNames(queryText, "from")
		pattern.TargetTable = s.extractTableName(queryText, "insert into")
		return pattern
	}

	return nil
}

/* hashQuery creates a hash of query for pattern matching */
func (s *QueryAnalysisService) hashQuery(queryText string) string {
	// Normalize query first
	normalized := normalizeQuery(queryText)
	// Create SHA256 hash
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

/* normalizeQuery normalizes SQL query for pattern matching */
func normalizeQuery(query string) string {
	// Remove comments, normalize whitespace, lowercase
	// Simplified implementation - in production would use SQL parser
	normalized := strings.TrimSpace(strings.ToLower(query))
	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")
	// Remove semicolons and trailing whitespace
	normalized = strings.TrimRight(normalized, ";")
	return normalized
}

/* extractTableNames extracts table names from SQL query after a keyword */
func (s *QueryAnalysisService) extractTableNames(query, keyword string) []string {
	// Simplified table name extraction - in production would use SQL parser
	// This is a placeholder that would parse SQL properly
	lowerQuery := strings.ToLower(query)
	keywordPos := strings.Index(lowerQuery, strings.ToLower(keyword))
	if keywordPos == -1 {
		return []string{}
	}

	// Extract table names (simplified - would need proper SQL parsing)
	// For now, return empty - proper implementation would parse SQL
	return []string{}
}

/* extractTableName extracts a table name from SQL query after a keyword */
func (s *QueryAnalysisService) extractTableName(query, keyword string) string {
	// Simplified table name extraction - in production would use SQL parser
	// This is a placeholder that would parse SQL properly
	lowerQuery := strings.ToLower(query)
	keywordPos := strings.Index(lowerQuery, strings.ToLower(keyword))
	if keywordPos == -1 {
		return ""
	}

	// Extract table name (simplified - would need proper SQL parsing)
	// For now, return empty - proper implementation would parse SQL
	return ""
}

/* DiscoverLineageFromQueries discovers lineage from query patterns */
func (s *QueryAnalysisService) DiscoverLineageFromQueries(ctx context.Context,
	patterns []QueryPattern) ([]DiscoveredLineage, error) {

	var discovered []DiscoveredLineage

	for _, pattern := range patterns {
		// Get schema ID from metadata
		schemaID := ""
		if schemaIDVal, ok := pattern.Metadata["schema_id"].(string); ok {
			schemaID = schemaIDVal
		}

		// Find or create nodes for source tables
		var sourceNodeIDs []uuid.UUID
		for _, tableName := range pattern.SourceTables {
			nodeID, err := s.findOrCreateNode(ctx, tableName, schemaID)
			if err == nil {
				sourceNodeIDs = append(sourceNodeIDs, nodeID)
			}
		}

		// Find or create node for target table
		targetNodeID, err := s.findOrCreateNode(ctx, pattern.TargetTable, schemaID)
		if err != nil {
			continue
		}

		// Create lineage edges
		for _, sourceNodeID := range sourceNodeIDs {
			discovered = append(discovered, DiscoveredLineage{
				ID:           uuid.New(),
				SourceNodeID: sourceNodeID,
				TargetNodeID: targetNodeID,
				EdgeType:     pattern.PatternType,
				Confidence:   0.8, // High confidence from actual queries
				Evidence: map[string]interface{}{
					"pattern_type": pattern.PatternType,
					"frequency":     pattern.Frequency,
					"first_seen":    pattern.FirstSeen,
					"last_seen":     pattern.LastSeen,
				},
			})
		}
	}

	return discovered, nil
}

/* findOrCreateNode finds or creates a lineage node */
func (s *QueryAnalysisService) findOrCreateNode(ctx context.Context,
	tableName, schemaID string) (uuid.UUID, error) {

	// Try to find existing node
	var nodeID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id
		FROM neuronip.lineage_nodes
		WHERE node_name = $1
		AND metadata->>'schema_id' = $2
		LIMIT 1`, tableName, schemaID,
	).Scan(&nodeID)

	if err == nil {
		return nodeID, nil
	}

	// Create new node
	nodeID = uuid.New()
	metadata := map[string]interface{}{
		"schema_id":     schemaID,
		"resource_type": "table",
		"resource_id":   tableName,
	}
	metadataJSON, _ := json.Marshal(metadata)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.lineage_nodes
		(id, node_type, node_name, metadata, created_at)
		VALUES ($1, 'table', $2, $3, NOW())`,
		nodeID, tableName, metadataJSON,
	)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create node: %w", err)
	}

	return nodeID, nil
}
