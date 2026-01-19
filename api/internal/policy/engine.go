package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PolicyEngine provides policy enforcement */
type PolicyEngine struct {
	pool *pgxpool.Pool
}

/* NewPolicyEngine creates a new policy engine */
func NewPolicyEngine(pool *pgxpool.Pool) *PolicyEngine {
	return &PolicyEngine{pool: pool}
}

/* Policy represents a policy definition */
type Policy struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     *string                `json:"description,omitempty"`
	PolicyType      string                 `json:"policy_type"`
	PolicyDefinition map[string]interface{} `json:"policy_definition"`
	Enabled         bool                   `json:"enabled"`
	Priority        int                    `json:"priority"`
	AppliesTo       []string               `json:"applies_to,omitempty"`
	Conditions      map[string]interface{} `json:"conditions,omitempty"`
	Actions         map[string]interface{} `json:"actions,omitempty"`
	CreatedBy       *string                `json:"created_by,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* EvaluatePolicy evaluates a policy against a request */
func (e *PolicyEngine) EvaluatePolicy(ctx context.Context, policyID uuid.UUID, request PolicyRequest) (*PolicyResult, error) {
	policy, err := e.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}
	
	if !policy.Enabled {
		return &PolicyResult{
			Allowed: true,
			Reason:  "policy_disabled",
		}, nil
	}
	
	// Check if policy applies to this resource type
	if len(policy.AppliesTo) > 0 {
		applies := false
		for _, resourceType := range policy.AppliesTo {
			if resourceType == request.ResourceType || resourceType == "*" {
				applies = true
				break
			}
		}
		if !applies {
			return &PolicyResult{
				Allowed: true,
				Reason:  "policy_not_applicable",
			}, nil
		}
	}
	
	// Evaluate conditions
	conditionsMet := e.evaluateConditions(policy.Conditions, request)
	if !conditionsMet {
		return &PolicyResult{
			Allowed: true,
			Reason:  "conditions_not_met",
		}, nil
	}
	
	// Apply policy based on type
	result := e.applyPolicy(policy, request)
	
	// Log enforcement
	e.logEnforcement(ctx, policyID, request, result)
	
	return result, nil
}

/* EvaluatePolicies evaluates all applicable policies */
func (e *PolicyEngine) EvaluatePolicies(ctx context.Context, request PolicyRequest) (*PolicyResult, error) {
	// Get all enabled policies for this resource type
	policies, err := e.getApplicablePolicies(ctx, request.ResourceType)
	if err != nil {
		return nil, err
	}
	
	// Evaluate policies in priority order
	for _, policy := range policies {
		result, err := e.EvaluatePolicy(ctx, policy.ID, request)
		if err != nil {
			continue
		}
		
		// If policy denies, return immediately
		if !result.Allowed {
			return result, nil
		}
		
		// If policy filters/modifies, apply and continue
		if result.Filtered || result.Modified {
			request = *result.ModifiedRequest
		}
	}
	
	return &PolicyResult{
		Allowed: true,
		Reason:  "all_policies_passed",
	}, nil
}

/* GetPolicy retrieves a policy */
func (e *PolicyEngine) GetPolicy(ctx context.Context, id uuid.UUID) (*Policy, error) {
	query := `
		SELECT id, name, description, policy_type, policy_definition, enabled, priority,
		       applies_to, conditions, actions, created_by, created_at, updated_at
		FROM neuronip.policies
		WHERE id = $1`
	
	var policy Policy
	var appliesToJSON, conditionsJSON, actionsJSON, definitionJSON json.RawMessage
	
	err := e.pool.QueryRow(ctx, query, id).Scan(
		&policy.ID, &policy.Name, &policy.Description, &policy.PolicyType, &definitionJSON,
		&policy.Enabled, &policy.Priority, &appliesToJSON, &conditionsJSON, &actionsJSON,
		&policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	
	if appliesToJSON != nil {
		json.Unmarshal(appliesToJSON, &policy.AppliesTo)
	}
	if conditionsJSON != nil {
		json.Unmarshal(conditionsJSON, &policy.Conditions)
	}
	if actionsJSON != nil {
		json.Unmarshal(actionsJSON, &policy.Actions)
	}
	if definitionJSON != nil {
		json.Unmarshal(definitionJSON, &policy.PolicyDefinition)
	}
	
	return &policy, nil
}

/* CreatePolicy creates a new policy */
func (e *PolicyEngine) CreatePolicy(ctx context.Context, policy Policy) (*Policy, error) {
	policy.ID = uuid.New()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	
	appliesToJSON, _ := json.Marshal(policy.AppliesTo)
	conditionsJSON, _ := json.Marshal(policy.Conditions)
	actionsJSON, _ := json.Marshal(policy.Actions)
	definitionJSON, _ := json.Marshal(policy.PolicyDefinition)
	
	query := `
		INSERT INTO neuronip.policies 
		(id, name, description, policy_type, policy_definition, enabled, priority,
		 applies_to, conditions, actions, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`
	
	err := e.pool.QueryRow(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.PolicyType, definitionJSON,
		policy.Enabled, policy.Priority, appliesToJSON, conditionsJSON, actionsJSON,
		policy.CreatedBy, policy.CreatedAt, policy.UpdatedAt,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}
	
	return &policy, nil
}

/* getApplicablePolicies gets policies applicable to a resource type */
func (e *PolicyEngine) getApplicablePolicies(ctx context.Context, resourceType string) ([]Policy, error) {
	query := `
		SELECT id, name, description, policy_type, policy_definition, enabled, priority,
		       applies_to, conditions, actions, created_by, created_at, updated_at
		FROM neuronip.policies
		WHERE enabled = true
		AND (applies_to = '{}' OR $1 = ANY(applies_to) OR '*' = ANY(applies_to))
		ORDER BY priority DESC, created_at ASC`
	
	rows, err := e.pool.Query(ctx, query, resourceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	policies := make([]Policy, 0)
	for rows.Next() {
		var policy Policy
		var appliesToJSON, conditionsJSON, actionsJSON, definitionJSON json.RawMessage
		
		err := rows.Scan(
			&policy.ID, &policy.Name, &policy.Description, &policy.PolicyType, &definitionJSON,
			&policy.Enabled, &policy.Priority, &appliesToJSON, &conditionsJSON, &actionsJSON,
			&policy.CreatedBy, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if appliesToJSON != nil {
			json.Unmarshal(appliesToJSON, &policy.AppliesTo)
		}
		if conditionsJSON != nil {
			json.Unmarshal(conditionsJSON, &policy.Conditions)
		}
		if actionsJSON != nil {
			json.Unmarshal(actionsJSON, &policy.Actions)
		}
		if definitionJSON != nil {
			json.Unmarshal(definitionJSON, &policy.PolicyDefinition)
		}
		
		policies = append(policies, policy)
	}
	
	return policies, nil
}

/* evaluateConditions evaluates policy conditions */
func (e *PolicyEngine) evaluateConditions(conditions map[string]interface{}, request PolicyRequest) bool {
	if len(conditions) == 0 {
		return true
	}
	
	// Simple condition evaluation
	// In production, would use a proper expression evaluator
	
	// Check user conditions
	if userCond, ok := conditions["user"].(map[string]interface{}); ok {
		if userID, ok := userCond["user_id"].(string); ok {
			if request.UserID != userID {
				return false
			}
		}
		if roles, ok := userCond["roles"].([]interface{}); ok {
			// Check if user has one of the required roles
			// This would need RBAC integration
		}
	}
	
	// Check resource conditions
	if resourceCond, ok := conditions["resource"].(map[string]interface{}); ok {
		if resourceType, ok := resourceCond["type"].(string); ok {
			if request.ResourceType != resourceType {
				return false
			}
		}
	}
	
	return true
}

/* applyPolicy applies a policy and returns the result */
func (e *PolicyEngine) applyPolicy(policy Policy, request PolicyRequest) *PolicyResult {
	switch policy.PolicyType {
	case "data_access":
		return e.applyDataAccessPolicy(policy, request)
	case "query_filter":
		return e.applyQueryFilterPolicy(policy, request)
	case "result_filter":
		return e.applyResultFilterPolicy(policy, request)
	default:
		return &PolicyResult{
			Allowed: true,
			Reason:  "unknown_policy_type",
		}
	}
}

/* applyDataAccessPolicy applies a data access policy */
func (e *PolicyEngine) applyDataAccessPolicy(policy Policy, request PolicyRequest) *PolicyResult {
	// Check if access is allowed
	if allow, ok := policy.PolicyDefinition["allow"].(bool); ok && !allow {
		return &PolicyResult{
			Allowed: false,
			Reason:  "access_denied_by_policy",
			Message: policy.Description,
		}
	}
	
	return &PolicyResult{
		Allowed: true,
		Reason:  "access_allowed",
	}
}

/* applyQueryFilterPolicy applies a query filter policy */
func (e *PolicyEngine) applyQueryFilterPolicy(policy Policy, request PolicyRequest) *PolicyResult {
	// Modify query based on policy
	modifiedRequest := request
	if filters, ok := policy.PolicyDefinition["filters"].(map[string]interface{}); ok {
		// Apply filters to query
		// This would modify the SQL or add WHERE clauses
		modifiedRequest.QueryFilters = filters
	}
	
	return &PolicyResult{
		Allowed:       true,
		Filtered:      true,
		ModifiedRequest: &modifiedRequest,
		Reason:        "query_filtered",
	}
}

/* applyResultFilterPolicy applies a result filter policy */
func (e *PolicyEngine) applyResultFilterPolicy(policy Policy, request PolicyRequest) *PolicyResult {
	// Policy will filter results after query execution
	return &PolicyResult{
		Allowed: true,
		Filtered: true,
		Reason:  "results_will_be_filtered",
	}
}

/* logEnforcement logs a policy enforcement */
func (e *PolicyEngine) logEnforcement(ctx context.Context, policyID uuid.UUID, request PolicyRequest, result *PolicyResult) {
	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"request": request,
		"result":  result,
	})
	
	query := `
		INSERT INTO neuronip.policy_enforcements 
		(policy_id, user_id, resource_type, resource_id, action, enforcement_result, enforcement_details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	
	userID := request.UserID
	e.pool.Exec(ctx, query, policyID, userID, request.ResourceType, request.ResourceID, request.Action, result.EnforcementResult(), detailsJSON)
}

/* PolicyRequest represents a policy evaluation request */
type PolicyRequest struct {
	UserID       string
	ResourceType string
	ResourceID   string
	Action       string
	Query        string
	QueryFilters map[string]interface{}
}

/* PolicyResult represents the result of policy evaluation */
type PolicyResult struct {
	Allowed        bool
	Filtered       bool
	Modified       bool
	ModifiedRequest *PolicyRequest
	Reason         string
	Message        *string
}

/* EnforcementResult returns the enforcement result string */
func (r *PolicyResult) EnforcementResult() string {
	if !r.Allowed {
		return "denied"
	}
	if r.Filtered {
		return "filtered"
	}
	if r.Modified {
		return "modified"
	}
	return "allowed"
}
