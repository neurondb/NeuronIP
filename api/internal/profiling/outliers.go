package profiling

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* OutlierService provides outlier detection functionality */
type OutlierService struct {
	pool *pgxpool.Pool
}

/* NewOutlierService creates a new outlier service */
func NewOutlierService(pool *pgxpool.Pool) *OutlierService {
	return &OutlierService{pool: pool}
}

/* OutlierDetection represents an outlier detection result */
type OutlierDetection struct {
	ID            uuid.UUID              `json:"id"`
	ConnectorID   uuid.UUID              `json:"connector_id"`
	SchemaName    string                 `json:"schema_name"`
	TableName     string                 `json:"table_name"`
	ColumnName    string                 `json:"column_name"`
	OutlierType   string                 `json:"outlier_type"` // "statistical", "temporal", "pattern"
	OutlierValue  interface{}            `json:"outlier_value"`
	ExpectedValue interface{}            `json:"expected_value,omitempty"`
	Deviation     float64                `json:"deviation"` // Number of standard deviations
	Severity      string                 `json:"severity"` // "low", "medium", "high", "critical"
	DetectedAt    time.Time              `json:"detected_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* DetectOutliers detects outliers in column data */
func (s *OutlierService) DetectOutliers(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName, columnName string,
	detectionType string) ([]OutlierDetection, error) {

	var outliers []OutlierDetection

	switch detectionType {
	case "statistical":
		statOutliers, err := s.detectStatisticalOutliers(ctx, connectorID, schemaName, tableName, columnName)
		if err == nil {
			outliers = append(outliers, statOutliers...)
		}

	case "temporal":
		tempOutliers, err := s.detectTemporalOutliers(ctx, connectorID, schemaName, tableName, columnName)
		if err == nil {
			outliers = append(outliers, tempOutliers...)
		}

	case "pattern":
		patternOutliers, err := s.detectPatternOutliers(ctx, connectorID, schemaName, tableName, columnName)
		if err == nil {
			outliers = append(outliers, patternOutliers...)
		}

	case "all":
		statOutliers, _ := s.detectStatisticalOutliers(ctx, connectorID, schemaName, tableName, columnName)
		tempOutliers, _ := s.detectTemporalOutliers(ctx, connectorID, schemaName, tableName, columnName)
		patternOutliers, _ := s.detectPatternOutliers(ctx, connectorID, schemaName, tableName, columnName)
		outliers = append(outliers, statOutliers...)
		outliers = append(outliers, tempOutliers...)
		outliers = append(outliers, patternOutliers...)
	}

	// Save detected outliers
	for i := range outliers {
		outliers[i].ID = uuid.New()
		outliers[i].DetectedAt = time.Now()

		valueJSON, _ := json.Marshal(outliers[i].OutlierValue)
		expectedJSON, _ := json.Marshal(outliers[i].ExpectedValue)
		metadataJSON, _ := json.Marshal(outliers[i].Metadata)

		s.pool.Exec(ctx, `
			INSERT INTO neuronip.outlier_detections
			(id, connector_id, schema_name, table_name, column_name, outlier_type,
			 outlier_value, expected_value, deviation, severity, detected_at, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			outliers[i].ID, outliers[i].ConnectorID, outliers[i].SchemaName,
			outliers[i].TableName, outliers[i].ColumnName, outliers[i].OutlierType,
			valueJSON, expectedJSON, outliers[i].Deviation, outliers[i].Severity,
			outliers[i].DetectedAt, metadataJSON,
		)
	}

	return outliers, nil
}

/* detectStatisticalOutliers detects outliers using statistical methods (Z-score, IQR) */
func (s *OutlierService) detectStatisticalOutliers(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName, columnName string) ([]OutlierDetection, error) {

	// Calculate statistics (mean, stddev, min, max)
	query := fmt.Sprintf(`
		SELECT 
			AVG(%s::numeric) as mean,
			STDDEV(%s::numeric) as stddev,
			MIN(%s::numeric) as min_val,
			MAX(%s::numeric) as max_val,
			PERCENTILE_CONT(0.25) WITHIN GROUP (ORDER BY %s::numeric) as q1,
			PERCENTILE_CONT(0.75) WITHIN GROUP (ORDER BY %s::numeric) as q3
		FROM %s.%s`, columnName, columnName, columnName, columnName, columnName, columnName, schemaName, tableName)

	var mean, stddev, minVal, maxVal, q1, q3 *float64
	err := s.pool.QueryRow(ctx, query).Scan(&mean, &stddev, &minVal, &maxVal, &q1, &q3)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate statistics: %w", err)
	}

	if mean == nil || stddev == nil {
		return []OutlierDetection{}, nil
	}

	var outliers []OutlierDetection

	// Z-score method (values > 3 standard deviations)
	threshold := 3.0
	zScoreQuery := fmt.Sprintf(`
		SELECT %s, ABS((%s::numeric - $1) / NULLIF($2, 0)) as z_score
		FROM %s.%s
		WHERE ABS((%s::numeric - $1) / NULLIF($2, 0)) > $3
		LIMIT 100`, columnName, columnName, schemaName, tableName, columnName)

	rows, _ := s.pool.Query(ctx, zScoreQuery, *mean, *stddev, threshold)
	defer rows.Close()

	for rows.Next() {
		var value float64
		var zScore float64

		err := rows.Scan(&value, &zScore)
		if err != nil {
			continue
		}

		outlier := OutlierDetection{
			ConnectorID: connectorID,
			SchemaName:  schemaName,
			TableName:   tableName,
			ColumnName:  columnName,
			OutlierType: "statistical",
			OutlierValue: value,
			ExpectedValue: *mean,
			Deviation:   zScore,
			Severity:    s.determineSeverity(zScore),
		}

		outliers = append(outliers, outlier)
	}

	// IQR method
	if q1 != nil && q3 != nil {
		iqr := *q3 - *q1
		lowerBound := *q1 - 1.5*iqr
		upperBound := *q3 + 1.5*iqr

		iqrQuery := fmt.Sprintf(`
			SELECT %s
			FROM %s.%s
			WHERE %s::numeric < $1 OR %s::numeric > $2
			LIMIT 100`, columnName, schemaName, tableName, columnName, columnName)

		iqrRows, _ := s.pool.Query(ctx, iqrQuery, lowerBound, upperBound)
		defer iqrRows.Close()

		for iqrRows.Next() {
			var value float64

			err := iqrRows.Scan(&value)
			if err != nil {
				continue
			}

			// Check if not already detected by Z-score
			found := false
			for _, existing := range outliers {
				if existing.OutlierValue == value {
					found = true
					break
				}
			}

			if !found {
				deviation := math.Abs(value - *mean)
				if *stddev > 0 {
					deviation = deviation / *stddev
				}

				outlier := OutlierDetection{
					ConnectorID:  connectorID,
					SchemaName:   schemaName,
					TableName:    tableName,
					ColumnName:   columnName,
					OutlierType:  "statistical",
					OutlierValue: value,
					ExpectedValue: *mean,
					Deviation:    deviation,
					Severity:     s.determineSeverity(deviation),
				}

				outliers = append(outliers, outlier)
			}
		}
	}

	return outliers, nil
}

/* detectTemporalOutliers detects temporal anomalies */
func (s *OutlierService) detectTemporalOutliers(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName, columnName string) ([]OutlierDetection, error) {

	// Detect sudden spikes or drops in time series data
	// This would require a timestamp column, which we'll assume exists
	var outliers []OutlierDetection

	// Implementation would analyze temporal patterns
	// For now, return empty

	return outliers, nil
}

/* detectPatternOutliers detects pattern-based outliers */
func (s *OutlierService) detectPatternOutliers(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName, columnName string) ([]OutlierDetection, error) {

	// Detect unusual patterns (e.g., unexpected null values, format violations)
	var outliers []OutlierDetection

	// Implementation would analyze patterns
	// For now, return empty

	return outliers, nil
}

/* determineSeverity determines outlier severity based on deviation */
func (s *OutlierService) determineSeverity(deviation float64) string {
	if deviation >= 5.0 {
		return "critical"
	} else if deviation >= 3.0 {
		return "high"
	} else if deviation >= 2.0 {
		return "medium"
	}
	return "low"
}

/* GetOutlierHistory retrieves outlier detection history */
func (s *OutlierService) GetOutlierHistory(ctx context.Context,
	connectorID uuid.UUID, limit int) ([]OutlierDetection, error) {

	if limit == 0 {
		limit = 100
	}

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name,
		       outlier_type, outlier_value, expected_value, deviation,
		       severity, detected_at, metadata
		FROM neuronip.outlier_detections
		WHERE connector_id = $1
		ORDER BY detected_at DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, connectorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get outlier history: %w", err)
	}
	defer rows.Close()

	var outliers []OutlierDetection
	for rows.Next() {
		var outlier OutlierDetection
		var valueJSON, expectedJSON, metadataJSON []byte

		err := rows.Scan(&outlier.ID, &outlier.ConnectorID, &outlier.SchemaName,
			&outlier.TableName, &outlier.ColumnName, &outlier.OutlierType,
			&valueJSON, &expectedJSON, &outlier.Deviation, &outlier.Severity,
			&outlier.DetectedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(valueJSON, &outlier.OutlierValue)
		json.Unmarshal(expectedJSON, &outlier.ExpectedValue)
		json.Unmarshal(metadataJSON, &outlier.Metadata)

		outliers = append(outliers, outlier)
	}

	return outliers, nil
}
