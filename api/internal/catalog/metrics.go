package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* MetricsService provides metric catalog functionality */
type MetricsService struct {
	pool *pgxpool.Pool
}

/* NewMetricsService creates a new metrics service */
func NewMetricsService(pool *pgxpool.Pool) *MetricsService {
	return &MetricsService{pool: pool}
}

/* Metric represents a metric definition */
type Metric struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Description     *string                `json:"description,omitempty"`
	SQLExpression   string                 `json:"sql_expression"`
	MetricType      string                 `json:"metric_type"`
	Unit            *string                `json:"unit,omitempty"`
	Category        *string                `json:"category,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	OwnerID         *string                `json:"owner_id,omitempty"`
	Version         string                 `json:"version"`
	Status          string                 `json:"status"`
	ApprovalWorkflow map[string]interface{} `json:"approval_workflow,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ApprovedAt      *time.Time             `json:"approved_at,omitempty"`
	ApprovedBy      *string                `json:"approved_by,omitempty"`
}

/* CreateMetric creates a new metric */
func (s *MetricsService) CreateMetric(ctx context.Context, metric Metric) (*Metric, error) {
	metric.ID = uuid.New()
	metric.CreatedAt = time.Now()
	metric.UpdatedAt = time.Now()
	
	if metric.Version == "" {
		metric.Version = "1.0.0"
	}
	if metric.Status == "" {
		metric.Status = "draft"
	}
	
	tagsJSON, _ := json.Marshal(metric.Tags)
	workflowJSON, _ := json.Marshal(metric.ApprovalWorkflow)
	
	query := `
		INSERT INTO neuronip.metric_catalog 
		(id, name, display_name, description, sql_expression, metric_type, unit, category, 
		 tags, owner_id, version, status, approval_workflow, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at, approved_at, approved_by`
	
	err := s.pool.QueryRow(ctx, query,
		metric.ID, metric.Name, metric.DisplayName, metric.Description, metric.SQLExpression,
		metric.MetricType, metric.Unit, metric.Category, tagsJSON, metric.OwnerID,
		metric.Version, metric.Status, workflowJSON, metric.CreatedAt, metric.UpdatedAt,
	).Scan(&metric.ID, &metric.CreatedAt, &metric.UpdatedAt, &metric.ApprovedAt, &metric.ApprovedBy)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}
	
	return &metric, nil
}

/* GetMetric retrieves a metric by ID */
func (s *MetricsService) GetMetric(ctx context.Context, id uuid.UUID) (*Metric, error) {
	query := `
		SELECT id, name, display_name, description, sql_expression, metric_type, unit, category,
		       tags, owner_id, version, status, approval_workflow, created_at, updated_at, approved_at, approved_by
		FROM neuronip.metric_catalog
		WHERE id = $1`
	
	var metric Metric
	var tagsJSON, workflowJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&metric.ID, &metric.Name, &metric.DisplayName, &metric.Description, &metric.SQLExpression,
		&metric.MetricType, &metric.Unit, &metric.Category, &tagsJSON, &metric.OwnerID,
		&metric.Version, &metric.Status, &workflowJSON, &metric.CreatedAt, &metric.UpdatedAt,
		&metric.ApprovedAt, &metric.ApprovedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}
	
	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &metric.Tags)
	}
	if workflowJSON != nil {
		json.Unmarshal(workflowJSON, &metric.ApprovalWorkflow)
	}
	
	return &metric, nil
}

/* ListMetrics lists all metrics */
func (s *MetricsService) ListMetrics(ctx context.Context, status *string, category *string, limit int) ([]Metric, error) {
	query := `SELECT id, name, display_name, description, sql_expression, metric_type, unit, category,
		       tags, owner_id, version, status, approval_workflow, created_at, updated_at, approved_at, approved_by
		FROM neuronip.metric_catalog WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}
	
	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, *category)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics: %w", err)
	}
	defer rows.Close()
	
	metrics := make([]Metric, 0)
	for rows.Next() {
		var metric Metric
		var tagsJSON, workflowJSON json.RawMessage
		
		err := rows.Scan(
			&metric.ID, &metric.Name, &metric.DisplayName, &metric.Description, &metric.SQLExpression,
			&metric.MetricType, &metric.Unit, &metric.Category, &tagsJSON, &metric.OwnerID,
			&metric.Version, &metric.Status, &workflowJSON, &metric.CreatedAt, &metric.UpdatedAt,
			&metric.ApprovedAt, &metric.ApprovedBy,
		)
		if err != nil {
			continue
		}
		
		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &metric.Tags)
		}
		if workflowJSON != nil {
			json.Unmarshal(workflowJSON, &metric.ApprovalWorkflow)
		}
		
		metrics = append(metrics, metric)
	}
	
	return metrics, nil
}

/* AddMetricLineage adds a lineage relationship */
func (s *MetricsService) AddMetricLineage(ctx context.Context, metricID uuid.UUID, dependsOnMetricID *uuid.UUID, dependsOnTable *string, dependsOnColumn *string, relationshipType string) error {
	query := `
		INSERT INTO neuronip.metric_lineage 
		(metric_id, depends_on_metric_id, depends_on_table, depends_on_column, relationship_type, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`
	
	_, err := s.pool.Exec(ctx, query, metricID, dependsOnMetricID, dependsOnTable, dependsOnColumn, relationshipType)
	return err
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
	
	lineage := make([]MetricLineage, 0)
	for rows.Next() {
		var l MetricLineage
		err := rows.Scan(
			&l.ID, &l.MetricID, &l.DependsOnMetricID, &l.DependsOnTable, &l.DependsOnColumn,
			&l.RelationshipType, &l.CreatedAt,
		)
		if err != nil {
			continue
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
