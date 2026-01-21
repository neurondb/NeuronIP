package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* MetricsService provides metrics and semantic layer functionality */
type MetricsService struct {
	pool *pgxpool.Pool
}

/* NewMetricsService creates a new metrics service */
func NewMetricsService(pool *pgxpool.Pool) *MetricsService {
	return &MetricsService{pool: pool}
}

/* Metric represents a metric definition */
type Metric struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Definition  string                 `json:"definition"`
	KPIType     *string                `json:"kpi_type,omitempty"`
	BusinessTerm *string               `json:"business_term,omitempty"`
	Reusable    bool                   `json:"reusable"`
	Dimensions  []MetricDimension      `json:"dimensions,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* MetricDimension represents a metric dimension */
type MetricDimension struct {
	ID            uuid.UUID `json:"id"`
	MetricID      uuid.UUID `json:"metric_id"`
	DimensionName string    `json:"dimension_name"`
	DimensionType string    `json:"dimension_type"`
	CreatedAt     time.Time `json:"created_at"`
}

/* CreateMetric creates a new metric */
func (s *MetricsService) CreateMetric(ctx context.Context, m Metric) (*Metric, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO neuronip.metrics (id, name, definition, kpi_type, business_term, reusable, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, definition, kpi_type, business_term, reusable, created_at, updated_at`

	var result Metric
	err := s.pool.QueryRow(ctx, query,
		id, m.Name, m.Definition, m.KPIType, m.BusinessTerm, m.Reusable, now, now,
	).Scan(
		&result.ID, &result.Name, &result.Definition, &result.KPIType,
		&result.BusinessTerm, &result.Reusable, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}

	// Add dimensions if provided
	if len(m.Dimensions) > 0 {
		for _, dim := range m.Dimensions {
			dimQuery := `
				INSERT INTO neuronip.metric_dimensions (id, metric_id, dimension_name, dimension_type, created_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4)`
			s.pool.Exec(ctx, dimQuery, result.ID, dim.DimensionName, dim.DimensionType, now)
		}
		// Reload dimensions
		dims, _ := s.getDimensions(ctx, result.ID)
		result.Dimensions = dims
	}

	return &result, nil
}

/* GetMetric retrieves a metric by ID */
func (s *MetricsService) GetMetric(ctx context.Context, id uuid.UUID) (*Metric, error) {
	query := `
		SELECT id, name, definition, kpi_type, business_term, reusable, created_at, updated_at
		FROM neuronip.metrics
		WHERE id = $1`

	var result Metric
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.Name, &result.Definition, &result.KPIType,
		&result.BusinessTerm, &result.Reusable, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	dims, _ := s.getDimensions(ctx, result.ID)
	result.Dimensions = dims

	return &result, nil
}

/* ListMetrics lists all metrics */
func (s *MetricsService) ListMetrics(ctx context.Context) ([]Metric, error) {
	query := `
		SELECT id, name, definition, kpi_type, business_term, reusable, created_at, updated_at
		FROM neuronip.metrics
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics: %w", err)
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		err := rows.Scan(
			&m.ID, &m.Name, &m.Definition, &m.KPIType,
			&m.BusinessTerm, &m.Reusable, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			continue
		}

		dims, _ := s.getDimensions(ctx, m.ID)
		m.Dimensions = dims
		metrics = append(metrics, m)
	}

	return metrics, nil
}

/* UpdateMetric updates a metric */
func (s *MetricsService) UpdateMetric(ctx context.Context, id uuid.UUID, m Metric) (*Metric, error) {
	query := `
		UPDATE neuronip.metrics
		SET name = $1, definition = $2, kpi_type = $3, business_term = $4, reusable = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, definition, kpi_type, business_term, reusable, created_at, updated_at`

	var result Metric
	err := s.pool.QueryRow(ctx, query,
		m.Name, m.Definition, m.KPIType, m.BusinessTerm, m.Reusable, id,
	).Scan(
		&result.ID, &result.Name, &result.Definition, &result.KPIType,
		&result.BusinessTerm, &result.Reusable, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update metric: %w", err)
	}

	dims, _ := s.getDimensions(ctx, result.ID)
	result.Dimensions = dims

	return &result, nil
}

/* DeleteMetric deletes a metric */
func (s *MetricsService) DeleteMetric(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.metrics WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

/* SearchMetrics searches metrics by business terms */
func (s *MetricsService) SearchMetrics(ctx context.Context, query string) ([]Metric, error) {
	searchQuery := `
		SELECT id, name, definition, kpi_type, business_term, reusable, created_at, updated_at
		FROM neuronip.metrics
		WHERE name ILIKE $1 OR definition ILIKE $1 OR business_term ILIKE $1
		ORDER BY created_at DESC`

	pattern := "%" + query + "%"
	rows, err := s.pool.Query(ctx, searchQuery, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search metrics: %w", err)
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		err := rows.Scan(
			&m.ID, &m.Name, &m.Definition, &m.KPIType,
			&m.BusinessTerm, &m.Reusable, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			continue
		}

		dims, _ := s.getDimensions(ctx, m.ID)
		m.Dimensions = dims
		metrics = append(metrics, m)
	}

	return metrics, nil
}

/* getDimensions retrieves dimensions for a metric */
func (s *MetricsService) getDimensions(ctx context.Context, metricID uuid.UUID) ([]MetricDimension, error) {
	query := `
		SELECT id, metric_id, dimension_name, dimension_type, created_at
		FROM neuronip.metric_dimensions
		WHERE metric_id = $1
		ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dimensions []MetricDimension
	for rows.Next() {
		var dim MetricDimension
		err := rows.Scan(
			&dim.ID, &dim.MetricID, &dim.DimensionName, &dim.DimensionType, &dim.CreatedAt,
		)
		if err != nil {
			continue
		}
		dimensions = append(dimensions, dim)
	}

	return dimensions, nil
}

/* CalculateMetric calculates a metric value by evaluating its SQL expression */
func (s *MetricsService) CalculateMetric(ctx context.Context, metricID uuid.UUID, filters map[string]interface{}) (interface{}, error) {
	metric, err := s.GetMetric(ctx, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	// Build SQL query from definition
	sqlQuery := metric.Definition

	// Apply filters if provided
	if len(filters) > 0 {
		whereClauses := []string{}
		args := []interface{}{}
		argIndex := 1

		for key, value := range filters {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}

		if len(whereClauses) > 0 {
			if strings.Contains(strings.ToUpper(sqlQuery), "WHERE") {
				sqlQuery += " AND " + strings.Join(whereClauses, " AND ")
			} else {
				sqlQuery += " WHERE " + strings.Join(whereClauses, " AND ")
			}
		}
	}

	// Execute SQL query
	rows, err := s.pool.Query(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute metric calculation: %w", err)
	}
	defer rows.Close()

	// Get first row result
	if !rows.Next() {
		return nil, fmt.Errorf("metric calculation returned no results")
	}

	var result interface{}
	err = rows.Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to scan metric result: %w", err)
	}

	return result, nil
}

/* AddMetricLineage adds a lineage relationship for a metric */
func (s *MetricsService) AddMetricLineage(ctx context.Context, metricID uuid.UUID, dependsOnMetricID *uuid.UUID, dependsOnTable *string, dependsOnColumn *string, relationshipType string) error {
	query := `
		INSERT INTO neuronip.metric_lineage 
		(metric_id, depends_on_metric_id, depends_on_table, depends_on_column, relationship_type, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`

	_, err := s.pool.Exec(ctx, query, metricID, dependsOnMetricID, dependsOnTable, dependsOnColumn, relationshipType)
	if err != nil {
		return fmt.Errorf("failed to add metric lineage: %w", err)
	}

	return nil
}

/* GetMetricLineage retrieves lineage for a metric */
func (s *MetricsService) GetMetricLineage(ctx context.Context, metricID uuid.UUID) ([]MetricLineage, error) {
	query := `
		SELECT id, metric_id, depends_on_metric_id, depends_on_table, depends_on_column, 
		       relationship_type, created_at
		FROM neuronip.metric_lineage
		WHERE metric_id = $1`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric lineage: %w", err)
	}
	defer rows.Close()

	var lineage []MetricLineage
	for rows.Next() {
		var l MetricLineage
		var dependsOnMetricID sql.NullString
		var dependsOnTable sql.NullString
		var dependsOnColumn sql.NullString

		err := rows.Scan(
			&l.ID, &l.MetricID, &dependsOnMetricID, &dependsOnTable, &dependsOnColumn,
			&l.RelationshipType, &l.CreatedAt,
		)
		if err != nil {
			continue
		}

		if dependsOnMetricID.Valid {
			if id, err := uuid.Parse(dependsOnMetricID.String); err == nil {
				l.DependsOnMetricID = &id
			}
		}
		if dependsOnTable.Valid {
			l.DependsOnTable = &dependsOnTable.String
		}
		if dependsOnColumn.Valid {
			l.DependsOnColumn = &dependsOnColumn.String
		}

		lineage = append(lineage, l)
	}

	return lineage, nil
}

/* MetricLineage represents a metric lineage relationship */
type MetricLineage struct {
	ID               uuid.UUID  `json:"id"`
	MetricID         uuid.UUID  `json:"metric_id"`
	DependsOnMetricID *uuid.UUID `json:"depends_on_metric_id,omitempty"`
	DependsOnTable   *string    `json:"depends_on_table,omitempty"`
	DependsOnColumn  *string    `json:"depends_on_column,omitempty"`
	RelationshipType string     `json:"relationship_type"`
	CreatedAt        time.Time  `json:"created_at"`
}

/* ResolveMetricDependencies resolves all dependencies for a metric */
func (s *MetricsService) ResolveMetricDependencies(ctx context.Context, metricID uuid.UUID) ([]Metric, error) {
	visited := make(map[uuid.UUID]bool)
	dependencies := []Metric{}

	var resolve func(uuid.UUID) error
	resolve = func(id uuid.UUID) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		lineage, err := s.GetMetricLineage(ctx, id)
		if err != nil {
			return err
		}

		for _, l := range lineage {
			if l.DependsOnMetricID != nil {
				depMetric, err := s.GetMetric(ctx, *l.DependsOnMetricID)
				if err != nil {
					continue
				}
				dependencies = append(dependencies, *depMetric)
				if err := resolve(*l.DependsOnMetricID); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if err := resolve(metricID); err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	return dependencies, nil
}

/* ApproveMetric approves a metric in the catalog */
func (s *MetricsService) ApproveMetric(ctx context.Context, metricID uuid.UUID, approvedBy string) error {
	query := `
		UPDATE neuronip.metric_catalog
		SET status = 'approved', approved_at = NOW(), approved_by = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := s.pool.Exec(ctx, query, approvedBy, metricID)
	if err != nil {
		return fmt.Errorf("failed to approve metric: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("metric not found")
	}

	return nil
}

/* CreateMetricVersion creates a new version of a metric */
func (s *MetricsService) CreateMetricVersion(ctx context.Context, metricID uuid.UUID, newVersion string, changes map[string]interface{}) (*Metric, error) {
	// Get current metric
	current, err := s.GetMetric(ctx, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current metric: %w", err)
	}

	// Create new version
	newMetric := *current
	newMetric.ID = uuid.New()
	newMetric.CreatedAt = time.Now()
	newMetric.UpdatedAt = time.Now()

	// Apply changes
	if name, ok := changes["name"].(string); ok {
		newMetric.Name = name
	}
	if definition, ok := changes["definition"].(string); ok {
		newMetric.Definition = definition
	}
	if kpiType, ok := changes["kpi_type"].(string); ok {
		newMetric.KPIType = &kpiType
	}
	if businessTerm, ok := changes["business_term"].(string); ok {
		newMetric.BusinessTerm = &businessTerm
	}
	if reusable, ok := changes["reusable"].(bool); ok {
		newMetric.Reusable = reusable
	}

	// Create new metric
	return s.CreateMetric(ctx, newMetric)
}

/* DiscoverMetrics discovers metrics based on search criteria */
func (s *MetricsService) DiscoverMetrics(ctx context.Context, criteria MetricDiscoveryCriteria) ([]Metric, error) {
	query := `
		SELECT id, name, definition, kpi_type, business_term, reusable, created_at, updated_at
		FROM neuronip.metrics
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if criteria.KPIType != nil {
		query += fmt.Sprintf(" AND kpi_type = $%d", argIndex)
		args = append(args, *criteria.KPIType)
		argIndex++
	}

	if criteria.BusinessTerm != nil {
		query += fmt.Sprintf(" AND business_term ILIKE $%d", argIndex)
		args = append(args, "%"+*criteria.BusinessTerm+"%")
		argIndex++
	}

	if criteria.Reusable != nil {
		query += fmt.Sprintf(" AND reusable = $%d", argIndex)
		args = append(args, *criteria.Reusable)
		argIndex++
	}

	if criteria.SearchQuery != nil {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR definition ILIKE $%d OR business_term ILIKE $%d)", argIndex, argIndex, argIndex)
		pattern := "%" + *criteria.SearchQuery + "%"
		args = append(args, pattern)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if criteria.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, criteria.Limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to discover metrics: %w", err)
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		err := rows.Scan(
			&m.ID, &m.Name, &m.Definition, &m.KPIType,
			&m.BusinessTerm, &m.Reusable, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			continue
		}

		dims, _ := s.getDimensions(ctx, m.ID)
		m.Dimensions = dims
		metrics = append(metrics, m)
	}

	return metrics, nil
}

/* MetricDiscoveryCriteria represents criteria for metric discovery */
type MetricDiscoveryCriteria struct {
	KPIType      *string
	BusinessTerm *string
	Reusable     *bool
	SearchQuery  *string
	Limit        int
}
