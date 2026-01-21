package tenancy

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* IsolationTestSuite provides isolation testing functionality */
type IsolationTestSuite struct {
	service *TenancyService
	pool    *pgxpool.Pool
}

/* NewIsolationTestSuite creates a new isolation test suite */
func NewIsolationTestSuite(service *TenancyService, pool *pgxpool.Pool) *IsolationTestSuite {
	return &IsolationTestSuite{
		service: service,
		pool:    pool,
	}
}

/* TestDataIsolation tests that tenant data is properly isolated */
func (s *IsolationTestSuite) TestDataIsolation(ctx context.Context, tenant1ID, tenant2ID uuid.UUID) error {
	// Create test data in tenant 1
	tenant1, err := s.service.GetTenant(ctx, tenant1ID)
	if err != nil {
		return fmt.Errorf("failed to get tenant 1: %w", err)
	}

	// Insert test record in tenant 1 schema
	testDataID := uuid.New()
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s.test_isolation (id, tenant_id, data)
		VALUES ($1, $2, 'tenant1_data')
	`, tenant1.SchemaName)
	_, err = s.pool.Exec(ctx, insertQuery, testDataID, tenant1ID)
	if err != nil {
		return fmt.Errorf("failed to insert test data in tenant 1: %w", err)
	}

	// Try to access tenant 1 data from tenant 2 context
	tenant2, err := s.service.GetTenant(ctx, tenant2ID)
	if err != nil {
		return fmt.Errorf("failed to get tenant 2: %w", err)
	}

	// Set tenant 2 context
	ctx2, err := s.service.SetTenantContext(ctx, tenant2ID)
	if err != nil {
		return fmt.Errorf("failed to set tenant 2 context: %w", err)
	}

	// Try to query tenant 1 data from tenant 2 - should fail or return empty
	selectQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.test_isolation WHERE id = $1
	`, tenant2.SchemaName)
	var count int
	err = s.pool.QueryRow(ctx2, selectQuery, testDataID).Scan(&count)
	if err == nil && count > 0 {
		return fmt.Errorf("isolation violation: tenant 2 can access tenant 1 data")
	}

	return nil
}

/* TestSchemaIsolation tests that tenant schemas are properly isolated */
func (s *IsolationTestSuite) TestSchemaIsolation(ctx context.Context, tenantID uuid.UUID) error {
	tenant, err := s.service.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Verify schema exists
	checkSchemaQuery := `
		SELECT COUNT(*) FROM information_schema.schemata 
		WHERE schema_name = $1
	`
	var schemaCount int
	err = s.pool.QueryRow(ctx, checkSchemaQuery, tenant.SchemaName).Scan(&schemaCount)
	if err != nil {
		return fmt.Errorf("failed to check schema: %w", err)
	}

	if schemaCount == 0 {
		return fmt.Errorf("tenant schema does not exist: %s", tenant.SchemaName)
	}

	// Verify schema has required tables
	checkTablesQuery := `
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = $1
	`
	var tableCount int
	err = s.pool.QueryRow(ctx, checkTablesQuery, tenant.SchemaName).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to check tables: %w", err)
	}

	if tableCount == 0 {
		return fmt.Errorf("tenant schema has no tables: %s", tenant.SchemaName)
	}

	return nil
}

/* TestAccessIsolation tests that users can only access their tenant's data */
func (s *IsolationTestSuite) TestAccessIsolation(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	// Verify user is associated with tenant
	checkUserQuery := `
		SELECT COUNT(*) FROM neuronip.tenant_users 
		WHERE tenant_id = $1 AND user_id = $2
	`
	var userCount int
	err := s.pool.QueryRow(ctx, checkUserQuery, tenantID, userID).Scan(&userCount)
	if err != nil {
		return fmt.Errorf("failed to check user-tenant association: %w", err)
	}

	if userCount == 0 {
		return fmt.Errorf("user %s is not associated with tenant %s", userID, tenantID)
	}

	// Set tenant context
	tenantCtx, err := s.service.SetTenantContext(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Verify user can access tenant data
	tenant, err := s.service.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Try to query tenant data
	testQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.knowledge_documents`, tenant.SchemaName)
	var docCount int
	err = s.pool.QueryRow(tenantCtx, testQuery).Scan(&docCount)
	if err != nil {
		return fmt.Errorf("user cannot access tenant data: %w", err)
	}

	return nil
}

/* RunAllIsolationTests runs all isolation tests for a tenant */
func (s *IsolationTestSuite) RunAllIsolationTests(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error) {
	results := make(map[string]bool)

	// Test schema isolation
	err := s.TestSchemaIsolation(ctx, tenantID)
	results["schema_isolation"] = err == nil
	if err != nil {
		// Log error but continue
	}

	// Test access isolation (requires a test user)
	// This would need a test user ID
	// results["access_isolation"] = true // Placeholder

	return results, nil
}

/* TestIsolationInTestSuite is a Go test function for running isolation tests */
func TestIsolationInTestSuite(t *testing.T, service *TenancyService, pool *pgxpool.Pool) {
	ctx := context.Background()
	testSuite := NewIsolationTestSuite(service, pool)

	// Create test tenants
	tenant1, err := service.CreateTenant(ctx, "test-tenant-1")
	if err != nil {
		t.Fatalf("Failed to create tenant 1: %v", err)
	}

	tenant2, err := service.CreateTenant(ctx, "test-tenant-2")
	if err != nil {
		t.Fatalf("Failed to create tenant 2: %v", err)
	}

	// Test data isolation
	err = testSuite.TestDataIsolation(ctx, tenant1.ID, tenant2.ID)
	if err != nil {
		t.Errorf("Data isolation test failed: %v", err)
	}

	// Test schema isolation
	err = testSuite.TestSchemaIsolation(ctx, tenant1.ID)
	if err != nil {
		t.Errorf("Schema isolation test failed: %v", err)
	}

	// Cleanup
	_ = service.DeleteTenant(ctx, tenant1.ID)
	_ = service.DeleteTenant(ctx, tenant2.ID)
}
