package semantic

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/auth"
)

/* PolicyAwareService provides policy-aware retrieval for semantic search */
type PolicyAwareService struct {
	pool           *pgxpool.Pool
	rowSecurityService *auth.RowSecurityService
}

/* NewPolicyAwareService creates a new policy-aware service */
func NewPolicyAwareService(pool *pgxpool.Pool, rowSecurityService *auth.RowSecurityService) *PolicyAwareService {
	return &PolicyAwareService{
		pool: pool,
		rowSecurityService: rowSecurityService,
	}
}

/* ApplyRLSToQuery applies RLS policies to a query */
func (s *PolicyAwareService) ApplyRLSToQuery(ctx context.Context, userID string, userRoles []string, query string, connectorID *uuid.UUID, schema string, table string) (string, error) {
	// Get RLS policies for this table
	policies, err := s.rowSecurityService.GetRowPolicies(ctx, connectorID, schema, table)
	if err != nil {
		return query, fmt.Errorf("failed to get RLS policies: %w", err)
	}

	if len(policies) == 0 {
		return query, nil // No policies, return original query
	}

	// Build filter expression from applicable policies
	var filterExpressions []string
	for _, policy := range policies {
		// Check if user roles match
		if len(policy.UserRoles) > 0 {
			roleMatch := false
			for _, role := range userRoles {
				for _, policyRole := range policy.UserRoles {
					if role == policyRole {
						roleMatch = true
						break
					}
				}
				if roleMatch {
					break
				}
			}
			if !roleMatch {
				continue // Skip this policy
			}
		}

		// Replace user context variables in filter expression
		filterExpr := policy.FilterExpression
		filterExpr = strings.ReplaceAll(filterExpr, "current_user_id()", fmt.Sprintf("'%s'", userID))
		filterExpr = strings.ReplaceAll(filterExpr, "current_user_roles()", fmt.Sprintf("ARRAY[%s]", strings.Join(userRoles, ",")))

		filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", filterExpr))
	}

	if len(filterExpressions) == 0 {
		return query, nil // No applicable policies
	}

	// Combine filters with OR (user can access if any policy allows)
	combinedFilter := strings.Join(filterExpressions, " OR ")

	// Add WHERE clause to query
	// This is simplified - in production, would use proper SQL parsing
	if strings.Contains(strings.ToUpper(query), "WHERE") {
		// Add to existing WHERE clause
		query = strings.Replace(query, "WHERE", fmt.Sprintf("WHERE (%s) AND", combinedFilter), 1)
	} else {
		// Add new WHERE clause
		query = fmt.Sprintf("%s WHERE %s", query, combinedFilter)
	}

	return query, nil
}

/* FilterDocumentsByPolicy filters retrieved documents based on RLS policies */
func (s *PolicyAwareService) FilterDocumentsByPolicy(ctx context.Context, userID string, userRoles []string, documents []map[string]interface{}) ([]map[string]interface{}, error) {
	// Get all RLS policies
	// In production, would cache policies and filter efficiently
	// For now, return all documents (policy application happens at query time)
	return documents, nil
}
