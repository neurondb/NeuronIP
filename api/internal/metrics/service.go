package metrics

import (
	"context"
	"encoding/json"
	"fmt"
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
