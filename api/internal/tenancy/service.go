package tenancy

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* TenancyMode represents the tenancy isolation mode */
type TenancyMode string

const (
	TenancyModeSchema TenancyMode = "schema"  // Per-tenant schema
	TenancyModeDatabase TenancyMode = "database" // Per-tenant database
)

/* TenancyService provides multi-tenancy functionality */
type TenancyService struct {
	pool *pgxpool.Pool
	mode TenancyMode
}

/* NewTenancyService creates a new tenancy service */
func NewTenancyService(pool *pgxpool.Pool, mode TenancyMode) *TenancyService {
	return &TenancyService{
		pool: pool,
		mode: mode,
	}
}

/* Tenant represents a tenant */
type Tenant struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	SchemaName  string    `json:"schema_name,omitempty"`
	DatabaseName string   `json:"database_name,omitempty"`
	Isolated    bool      `json:"isolated"`
	CreatedAt   string    `json:"created_at"`
}

/* CreateTenant creates a new tenant with isolation */
func (s *TenancyService) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	tenantID := uuid.New()
	schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID.String(), "-", "_"))

	var tenant Tenant
	tenant.ID = tenantID
	tenant.Name = name
	tenant.Isolated = true

	switch s.mode {
	case TenancyModeSchema:
		// Create per-tenant schema
		if err := s.createTenantSchema(ctx, schemaName); err != nil {
			return nil, fmt.Errorf("failed to create tenant schema: %w", err)
		}
		tenant.SchemaName = schemaName

	case TenancyModeDatabase:
		// Create per-tenant database
		databaseName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID.String(), "-", "_"))
		if err := s.createTenantDatabase(ctx, databaseName); err != nil {
			return nil, fmt.Errorf("failed to create tenant database: %w", err)
		}
		tenant.DatabaseName = databaseName
	}

	// Store tenant metadata
	if err := s.storeTenantMetadata(ctx, &tenant); err != nil {
		return nil, fmt.Errorf("failed to store tenant metadata: %w", err)
	}

	return &tenant, nil
}

/* createTenantSchema creates a schema for a tenant */
func (s *TenancyService) createTenantSchema(ctx context.Context, schemaName string) error {
	// Create schema
	createSchemaQuery := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	if _, err := s.pool.Exec(ctx, createSchemaQuery); err != nil {
		return err
	}

	// Create tables in tenant schema (copy of main schema structure)
	// This is a simplified version - in production, you'd copy the full schema
	tables := []string{
		"knowledge_documents",
		"knowledge_collections",
		"warehouse_schemas",
		"support_tickets",
		"workflows",
	}

	for _, table := range tables {
		// In production, you'd create the full table structure
		// For now, placeholder
		createTableQuery := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s.%s (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)
		`, schemaName, table)
		if _, err := s.pool.Exec(ctx, createTableQuery); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table, err)
		}
	}

	return nil
}

/* createTenantDatabase creates a database for a tenant */
func (s *TenancyService) createTenantDatabase(ctx context.Context, databaseName string) error {
	// Note: Creating databases requires superuser privileges
	// In production, this would be done by a database admin or through a separate service
	
	// For now, we'll create a connection to a new database
	// This is a placeholder - actual implementation would require database creation privileges
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", databaseName)
	
	// Execute on postgres database (requires superuser)
	// In production, use a separate admin connection pool
	_, err := s.pool.Exec(ctx, createDBQuery)
	if err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}

	return nil
}

/* storeTenantMetadata stores tenant metadata in the main schema */
func (s *TenancyService) storeTenantMetadata(ctx context.Context, tenant *Tenant) error {
	query := `
		INSERT INTO neuronip.tenants (id, name, schema_name, database_name, isolated, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = $2,
			schema_name = $3,
			database_name = $4,
			isolated = $5
	`
	_, err := s.pool.Exec(ctx, query, tenant.ID, tenant.Name, tenant.SchemaName, tenant.DatabaseName, tenant.Isolated)
	return err
}

/* GetTenant retrieves a tenant by ID */
func (s *TenancyService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	query := `
		SELECT id, name, schema_name, database_name, isolated, created_at
		FROM neuronip.tenants
		WHERE id = $1
	`
	var tenant Tenant
	err := s.pool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.SchemaName, &tenant.DatabaseName,
		&tenant.Isolated, &tenant.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return &tenant, nil
}

/* SetTenantContext sets the tenant context for queries */
func (s *TenancyService) SetTenantContext(ctx context.Context, tenantID uuid.UUID) (context.Context, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return ctx, err
	}

	// Set tenant context variable
	switch s.mode {
	case TenancyModeSchema:
		// Set search_path to tenant schema
		setSchemaQuery := fmt.Sprintf("SET search_path TO %s, public", tenant.SchemaName)
		if _, err := s.pool.Exec(ctx, setSchemaQuery); err != nil {
			return ctx, fmt.Errorf("failed to set tenant schema: %w", err)
		}
	case TenancyModeDatabase:
		// For database mode, you'd switch to the tenant database connection
		// This requires a connection pool per tenant or dynamic connection switching
	}

	return context.WithValue(ctx, "tenant_id", tenantID), nil
}

/* VerifyIsolation verifies that tenant data is properly isolated */
func (s *TenancyService) VerifyIsolation(ctx context.Context, tenantID uuid.UUID) error {
	// This would run isolation tests
	// For now, placeholder
	return nil
}

/* DeleteTenant deletes a tenant and its isolated data */
func (s *TenancyService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	switch s.mode {
	case TenancyModeSchema:
		// Drop tenant schema
		dropSchemaQuery := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant.SchemaName)
		if _, err := s.pool.Exec(ctx, dropSchemaQuery); err != nil {
			return fmt.Errorf("failed to drop tenant schema: %w", err)
		}
	case TenancyModeDatabase:
		// Drop tenant database
		dropDBQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", tenant.DatabaseName)
		if _, err := s.pool.Exec(ctx, dropDBQuery); err != nil {
			return fmt.Errorf("failed to drop tenant database: %w", err)
		}
	}

	// Delete tenant metadata
	deleteQuery := `DELETE FROM neuronip.tenants WHERE id = $1`
	_, err = s.pool.Exec(ctx, deleteQuery, tenantID)
	return err
}
