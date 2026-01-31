package enterprise

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* EnterpriseIngestionService provides enterprise-grade ingestion capabilities */
type EnterpriseIngestionService struct {
	pool            *pgxpool.Pool
	ingestionService *ingestion.IngestionService
}

/* NewService creates a new enterprise ingestion service */
func NewService(pool *pgxpool.Pool, ingestionService *ingestion.IngestionService) *EnterpriseIngestionService {
	return &EnterpriseIngestionService{
		pool:            pool,
		ingestionService: ingestionService,
	}
}

/* CreateScheduledIngestion creates a scheduled ingestion job */
func (s *EnterpriseIngestionService) CreateScheduledIngestion(ctx context.Context, dataSourceID uuid.UUID, schedule ScheduleConfig) (*ScheduledIngestion, error) {
	scheduleID := uuid.New()
	scheduleJSON, _ := json.Marshal(schedule)

	// Calculate next run time
	nextRun := s.calculateNextRun(schedule)

	query := `
		INSERT INTO neuronip.ingestion_schedules 
		(id, data_source_id, schedule_type, schedule_config, enabled, next_run_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, data_source_id, schedule_type, schedule_config, enabled, next_run_at, last_run_at, created_at, updated_at`

	var scheduled ScheduledIngestion
	var scheduleJSONRaw json.RawMessage
	var nextRunAt, lastRunAt *time.Time

	err := s.pool.QueryRow(ctx, query, scheduleID, dataSourceID, schedule.Type, scheduleJSON, schedule.Enabled, nextRun).Scan(
		&scheduled.ID, &scheduled.DataSourceID, &scheduled.ScheduleType, &scheduleJSONRaw,
		&scheduled.Enabled, &nextRunAt, &lastRunAt, &scheduled.CreatedAt, &scheduled.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduled ingestion: %w", err)
	}

	if scheduleJSONRaw != nil {
		json.Unmarshal(scheduleJSONRaw, &scheduled.ScheduleConfig)
	}
	if nextRunAt != nil {
		scheduled.NextRunAt = nextRunAt
	}
	if lastRunAt != nil {
		scheduled.LastRunAt = lastRunAt
	}

	return &scheduled, nil
}

/* RecordIngestionMetric records an ingestion metric */
func (s *EnterpriseIngestionService) RecordIngestionMetric(ctx context.Context, jobID *uuid.UUID, dataSourceID uuid.UUID, metricName string, metricValue float64, unit *string, metadata map[string]interface{}) error {
	metricID := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO neuronip.ingestion_metrics 
		(id, job_id, data_source_id, metric_name, metric_value, metric_unit, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`

	_, err := s.pool.Exec(ctx, query, metricID, jobID, dataSourceID, metricName, metricValue, unit, metadataJSON)
	return err
}

/* GetIngestionMetrics retrieves ingestion metrics */
func (s *EnterpriseIngestionService) GetIngestionMetrics(ctx context.Context, filters MetricFilters) ([]IngestionMetric, error) {
	query := `
		SELECT id, job_id, data_source_id, metric_name, metric_value, metric_unit, metadata, timestamp
		FROM neuronip.ingestion_metrics
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if filters.JobID != nil {
		query += fmt.Sprintf(" AND job_id = $%d", argIndex)
		args = append(args, *filters.JobID)
		argIndex++
	}

	if filters.DataSourceID != nil {
		query += fmt.Sprintf(" AND data_source_id = $%d", argIndex)
		args = append(args, *filters.DataSourceID)
		argIndex++
	}

	if filters.MetricName != nil {
		query += fmt.Sprintf(" AND metric_name = $%d", argIndex)
		args = append(args, *filters.MetricName)
		argIndex++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get ingestion metrics: %w", err)
	}
	defer rows.Close()

	var metrics []IngestionMetric
	for rows.Next() {
		var metric IngestionMetric
		var jobID *uuid.UUID
		var metadataJSON json.RawMessage
		var unit *string

		err := rows.Scan(&metric.ID, &jobID, &metric.DataSourceID, &metric.MetricName,
			&metric.MetricValue, &unit, &metadataJSON, &metric.Timestamp)
		if err != nil {
			continue
		}

		metric.JobID = jobID
		metric.Unit = unit
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &metric.Metadata)
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/* calculateNextRun calculates the next run time based on schedule */
func (s *EnterpriseIngestionService) calculateNextRun(schedule ScheduleConfig) time.Time {
	now := time.Now()

	switch schedule.Type {
	case "cron":
		// In production, use a cron parser library
		// For now, return now + 1 hour as default
		return now.Add(1 * time.Hour)

	case "interval":
		interval := schedule.Interval
		if interval == "" {
			interval = "1h"
		}
		duration, err := time.ParseDuration(interval)
		if err != nil {
			duration = 1 * time.Hour
		}
		return now.Add(duration)

	case "event_driven":
		// Event-driven schedules don't have a fixed next run
		return now

	default:
		return now.Add(1 * time.Hour)
	}
}

/* ScheduleConfig represents ingestion schedule configuration */
type ScheduleConfig struct {
	Type     string                 `json:"type"` // "cron", "interval", "event_driven"
	CronExpr string                 `json:"cron_expr,omitempty"`
	Interval string                 `json:"interval,omitempty"` // e.g., "1h", "30m"
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

/* ScheduledIngestion represents a scheduled ingestion */
type ScheduledIngestion struct {
	ID            uuid.UUID              `json:"id"`
	DataSourceID  uuid.UUID              `json:"data_source_id"`
	ScheduleType  string                 `json:"schedule_type"`
	ScheduleConfig map[string]interface{} `json:"schedule_config"`
	Enabled       bool                   `json:"enabled"`
	NextRunAt     *time.Time             `json:"next_run_at,omitempty"`
	LastRunAt     *time.Time             `json:"last_run_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* IngestionMetric represents an ingestion metric */
type IngestionMetric struct {
	ID           uuid.UUID              `json:"id"`
	JobID        *uuid.UUID             `json:"job_id,omitempty"`
	DataSourceID uuid.UUID              `json:"data_source_id"`
	MetricName   string                 `json:"metric_name"`
	MetricValue  float64                `json:"metric_value"`
	Unit         *string                `json:"unit,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

/* MetricFilters filters for metric queries */
type MetricFilters struct {
	JobID        *uuid.UUID
	DataSourceID *uuid.UUID
	MetricName   *string
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
}
