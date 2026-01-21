package lineage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ImpactService provides enhanced impact analysis functionality */
type ImpactService struct {
	pool *pgxpool.Pool
}

/* NewImpactService creates a new impact service */
func NewImpactService(pool *pgxpool.Pool) *ImpactService {
	return &ImpactService{pool: pool}
}

/* ImpactAnalysis represents enhanced impact analysis results */
type ImpactAnalysis struct {
	ID              uuid.UUID              `json:"id"`
	ResourceID      string                 `json:"resource_id"`
	ResourceType    string                 `json:"resource_type"`
	ImpactType      string                 `json:"impact_type"` // "upstream", "downstream", "both"
	AffectedResources []AffectedResource  `json:"affected_resources"`
	ImpactScore     float64                `json:"impact_score"` // 0.0 to 1.0
	CriticalPaths   []LineagePath         `json:"critical_paths,omitempty"`
	RiskLevel       string                 `json:"risk_level"` // "low", "medium", "high", "critical"
	Recommendations []string               `json:"recommendations,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

/* AffectedResource represents a resource affected by a change */
type AffectedResource struct {
	ResourceID    string                 `json:"resource_id"`
	ResourceType  string                 `json:"resource_type"`
	ResourceName  string                 `json:"resource_name"`
	System        string                 `json:"system"`
	ImpactLevel   string                 `json:"impact_level"` // "direct", "indirect", "transitive"
	HopDistance   int                    `json:"hop_distance"`
	Criticality   string                 `json:"criticality"` // "critical", "high", "medium", "low"
	Dependencies  []string               `json:"dependencies,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* AnalyzeImpact performs comprehensive impact analysis */
func (s *ImpactService) AnalyzeImpact(ctx context.Context,
	resourceID string, resourceType string,
	impactType string) (*ImpactAnalysis, error) {

	analysis := &ImpactAnalysis{
		ID:           uuid.New(),
		ResourceID:   resourceID,
		ResourceType: resourceType,
		ImpactType:   impactType,
		CreatedAt:    time.Now(),
	}

	if impactType == "upstream" || impactType == "both" {
		upstream, err := s.analyzeUpstream(ctx, resourceID, resourceType)
		if err == nil {
			analysis.AffectedResources = append(analysis.AffectedResources, upstream...)
		}
	}

	if impactType == "downstream" || impactType == "both" {
		downstream, err := s.analyzeDownstream(ctx, resourceID, resourceType)
		if err == nil {
			analysis.AffectedResources = append(analysis.AffectedResources, downstream...)
		}
	}

	// Calculate impact score
	analysis.ImpactScore = s.calculateImpactScore(analysis.AffectedResources)

	// Identify critical paths
	analysis.CriticalPaths = s.identifyCriticalPaths(ctx, resourceID, resourceType)

	// Determine risk level
	analysis.RiskLevel = s.determineRiskLevel(analysis.ImpactScore, len(analysis.AffectedResources))

	// Generate recommendations
	analysis.Recommendations = s.generateRecommendations(analysis)

	return analysis, nil
}

/* analyzeUpstream analyzes upstream dependencies */
func (s *ImpactService) analyzeUpstream(ctx context.Context,
	resourceID string, resourceType string) ([]AffectedResource, error) {

	// Find all upstream resources (resources that this depends on)
	query := `
		WITH RECURSIVE upstream_tree AS (
			-- Start with source node
			SELECT 
				ln.id,
				ln.node_name,
				ln.metadata->>'resource_type' as resource_type,
				ln.metadata->>'resource_id' as resource_id,
				ln.metadata->>'system' as system,
				0 as hop_distance,
				ARRAY[ln.id] as visited_nodes
			FROM neuronip.lineage_nodes ln
			WHERE ln.metadata->>'resource_type' = $1 
			AND (ln.metadata->>'resource_id' = $2 OR ln.id::text = $2)
			
			UNION ALL
			
			-- Recursively find upstream nodes
			SELECT 
				source_ln.id,
				source_ln.node_name,
				source_ln.metadata->>'resource_type' as resource_type,
				source_ln.metadata->>'resource_id' as resource_id,
				source_ln.metadata->>'system' as system,
				ut.hop_distance + 1,
				ut.visited_nodes || source_ln.id
			FROM upstream_tree ut
			JOIN neuronip.lineage_edges le ON le.target_node_id = ut.id
			JOIN neuronip.lineage_nodes source_ln ON source_ln.id = le.source_node_id
			WHERE NOT (source_ln.id = ANY(ut.visited_nodes))
			AND ut.hop_distance < 10
		)
		SELECT id, node_name, resource_type, resource_id, system, hop_distance
		FROM upstream_tree
		WHERE hop_distance > 0
		ORDER BY hop_distance ASC`

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze upstream: %w", err)
	}
	defer rows.Close()

	var affected []AffectedResource
	for rows.Next() {
		var res AffectedResource
		var nodeID uuid.UUID

		err := rows.Scan(&nodeID, &res.ResourceName, &res.ResourceType,
			&res.ResourceID, &res.System, &res.HopDistance)
		if err != nil {
			continue
		}

		if res.ResourceID == "" {
			res.ResourceID = nodeID.String()
		}

		res.ImpactLevel = s.determineImpactLevel(res.HopDistance)
		res.Criticality = s.determineCriticality(res.HopDistance, res.System)
		affected = append(affected, res)
	}

	return affected, nil
}

/* analyzeDownstream analyzes downstream dependencies */
func (s *ImpactService) analyzeDownstream(ctx context.Context,
	resourceID string, resourceType string) ([]AffectedResource, error) {

	// Find all downstream resources (resources that depend on this)
	query := `
		WITH RECURSIVE downstream_tree AS (
			-- Start with source node
			SELECT 
				ln.id,
				ln.node_name,
				ln.metadata->>'resource_type' as resource_type,
				ln.metadata->>'resource_id' as resource_id,
				ln.metadata->>'system' as system,
				0 as hop_distance,
				ARRAY[ln.id] as visited_nodes
			FROM neuronip.lineage_nodes ln
			WHERE ln.metadata->>'resource_type' = $1 
			AND (ln.metadata->>'resource_id' = $2 OR ln.id::text = $2)
			
			UNION ALL
			
			-- Recursively find downstream nodes
			SELECT 
				target_ln.id,
				target_ln.node_name,
				target_ln.metadata->>'resource_type' as resource_type,
				target_ln.metadata->>'resource_id' as resource_id,
				target_ln.metadata->>'system' as system,
				dt.hop_distance + 1,
				dt.visited_nodes || target_ln.id
			FROM downstream_tree dt
			JOIN neuronip.lineage_edges le ON le.source_node_id = dt.id
			JOIN neuronip.lineage_nodes target_ln ON target_ln.id = le.target_node_id
			WHERE NOT (target_ln.id = ANY(dt.visited_nodes))
			AND dt.hop_distance < 10
		)
		SELECT id, node_name, resource_type, resource_id, system, hop_distance
		FROM downstream_tree
		WHERE hop_distance > 0
		ORDER BY hop_distance ASC`

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze downstream: %w", err)
	}
	defer rows.Close()

	var affected []AffectedResource
	for rows.Next() {
		var res AffectedResource
		var nodeID uuid.UUID

		err := rows.Scan(&nodeID, &res.ResourceName, &res.ResourceType,
			&res.ResourceID, &res.System, &res.HopDistance)
		if err != nil {
			continue
		}

		if res.ResourceID == "" {
			res.ResourceID = nodeID.String()
		}

		res.ImpactLevel = s.determineImpactLevel(res.HopDistance)
		res.Criticality = s.determineCriticality(res.HopDistance, res.System)
		affected = append(affected, res)
	}

	return affected, nil
}

/* determineImpactLevel determines impact level based on hop distance */
func (s *ImpactService) determineImpactLevel(hopDistance int) string {
	if hopDistance == 1 {
		return "direct"
	} else if hopDistance <= 3 {
		return "indirect"
	}
	return "transitive"
}

/* determineCriticality determines resource criticality */
func (s *ImpactService) determineCriticality(hopDistance int, system string) string {
	if hopDistance == 1 {
		return "critical"
	} else if hopDistance <= 2 {
		return "high"
	} else if hopDistance <= 4 {
		return "medium"
	}
	return "low"
}

/* calculateImpactScore calculates overall impact score */
func (s *ImpactService) calculateImpactScore(affected []AffectedResource) float64 {
	if len(affected) == 0 {
		return 0.0
	}

	score := 0.0
	totalWeight := 0.0

	for _, res := range affected {
		weight := 1.0 / float64(res.HopDistance)
		totalWeight += weight

		var criticalityWeight float64
		switch res.Criticality {
		case "critical":
			criticalityWeight = 1.0
		case "high":
			criticalityWeight = 0.75
		case "medium":
			criticalityWeight = 0.5
		default:
			criticalityWeight = 0.25
		}

		score += weight * criticalityWeight
	}

	if totalWeight > 0 {
		score = score / totalWeight
	}

	if score > 1.0 {
		return 1.0
	}

	return score
}

/* identifyCriticalPaths identifies critical paths in the lineage graph */
func (s *ImpactService) identifyCriticalPaths(ctx context.Context,
	resourceID string, resourceType string) []LineagePath {

	// Find paths with highest impact (shortest paths to critical resources)
	query := `
		WITH RECURSIVE critical_paths AS (
			SELECT 
				ln.id as start_node_id,
				ln.metadata->>'resource_id' as start_resource_id,
				target_ln.id as end_node_id,
				target_ln.metadata->>'resource_id' as end_resource_id,
				target_ln.metadata->>'system' as end_system,
				1 as path_length,
				ARRAY[ln.id, target_ln.id] as path_nodes
			FROM neuronip.lineage_nodes ln
			JOIN neuronip.lineage_edges le ON le.source_node_id = ln.id
			JOIN neuronip.lineage_nodes target_ln ON target_ln.id = le.target_node_id
			WHERE ln.metadata->>'resource_type' = $1 
			AND (ln.metadata->>'resource_id' = $2 OR ln.id::text = $2)
			
			UNION ALL
			
			SELECT 
				cp.start_node_id,
				cp.start_resource_id,
				target_ln.id as end_node_id,
				target_ln.metadata->>'resource_id' as end_resource_id,
				target_ln.metadata->>'system' as end_system,
				cp.path_length + 1,
				cp.path_nodes || target_ln.id
			FROM critical_paths cp
			JOIN neuronip.lineage_edges le ON le.source_node_id = cp.end_node_id
			JOIN neuronip.lineage_nodes target_ln ON target_ln.id = le.target_node_id
			WHERE NOT (target_ln.id = ANY(cp.path_nodes))
			AND cp.path_length < 5
		)
		SELECT DISTINCT path_length, end_system
		FROM critical_paths
		ORDER BY path_length ASC
		LIMIT 10`

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return []LineagePath{}
	}
	defer rows.Close()

	var paths []LineagePath
	for rows.Next() {
		var length int
		var system string

		err := rows.Scan(&length, &system)
		if err != nil {
			continue
		}

		path := LineagePath{
			ID:          uuid.New(),
			TotalHops:   length,
			TargetSystem: system,
			Confidence: 1.0 - (float64(length) * 0.1),
			CreatedAt:   time.Now(),
		}

		if path.Confidence < 0.5 {
			path.Confidence = 0.5
		}

		paths = append(paths, path)
	}

	return paths
}

/* determineRiskLevel determines overall risk level */
func (s *ImpactService) determineRiskLevel(impactScore float64, affectedCount int) string {
	if impactScore >= 0.8 || affectedCount >= 50 {
		return "critical"
	} else if impactScore >= 0.6 || affectedCount >= 20 {
		return "high"
	} else if impactScore >= 0.4 || affectedCount >= 10 {
		return "medium"
	}
	return "low"
}

/* generateRecommendations generates recommendations based on impact analysis */
func (s *ImpactService) generateRecommendations(analysis *ImpactAnalysis) []string {
	var recommendations []string

	if analysis.RiskLevel == "critical" || analysis.RiskLevel == "high" {
		recommendations = append(recommendations,
			"High impact detected. Consider implementing change management process.",
			"Notify all affected downstream systems before making changes.",
			"Run comprehensive tests on affected resources.")
	}

	if len(analysis.CriticalPaths) > 5 {
		recommendations = append(recommendations,
			"Multiple critical paths detected. Consider breaking dependencies.")
	}

	if analysis.ImpactScore > 0.7 {
		recommendations = append(recommendations,
			"Consider implementing feature flags or gradual rollout.",
			"Document all dependencies for future reference.")
	}

	crossSystemCount := 0
	for _, res := range analysis.AffectedResources {
		if res.System != "" {
			crossSystemCount++
		}
	}

	if crossSystemCount > 0 {
		recommendations = append(recommendations,
			"Cross-system dependencies detected. Coordinate with other teams.")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"Impact is manageable. Proceed with standard change process.")
	}

	return recommendations
}

/* GetImpactHistory retrieves impact analysis history */
func (s *ImpactService) GetImpactHistory(ctx context.Context,
	resourceID string, limit int) ([]ImpactAnalysis, error) {

	if limit == 0 {
		limit = 10
	}

	query := `
		SELECT id, resource_id, resource_type, impact_type, impact_score,
		       risk_level, recommendations, created_at, metadata
		FROM neuronip.impact_analysis
		WHERE resource_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get impact history: %w", err)
	}
	defer rows.Close()

	var analyses []ImpactAnalysis
	for rows.Next() {
		var analysis ImpactAnalysis
		var recommendationsJSON, metadataJSON []byte

		err := rows.Scan(&analysis.ID, &analysis.ResourceID, &analysis.ResourceType,
			&analysis.ImpactType, &analysis.ImpactScore, &analysis.RiskLevel,
			&recommendationsJSON, &analysis.CreatedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(recommendationsJSON, &analysis.Recommendations)
		json.Unmarshal(metadataJSON, &analysis.Metadata)
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}
