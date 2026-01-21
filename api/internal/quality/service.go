package quality

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides data quality functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new quality service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* QualityRule represents a data quality rule */
type QualityRule struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	RuleType      string                 `json:"rule_type"`
	ConnectorID   *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName    *string                `json:"schema_name,omitempty"`
	TableName     *string                `json:"table_name,omitempty"`
	ColumnName    *string                `json:"column_name,omitempty"`
	RuleExpression string                `json:"rule_expression"`
	Threshold     *float64               `json:"threshold,omitempty"`
	Enabled       bool                   `json:"enabled"`
	ScheduleCron  *string                `json:"schedule_cron,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy     *string                `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* QualityCheck represents a quality check execution result */
type QualityCheck struct {
	ID             uuid.UUID              `json:"id"`
	RuleID         uuid.UUID              `json:"rule_id"`
	ConnectorID    *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName     *string                `json:"schema_name,omitempty"`
	TableName      *string                `json:"table_name,omitempty"`
	ColumnName     *string                `json:"column_name,omitempty"`
	Status         string                 `json:"status"`
	Score          float64                `json:"score"`
	PassedCount    int64                  `json:"passed_count"`
	FailedCount    int64                  `json:"failed_count"`
	TotalCount     int64                  `json:"total_count"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	ExecutionTimeMs *int                  `json:"execution_time_ms,omitempty"`
	ExecutedAt     time.Time              `json:"executed_at"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

/* QualityScore represents an aggregated quality score */
type QualityScore struct {
	ID              uuid.UUID              `json:"id"`
	ConnectorID     *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName      *string                `json:"schema_name,omitempty"`
	TableName       *string                `json:"table_name,omitempty"`
	ColumnName      *string                `json:"column_name,omitempty"`
	Score           float64                `json:"score"`
	ScoreBreakdown  map[string]interface{} `json:"score_breakdown"`
	RuleCount       int                    `json:"rule_count"`
	PassedRules     int                    `json:"passed_rules"`
	FailedRules     int                    `json:"failed_rules"`
	LastCalculatedAt time.Time             `json:"last_calculated_at"`
	CalculatedAt    time.Time              `json:"calculated_at"`
}

/* QualityViolation represents a quality violation */
type QualityViolation struct {
	ID              uuid.UUID `json:"id"`
	CheckID         uuid.UUID `json:"check_id"`
	RuleID          uuid.UUID `json:"rule_id"`
	RowIdentifier   *string   `json:"row_identifier,omitempty"`
	ColumnValue     *string   `json:"column_value,omitempty"`
	ViolationType   string    `json:"violation_type"`
	ViolationMessage string   `json:"violation_message"`
	Severity        string    `json:"severity"`
	CreatedAt       time.Time `json:"created_at"`
}

/* CreateRule creates a new quality rule */
func (s *Service) CreateRule(ctx context.Context, rule QualityRule) (*QualityRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(rule.Metadata)

	query := `
		INSERT INTO neuronip.data_quality_rules
		(id, name, description, rule_type, connector_id, schema_name, table_name, column_name,
		 rule_expression, threshold, enabled, schedule_cron, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.RuleType,
		rule.ConnectorID, rule.SchemaName, rule.TableName, rule.ColumnName,
		rule.RuleExpression, rule.Threshold, rule.Enabled, rule.ScheduleCron,
		metadataJSON, rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return &rule, nil
}

/* GetRule retrieves a quality rule by ID */
func (s *Service) GetRule(ctx context.Context, id uuid.UUID) (*QualityRule, error) {
	var rule QualityRule
	var connectorID sql.NullString
	var schemaName, tableName, columnName sql.NullString
	var threshold sql.NullFloat64
	var scheduleCron sql.NullString
	var createdBy sql.NullString
	var metadataJSON []byte

	query := `
		SELECT id, name, description, rule_type, connector_id, schema_name, table_name,
		       column_name, rule_expression, threshold, enabled, schedule_cron, metadata,
		       created_by, created_at, updated_at
		FROM neuronip.data_quality_rules
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
		&connectorID, &schemaName, &tableName, &columnName,
		&rule.RuleExpression, &threshold, &rule.Enabled, &scheduleCron,
		&metadataJSON, &createdBy, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	if connectorID.Valid {
		id, _ := uuid.Parse(connectorID.String)
		rule.ConnectorID = &id
	}
	if schemaName.Valid {
		rule.SchemaName = &schemaName.String
	}
	if tableName.Valid {
		rule.TableName = &tableName.String
	}
	if columnName.Valid {
		rule.ColumnName = &columnName.String
	}
	if threshold.Valid {
		rule.Threshold = &threshold.Float64
	}
	if scheduleCron.Valid {
		rule.ScheduleCron = &scheduleCron.String
	}
	if createdBy.Valid {
		rule.CreatedBy = &createdBy.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &rule.Metadata)
	}

	return &rule, nil
}

/* ListRules lists quality rules */
func (s *Service) ListRules(ctx context.Context, enabled *bool, ruleType *string) ([]QualityRule, error) {
	query := `
		SELECT id, name, description, rule_type, connector_id, schema_name, table_name,
		       column_name, rule_expression, threshold, enabled, schedule_cron, metadata,
		       created_by, created_at, updated_at
		FROM neuronip.data_quality_rules
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if enabled != nil {
		query += fmt.Sprintf(" AND enabled = $%d", argIdx)
		args = append(args, *enabled)
		argIdx++
	}
	if ruleType != nil {
		query += fmt.Sprintf(" AND rule_type = $%d", argIdx)
		args = append(args, *ruleType)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}
	defer rows.Close()

	var rules []QualityRule
	for rows.Next() {
		var rule QualityRule
		var connectorID sql.NullString
		var schemaName, tableName, columnName sql.NullString
		var threshold sql.NullFloat64
		var scheduleCron sql.NullString
		var createdBy sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType,
			&connectorID, &schemaName, &tableName, &columnName,
			&rule.RuleExpression, &threshold, &rule.Enabled, &scheduleCron,
			&metadataJSON, &createdBy, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			rule.ConnectorID = &id
		}
		if schemaName.Valid {
			rule.SchemaName = &schemaName.String
		}
		if tableName.Valid {
			rule.TableName = &tableName.String
		}
		if columnName.Valid {
			rule.ColumnName = &columnName.String
		}
		if threshold.Valid {
			rule.Threshold = &threshold.Float64
		}
		if scheduleCron.Valid {
			rule.ScheduleCron = &scheduleCron.String
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

/* ExecuteRule executes a quality rule and returns a check result */
func (s *Service) ExecuteRule(ctx context.Context, ruleID uuid.UUID) (*QualityCheck, error) {
	rule, err := s.GetRule(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	if !rule.Enabled {
		return nil, fmt.Errorf("rule is disabled")
	}

	startTime := time.Now()
	check := QualityCheck{
		ID:         uuid.New(),
		RuleID:     rule.ID,
		ConnectorID: rule.ConnectorID,
		SchemaName:  rule.SchemaName,
		TableName:   rule.TableName,
		ColumnName:  rule.ColumnName,
		ExecutedAt:  time.Now(),
	}

	// Execute rule based on type
	switch rule.RuleType {
	case "completeness":
		result, err := s.executeCompletenessRule(ctx, rule)
		if err != nil {
			status := "error"
			msg := err.Error()
			check.Status = status
			check.ErrorMessage = &msg
			check.Score = 0
		} else {
			check.Status = result.Status
			check.Score = result.Score
			check.PassedCount = result.PassedCount
			check.FailedCount = result.FailedCount
			check.TotalCount = result.TotalCount
			check.Metadata = result.Metadata
		}
	case "uniqueness":
		result, err := s.executeUniquenessRule(ctx, rule)
		if err != nil {
			status := "error"
			msg := err.Error()
			check.Status = status
			check.ErrorMessage = &msg
			check.Score = 0
		} else {
			check.Status = result.Status
			check.Score = result.Score
			check.PassedCount = result.PassedCount
			check.FailedCount = result.FailedCount
			check.TotalCount = result.TotalCount
			check.Metadata = result.Metadata
		}
	case "validity":
		result, err := s.executeValidityRule(ctx, rule)
		if err != nil {
			status := "error"
			msg := err.Error()
			check.Status = status
			check.ErrorMessage = &msg
			check.Score = 0
		} else {
			check.Status = result.Status
			check.Score = result.Score
			check.PassedCount = result.PassedCount
			check.FailedCount = result.FailedCount
			check.TotalCount = result.TotalCount
			check.Metadata = result.Metadata
		}
	default:
		return nil, fmt.Errorf("unsupported rule type: %s", rule.RuleType)
	}

	executionTime := int(time.Since(startTime).Milliseconds())
	check.ExecutionTimeMs = &executionTime

	// Store check result
	metadataJSON, _ := json.Marshal(check.Metadata)
	checkQuery := `
		INSERT INTO neuronip.data_quality_checks
		(id, rule_id, connector_id, schema_name, table_name, column_name, status, score,
		 passed_count, failed_count, total_count, error_message, execution_time_ms, executed_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err = s.pool.Exec(ctx, checkQuery,
		check.ID, check.RuleID, check.ConnectorID, check.SchemaName,
		check.TableName, check.ColumnName, check.Status, check.Score,
		check.PassedCount, check.FailedCount, check.TotalCount,
		check.ErrorMessage, check.ExecutionTimeMs, check.ExecutedAt, metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to store check: %w", err)
	}

	// Update aggregated score
	s.updateQualityScore(ctx, check)

	return &check, nil
}

/* executeCompletenessRule executes a completeness rule */
func (s *Service) executeCompletenessRule(ctx context.Context, rule *QualityRule) (*QualityCheck, error) {
	if rule.TableName == nil || *rule.TableName == "" {
		return nil, fmt.Errorf("table_name is required for completeness rule")
	}

	var query string
	if rule.ColumnName != nil && *rule.ColumnName != "" {
		// Column-level completeness
		schema := "public"
		if rule.SchemaName != nil {
			schema = *rule.SchemaName
		}
		query = fmt.Sprintf(`
			SELECT 
				COUNT(*) as total_count,
				COUNT(%s) as non_null_count,
				COUNT(*) - COUNT(%s) as null_count
			FROM %s.%s`,
			*rule.ColumnName, *rule.ColumnName, schema, *rule.TableName)
	} else {
		return nil, fmt.Errorf("column_name is required for completeness rule")
	}

	var totalCount, nonNullCount, nullCount int64
	err := s.pool.QueryRow(ctx, query).Scan(&totalCount, &nonNullCount, &nullCount)
	if err != nil {
		return nil, fmt.Errorf("failed to execute completeness check: %w", err)
	}

	// Calculate score
	var score float64
	if totalCount > 0 {
		score = float64(nonNullCount) / float64(totalCount) * 100
	} else {
		score = 100.0
	}

	// Determine status
	status := "pass"
	if rule.Threshold != nil {
		if score < *rule.Threshold {
			status = "fail"
		} else if score < *rule.Threshold+10 {
			status = "warning"
		}
	} else {
		// Default threshold: 95%
		if score < 95.0 {
			status = "fail"
		}
	}

	return &QualityCheck{
		Status:      status,
		Score:       score,
		PassedCount: nonNullCount,
		FailedCount: nullCount,
		TotalCount:  totalCount,
		Metadata: map[string]interface{}{
			"non_null_count": nonNullCount,
			"null_count":     nullCount,
			"completeness_pct": score,
		},
	}, nil
}

/* executeUniquenessRule executes a uniqueness rule */
func (s *Service) executeUniquenessRule(ctx context.Context, rule *QualityRule) (*QualityCheck, error) {
	if rule.TableName == nil || rule.ColumnName == nil {
		return nil, fmt.Errorf("table_name and column_name are required for uniqueness rule")
	}

	schema := "public"
	if rule.SchemaName != nil {
		schema = *rule.SchemaName
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			COUNT(DISTINCT %s) as unique_count
		FROM %s.%s`,
		*rule.ColumnName, schema, *rule.TableName)

	var totalCount, uniqueCount int64
	err := s.pool.QueryRow(ctx, query).Scan(&totalCount, &uniqueCount)
	if err != nil {
		return nil, fmt.Errorf("failed to execute uniqueness check: %w", err)
	}

	// Calculate score (uniqueness percentage)
	var score float64
	if totalCount > 0 {
		score = float64(uniqueCount) / float64(totalCount) * 100
	} else {
		score = 100.0
	}

	duplicateCount := totalCount - uniqueCount

	// Determine status
	status := "pass"
	if rule.Threshold != nil {
		if score < *rule.Threshold {
			status = "fail"
		}
	} else {
		// Default: 100% uniqueness expected
		if duplicateCount > 0 {
			status = "fail"
		}
	}

	return &QualityCheck{
		Status:      status,
		Score:       score,
		PassedCount: uniqueCount,
		FailedCount: duplicateCount,
		TotalCount:  totalCount,
		Metadata: map[string]interface{}{
			"unique_count":    uniqueCount,
			"duplicate_count": duplicateCount,
			"uniqueness_pct":  score,
		},
	}, nil
}

/* executeValidityRule executes a validity rule using the rule expression */
func (s *Service) executeValidityRule(ctx context.Context, rule *QualityRule) (*QualityCheck, error) {
	if rule.TableName == nil {
		return nil, fmt.Errorf("table_name is required for validity rule")
	}

	schema := "public"
	if rule.SchemaName != nil {
		schema = *rule.SchemaName
	}

	// Build validation query from rule expression
	// For now, we'll use the rule_expression directly as a WHERE condition
	whereClause := rule.RuleExpression
	if whereClause == "" {
		whereClause = "1=1"
	}

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN %s THEN 1 ELSE 0 END) as valid_count,
			SUM(CASE WHEN NOT (%s) THEN 1 ELSE 0 END) as invalid_count
		FROM %s.%s`,
		whereClause, whereClause, schema, *rule.TableName)

	var totalCount, validCount, invalidCount int64
	err := s.pool.QueryRow(ctx, query).Scan(&totalCount, &validCount, &invalidCount)
	if err != nil {
		return nil, fmt.Errorf("failed to execute validity check: %w", err)
	}

	// Calculate score
	var score float64
	if totalCount > 0 {
		score = float64(validCount) / float64(totalCount) * 100
	} else {
		score = 100.0
	}

	// Determine status
	status := "pass"
	if rule.Threshold != nil {
		if score < *rule.Threshold {
			status = "fail"
		} else if score < *rule.Threshold+10 {
			status = "warning"
		}
	} else {
		// Default threshold: 95%
		if score < 95.0 {
			status = "fail"
		}
	}

	return &QualityCheck{
		Status:      status,
		Score:       score,
		PassedCount: validCount,
		FailedCount: invalidCount,
		TotalCount:  totalCount,
		Metadata: map[string]interface{}{
			"valid_count":   validCount,
			"invalid_count": invalidCount,
			"validity_pct":  score,
		},
	}, nil
}

/* updateQualityScore updates the aggregated quality score */
func (s *Service) updateQualityScore(ctx context.Context, check QualityCheck) {
	// Calculate aggregated score for the resource
	query := `
		SELECT 
			COUNT(*) as rule_count,
			SUM(CASE WHEN status = 'pass' THEN 1 ELSE 0 END) as passed_rules,
			SUM(CASE WHEN status IN ('fail', 'warning') THEN 1 ELSE 0 END) as failed_rules,
			AVG(score) as avg_score
		FROM neuronip.data_quality_checks
		WHERE connector_id IS NOT DISTINCT FROM $1
		  AND schema_name IS NOT DISTINCT FROM $2
		  AND table_name IS NOT DISTINCT FROM $3
		  AND column_name IS NOT DISTINCT FROM $4
		  AND executed_at >= NOW() - INTERVAL '7 days'`

	var ruleCount, passedRules, failedRules int
	var avgScore sql.NullFloat64

	err := s.pool.QueryRow(ctx, query,
		check.ConnectorID, check.SchemaName, check.TableName, check.ColumnName,
	).Scan(&ruleCount, &passedRules, &failedRules, &avgScore)

	if err != nil {
		return // Fail silently
	}

	score := 0.0
	if avgScore.Valid {
		score = avgScore.Float64
	}

	scoreBreakdownJSON, _ := json.Marshal(map[string]interface{}{
		"passed_rules": passedRules,
		"failed_rules": failedRules,
		"rule_count":   ruleCount,
	})

	upsertQuery := `
		INSERT INTO neuronip.data_quality_scores
		(connector_id, schema_name, table_name, column_name, score, score_breakdown,
		 rule_count, passed_rules, failed_rules, last_calculated_at, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (connector_id, schema_name, table_name, column_name)
		DO UPDATE SET
			score = EXCLUDED.score,
			score_breakdown = EXCLUDED.score_breakdown,
			rule_count = EXCLUDED.rule_count,
			passed_rules = EXCLUDED.passed_rules,
			failed_rules = EXCLUDED.failed_rules,
			last_calculated_at = EXCLUDED.last_calculated_at,
			calculated_at = EXCLUDED.calculated_at`

	s.pool.Exec(ctx, upsertQuery,
		check.ConnectorID, check.SchemaName, check.TableName, check.ColumnName,
		score, scoreBreakdownJSON, ruleCount, passedRules, failedRules,
	)
}

/* GetQualityScore retrieves quality score for a resource */
func (s *Service) GetQualityScore(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName, columnName *string) (*QualityScore, error) {
	var score QualityScore
	var scoreBreakdownJSON []byte

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, score,
		       score_breakdown, rule_count, passed_rules, failed_rules,
		       last_calculated_at, calculated_at
		FROM neuronip.data_quality_scores
		WHERE connector_id IS NOT DISTINCT FROM $1
		  AND schema_name IS NOT DISTINCT FROM $2
		  AND table_name IS NOT DISTINCT FROM $3
		  AND column_name IS NOT DISTINCT FROM $4`

	err := s.pool.QueryRow(ctx, query, connectorID, schemaName, tableName, columnName).Scan(
		&score.ID, &score.ConnectorID, &score.SchemaName, &score.TableName,
		&score.ColumnName, &score.Score, &scoreBreakdownJSON,
		&score.RuleCount, &score.PassedRules, &score.FailedRules,
		&score.LastCalculatedAt, &score.CalculatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No score yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quality score: %w", err)
	}

	if scoreBreakdownJSON != nil {
		json.Unmarshal(scoreBreakdownJSON, &score.ScoreBreakdown)
	}

	return &score, nil
}

/* ListQualityChecks lists quality checks */
func (s *Service) ListQualityChecks(ctx context.Context, ruleID *uuid.UUID, status *string, limit int) ([]QualityCheck, error) {
	query := `
		SELECT id, rule_id, connector_id, schema_name, table_name, column_name, status, score,
		       passed_count, failed_count, total_count, error_message, execution_time_ms,
		       executed_at, metadata
		FROM neuronip.data_quality_checks
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if ruleID != nil {
		query += fmt.Sprintf(" AND rule_id = $%d", argIdx)
		args = append(args, *ruleID)
		argIdx++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}

	query += " ORDER BY executed_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list checks: %w", err)
	}
	defer rows.Close()

	var checks []QualityCheck
	for rows.Next() {
		var check QualityCheck
		var connectorID sql.NullString
		var schemaName, tableName, columnName sql.NullString
		var errorMessage sql.NullString
		var executionTimeMs sql.NullInt32
		var metadataJSON []byte

		err := rows.Scan(
			&check.ID, &check.RuleID, &connectorID, &schemaName,
			&tableName, &columnName, &check.Status, &check.Score,
			&check.PassedCount, &check.FailedCount, &check.TotalCount,
			&errorMessage, &executionTimeMs, &check.ExecutedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			check.ConnectorID = &id
		}
		if schemaName.Valid {
			check.SchemaName = &schemaName.String
		}
		if tableName.Valid {
			check.TableName = &tableName.String
		}
		if columnName.Valid {
			check.ColumnName = &columnName.String
		}
		if errorMessage.Valid {
			check.ErrorMessage = &errorMessage.String
		}
		if executionTimeMs.Valid {
			timeMs := int(executionTimeMs.Int32)
			check.ExecutionTimeMs = &timeMs
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &check.Metadata)
		}

		checks = append(checks, check)
	}

	return checks, nil
}
