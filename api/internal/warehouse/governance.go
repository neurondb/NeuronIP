package warehouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* GovernanceService provides query governance functionality */
type GovernanceService struct {
	pool *pgxpool.Pool
}

/* NewGovernanceService creates a new governance service */
func NewGovernanceService(pool *pgxpool.Pool) *GovernanceService {
	return &GovernanceService{pool: pool}
}

/* ValidateQuery validates a query against governance rules */
func (s *GovernanceService) ValidateQuery(ctx context.Context, queryText, userRole string, userID *string) (*QueryValidationResult, error) {
	result := &QueryValidationResult{
		Allowed: true,
		Warnings: []string{},
	}

	// Check read-only mode
	if s.isReadOnlyMode(ctx, userRole) {
		if s.containsWriteOperations(queryText) {
			result.Allowed = false
			result.BlockedReason = "Read-only mode: write operations not allowed"
			return result, nil
		}
	}

	// Check function allow-list
	allowed, reason := s.checkFunctionAllowList(ctx, queryText, userRole)
	if !allowed {
		result.Allowed = false
		result.BlockedReason = reason
		return result, nil
	}

	// Check sandbox role restrictions
	if s.isSandboxRole(ctx, userRole) {
		sandboxResult := s.checkSandboxRestrictions(ctx, queryText, userRole)
		if !sandboxResult.Allowed {
			result.Allowed = false
			result.BlockedReason = sandboxResult.BlockedReason
			return result, nil
		}
		result.Warnings = append(result.Warnings, sandboxResult.Warnings...)
	}

	// Estimate query cost
	cost, err := s.estimateQueryCost(ctx, queryText)
	if err == nil {
		result.EstimatedCost = cost
		// Check cost limits
		if cost > s.getCostLimit(ctx, userRole) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Query cost (%v) exceeds recommended limit", cost))
		}
	}

	return result, nil
}

/* QueryValidationResult represents query validation result */
type QueryValidationResult struct {
	Allowed        bool
	BlockedReason  string
	Warnings       []string
	EstimatedCost  float64
}

/* isReadOnlyMode checks if user role is in read-only mode */
func (s *GovernanceService) isReadOnlyMode(ctx context.Context, role string) bool {
	query := `SELECT read_only FROM neuronip.sandbox_roles WHERE role_name = $1`
	var readOnly bool
	err := s.pool.QueryRow(ctx, query, role).Scan(&readOnly)
	if err != nil {
		return false
	}
	return readOnly
}

/* containsWriteOperations checks if query contains write operations */
func (s *GovernanceService) containsWriteOperations(queryText string) bool {
	upperQuery := strings.ToUpper(queryText)
	writeOps := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE"}
	for _, op := range writeOps {
		if strings.Contains(upperQuery, op) {
			return true
		}
	}
	return false
}

/* checkFunctionAllowList checks if query uses only allowed functions */
func (s *GovernanceService) checkFunctionAllowList(ctx context.Context, queryText, role string) (bool, string) {
	// Extract function names from query (simplified)
	// In production, use proper SQL parser
	query := `
		SELECT function_name, allowed_for_roles
		FROM neuronip.allowed_functions
		WHERE is_safe = true
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return true, "" // Allow if check fails
	}
	defer rows.Close()

	allowedFunctions := make(map[string]bool)
	for rows.Next() {
		var funcName string
		var allowedRoles []string
		rows.Scan(&funcName, &allowedRoles)
		// Check if role is allowed
		allowed := len(allowedRoles) == 0 // Empty means all roles
		for _, r := range allowedRoles {
			if r == role {
				allowed = true
				break
			}
		}
		allowedFunctions[strings.ToUpper(funcName)] = allowed
	}

	// Check query for disallowed functions (simplified check)
	upperQuery := strings.ToUpper(queryText)
	for funcName, allowed := range allowedFunctions {
		if !allowed && strings.Contains(upperQuery, funcName+"(") {
			return false, fmt.Sprintf("Function %s is not allowed for role %s", funcName, role)
		}
	}

	return true, ""
}

/* isSandboxRole checks if role is a sandbox role */
func (s *GovernanceService) isSandboxRole(ctx context.Context, role string) bool {
	query := `SELECT COUNT(*) FROM neuronip.sandbox_roles WHERE role_name = $1`
	var count int
	s.pool.QueryRow(ctx, query, role).Scan(&count)
	return count > 0
}

/* checkSandboxRestrictions checks sandbox role restrictions */
func (s *GovernanceService) checkSandboxRestrictions(ctx context.Context, queryText, role string) *QueryValidationResult {
	result := &QueryValidationResult{Allowed: true, Warnings: []string{}}

	query := `
		SELECT max_query_time_seconds, max_result_rows, max_query_cost,
		       allowed_schemas, blocked_schemas, read_only
		FROM neuronip.sandbox_roles
		WHERE role_name = $1
	`
	var maxTime, maxRows int
	var maxCost *float64
	var allowedSchemas, blockedSchemas []string
	var readOnly bool

	err := s.pool.QueryRow(ctx, query, role).Scan(
		&maxTime, &maxRows, &maxCost, &allowedSchemas, &blockedSchemas, &readOnly,
	)
	if err != nil {
		return result
	}

	// Check schema access
	if len(blockedSchemas) > 0 {
		for _, schema := range blockedSchemas {
			if strings.Contains(strings.ToUpper(queryText), strings.ToUpper(schema)+".") {
				result.Allowed = false
				result.BlockedReason = fmt.Sprintf("Access to schema %s is blocked for sandbox role", schema)
				return result
			}
		}
	}

	if len(allowedSchemas) > 0 {
		hasAllowedSchema := false
		for _, schema := range allowedSchemas {
			if strings.Contains(strings.ToUpper(queryText), strings.ToUpper(schema)+".") {
				hasAllowedSchema = true
				break
			}
		}
		if !hasAllowedSchema {
			result.Warnings = append(result.Warnings, "Query may access schemas outside allowed list")
		}
	}

	// Check read-only
	if readOnly && s.containsWriteOperations(queryText) {
		result.Allowed = false
		result.BlockedReason = "Sandbox role is read-only"
		return result
	}

	return result
}

/* estimateQueryCost estimates query execution cost */
func (s *GovernanceService) estimateQueryCost(ctx context.Context, queryText string) (float64, error) {
	// Simplified cost estimation
	// In production, use EXPLAIN to get actual cost estimate
	words := strings.Fields(queryText)
	complexity := float64(len(words))
	
	// Check for expensive operations
	if strings.Contains(strings.ToUpper(queryText), "JOIN") {
		complexity *= 1.5
	}
	if strings.Contains(strings.ToUpper(queryText), "GROUP BY") {
		complexity *= 1.3
	}
	if strings.Contains(strings.ToUpper(queryText), "ORDER BY") {
		complexity *= 1.2
	}

	return complexity, nil
}

/* getCostLimit gets cost limit for a role */
func (s *GovernanceService) getCostLimit(ctx context.Context, role string) float64 {
	query := `SELECT max_query_cost FROM neuronip.sandbox_roles WHERE role_name = $1`
	var maxCost *float64
	s.pool.QueryRow(ctx, query, role).Scan(&maxCost)
	if maxCost != nil {
		return *maxCost
	}
	return 1000.0 // Default limit
}

/* SanitizeQuery sanitizes a query for safe execution */
func (s *GovernanceService) SanitizeQuery(queryText string) string {
	// Remove comments
	lines := strings.Split(queryText, "\n")
	var sanitized []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			sanitized = append(sanitized, line)
		}
	}
	return strings.Join(sanitized, "\n")
}
