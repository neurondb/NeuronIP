package reporting

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ReportingService provides advanced reporting functionality */
type ReportingService struct {
	pool *pgxpool.Pool
}

/* NewReportingService creates a new reporting service */
func NewReportingService(pool *pgxpool.Pool) *ReportingService {
	return &ReportingService{pool: pool}
}

/* Report represents a custom report */
type Report struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"` // "analytics", "compliance", "quality", "usage"
	Definition    ReportDefinition       `json:"definition"`
	Schedule      *ReportSchedule        `json:"schedule,omitempty"`
	Format        string                 `json:"format"` // "pdf", "csv", "json", "html"
	Enabled       bool                   `json:"enabled"`
	CreatedBy     uuid.UUID              `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* ReportDefinition represents the structure of a report */
type ReportDefinition struct {
	Sections    []ReportSection       `json:"sections"`
	Filters     []ReportFilter        `json:"filters,omitempty"`
	Aggregations []ReportAggregation  `json:"aggregations,omitempty"`
	Visualizations []ReportVisualization `json:"visualizations,omitempty"`
}

/* ReportSection represents a section in a report */
type ReportSection struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // "table", "chart", "summary", "custom"
	DataSource  string                 `json:"data_source"` // SQL query, API endpoint, etc.
	Columns     []string               `json:"columns,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

/* ReportFilter represents a filter for a report */
type ReportFilter struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // "equals", "contains", "greater_than", etc.
	Value     interface{} `json:"value"`
}

/* ReportAggregation represents an aggregation in a report */
type ReportAggregation struct {
	Field       string `json:"field"`
	Function    string `json:"function"` // "sum", "avg", "count", "min", "max"
	Alias       string `json:"alias,omitempty"`
}

/* ReportVisualization represents a visualization in a report */
type ReportVisualization struct {
	Type    string                 `json:"type"` // "bar", "line", "pie", "table"
	Title   string                 `json:"title"`
	Data    map[string]interface{} `json:"data"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

/* ReportSchedule represents a schedule for automatic report generation */
type ReportSchedule struct {
	Frequency  string    `json:"frequency"` // "daily", "weekly", "monthly", "custom"
	Cron       string    `json:"cron,omitempty"` // For custom schedules
	Time       string    `json:"time,omitempty"` // HH:MM format
	DayOfWeek  int       `json:"day_of_week,omitempty"` // 0-6 (Sunday-Saturday)
	DayOfMonth int       `json:"day_of_month,omitempty"` // 1-31
	NextRun    time.Time `json:"next_run"`
	Recipients []string  `json:"recipients,omitempty"`
}

/* CreateReport creates a new report */
func (s *ReportingService) CreateReport(ctx context.Context, report Report) (*Report, error) {
	report.ID = uuid.New()
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	defJSON, _ := json.Marshal(report.Definition)
	scheduleJSON, _ := json.Marshal(report.Schedule)
	metadataJSON, _ := json.Marshal(report.Metadata)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.reports
		(id, name, description, category, definition, schedule, format,
		 enabled, created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		report.ID, report.Name, report.Description, report.Category,
		defJSON, scheduleJSON, report.Format, report.Enabled,
		report.CreatedBy, report.CreatedAt, report.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return &report, nil
}

/* GenerateReport generates a report and returns the result */
func (s *ReportingService) GenerateReport(ctx context.Context, reportID uuid.UUID) (*ReportResult, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	result := &ReportResult{
		ID:         uuid.New(),
		ReportID:   reportID,
		GeneratedAt: time.Now(),
		Status:     "completed",
	}

	// Generate report sections
	sections := []ReportSectionResult{}
	for _, section := range report.Definition.Sections {
		sectionResult, err := s.generateSection(ctx, section, report.Definition.Filters)
		if err != nil {
			result.Status = "partial"
			continue
		}
		sections = append(sections, sectionResult)
	}

	result.Sections = sections

	// Save report result
	resultJSON, _ := json.Marshal(result)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.report_results
		(id, report_id, generated_at, status, result)
		VALUES ($1, $2, $3, $4, $5)`,
		result.ID, result.ReportID, result.GeneratedAt, result.Status, resultJSON,
	)

	return result, nil
}

/* generateSection generates a single report section */
func (s *ReportingService) generateSection(ctx context.Context,
	section ReportSection, filters []ReportFilter) (ReportSectionResult, error) {

	result := ReportSectionResult{
		SectionID: section.ID,
		Title:     section.Title,
		Type:      section.Type,
	}

	// Execute data source query
	if section.DataSource != "" {
		rows, err := s.pool.Query(ctx, section.DataSource)
		if err != nil {
			return result, fmt.Errorf("failed to execute query: %w", err)
		}
		defer rows.Close()

		// Get field descriptions from rows
		fieldDescriptions := rows.FieldDescriptions()
		columnCount := len(fieldDescriptions)

		// Convert rows to data
		data := []map[string]interface{}{}
		for rows.Next() {
			values := make([]interface{}, columnCount)
			valuePtrs := make([]interface{}, columnCount)
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			row := make(map[string]interface{})
			for i, field := range fieldDescriptions {
				row[field.Name] = values[i]
			}
			data = append(data, row)
		}

		result.Data = data
	}

	return result, nil
}

/* ReportResult represents the result of generating a report */
type ReportResult struct {
	ID          uuid.UUID              `json:"id"`
	ReportID    uuid.UUID              `json:"report_id"`
	GeneratedAt time.Time              `json:"generated_at"`
	Status      string                 `json:"status"` // "completed", "partial", "failed"
	Sections    []ReportSectionResult  `json:"sections"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* ReportSectionResult represents the result of a report section */
type ReportSectionResult struct {
	SectionID string                   `json:"section_id"`
	Title     string                   `json:"title"`
	Type      string                   `json:"type"`
	Data      []map[string]interface{} `json:"data,omitempty"`
}

/* GetReport retrieves a report */
func (s *ReportingService) GetReport(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	var report Report
	var defJSON, scheduleJSON, metadataJSON []byte
	var createdBy *uuid.UUID

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, description, category, definition, schedule, format,
		       enabled, created_by, created_at, updated_at, metadata
		FROM neuronip.reports
		WHERE id = $1`, reportID,
	).Scan(&report.ID, &report.Name, &report.Description, &report.Category,
		&defJSON, &scheduleJSON, &report.Format, &report.Enabled,
		&createdBy, &report.CreatedAt, &report.UpdatedAt, &metadataJSON)

	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	json.Unmarshal(defJSON, &report.Definition)
	json.Unmarshal(scheduleJSON, &report.Schedule)
	json.Unmarshal(metadataJSON, &report.Metadata)
	if createdBy != nil {
		report.CreatedBy = *createdBy
	}

	return &report, nil
}

/* ListReports lists all reports */
func (s *ReportingService) ListReports(ctx context.Context, category *string, limit int) ([]Report, error) {
	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT id, name, description, category, definition, schedule, format,
		       enabled, created_by, created_at, updated_at, metadata
		FROM neuronip.reports
		WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *category)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var report Report
		var defJSON, scheduleJSON, metadataJSON []byte
		var createdBy *uuid.UUID

		err := rows.Scan(&report.ID, &report.Name, &report.Description, &report.Category,
			&defJSON, &scheduleJSON, &report.Format, &report.Enabled,
			&createdBy, &report.CreatedAt, &report.UpdatedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(defJSON, &report.Definition)
		json.Unmarshal(scheduleJSON, &report.Schedule)
		json.Unmarshal(metadataJSON, &report.Metadata)
		if createdBy != nil {
			report.CreatedBy = *createdBy
		}

		reports = append(reports, report)
	}

	return reports, nil
}
