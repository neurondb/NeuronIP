package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* ColumnSecurityPolicy represents a column-level security policy */
type ColumnSecurityPolicy struct {
	ID           uuid.UUID              `json:"id"`
	ConnectorID  *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName   string                 `json:"schema_name"`
	TableName    string                 `json:"table_name"`
	ColumnName   string                 `json:"column_name"`
	PolicyType   string                 `json:"policy_type"` // "mask", "hide", "redact"
	MaskingRule  *string                `json:"masking_rule,omitempty"`
	UserRoles    []string               `json:"user_roles"` // Roles that can see this column
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    sql.NullTime           `json:"created_at"`
	UpdatedAt    sql.NullTime           `json:"updated_at"`
}

/* ColumnSecurityService provides column-level security */
type ColumnSecurityService struct {
	queries *db.Queries
}

/* NewColumnSecurityService creates a new column security service */
func NewColumnSecurityService(queries *db.Queries) *ColumnSecurityService {
	return &ColumnSecurityService{queries: queries}
}

/* CreateColumnPolicy creates a column security policy */
func (s *ColumnSecurityService) CreateColumnPolicy(ctx context.Context, policy ColumnSecurityPolicy) (*ColumnSecurityPolicy, error) {
	policy.ID = uuid.New()
	metadataJSON, _ := json.Marshal(policy.Metadata)
	userRolesJSON, _ := json.Marshal(policy.UserRoles)

	var connectorID sql.NullString
	if policy.ConnectorID != nil {
		connectorID = sql.NullString{String: policy.ConnectorID.String(), Valid: true}
	}

	var maskingRule sql.NullString
	if policy.MaskingRule != nil {
		maskingRule = sql.NullString{String: *policy.MaskingRule, Valid: true}
	}

	query := `
		INSERT INTO neuronip.column_security_policies
		(id, connector_id, schema_name, table_name, column_name, policy_type,
		 masking_rule, user_roles, metadata, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := s.queries.DB.QueryRow(ctx, query,
		policy.ID, connectorID, policy.SchemaName, policy.TableName, policy.ColumnName,
		policy.PolicyType, maskingRule, userRolesJSON, metadataJSON, policy.Enabled,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create column policy: %w", err)
	}

	return &policy, nil
}

/* GetColumnPolicy gets a column security policy */
func (s *ColumnSecurityService) GetColumnPolicy(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName, columnName string) (*ColumnSecurityPolicy, error) {
	var policy ColumnSecurityPolicy
	var connectorIDStr sql.NullString
	var maskingRule sql.NullString
	var userRolesJSON, metadataJSON []byte

	var connectorIDParam sql.NullString
	if connectorID != nil {
		connectorIDParam = sql.NullString{String: connectorID.String(), Valid: true}
	}

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, policy_type,
		       masking_rule, user_roles, metadata, enabled, created_at, updated_at
		FROM neuronip.column_security_policies
		WHERE (connector_id = $1 OR ($1 IS NULL AND connector_id IS NULL))
		  AND schema_name = $2 AND table_name = $3 AND column_name = $4
		  AND enabled = true
		LIMIT 1`

	err := s.queries.DB.QueryRow(ctx, query, connectorIDParam, schemaName, tableName, columnName).Scan(
		&policy.ID, &connectorIDStr, &policy.SchemaName, &policy.TableName, &policy.ColumnName,
		&policy.PolicyType, &maskingRule, &userRolesJSON, &metadataJSON, &policy.Enabled,
		&policy.CreatedAt, &policy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("column policy not found: %w", err)
	}

	if connectorIDStr.Valid {
		connUUID, _ := uuid.Parse(connectorIDStr.String)
		policy.ConnectorID = &connUUID
	}
	if maskingRule.Valid {
		policy.MaskingRule = &maskingRule.String
	}
	if userRolesJSON != nil {
		json.Unmarshal(userRolesJSON, &policy.UserRoles)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &policy.Metadata)
	}

	return &policy, nil
}

/* CheckColumnAccess checks if a user can access a column */
func (s *ColumnSecurityService) CheckColumnAccess(ctx context.Context, userID string, userRole string, connectorID *uuid.UUID, schemaName, tableName, columnName string) (bool, error) {
	policy, err := s.GetColumnPolicy(ctx, connectorID, schemaName, tableName, columnName)
	if err != nil {
		// No policy means column is accessible
		return true, nil
	}

	// Check if user role is in allowed roles
	for _, allowedRole := range policy.UserRoles {
		if allowedRole == userRole || allowedRole == "*" {
			return true, nil
		}
	}

	// Check if user has admin permission
	rbacService := NewRBACService(s.queries)
	hasAdmin, _ := rbacService.HasPermission(ctx, userID, PermissionAdmin)
	if hasAdmin {
		return true, nil
	}

	return false, nil
}

/* ApplyColumnMasking applies masking to a column value based on policy */
func (s *ColumnSecurityService) ApplyColumnMasking(ctx context.Context, userID string, userRole string, connectorID *uuid.UUID, schemaName, tableName, columnName string, value interface{}) (interface{}, error) {
	hasAccess, err := s.CheckColumnAccess(ctx, userID, userRole, connectorID, schemaName, tableName, columnName)
	if err != nil {
		return value, err
	}

	if hasAccess {
		return value, nil
	}

	policy, err := s.GetColumnPolicy(ctx, connectorID, schemaName, tableName, columnName)
	if err != nil {
		return value, nil
	}

	// Apply masking based on policy type
	switch policy.PolicyType {
	case "hide":
		return nil, nil
	case "mask":
		return s.applyMaskingRule(value, policy.MaskingRule)
	case "redact":
		return "[REDACTED]", nil
	default:
		return value, nil
	}
}

/* applyMaskingRule applies a masking rule to a value */
func (s *ColumnSecurityService) applyMaskingRule(value interface{}, rule *string) (interface{}, error) {
	if rule == nil {
		return "[MASKED]", nil
	}

	// Simple masking rules
	valueStr := fmt.Sprintf("%v", value)
	switch *rule {
	case "email":
		// Mask email: user@domain.com -> u***@d***.com
		if len(valueStr) > 0 {
			parts := splitEmail(valueStr)
			if len(parts) == 2 {
				return maskString(parts[0], 1) + "@" + maskString(parts[1], 1), nil
			}
		}
		return maskString(valueStr, 3), nil
	case "phone":
		// Mask phone: (123) 456-7890 -> (***) ***-7890
		return maskString(valueStr, len(valueStr)-4), nil
	case "ssn":
		// Mask SSN: 123-45-6789 -> ***-**-6789
		return "***-**-" + valueStr[len(valueStr)-4:], nil
	case "partial":
		// Show first and last 2 characters
		if len(valueStr) > 4 {
			return valueStr[:2] + "***" + valueStr[len(valueStr)-2:], nil
		}
		return "***", nil
	default:
		return "[MASKED]", nil
	}
}

/* maskString masks a string showing only first n characters */
func maskString(s string, showChars int) string {
	if len(s) <= showChars {
		return "***"
	}
	return s[:showChars] + "***"
}

/* splitEmail splits an email into local and domain parts */
func splitEmail(email string) []string {
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			return []string{email[:i], email[i+1:]}
		}
	}
	return []string{email}
}
