package semantic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* LineageService provides metric dependency tracking and lineage analysis */
type LineageService struct {
	pool *pgxpool.Pool
}

/* NewLineageService creates a new lineage service */
func NewLineageService(pool *pgxpool.Pool) *LineageService {
	return &LineageService{pool: pool}
}

/* MetricDependency represents a metric dependency */
type MetricDependency struct {
	ID               uuid.UUID `json:"id"`
	MetricID         uuid.UUID `json:"metric_id"`
	DependsOnMetricID *uuid.UUID `json:"depends_on_metric_id,omitempty"`
	DependsOnDatasetID *uuid.UUID `json:"depends_on_dataset_id,omitempty"`
	DependencyType   string    `json:"dependency_type"` // metric, dataset, table, column
	CreatedAt        string    `json:"created_at"`
}

/* AddMetricDependency adds a dependency for a metric */
func (s *LineageService) AddMetricDependency(ctx context.Context, metricID uuid.UUID, dependsOnMetricID *uuid.UUID, dependsOnDatasetID *uuid.UUID, dependencyType string) error {
	id := uuid.New()

	query := `
		INSERT INTO neuronip.metric_lineage 
		(id, metric_id, depends_on_metric_id, depends_on_dataset_id, dependency_type, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT DO NOTHING`

	_, err := s.pool.Exec(ctx, query, id, metricID, dependsOnMetricID, dependsOnDatasetID, dependencyType)
	if err != nil {
		return fmt.Errorf("failed to add metric dependency: %w", err)
	}

	return nil
}

/* GetMetricDependencies retrieves all dependencies for a metric */
func (s *LineageService) GetMetricDependencies(ctx context.Context, metricID uuid.UUID) ([]MetricDependency, error) {
	query := `
		SELECT id, metric_id, depends_on_metric_id, depends_on_dataset_id, dependency_type, created_at
		FROM neuronip.metric_lineage
		WHERE metric_id = $1
		ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []MetricDependency
	for rows.Next() {
		var dep MetricDependency
		var dependsOnMetricID, dependsOnDatasetID *uuid.UUID
		var createdAt string

		err := rows.Scan(&dep.ID, &dep.MetricID, &dependsOnMetricID, &dependsOnDatasetID, &dep.DependencyType, &createdAt)
		if err != nil {
			continue
		}

		dep.DependsOnMetricID = dependsOnMetricID
		dep.DependsOnDatasetID = dependsOnDatasetID
		dep.CreatedAt = createdAt
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

/* GetMetricLineage retrieves the full lineage graph for a metric */
func (s *LineageService) GetMetricLineage(ctx context.Context, metricID uuid.UUID, maxDepth int) (map[string]interface{}, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}

	// Use recursive CTE to traverse the dependency graph
	query := `
		WITH RECURSIVE lineage_tree AS (
			-- Base case: start with the metric
			SELECT 
				ml.id,
				ml.metric_id,
				ml.depends_on_metric_id,
				ml.depends_on_dataset_id,
				ml.dependency_type,
				0 as depth,
				ARRAY[ml.metric_id] as path
			FROM neuronip.metric_lineage ml
			WHERE ml.metric_id = $1

			UNION ALL

			-- Recursive case: follow dependencies
			SELECT 
				ml.id,
				ml.metric_id,
				ml.depends_on_metric_id,
				ml.depends_on_dataset_id,
				ml.dependency_type,
				lt.depth + 1,
				lt.path || ml.metric_id
			FROM neuronip.metric_lineage ml
			INNER JOIN lineage_tree lt ON ml.metric_id = lt.depends_on_metric_id
			WHERE lt.depth < $2
				AND ml.metric_id != ALL(lt.path) -- Prevent cycles
		)
		SELECT 
			id,
			metric_id,
			depends_on_metric_id,
			depends_on_dataset_id,
			dependency_type,
			depth
		FROM lineage_tree
		ORDER BY depth, metric_id`

	rows, err := s.pool.Query(ctx, query, metricID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric lineage: %w", err)
	}
	defer rows.Close()

	nodes := make(map[uuid.UUID]bool)
	edges := []map[string]interface{}{}

	for rows.Next() {
		var id, metricID uuid.UUID
		var dependsOnMetricID, dependsOnDatasetID *uuid.UUID
		var dependencyType string
		var depth int

		err := rows.Scan(&id, &metricID, &dependsOnMetricID, &dependsOnDatasetID, &dependencyType, &depth)
		if err != nil {
			continue
		}

		nodes[metricID] = true
		if dependsOnMetricID != nil {
			nodes[*dependsOnMetricID] = true
			edges = append(edges, map[string]interface{}{
				"from": metricID,
				"to":   *dependsOnMetricID,
				"type": dependencyType,
				"depth": depth,
			})
		}

		if dependsOnDatasetID != nil {
			edges = append(edges, map[string]interface{}{
				"from": metricID,
				"to":   *dependsOnDatasetID,
				"type": dependencyType,
				"depth": depth,
			})
		}
	}

	nodeList := make([]uuid.UUID, 0, len(nodes))
	for node := range nodes {
		nodeList = append(nodeList, node)
	}

	return map[string]interface{}{
		"metric_id": metricID,
		"nodes":     nodeList,
		"edges":     edges,
		"max_depth": maxDepth,
	}, nil
}

/* GetImpactAnalysis analyzes the impact of changing a metric */
func (s *LineageService) GetImpactAnalysis(ctx context.Context, metricID uuid.UUID) (map[string]interface{}, error) {
	// Find all metrics that depend on this metric
	query := `
		SELECT DISTINCT ml.metric_id, bm.name, bm.display_name
		FROM neuronip.metric_lineage ml
		JOIN neuronip.business_metrics bm ON ml.metric_id = bm.id
		WHERE ml.depends_on_metric_id = $1`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get impact analysis: %w", err)
	}
	defer rows.Close()

	var affectedMetrics []map[string]interface{}
	for rows.Next() {
		var affectedID uuid.UUID
		var name, displayName string

		err := rows.Scan(&affectedID, &name, &displayName)
		if err != nil {
			continue
		}

		affectedMetrics = append(affectedMetrics, map[string]interface{}{
			"id":           affectedID,
			"name":         name,
			"display_name": displayName,
		})
	}

	// Get datasets that this metric depends on
	datasetQuery := `
		SELECT DISTINCT depends_on_dataset_id
		FROM neuronip.metric_lineage
		WHERE metric_id = $1 AND depends_on_dataset_id IS NOT NULL`

	datasetRows, err := s.pool.Query(ctx, datasetQuery, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get datasets: %w", err)
	}
	defer datasetRows.Close()

	var datasets []uuid.UUID
	for datasetRows.Next() {
		var datasetID uuid.UUID
		err := datasetRows.Scan(&datasetID)
		if err != nil {
			continue
		}
		datasets = append(datasets, datasetID)
	}

	return map[string]interface{}{
		"metric_id":        metricID,
		"affected_metrics": affectedMetrics,
		"affected_count":   len(affectedMetrics),
		"datasets":         datasets,
	}, nil
}

/* TraverseGraph performs graph traversal from a starting metric */
func (s *LineageService) TraverseGraph(ctx context.Context, startMetricID uuid.UUID, direction string, maxDepth int) (map[string]interface{}, error) {
	if direction != "upstream" && direction != "downstream" {
		direction = "downstream"
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}

	var query string
	if direction == "downstream" {
		// Find metrics that depend on this metric
		query = `
			WITH RECURSIVE downstream AS (
				SELECT metric_id, depends_on_metric_id, dependency_type, 0 as depth
				FROM neuronip.metric_lineage
				WHERE depends_on_metric_id = $1
				UNION ALL
				SELECT ml.metric_id, ml.depends_on_metric_id, ml.dependency_type, d.depth + 1
				FROM neuronip.metric_lineage ml
				JOIN downstream d ON ml.depends_on_metric_id = d.metric_id
				WHERE d.depth < $2
			)
			SELECT DISTINCT metric_id, dependency_type, depth
			FROM downstream
			ORDER BY depth`
	} else {
		// Find metrics that this metric depends on
		query = `
			WITH RECURSIVE upstream AS (
				SELECT metric_id, depends_on_metric_id, dependency_type, 0 as depth
				FROM neuronip.metric_lineage
				WHERE metric_id = $1
				UNION ALL
				SELECT ml.metric_id, ml.depends_on_metric_id, ml.dependency_type, u.depth + 1
				FROM neuronip.metric_lineage ml
				JOIN upstream u ON ml.metric_id = u.depends_on_metric_id
				WHERE u.depth < $2
			)
			SELECT DISTINCT COALESCE(depends_on_metric_id, metric_id) as metric_id, dependency_type, depth
			FROM upstream
			ORDER BY depth`
	}

	rows, err := s.pool.Query(ctx, query, startMetricID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to traverse graph: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var metricID uuid.UUID
		var dependencyType string
		var depth int

		err := rows.Scan(&metricID, &dependencyType, &depth)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"metric_id":      metricID,
			"dependency_type": dependencyType,
			"depth":          depth,
		})
	}

	return map[string]interface{}{
		"start_metric_id": startMetricID,
		"direction":       direction,
		"results":         results,
		"max_depth":       maxDepth,
	}, nil
}
