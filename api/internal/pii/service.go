package pii

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides PII detection functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new PII detection service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* PIIType represents a PII type */
type PIIType struct {
	Type         string                 `json:"type"`
	Category     string                 `json:"category"`
	Confidence   float64                `json:"confidence"`
	Pattern      *string                `json:"pattern,omitempty"`
	SampleValues []string               `json:"sample_values,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* PIIDetectionResult represents a PII detection result */
type PIIDetectionResult struct {
	ID               uuid.UUID              `json:"id"`
	ConnectorID      *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName       string                 `json:"schema_name"`
	TableName        string                 `json:"table_name"`
	ColumnName       string                 `json:"column_name"`
	PIITypes         []PIIType              `json:"pii_types"`
	Classification   string                 `json:"classification"` // "sensitive", "public", "internal"
	RiskLevel        string                 `json:"risk_level"`     // "low", "medium", "high", "critical"
	DetectedAt       time.Time              `json:"detected_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

/* PII patterns for common types */
var piiPatterns = map[string]struct {
	Pattern   *regexp.Regexp
	Category  string
	RiskLevel string
}{
	"email": {
		Pattern:   regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`),
		Category:  "contact",
		RiskLevel: "medium",
	},
	"ssn": {
		Pattern:   regexp.MustCompile(`\b\d{3}-?\d{2}-?\d{4}\b`),
		Category:  "identity",
		RiskLevel: "critical",
	},
	"credit_card": {
		Pattern:   regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
		Category:  "financial",
		RiskLevel: "critical",
	},
	"phone": {
		Pattern:   regexp.MustCompile(`\b(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}\b`),
		Category:  "contact",
		RiskLevel: "medium",
	},
	"ip_address": {
		Pattern:   regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
		Category:  "network",
		RiskLevel: "low",
	},
	"date_of_birth": {
		Pattern:   regexp.MustCompile(`\b(?:\d{1,2}[/-])?(?:\d{1,2}[/-])?(?:19|20)\d{2}\b`),
		Category:  "identity",
		RiskLevel: "high",
	},
}

/* Keyword-based detection */
var piiKeywords = map[string]struct {
	Category  string
	RiskLevel string
}{
	"email":         {"contact", "medium"},
	"phone":         {"contact", "medium"},
	"ssn":           {"identity", "critical"},
	"social_security": {"identity", "critical"},
	"credit_card":   {"financial", "critical"},
	"card_number":   {"financial", "critical"},
	"date_of_birth": {"identity", "high"},
	"dob":           {"identity", "high"},
	"birth_date":    {"identity", "high"},
	"passport":      {"identity", "critical"},
	"driver_license": {"identity", "high"},
	"address":       {"location", "medium"},
	"street":        {"location", "medium"},
	"city":          {"location", "low"},
	"zip":           {"location", "low"},
	"postal":        {"location", "low"},
	"first_name":    {"name", "medium"},
	"last_name":     {"name", "medium"},
	"full_name":     {"name", "medium"},
	"username":      {"identity", "medium"},
	"user_id":       {"identity", "low"},
}

/* DetectPII detects PII in a column */
func (s *Service) DetectPII(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName, columnName string, sampleSize int) (*PIIDetectionResult, error) {
	result := PIIDetectionResult{
		ID:         uuid.New(),
		ConnectorID: connectorID,
		SchemaName:  schemaName,
		TableName:   tableName,
		ColumnName:  columnName,
		DetectedAt:  time.Now(),
		PIITypes:    []PIIType{},
	}

	// Get sample values
	limit := sampleSize
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT %s::text as value
		FROM %s.%s
		WHERE %s IS NOT NULL
		LIMIT $1`,
		columnName, schemaName, tableName, columnName)

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sample values: %w", err)
	}
	defer rows.Close()

	var sampleValues []string
	var allValues []string

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			continue
		}
		sampleValues = append(sampleValues, value)
		allValues = append(allValues, strings.ToLower(value))
	}

	if len(sampleValues) == 0 {
		return nil, fmt.Errorf("no sample values found")
	}

	// Check column name for keywords
	columnLower := strings.ToLower(columnName)
	detectedTypes := make(map[string]PIIType)

	// Check keywords in column name
	for keyword, info := range piiKeywords {
		if strings.Contains(columnLower, keyword) {
			detectedTypes[keyword] = PIIType{
				Type:       keyword,
				Category:   info.Category,
				Confidence: 0.7,
				Metadata: map[string]interface{}{
					"detection_method": "keyword_match",
					"matched_keyword":  keyword,
				},
			}
		}
	}

	// Check patterns in values
	for piiType, patternInfo := range piiPatterns {
		if patternInfo.Pattern == nil {
			continue
		}

		matches := []string{}
		for _, value := range allValues {
			if patternInfo.Pattern.MatchString(value) {
				// Extract matches
				matches = append(matches, patternInfo.Pattern.FindString(value))
			}
		}

		if len(matches) > 0 {
			confidence := float64(len(matches)) / float64(len(allValues))
			
			// Update or add detection
			if existing, ok := detectedTypes[piiType]; ok {
				existing.Confidence = (existing.Confidence + confidence) / 2
				if len(matches) > len(existing.SampleValues) {
					existing.SampleValues = matches[:min(5, len(matches))]
				}
				detectedTypes[piiType] = existing
			} else {
				patternStr := patternInfo.Pattern.String()
				detectedTypes[piiType] = PIIType{
					Type:         piiType,
					Category:     patternInfo.Category,
					Confidence:   confidence,
					Pattern:      &patternStr,
					SampleValues: matches[:min(5, len(matches))],
					Metadata: map[string]interface{}{
						"detection_method": "pattern_match",
						"matches_found":    len(matches),
					},
				}
			}
		}
	}

	// Convert map to slice
	for _, piiType := range detectedTypes {
		result.PIITypes = append(result.PIITypes, piiType)
	}

	// Determine overall classification and risk level
	if len(result.PIITypes) == 0 {
		result.Classification = "public"
		result.RiskLevel = "low"
	} else {
		result.Classification = "sensitive"
		
		// Determine highest risk level
		maxRisk := "low"
		for _, piiType := range result.PIITypes {
			var risk string
			if info, ok := piiKeywords[piiType.Type]; ok {
				risk = info.RiskLevel
			} else if info, ok := piiPatterns[piiType.Type]; ok {
				risk = info.RiskLevel
			} else {
				risk = "medium"
			}
			
			if riskLevelOrder(risk) > riskLevelOrder(maxRisk) {
				maxRisk = risk
			}
		}
		result.RiskLevel = maxRisk
	}

	// Store detection result
	metadataJSON, _ := json.Marshal(result.Metadata)
	piiTypesJSON, _ := json.Marshal(result.PIITypes)

	insertQuery := `
		INSERT INTO neuronip.pii_detections
		(id, connector_id, schema_name, table_name, column_name, pii_types,
		 classification, risk_level, detected_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, detected_at`

	err = s.pool.QueryRow(ctx, insertQuery,
		result.ID, result.ConnectorID, result.SchemaName, result.TableName,
		result.ColumnName, piiTypesJSON, result.Classification,
		result.RiskLevel, result.DetectedAt, metadataJSON,
	).Scan(&result.ID, &result.DetectedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store detection: %w", err)
	}

	return &result, nil
}

/* GetDetection retrieves a PII detection result by ID */
func (s *Service) GetDetection(ctx context.Context, id uuid.UUID) (*PIIDetectionResult, error) {
	var result PIIDetectionResult
	var connectorID sql.NullString
	var piiTypesJSON []byte
	var metadataJSON []byte

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, pii_types,
		       classification, risk_level, detected_at, metadata
		FROM neuronip.pii_detections
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &connectorID, &result.SchemaName, &result.TableName,
		&result.ColumnName, &piiTypesJSON, &result.Classification,
		&result.RiskLevel, &result.DetectedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("detection not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get detection: %w", err)
	}

	if connectorID.Valid {
		id, _ := uuid.Parse(connectorID.String)
		result.ConnectorID = &id
	}
	if piiTypesJSON != nil {
		json.Unmarshal(piiTypesJSON, &result.PIITypes)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &result.Metadata)
	}

	return &result, nil
}

/* ListDetections lists PII detection results */
func (s *Service) ListDetections(ctx context.Context, connectorID *uuid.UUID, riskLevel *string, limit int) ([]PIIDetectionResult, error) {
	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, pii_types,
		       classification, risk_level, detected_at, metadata
		FROM neuronip.pii_detections
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if connectorID != nil {
		query += fmt.Sprintf(" AND connector_id = $%d", argIdx)
		args = append(args, *connectorID)
		argIdx++
	}
	if riskLevel != nil {
		query += fmt.Sprintf(" AND risk_level = $%d", argIdx)
		args = append(args, *riskLevel)
		argIdx++
	}

	query += " ORDER BY detected_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list detections: %w", err)
	}
	defer rows.Close()

	var results []PIIDetectionResult
	for rows.Next() {
		var result PIIDetectionResult
		var connectorID sql.NullString
		var piiTypesJSON []byte
		var metadataJSON []byte

		err := rows.Scan(
			&result.ID, &connectorID, &result.SchemaName, &result.TableName,
			&result.ColumnName, &piiTypesJSON, &result.Classification,
			&result.RiskLevel, &result.DetectedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			result.ConnectorID = &id
		}
		if piiTypesJSON != nil {
			json.Unmarshal(piiTypesJSON, &result.PIITypes)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &result.Metadata)
		}

		results = append(results, result)
	}

	return results, nil
}

/* Helper functions */
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func riskLevelOrder(level string) int {
	switch level {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
