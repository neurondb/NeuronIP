package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* RowSecurityPolicy represents a row-level security policy */
type RowSecurityPolicy struct {
	ID           uuid.UUID              `json:"id"`
	ConnectorID  *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName   string                 `json:"schema_name"`
	TableName    string                 `json:"table_name"`
	PolicyName   string                 `json:"policy_name"`
	FilterExpression string             `json:"filter_expression"` // SQL WHERE clause
	UserRoles    []string               `json:"user_roles"` // Roles this policy applies to
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    sql.NullTime           `json:"created_at"`
	UpdatedAt    sql.NullTime           `json:"updated_at"`
}

/* RowSecurityService provides row-level security */
type RowSecurityService struct {
	queries *db.Queries
}

/* NewRowSecurityService creates a new row security service */
func NewRowSecurityService(queries *db.Queries) *RowSecurityService {
	return &RowSecurityService{queries: queries}
}

/* CreateRowPolicy creates a row security policy */
func (s *RowSecurityService) CreateRowPolicy(ctx context.Context, policy RowSecurityPolicy) (*RowSecurityPolicy, error) {
	policy.ID = uuid.New()
	metadataJSON, _ := json.Marshal(policy.Metadata)
	userRolesJSON, _ := json.Marshal(policy.UserRoles)

	var connectorID sql.NullString
	if policy.ConnectorID != nil {
		connectorID = sql.NullString{String: policy.ConnectorID.String(), Valid: true}
	}

	query := `
		INSERT INTO neuronip.row_security_policies
		(id, connector_id, schema_name, table_name, policy_name, filter_expression,
		 user_roles, metadata, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := s.queries.DB.QueryRow(ctx, query,
		policy.ID, connectorID, policy.SchemaName, policy.TableName, policy.PolicyName,
		policy.FilterExpression, userRolesJSON, metadataJSON, policy.Enabled,
	).Scan(&policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create row policy: %w", err)
	}

	return &policy, nil
}

/* GetRowPolicies gets row security policies for a table */
func (s *RowSecurityService) GetRowPolicies(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName string) ([]RowSecurityPolicy, error) {
	var connectorIDParam sql.NullString
	if connectorID != nil {
		connectorIDParam = sql.NullString{String: connectorID.String(), Valid: true}
	}

	query := `
		SELECT id, connector_id, schema_name, table_name, policy_name, filter_expression,
		       user_roles, metadata, enabled, created_at, updated_at
		FROM neuronip.row_security_policies
		WHERE (connector_id = $1 OR ($1 IS NULL AND connector_id IS NULL))
		  AND schema_name = $2 AND table_name = $3
		  AND enabled = true`

	rows, err := s.queries.DB.Query(ctx, query, connectorIDParam, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get row policies: %w", err)
	}
	defer rows.Close()

	policies := []RowSecurityPolicy{}
	for rows.Next() {
		var policy RowSecurityPolicy
		var connectorIDStr sql.NullString
		var userRolesJSON, metadataJSON []byte

		err := rows.Scan(
			&policy.ID, &connectorIDStr, &policy.SchemaName, &policy.TableName,
			&policy.PolicyName, &policy.FilterExpression, &userRolesJSON, &metadataJSON,
			&policy.Enabled, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if connectorIDStr.Valid {
			connUUID, _ := uuid.Parse(connectorIDStr.String)
			policy.ConnectorID = &connUUID
		}
		if userRolesJSON != nil {
			json.Unmarshal(userRolesJSON, &policy.UserRoles)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &policy.Metadata)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

/* ApplyRowFilter applies row-level filtering to a query */
func (s *RowSecurityService) ApplyRowFilter(ctx context.Context, userID string, userRole string, connectorID *uuid.UUID, schemaName, tableName string, baseQuery string) (string, error) {
	policies, err := s.GetRowPolicies(ctx, connectorID, schemaName, tableName)
	if err != nil || len(policies) == 0 {
		return baseQuery, nil
	}

	// Find applicable policies for this user role
	applicableFilters := []string{}
	for _, policy := range policies {
		// Check if policy applies to this user role
		applies := false
		for _, allowedRole := range policy.UserRoles {
			if allowedRole == userRole || allowedRole == "*" {
				applies = true
				break
			}
		}

		// Check if user has admin permission (admins bypass row security)
		if !applies {
			rbacService := NewRBACService(s.queries)
			hasAdmin, _ := rbacService.HasPermission(ctx, userID, PermissionAdmin)
			if hasAdmin {
				return baseQuery, nil // Admins see all rows
			}
		}

		if applies {
			// Replace placeholders in filter expression
			filter := s.replaceFilterPlaceholders(policy.FilterExpression, userID, userRole)
			applicableFilters = append(applicableFilters, filter)
		}
	}

	if len(applicableFilters) == 0 {
		return baseQuery, nil
	}

	// Combine filters with AND
	combinedFilter := strings.Join(applicableFilters, " AND ")

	// Add WHERE clause to query
	if strings.Contains(strings.ToUpper(baseQuery), "WHERE") {
		// Query already has WHERE, add AND
		baseQuery += " AND (" + combinedFilter + ")"
	} else {
		// Add WHERE clause
		baseQuery += " WHERE " + combinedFilter
	}

	return baseQuery, nil
}

/* replaceFilterPlaceholders replaces placeholders in filter expression */
func (s *RowSecurityService) replaceFilterPlaceholders(filter string, userID string, userRole string) string {
	// Replace {user_id} with actual user ID
	filter = strings.ReplaceAll(filter, "{user_id}", fmt.Sprintf("'%s'", userID))
	
	// Replace {user_role} with actual user role
	filter = strings.ReplaceAll(filter, "{user_role}", fmt.Sprintf("'%s'", userRole))

	// Replace {tenant_id} - would need to get from context in production
	// For now, leave as placeholder

	return filter
}
