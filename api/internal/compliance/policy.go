package compliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* PolicyService provides policy management functionality */
type PolicyService struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewPolicyService creates a new policy service */
func NewPolicyService(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *PolicyService {
	return &PolicyService{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* Policy represents a compliance policy */
type Policy struct {
	ID          uuid.UUID              `json:"id"`
	PolicyName  string                 `json:"policy_name"`
	PolicyType  string                 `json:"policy_type"`
	PolicyText  string                 `json:"policy_text"`
	Rules       []map[string]interface{} `json:"rules,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Embedding   *string                `json:"embedding,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* CreatePolicy creates a new compliance policy */
func (s *PolicyService) CreatePolicy(ctx context.Context, policy Policy) (*Policy, error) {
	policy.ID = uuid.UUID{}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	// Generate embedding for policy text
	var embedding *string
	if policy.PolicyText != "" {
		emb, err := s.neurondbClient.GenerateEmbedding(ctx, policy.PolicyText, "sentence-transformers/all-MiniLM-L6-v2")
		if err == nil {
			embedding = &emb
		}
	}

	rulesJSON, _ := json.Marshal(policy.Rules)

	query := `
		INSERT INTO neuronip.compliance_policies 
		(id, policy_name, policy_type, policy_text, embedding, rules, enabled, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4::vector, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		policy.PolicyName, policy.PolicyType, policy.PolicyText, embedding, rulesJSON,
		policy.Enabled, policy.CreatedAt, policy.UpdatedAt,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return &policy, nil
}

/* GetPolicy retrieves a compliance policy */
func (s *PolicyService) GetPolicy(ctx context.Context, id uuid.UUID) (*Policy, error) {
	query := `
		SELECT id, policy_name, policy_type, policy_text, embedding, rules, enabled, created_at, updated_at
		FROM neuronip.compliance_policies
		WHERE id = $1`

	var policy Policy
	var rulesJSON json.RawMessage
	var embedding sql.NullString

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&policy.ID, &policy.PolicyName, &policy.PolicyType, &policy.PolicyText,
		&embedding, &rulesJSON, &policy.Enabled, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if embedding.Valid {
		policy.Embedding = &embedding.String
	}
	if rulesJSON != nil {
		json.Unmarshal(rulesJSON, &policy.Rules)
	}

	return &policy, nil
}

/* ListPolicies lists all compliance policies */
func (s *PolicyService) ListPolicies(ctx context.Context, policyType *string, enabled *bool) ([]Policy, error) {
	query := `
		SELECT id, policy_name, policy_type, policy_text, embedding, rules, enabled, created_at, updated_at
		FROM neuronip.compliance_policies
		WHERE ($1 IS NULL OR policy_type = $1) AND ($2 IS NULL OR enabled = $2)
		ORDER BY created_at DESC`

	args := []interface{}{policyType, enabled}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var policy Policy
		var rulesJSON json.RawMessage
		var embedding sql.NullString

		err := rows.Scan(
			&policy.ID, &policy.PolicyName, &policy.PolicyType, &policy.PolicyText,
			&embedding, &rulesJSON, &policy.Enabled, &policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if embedding.Valid {
			policy.Embedding = &embedding.String
		}
		if rulesJSON != nil {
			json.Unmarshal(rulesJSON, &policy.Rules)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

/* UpdatePolicy updates a compliance policy */
func (s *PolicyService) UpdatePolicy(ctx context.Context, id uuid.UUID, updates Policy) (*Policy, error) {
	// Generate new embedding if policy text changed
	var embedding *string
	if updates.PolicyText != "" {
		emb, err := s.neurondbClient.GenerateEmbedding(ctx, updates.PolicyText, "sentence-transformers/all-MiniLM-L6-v2")
		if err == nil {
			embedding = &emb
		}
	}

	rulesJSON, _ := json.Marshal(updates.Rules)

	query := `
		UPDATE neuronip.compliance_policies
		SET policy_name = $1, policy_type = $2, policy_text = $3, 
		    embedding = COALESCE($4::vector, embedding), rules = $5, enabled = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING id, policy_name, policy_type, policy_text, embedding, rules, enabled, created_at, updated_at`

	var policy Policy
	var rulesJSONResult json.RawMessage
	var embeddingResult sql.NullString

	err := s.pool.QueryRow(ctx, query,
		updates.PolicyName, updates.PolicyType, updates.PolicyText, embedding, rulesJSON,
		updates.Enabled, id,
	).Scan(
		&policy.ID, &policy.PolicyName, &policy.PolicyType, &policy.PolicyText,
		&embeddingResult, &rulesJSONResult, &policy.Enabled, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("policy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	if embeddingResult.Valid {
		policy.Embedding = &embeddingResult.String
	}
	if rulesJSONResult != nil {
		json.Unmarshal(rulesJSONResult, &policy.Rules)
	}

	return &policy, nil
}

/* DeletePolicy deletes a compliance policy */
func (s *PolicyService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.compliance_policies WHERE id = $1`
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

/* GetComplianceReport generates a compliance report */
func (s *PolicyService) GetComplianceReport(ctx context.Context, startTime, endTime time.Time, entityType *string) (*ComplianceReport, error) {
	query := `
		SELECT 
			cm.policy_id,
			cp.policy_name,
			cp.policy_type,
			COUNT(*) as violation_count,
			AVG(cm.match_score) as avg_match_score,
			MAX(cm.match_score) as max_match_score
		FROM neuronip.compliance_matches cm
		JOIN neuronip.compliance_policies cp ON cp.id = cm.policy_id
		WHERE cm.created_at >= $1 AND cm.created_at <= $2
			AND ($3 IS NULL OR cm.entity_type = $3)
			AND cm.status = 'violation'
		GROUP BY cm.policy_id, cp.policy_name, cp.policy_type
		ORDER BY violation_count DESC`

	args := []interface{}{startTime, endTime, entityType}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compliance report: %w", err)
	}
	defer rows.Close()

	report := &ComplianceReport{
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		Violations:  []PolicyViolation{},
		TotalViolations: 0,
	}

	for rows.Next() {
		var violation PolicyViolation
		var policyID uuid.UUID
		var violationCount int

		err := rows.Scan(
			&policyID, &violation.PolicyName, &violation.PolicyType,
			&violationCount, &violation.AvgMatchScore, &violation.MaxMatchScore,
		)
		if err != nil {
			continue
		}

		violation.PolicyID = policyID
		violation.Count = violationCount
		report.Violations = append(report.Violations, violation)
		report.TotalViolations += violationCount
	}

	return report, nil
}

/* ComplianceReport represents a compliance report */
type ComplianceReport struct {
	PeriodStart    time.Time         `json:"period_start"`
	PeriodEnd      time.Time         `json:"period_end"`
	TotalViolations int              `json:"total_violations"`
	Violations     []PolicyViolation `json:"violations"`
}

/* PolicyViolation represents a policy violation */
type PolicyViolation struct {
	PolicyID      uuid.UUID `json:"policy_id"`
	PolicyName    string    `json:"policy_name"`
	PolicyType    string    `json:"policy_type"`
	Count         int       `json:"count"`
	AvgMatchScore float64   `json:"avg_match_score"`
	MaxMatchScore float64   `json:"max_match_score"`
}
