package profiling

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides data profiling functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new profiling service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* ProfileRequest represents a profiling request */
type ProfileRequest struct {
	ConnectorID uuid.UUID
	SchemaName  string
	TableName   string
	ColumnName  *string
	SampleSize  *int
}

/* ProfileResult represents profiling results */
type ProfileResult struct {
	ID            uuid.UUID              `json:"id"`
	ConnectorID   uuid.UUID              `json:"connector_id"`
	SchemaName    string                 `json:"schema_name"`
	TableName     string                 `json:"table_name"`
	ColumnName    *string                `json:"column_name,omitempty"`
	ProfileType   string                 `json:"profile_type"`
	Statistics    map[string]interface{} `json:"statistics"`
	DataType      *string                `json:"data_type,omitempty"`
	NullCount     int64                  `json:"null_count"`
	NonNullCount  int64                  `json:"non_null_count"`
	DistinctCount *int64                 `json:"distinct_count,omitempty"`
	MinValue      *string                `json:"min_value,omitempty"`
	MaxValue      *string                `json:"max_value,omitempty"`
	AvgValue      *float64               `json:"avg_value,omitempty"`
	MedianValue   *float64               `json:"median_value,omitempty"`
	Patterns      []DetectedPattern      `json:"patterns,omitempty"`
	SampleValues  []string               `json:"sample_values,omitempty"`
	ProfiledAt    time.Time              `json:"profiled_at"`
}

/* DetectedPattern represents a detected pattern */
type DetectedPattern struct {
	PatternType     string   `json:"pattern_type"`
	PatternRegex    *string  `json:"pattern_regex,omitempty"`
	MatchCount      int64    `json:"match_count"`
	MatchPercentage float64  `json:"match_percentage"`
	Confidence      float64  `json:"confidence"`
	Examples        []string `json:"examples,omitempty"`
}

/* ProfileTable profiles a table */
func (s *Service) ProfileTable(ctx context.Context, req ProfileRequest) (*ProfileResult, error) {
	// Get table statistics
	stats, err := s.getTableStatistics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get table statistics: %w", err)
	}

	// Profile each column
	columns, err := s.getTableColumns(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	columnProfiles := []map[string]interface{}{}
	for _, col := range columns {
		colReq := req
		colReq.ColumnName = &col
		colProfile, err := s.ProfileColumn(ctx, colReq)
		if err == nil {
			columnProfiles = append(columnProfiles, map[string]interface{}{
				"column_name": col,
				"profile":     colProfile,
			})
		}
	}

	result := &ProfileResult{
		ID:          uuid.New(),
		ConnectorID: req.ConnectorID,
		SchemaName:  req.SchemaName,
		TableName:   req.TableName,
		ProfileType: "table",
		Statistics: map[string]interface{}{
			"row_count":      stats.RowCount,
			"column_count":   len(columns),
			"size_bytes":     stats.SizeBytes,
			"column_profiles": columnProfiles,
		},
		ProfiledAt: time.Now(),
	}

	// Store profile
	if err := s.storeProfile(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to store profile: %w", err)
	}

	return result, nil
}

/* ProfileColumn profiles a column */
func (s *Service) ProfileColumn(ctx context.Context, req ProfileRequest) (*ProfileResult, error) {
	if req.ColumnName == nil {
		return nil, fmt.Errorf("column_name required for column profiling")
	}

	// Get column statistics
	stats, err := s.getColumnStatistics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get column statistics: %w", err)
	}

	// Detect patterns first (needed for data type detection)
	patterns := s.detectPatterns(ctx, req, stats)

	// Detect data type
	dataType := s.detectDataType(ctx, req, stats, patterns)

	// Get sample values
	sampleValues := s.getSampleValues(ctx, req)

	result := &ProfileResult{
		ID:            uuid.New(),
		ConnectorID:   req.ConnectorID,
		SchemaName:    req.SchemaName,
		TableName:     req.TableName,
		ColumnName:    req.ColumnName,
		ProfileType:   "column",
		Statistics:    stats.Statistics,
		DataType:      &dataType,
		NullCount:     stats.NullCount,
		NonNullCount:  stats.NonNullCount,
		DistinctCount: stats.DistinctCount,
		MinValue:      stats.MinValue,
		MaxValue:      stats.MaxValue,
		AvgValue:      stats.AvgValue,
		MedianValue:   stats.MedianValue,
		Patterns:      patterns,
		SampleValues:  sampleValues,
		ProfiledAt:    time.Now(),
	}

	// Store profile
	if err := s.storeProfile(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to store profile: %w", err)
	}

	// Store patterns
	if err := s.storePatterns(ctx, result.ID, patterns); err != nil {
		return nil, fmt.Errorf("failed to store patterns: %w", err)
	}

	return result, nil
}

/* ColumnStatistics represents column statistics */
type ColumnStatistics struct {
	Statistics    map[string]interface{}
	NullCount     int64
	NonNullCount  int64
	DistinctCount *int64
	MinValue      *string
	MaxValue      *string
	AvgValue      *float64
	MedianValue   *float64
}

/* getColumnStatistics gets column statistics */
func (s *Service) getColumnStatistics(ctx context.Context, req ProfileRequest) (*ColumnStatistics, error) {
	stats := &ColumnStatistics{
		Statistics: make(map[string]interface{}),
	}

	// Count nulls and non-nulls
	nullQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) FILTER (WHERE %s IS NULL) as null_count,
			COUNT(*) FILTER (WHERE %s IS NOT NULL) as non_null_count,
			COUNT(DISTINCT %s) as distinct_count
		FROM %s.%s`,
		*req.ColumnName, *req.ColumnName, *req.ColumnName, req.SchemaName, req.TableName)

	err := s.pool.QueryRow(ctx, nullQuery).Scan(
		&stats.NullCount, &stats.NonNullCount, &stats.DistinctCount)
	if err != nil {
		return nil, err
	}

	// Get min/max for numeric/text columns
	minMaxQuery := fmt.Sprintf(`
		SELECT 
			MIN(%s::text) as min_value,
			MAX(%s::text) as max_value
		FROM %s.%s
		WHERE %s IS NOT NULL`,
		*req.ColumnName, *req.ColumnName, req.SchemaName, req.TableName, *req.ColumnName)

	var minVal, maxVal sql.NullString
	err = s.pool.QueryRow(ctx, minMaxQuery).Scan(&minVal, &maxVal)
	if err == nil {
		if minVal.Valid {
			stats.MinValue = &minVal.String
		}
		if maxVal.Valid {
			stats.MaxValue = &maxVal.String
		}
	}

	// Try to get avg for numeric columns
	avgQuery := fmt.Sprintf(`
		SELECT AVG(%s::numeric) as avg_value
		FROM %s.%s
		WHERE %s IS NOT NULL`,
		*req.ColumnName, req.SchemaName, req.TableName, *req.ColumnName)

	var avgVal sql.NullFloat64
	err = s.pool.QueryRow(ctx, avgQuery).Scan(&avgVal)
	if err == nil && avgVal.Valid {
		stats.AvgValue = &avgVal.Float64
	}

	// Calculate statistics
	stats.Statistics["null_percentage"] = float64(stats.NullCount) / float64(stats.NullCount+stats.NonNullCount) * 100
	if stats.DistinctCount != nil {
		stats.Statistics["distinct_percentage"] = float64(*stats.DistinctCount) / float64(stats.NonNullCount) * 100
	}

	return stats, nil
}

/* TableStatistics represents table statistics */
type TableStatistics struct {
	RowCount  int64
	SizeBytes *int64
}

/* getTableStatistics gets table statistics */
func (s *Service) getTableStatistics(ctx context.Context, req ProfileRequest) (*TableStatistics, error) {
	stats := &TableStatistics{}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as row_count,
			pg_total_relation_size('%s.%s') as size_bytes
		FROM %s.%s`,
		req.SchemaName, req.TableName, req.SchemaName, req.TableName)

	var sizeBytes sql.NullInt64
	err := s.pool.QueryRow(ctx, query).Scan(&stats.RowCount, &sizeBytes)
	if err != nil {
		return nil, err
	}

	if sizeBytes.Valid {
		stats.SizeBytes = &sizeBytes.Int64
	}

	return stats, nil
}

/* getTableColumns gets column names for a table */
func (s *Service) getTableColumns(ctx context.Context, req ProfileRequest) ([]string, error) {
	query := `
		SELECT column_name
		FROM neuronip.catalog_columns
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3
		ORDER BY ordinal_position`

	rows, err := s.pool.Query(ctx, query, req.ConnectorID, req.SchemaName, req.TableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err == nil {
			columns = append(columns, col)
		}
	}

	return columns, nil
}

/* detectDataType detects data type from values */
func (s *Service) detectDataType(ctx context.Context, req ProfileRequest, stats *ColumnStatistics, patterns []DetectedPattern) string {
	// Check if numeric
	if stats.AvgValue != nil {
		return "numeric"
	}

	// Check patterns to determine type
	if len(patterns) > 0 {
		for _, pattern := range patterns {
			switch pattern.PatternType {
			case "email":
				return "email"
			case "phone":
				return "phone"
			case "date":
				return "date"
			case "url":
				return "url"
			case "ip_address":
				return "ip_address"
			}
		}
	}

	// Default to text
	return "text"
}

/* detectPatterns detects patterns in column data */
func (s *Service) detectPatterns(ctx context.Context, req ProfileRequest, stats *ColumnStatistics) []DetectedPattern {
	patterns := []DetectedPattern{}

	// Get sample values for pattern detection
	sampleQuery := fmt.Sprintf(`
		SELECT %s::text
		FROM %s.%s
		WHERE %s IS NOT NULL
		LIMIT 100`,
		*req.ColumnName, req.SchemaName, req.TableName, *req.ColumnName)

	rows, err := s.pool.Query(ctx, sampleQuery)
	if err != nil {
		return patterns
	}
	defer rows.Close()

	sampleValues := []string{}
	for rows.Next() {
		var val sql.NullString
		if err := rows.Scan(&val); err == nil && val.Valid {
			sampleValues = append(sampleValues, val.String)
		}
	}

	if len(sampleValues) == 0 {
		return patterns
	}

	// Pattern definitions
	patternDefs := []struct {
		Type  string
		Regex *regexp.Regexp
	}{
		{"email", regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)},
		{"phone", regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,}$`)},
		{"ssn", regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)},
		{"credit_card", regexp.MustCompile(`^\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}$`)},
		{"url", regexp.MustCompile(`^https?://[^\s]+$`)},
		{"ip_address", regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)},
		{"date", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)},
	}

	for _, def := range patternDefs {
		matches := 0
		examples := []string{}

		for _, val := range sampleValues {
			if def.Regex.MatchString(val) {
				matches++
				if len(examples) < 3 {
					examples = append(examples, val)
				}
			}
		}

		if matches > 0 {
			matchPct := float64(matches) / float64(len(sampleValues)) * 100
			confidence := math.Min(matchPct/100, 1.0)

			patterns = append(patterns, DetectedPattern{
				PatternType:     def.Type,
				MatchCount:      int64(matches),
				MatchPercentage: matchPct,
				Confidence:      confidence,
				Examples:        examples,
			})
		}
	}

	return patterns
}

/* getSampleValues gets sample values from column */
func (s *Service) getSampleValues(ctx context.Context, req ProfileRequest) []string {
	query := fmt.Sprintf(`
		SELECT DISTINCT %s::text
		FROM %s.%s
		WHERE %s IS NOT NULL
		LIMIT 10`,
		*req.ColumnName, req.SchemaName, req.TableName, *req.ColumnName)

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var samples []string
	for rows.Next() {
		var val sql.NullString
		if err := rows.Scan(&val); err == nil && val.Valid {
			samples = append(samples, val.String)
		}
	}

	return samples
}

/* storeProfile stores profile in database */
func (s *Service) storeProfile(ctx context.Context, profile *ProfileResult) error {
	statisticsJSON, _ := json.Marshal(profile.Statistics)
	patternsJSON, _ := json.Marshal(profile.Patterns)

	var columnName sql.NullString
	if profile.ColumnName != nil {
		columnName = sql.NullString{String: *profile.ColumnName, Valid: true}
	}
	var dataType sql.NullString
	if profile.DataType != nil {
		dataType = sql.NullString{String: *profile.DataType, Valid: true}
	}
	var distinctCount sql.NullInt64
	if profile.DistinctCount != nil {
		distinctCount = sql.NullInt64{Int64: *profile.DistinctCount, Valid: true}
	}
	var minValue, maxValue sql.NullString
	if profile.MinValue != nil {
		minValue = sql.NullString{String: *profile.MinValue, Valid: true}
	}
	if profile.MaxValue != nil {
		maxValue = sql.NullString{String: *profile.MaxValue, Valid: true}
	}
	var avgValue, medianValue sql.NullFloat64
	if profile.AvgValue != nil {
		avgValue = sql.NullFloat64{Float64: *profile.AvgValue, Valid: true}
	}
	if profile.MedianValue != nil {
		medianValue = sql.NullFloat64{Float64: *profile.MedianValue, Valid: true}
	}

	query := `
		INSERT INTO neuronip.data_profiles
		(id, connector_id, schema_name, table_name, column_name, profile_type,
		 statistics, data_type, null_count, non_null_count, distinct_count,
		 min_value, max_value, avg_value, median_value, patterns, sample_values,
		 profiled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		ON CONFLICT (connector_id, schema_name, table_name, column_name)
		DO UPDATE SET
			statistics = EXCLUDED.statistics,
			data_type = EXCLUDED.data_type,
			null_count = EXCLUDED.null_count,
			non_null_count = EXCLUDED.non_null_count,
			distinct_count = EXCLUDED.distinct_count,
			min_value = EXCLUDED.min_value,
			max_value = EXCLUDED.max_value,
			avg_value = EXCLUDED.avg_value,
			median_value = EXCLUDED.median_value,
			patterns = EXCLUDED.patterns,
			sample_values = EXCLUDED.sample_values,
			profiled_at = EXCLUDED.profiled_at,
			updated_at = EXCLUDED.updated_at`

	_, err := s.pool.Exec(ctx, query,
		profile.ID, profile.ConnectorID, profile.SchemaName, profile.TableName,
		columnName, profile.ProfileType, statisticsJSON, dataType,
		profile.NullCount, profile.NonNullCount, distinctCount,
		minValue, maxValue, avgValue, medianValue,
		patternsJSON, profile.SampleValues,
		profile.ProfiledAt, time.Now(), time.Now(),
	)

	return err
}

/* storePatterns stores detected patterns */
func (s *Service) storePatterns(ctx context.Context, profileID uuid.UUID, patterns []DetectedPattern) error {
	for _, pattern := range patterns {
		var patternRegex sql.NullString
		if pattern.PatternRegex != nil {
			patternRegex = sql.NullString{String: *pattern.PatternRegex, Valid: true}
		}

		query := `
			INSERT INTO neuronip.data_patterns
			(id, profile_id, pattern_type, pattern_regex, match_count,
			 match_percentage, confidence, examples, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW())`

		_, err := s.pool.Exec(ctx, query,
			profileID, pattern.PatternType, patternRegex,
			pattern.MatchCount, pattern.MatchPercentage, pattern.Confidence,
			pattern.Examples,
		)
		if err != nil {
			continue
		}
	}

	return nil
}
