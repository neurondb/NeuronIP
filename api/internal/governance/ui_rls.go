package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* UIRLSService provides UI-driven row-level security configuration */
type UIRLSService struct {
	pool *pgxpool.Pool
}

/* NewUIRLSService creates a new UI RLS service */
func NewUIRLSService(pool *pgxpool.Pool) *UIRLSService {
	return &UIRLSService{pool: pool}
}

/* RLSPolicy represents a UI-configured RLS policy */
type RLSPolicy struct {
	ID               uuid.UUID              `json:"id"`
	TableName        string                 `json:"table_name"`
	SchemaName       string                 `json:"schema_name"`
	PolicyName       string                 `json:"policy_name"`
	PolicyType       string                 `json:"policy_type"`                 // "select", "insert", "update", "delete", "all"
	Condition        string                 `json:"condition"`                   // SQL condition
	ConditionBuilder map[string]interface{} `json:"condition_builder,omitempty"` // UI builder config
	Enabled          bool                   `json:"enabled"`
	Description      string                 `json:"description,omitempty"`
	CreatedBy        string                 `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

/* CreateRLSPolicy creates a new RLS policy via UI */
func (urs *UIRLSService) CreateRLSPolicy(ctx context.Context, policy RLSPolicy) (*RLSPolicy, error) {
	policy.ID = uuid.New()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	builderJSON, _ := json.Marshal(policy.ConditionBuilder)

	query := `
		INSERT INTO neuronip.ui_rls_policies 
		(id, table_name, schema_name, policy_name, policy_type, condition, condition_builder, enabled, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := urs.pool.QueryRow(ctx, query,
		policy.ID, policy.TableName, policy.SchemaName, policy.PolicyName, policy.PolicyType,
		policy.Condition, builderJSON, policy.Enabled, policy.Description, policy.CreatedBy,
		policy.CreatedAt, policy.UpdatedAt,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create RLS policy: %w", err)
	}

	// Apply policy to database
	if err := urs.applyPolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to apply policy: %w", err)
	}

	return &policy, nil
}

/* applyPolicy applies an RLS policy to the database */
func (urs *UIRLSService) applyPolicy(ctx context.Context, policy RLSPolicy) error {
	// Enable RLS on table if not already enabled
	enableRLSQuery := fmt.Sprintf(`ALTER TABLE %s.%s ENABLE ROW LEVEL SECURITY`, policy.SchemaName, policy.TableName)
	urs.pool.Exec(ctx, enableRLSQuery) // Ignore error if already enabled

	// Create policy
	policyQuery := fmt.Sprintf(`
		CREATE POLICY %s ON %s.%s
		FOR %s
		USING (%s)
	`, policy.PolicyName, policy.SchemaName, policy.TableName, policy.PolicyType, policy.Condition)

	_, err := urs.pool.Exec(ctx, policyQuery)
	return err
}

/* GetRLSPolicies retrieves RLS policies for a table */
func (urs *UIRLSService) GetRLSPolicies(ctx context.Context, schemaName, tableName string) ([]RLSPolicy, error) {
	query := `
		SELECT id, table_name, schema_name, policy_name, policy_type, condition, condition_builder, enabled, description, created_by, created_at, updated_at
		FROM neuronip.ui_rls_policies
		WHERE schema_name = $1 AND table_name = $2
		ORDER BY created_at DESC
	`

	rows, err := urs.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get RLS policies: %w", err)
	}
	defer rows.Close()

	var policies []RLSPolicy
	for rows.Next() {
		var policy RLSPolicy
		var builderJSON json.RawMessage

		err := rows.Scan(
			&policy.ID, &policy.TableName, &policy.SchemaName, &policy.PolicyName, &policy.PolicyType,
			&policy.Condition, &builderJSON, &policy.Enabled, &policy.Description, &policy.CreatedBy,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if builderJSON != nil {
			json.Unmarshal(builderJSON, &policy.ConditionBuilder)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

/* UpdateRLSPolicy updates an RLS policy */
func (urs *UIRLSService) UpdateRLSPolicy(ctx context.Context, policyID uuid.UUID, updates map[string]interface{}) error {
	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if condition, ok := updates["condition"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("condition = $%d", argIndex))
		args = append(args, condition)
		argIndex++
	}

	if enabled, ok := updates["enabled"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("enabled = $%d", argIndex))
		args = append(args, enabled)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no updates provided")
	}

	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, policyID)

	query := fmt.Sprintf(`
		UPDATE neuronip.ui_rls_policies
		SET %s
		WHERE id = $%d
	`, fmt.Sprintf("%s, ", setParts[:len(setParts)-1]), argIndex)

	_, err := urs.pool.Exec(ctx, query, args...)
	return err
}

/* TogglePolicy enables or disables an RLS policy */
func (urs *UIRLSService) TogglePolicy(ctx context.Context, policyID uuid.UUID, enabled bool) error {
	query := `UPDATE neuronip.ui_rls_policies SET enabled = $1, updated_at = NOW() WHERE id = $2`
	_, err := urs.pool.Exec(ctx, query, enabled, policyID)
	return err
}

/* DeleteRLSPolicy deletes an RLS policy */
func (urs *UIRLSService) DeleteRLSPolicy(ctx context.Context, policyID uuid.UUID) error {
	// Get policy to drop from database
	var schemaName, tableName, policyName string
	err := urs.pool.QueryRow(ctx, `SELECT schema_name, table_name, policy_name FROM neuronip.ui_rls_policies WHERE id = $1`, policyID).Scan(
		&schemaName, &tableName, &policyName,
	)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Drop policy from database
	dropQuery := fmt.Sprintf(`DROP POLICY IF EXISTS %s ON %s.%s`, policyName, schemaName, tableName)
	urs.pool.Exec(ctx, dropQuery) // Ignore error

	// Delete from table
	_, err = urs.pool.Exec(ctx, `DELETE FROM neuronip.ui_rls_policies WHERE id = $1`, policyID)
	return err
}
