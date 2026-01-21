package classification

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

/* Service provides automated data classification functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new classification service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* ClassificationRule represents a classification rule */
type ClassificationRule struct {
	ID                  uuid.UUID              `json:"id"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description,omitempty"`
	ClassificationType  string                 `json:"classification_type"`
	RuleType            string                 `json:"rule_type"`
	RuleExpression      string                 `json:"rule_expression"`
	ConfidenceThreshold float64                `json:"confidence_threshold"`
	Enabled             bool                   `json:"enabled"`
	Priority            int                    `json:"priority"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy           *string                `json:"created_by,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

/* DataClassification represents a classification result */
type DataClassification struct {
	ID                uuid.UUID              `json:"id"`
	ConnectorID       uuid.UUID              `json:"connector_id"`
	SchemaName        string                 `json:"schema_name"`
	TableName         string                 `json:"table_name"`
	ColumnName        string                 `json:"column_name"`
	ClassificationType string                `json:"classification_type"`
	Confidence        float64                `json:"confidence"`
	DetectionMethod   string                 `json:"detection_method"`
	DetectedPatterns  []string               `json:"detected_patterns,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ClassifiedBy      *string                `json:"classified_by,omitempty"`
	ClassifiedAt      time.Time              `json:"classified_at"`
}

/* ClassifyColumn classifies a column */
func (s *Service) ClassifyColumn(ctx context.Context, connectorID uuid.UUID, schemaName, tableName, columnName string) (*DataClassification, error) {
	// Get classification rules
	rules, err := s.getEnabledRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	// Get column metadata
	columnInfo, err := s.getColumnInfo(ctx, connectorID, schemaName, tableName, columnName)
	if err != nil {
		return nil, fmt.Errorf("failed to get column info: %w", err)
	}

	// Get sample values
	sampleValues, err := s.getSampleValues(ctx, connectorID, schemaName, tableName, columnName)
	if err != nil {
		return nil, fmt.Errorf("failed to get sample values: %w", err)
	}

	// Apply rules in priority order
	bestMatch := &ClassificationMatch{
		ClassificationType: "public",
		Confidence:         0.0,
		DetectionMethod:    "rule",
		DetectedPatterns:  []string{},
	}

	for _, rule := range rules {
		match, err := s.applyRule(ctx, rule, columnInfo, sampleValues)
		if err != nil {
			continue
		}

		if match.Confidence > bestMatch.Confidence {
			bestMatch = match
		}
	}

	// Pattern-based detection
	patternMatch := s.detectPatterns(columnName, sampleValues)
	if patternMatch.Confidence > bestMatch.Confidence {
		bestMatch = patternMatch
	}

	// Create classification
	classification := &DataClassification{
		ID:                uuid.New(),
		ConnectorID:       connectorID,
		SchemaName:        schemaName,
		TableName:         tableName,
		ColumnName:        columnName,
		ClassificationType: bestMatch.ClassificationType,
		Confidence:        bestMatch.Confidence,
		DetectionMethod:   bestMatch.DetectionMethod,
		DetectedPatterns:  bestMatch.DetectedPatterns,
		ClassifiedAt:      time.Now(),
	}

	// Store classification
	if err := s.storeClassification(ctx, classification); err != nil {
		return nil, fmt.Errorf("failed to store classification: %w", err)
	}

	return classification, nil
}

/* ClassificationMatch represents a classification match */
type ClassificationMatch struct {
	ClassificationType string
	Confidence         float64
	DetectionMethod    string
	DetectedPatterns   []string
}

/* applyRule applies a classification rule */
func (s *Service) applyRule(ctx context.Context, rule ClassificationRule, columnInfo map[string]interface{}, sampleValues []string) (*ClassificationMatch, error) {
	match := &ClassificationMatch{
		ClassificationType: rule.ClassificationType,
		Confidence:         0.0,
		DetectionMethod:    "rule",
		DetectedPatterns:   []string{},
	}

	switch rule.RuleType {
	case "pattern":
		matched := s.matchPattern(rule.RuleExpression, sampleValues)
		if matched {
			match.Confidence = 0.9
			match.DetectedPatterns = []string{rule.RuleExpression}
		}
	case "keyword":
		keywords := strings.Split(rule.RuleExpression, ",")
		matched := s.matchKeywords(keywords, columnInfo, sampleValues)
		if matched {
			match.Confidence = 0.8
			match.DetectedPatterns = keywords
		}
	case "column_name":
		matched := s.matchColumnName(rule.RuleExpression, columnInfo["column_name"].(string))
		if matched {
			match.Confidence = 0.7
		}
	case "ml_model":
		// ML-based classification would go here
		match.Confidence = 0.6
	}

	if match.Confidence < rule.ConfidenceThreshold {
		return nil, fmt.Errorf("confidence below threshold")
	}

	return match, nil
}

/* detectPatterns detects patterns in column name and values */
func (s *Service) detectPatterns(columnName string, sampleValues []string) *ClassificationMatch {
	columnLower := strings.ToLower(columnName)

	// PII patterns
	piiPatterns := map[string]string{
		"email":        `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
		"ssn":          `^\d{3}-\d{2}-\d{4}$`,
		"credit_card": `^\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}$`,
		"phone":        `^\+?[\d\s\-\(\)]{10,}$`,
	}

	// Check column name
	for patternType := range piiPatterns {
		if strings.Contains(columnLower, patternType) {
			return &ClassificationMatch{
				ClassificationType: "pii",
				Confidence:         0.8,
				DetectionMethod:    "pattern",
				DetectedPatterns:  []string{patternType},
			}
		}
	}

	// Check sample values
	for _, value := range sampleValues {
		for patternType, patternRegex := range piiPatterns {
			matched, _ := regexp.MatchString(patternRegex, value)
			if matched {
				return &ClassificationMatch{
					ClassificationType: "pii",
					Confidence:         0.9,
					DetectionMethod:    "pattern",
					DetectedPatterns:  []string{patternType},
				}
			}
		}
	}

	// PHI patterns
	phiKeywords := []string{"patient", "medical", "diagnosis", "treatment", "health"}
	for _, keyword := range phiKeywords {
		if strings.Contains(columnLower, keyword) {
			return &ClassificationMatch{
				ClassificationType: "phi",
				Confidence:         0.7,
				DetectionMethod:    "pattern",
				DetectedPatterns:  []string{keyword},
			}
		}
	}

	// PCI patterns
	pciKeywords := []string{"card", "payment", "billing", "transaction"}
	for _, keyword := range pciKeywords {
		if strings.Contains(columnLower, keyword) {
			return &ClassificationMatch{
				ClassificationType: "pci",
				Confidence:         0.7,
				DetectionMethod:    "pattern",
				DetectedPatterns:  []string{keyword},
			}
		}
	}

	return &ClassificationMatch{
		ClassificationType: "public",
		Confidence:         0.5,
		DetectionMethod:    "pattern",
		DetectedPatterns:  []string{},
	}
}

/* matchPattern matches pattern against values */
func (s *Service) matchPattern(pattern string, values []string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	for _, value := range values {
		if re.MatchString(value) {
			return true
		}
	}

	return false
}

/* matchKeywords matches keywords */
func (s *Service) matchKeywords(keywords []string, columnInfo map[string]interface{}, sampleValues []string) bool {
	columnName, _ := columnInfo["column_name"].(string)
	columnLower := strings.ToLower(columnName)

	for _, keyword := range keywords {
		keywordLower := strings.ToLower(strings.TrimSpace(keyword))
		if strings.Contains(columnLower, keywordLower) {
			return true
		}

		for _, value := range sampleValues {
			if strings.Contains(strings.ToLower(value), keywordLower) {
				return true
			}
		}
	}

	return false
}

/* matchColumnName matches column name pattern */
func (s *Service) matchColumnName(pattern string, columnName string) bool {
	matched, _ := regexp.MatchString("(?i)"+pattern, columnName)
	return matched
}

/* getEnabledRules gets enabled classification rules */
func (s *Service) getEnabledRules(ctx context.Context) ([]ClassificationRule, error) {
	query := `
		SELECT id, name, description, classification_type, rule_type, rule_expression,
		       confidence_threshold, enabled, priority, metadata, created_by, created_at, updated_at
		FROM neuronip.classification_rules
		WHERE enabled = true
		ORDER BY priority DESC, created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []ClassificationRule
	for rows.Next() {
		var rule ClassificationRule
		var description, createdBy sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &description, &rule.ClassificationType,
			&rule.RuleType, &rule.RuleExpression, &rule.ConfidenceThreshold,
			&rule.Enabled, &rule.Priority, &metadataJSON, &createdBy,
			&rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if description.Valid {
			rule.Description = &description.String
		}
		if createdBy.Valid {
			rule.CreatedBy = &createdBy.String
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &rule.Metadata)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

/* getColumnInfo gets column information */
func (s *Service) getColumnInfo(ctx context.Context, connectorID uuid.UUID, schemaName, tableName, columnName string) (map[string]interface{}, error) {
	query := `
		SELECT column_name, column_type, is_nullable, is_primary_key, description
		FROM neuronip.catalog_columns
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3 AND column_name = $4`

	var colName, colType string
	var isNullable, isPrimaryKey bool
	var description sql.NullString

	err := s.pool.QueryRow(ctx, query, connectorID, schemaName, tableName, columnName).Scan(
		&colName, &colType, &isNullable, &isPrimaryKey, &description,
	)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"column_name":  colName,
		"column_type":  colType,
		"is_nullable":  isNullable,
		"is_primary_key": isPrimaryKey,
	}
	if description.Valid {
		info["description"] = description.String
	}

	return info, nil
}

/* getSampleValues gets sample values from column */
func (s *Service) getSampleValues(ctx context.Context, connectorID uuid.UUID, schemaName, tableName, columnName string) ([]string, error) {
	// Get connector to build query
	// For now, assume PostgreSQL
	query := fmt.Sprintf(`
		SELECT DISTINCT %s::text
		FROM %s.%s
		WHERE %s IS NOT NULL
		LIMIT 100`,
		columnName, schemaName, tableName, columnName)

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return []string{}, nil
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var val sql.NullString
		if err := rows.Scan(&val); err == nil && val.Valid {
			values = append(values, val.String)
		}
	}

	return values, nil
}

/* storeClassification stores classification result */
func (s *Service) storeClassification(ctx context.Context, classification *DataClassification) error {
	metadataJSON, _ := json.Marshal(classification.Metadata)
	var classifiedBy sql.NullString
	if classification.ClassifiedBy != nil {
		classifiedBy = sql.NullString{String: *classification.ClassifiedBy, Valid: true}
	}

	query := `
		INSERT INTO neuronip.data_classifications
		(id, connector_id, schema_name, table_name, column_name, classification_type,
		 confidence, detection_method, detected_patterns, metadata, classified_by, classified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (connector_id, schema_name, table_name, column_name)
		DO UPDATE SET
			classification_type = EXCLUDED.classification_type,
			confidence = EXCLUDED.confidence,
			detection_method = EXCLUDED.detection_method,
			detected_patterns = EXCLUDED.detected_patterns,
			metadata = EXCLUDED.metadata,
			classified_by = EXCLUDED.classified_by,
			classified_at = EXCLUDED.classified_at,
			updated_at = EXCLUDED.updated_at`

	_, err := s.pool.Exec(ctx, query,
		classification.ID, classification.ConnectorID, classification.SchemaName,
		classification.TableName, classification.ColumnName, classification.ClassificationType,
		classification.Confidence, classification.DetectionMethod, classification.DetectedPatterns,
		metadataJSON, classifiedBy, classification.ClassifiedAt, time.Now(), time.Now(),
	)

	return err
}

/* CreateRule creates a classification rule */
func (s *Service) CreateRule(ctx context.Context, rule ClassificationRule) (*ClassificationRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(rule.Metadata)
	var description, createdBy sql.NullString
	if rule.Description != nil {
		description = sql.NullString{String: *rule.Description, Valid: true}
	}
	if rule.CreatedBy != nil {
		createdBy = sql.NullString{String: *rule.CreatedBy, Valid: true}
	}

	query := `
		INSERT INTO neuronip.classification_rules
		(id, name, description, classification_type, rule_type, rule_expression,
		 confidence_threshold, enabled, priority, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, description, rule.ClassificationType, rule.RuleType,
		rule.RuleExpression, rule.ConfidenceThreshold, rule.Enabled, rule.Priority,
		metadataJSON, createdBy, rule.CreatedAt, rule.UpdatedAt,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return &rule, nil
}

/* ClassifyConnector classifies all columns in a connector */
func (s *Service) ClassifyConnector(ctx context.Context, connectorID uuid.UUID) error {
	// Get all columns from catalog
	query := `
		SELECT DISTINCT schema_name, table_name, column_name
		FROM neuronip.catalog_columns
		WHERE connector_id = $1`

	rows, err := s.pool.Query(ctx, query, connectorID)
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName, columnName string
		if err := rows.Scan(&schemaName, &tableName, &columnName); err != nil {
			continue
		}

		// Classify column (non-blocking)
		go func(schema, table, column string) {
			s.ClassifyColumn(context.Background(), connectorID, schema, table, column)
		}(schemaName, tableName, columnName)
	}

	return nil
}
