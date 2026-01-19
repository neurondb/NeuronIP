package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* RLSService provides row-level security management */
type RLSService struct {
	pool *pgxpool.Pool
}

/* NewRLSService creates a new RLS service */
func NewRLSService(pool *pgxpool.Pool) *RLSService {
	return &RLSService{pool: pool}
}

/* EnableRLS enables row-level security on a table */
func (s *RLSService) EnableRLS(ctx context.Context, schemaName, tableName string) error {
	query := fmt.Sprintf(`ALTER TABLE %s.%s ENABLE ROW LEVEL SECURITY`, schemaName, tableName)
	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable RLS on %s.%s: %w", schemaName, tableName, err)
	}
	return nil
}

/* CreateRLSPolicy creates a row-level security policy */
func (s *RLSService) CreateRLSPolicy(ctx context.Context, schemaName, tableName, policyName, policyDefinition string) error {
	query := fmt.Sprintf(`
		CREATE POLICY %s ON %s.%s
		%s
	`, policyName, schemaName, tableName, policyDefinition)
	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create RLS policy %s: %w", policyName, err)
	}
	return nil
}

/* SetupDefaultRLSPolicies sets up default RLS policies for NeuronIP tables */
func (s *RLSService) SetupDefaultRLSPolicies(ctx context.Context) error {
	// Enable RLS on key tables
	tables := []string{
		"knowledge_documents",
		"support_tickets",
		"warehouse_queries",
		"workflow_executions",
	}

	for _, table := range tables {
		err := s.EnableRLS(ctx, "neuronip", table)
		if err != nil {
			// Log but continue - table might already have RLS enabled
			continue
		}
	}

	// Create policy examples (actual policies would be based on user roles and data residency)
	// Example: Users can only see their own queries
	policyQuery := `
		CREATE POLICY IF NOT EXISTS users_own_queries ON neuronip.warehouse_queries
		FOR SELECT
		USING (user_id = current_setting('app.user_id', true))
	`
	s.pool.Exec(ctx, policyQuery)

	return nil
}

/* SetUserContext sets the current user context for RLS policies */
func (s *RLSService) SetUserContext(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `SET LOCAL app.user_id = $1`, userID)
	return err
}

/* SetDataResidency sets data residency context for RLS policies */
func (s *RLSService) SetDataResidency(ctx context.Context, region string) error {
	_, err := s.pool.Exec(ctx, `SET LOCAL app.data_region = $1`, region)
	return err
}
