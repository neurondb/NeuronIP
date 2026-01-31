package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Validator provides data quality validation functionality */
type Validator struct {
	pool *pgxpool.Pool
}

/* NewValidator creates a new data quality validator */
func NewValidator(pool *pgxpool.Pool) *Validator {
	return &Validator{pool: pool}
}

/* QualityRule represents a data quality validation rule */
type QualityRule struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	RuleType    string                 `json:"rule_type"` // "not_null", "unique", "range", "regex", "custom"
	TableName   string                 `json:"table_name"`
	ColumnName  *string                `json:"column_name,omitempty"`
	RuleConfig  map[string]interface{} `json:"rule_config"`
	Severity    string                 `json:"severity"` // "error", "warning", "info"
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* ValidationResult represents the result of a quality check */
type ValidationResult struct {
	RuleID    uuid.UUID              `json:"rule_id"`
	RuleName  string                 `json:"rule_name"`
	Passed    bool                   `json:"passed"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CheckedAt time.Time              `json:"checked_at"`
}

/* ValidateData validates data against quality rules */
func (v *Validator) ValidateData(ctx context.Context, tableName string, data map[string]interface{}) ([]ValidationResult, error) {
	// Get enabled rules for this table
	rules, err := v.GetRulesForTable(ctx, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	var results []ValidationResult
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		result := v.validateAgainstRule(ctx, rule, data)
		results = append(results, result)

		// Store validation result
		v.storeValidationResult(ctx, result, tableName, data)
	}

	return results, nil
}

/* validateAgainstRule validates data against a specific rule */
func (v *Validator) validateAgainstRule(ctx context.Context, rule QualityRule, data map[string]interface{}) ValidationResult {
	result := ValidationResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		Severity:  rule.Severity,
		CheckedAt: time.Now(),
	}

	// Get column value if column is specified
	var value interface{}
	if rule.ColumnName != nil {
		value = data[*rule.ColumnName]
	} else {
		value = data
	}

	// Apply rule based on type
	switch rule.RuleType {
	case "not_null":
		result.Passed = value != nil && value != ""
		if !result.Passed {
			result.Message = fmt.Sprintf("Column %s cannot be null", *rule.ColumnName)
		}

	case "unique":
		// Single-record uniqueness: value must be present; full uniqueness across dataset is enforced at load/DB level
		result.Passed = value != nil && value != ""
		if result.Passed {
			result.Message = "Uniqueness check (single record): value present"
		} else {
			result.Message = "Uniqueness check failed: value is empty"
		}

	case "range":
		result.Passed = v.validateRange(value, rule.RuleConfig)
		if !result.Passed {
			result.Message = fmt.Sprintf("Value %v is outside allowed range", value)
		}

	case "regex":
		if strValue, ok := value.(string); ok {
			pattern, _ := rule.RuleConfig["pattern"].(string)
			matched, _ := regexp.MatchString(pattern, strValue)
			result.Passed = matched
			if !result.Passed {
				result.Message = fmt.Sprintf("Value does not match pattern: %s", pattern)
			}
		} else {
			result.Passed = false
			result.Message = "Regex validation requires string value"
		}

	case "custom":
		// Custom validation logic
		result.Passed = v.validateCustom(value, rule.RuleConfig)
		if !result.Passed {
			result.Message = "Custom validation failed"
		}

	default:
		result.Passed = false
		result.Message = fmt.Sprintf("Unknown rule type: %s", rule.RuleType)
	}

	return result
}

/* validateRange validates a value against a range rule */
func (v *Validator) validateRange(value interface{}, config map[string]interface{}) bool {
	min, hasMin := config["min"]
	max, hasMax := config["max"]

	switch val := value.(type) {
	case int:
		if hasMin {
			if minVal, ok := min.(float64); ok && float64(val) < minVal {
				return false
			}
		}
		if hasMax {
			if maxVal, ok := max.(float64); ok && float64(val) > maxVal {
				return false
			}
		}
		return true

	case float64:
		if hasMin {
			if minVal, ok := min.(float64); ok && val < minVal {
				return false
			}
		}
		if hasMax {
			if maxVal, ok := max.(float64); ok && val > maxVal {
				return false
			}
		}
		return true

	default:
		return false
	}
}

/* validateCustom validates using custom logic. Custom rules are allow-list until a safe expression evaluator or MCP integration is added; until then we accept the value. */
func (v *Validator) validateCustom(value interface{}, config map[string]interface{}) bool {
	return true
}

/* GetRulesForTable retrieves quality rules for a table */
func (v *Validator) GetRulesForTable(ctx context.Context, tableName string) ([]QualityRule, error) {
	query := `
		SELECT id, name, description, rule_type, table_name, column_name, rule_config, severity, enabled, created_at, updated_at
		FROM neuronip.quality_rules
		WHERE table_name = $1 AND enabled = true
		ORDER BY created_at DESC
	`

	rows, err := v.pool.Query(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	defer rows.Close()

	var rules []QualityRule
	for rows.Next() {
		var rule QualityRule
		var columnName *string
		var configJSON json.RawMessage

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.RuleType, &rule.TableName,
			&columnName, &configJSON, &rule.Severity, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			continue
		}

		rule.ColumnName = columnName
		if configJSON != nil {
			json.Unmarshal(configJSON, &rule.RuleConfig)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

/* CreateRule creates a new quality rule */
func (v *Validator) CreateRule(ctx context.Context, rule QualityRule) (*QualityRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	configJSON, _ := json.Marshal(rule.RuleConfig)

	query := `
		INSERT INTO neuronip.quality_rules 
		(id, name, description, rule_type, table_name, column_name, rule_config, severity, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := v.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.RuleType, rule.TableName, rule.ColumnName,
		configJSON, rule.Severity, rule.Enabled, rule.CreatedAt, rule.UpdatedAt,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return &rule, nil
}

/* storeValidationResult stores a validation result */
func (v *Validator) storeValidationResult(ctx context.Context, result ValidationResult, tableName string, data map[string]interface{}) {
	resultID := uuid.New()
	detailsJSON, _ := json.Marshal(result.Details)
	dataJSON, _ := json.Marshal(data)

	query := `
		INSERT INTO neuronip.quality_validation_results 
		(id, rule_id, table_name, passed, message, severity, details, data_sample, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	v.pool.Exec(ctx, query,
		resultID, result.RuleID, tableName, result.Passed, result.Message, result.Severity,
		detailsJSON, dataJSON, result.CheckedAt,
	)
}
