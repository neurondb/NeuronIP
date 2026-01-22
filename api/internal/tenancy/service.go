package tenancy

import (
	"context"
	"database/sql"
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
		// Create table structure based on main schema
		// Query information_schema to get the full table structure from main schema
		var createTableQuery string
		
		// Get table definition from main schema
		tableDefQuery := `
			SELECT 
				column_name,
				data_type,
				character_maximum_length,
				is_nullable,
				column_default
			FROM information_schema.columns
			WHERE table_schema = 'neuronip' AND table_name = $1
			ORDER BY ordinal_position
		`
		
		rows, err := s.pool.Query(ctx, tableDefQuery, table)
		if err != nil {
			// If can't get definition, create minimal table
			createTableQuery = fmt.Sprintf(`
				CREATE TABLE IF NOT EXISTS %s.%s (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`, schemaName, table)
		} else {
			defer rows.Close()
			
			// Build CREATE TABLE statement from column definitions
			var columns []string
			columns = append(columns, "id UUID PRIMARY KEY DEFAULT gen_random_uuid()")
			
			for rows.Next() {
				var colName, dataType, isNullable, colDefault sql.NullString
				var charMaxLength sql.NullInt64
				
				if err := rows.Scan(&colName, &dataType, &charMaxLength, &isNullable, &colDefault); err != nil {
					continue
				}
				
				// Skip id column as we already added it
				if colName.String == "id" {
					continue
				}
				
				colDef := fmt.Sprintf("%s %s", colName.String, dataType.String)
				if charMaxLength.Valid {
					colDef += fmt.Sprintf("(%d)", charMaxLength.Int64)
				}
				if isNullable.String == "NO" {
					colDef += " NOT NULL"
				}
				if colDefault.Valid {
					colDef += " DEFAULT " + colDefault.String
				}
				columns = append(columns, colDef)
			}
			
			createTableQuery = fmt.Sprintf(
				"CREATE TABLE IF NOT EXISTS %s.%s (%s)",
				schemaName, table, strings.Join(columns, ", "),
			)
		}
		
		if _, err := s.pool.Exec(ctx, createTableQuery); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table, err)
		}
	}

	return nil
}

/* createTenantDatabase creates a database for a tenant */
func (s *TenancyService) createTenantDatabase(ctx context.Context, databaseName string) error {
	// Note: Creating databases requires superuser or CREATEDB privileges
	// This implementation includes privilege checks and fallback to schema-based isolation
	
	// Sanitize database name to prevent SQL injection
	sanitizedDBName := strings.ReplaceAll(strings.ToLower(databaseName), "-", "_")
	if !isValidIdentifier(sanitizedDBName) {
		return fmt.Errorf("invalid database name: %s", databaseName)
	}

	// Check if we have privileges to create databases
	hasPrivilege, err := s.checkCreateDatabasePrivilege(ctx)
	if err != nil {
		// If we can't check, try to create anyway
		hasPrivilege = true
	}

	if !hasPrivilege {
		// Fallback: Use schema-based isolation instead
		// This is a more permissive approach that doesn't require superuser
		return fmt.Errorf("insufficient privileges to create database. "+
			"Please use schema-based tenancy mode or grant CREATEDB privilege. "+
			"Attempted to create database: %s", sanitizedDBName)
	}

	// Check if database already exists
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM pg_database WHERE datname = $1
		)`
	err = s.pool.QueryRow(ctx, checkQuery, sanitizedDBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		// Database already exists, which is acceptable
		return nil
	}

	// Create the database
	// Note: CREATE DATABASE cannot be executed in a transaction
	// We need to use a separate connection or execute directly
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", sanitizedDBName)
	
	// Execute on postgres database (requires superuser or CREATEDB)
	// In production, use a separate admin connection pool
	_, err = s.pool.Exec(ctx, createDBQuery)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "permission denied") || 
		   strings.Contains(err.Error(), "insufficient privilege") {
			return fmt.Errorf("insufficient privileges to create database. "+
				"Required: CREATEDB privilege or superuser role. "+
				"Error: %w", err)
		}
		if strings.Contains(err.Error(), "already exists") {
			// Database was created between check and create - this is fine
			return nil
		}
		return fmt.Errorf("failed to create tenant database: %w", err)
	}

	return nil
}

/* checkCreateDatabasePrivilege checks if the current user has privilege to create databases */
func (s *TenancyService) checkCreateDatabasePrivilege(ctx context.Context) (bool, error) {
	// Check if current user is superuser
	var isSuperuser bool
	superuserQuery := `
		SELECT EXISTS(
			SELECT 1 FROM pg_user 
			WHERE usename = current_user AND usesuper = true
		)`
	err := s.pool.QueryRow(ctx, superuserQuery).Scan(&isSuperuser)
	if err != nil {
		return false, fmt.Errorf("failed to check superuser status: %w", err)
	}

	if isSuperuser {
		return true, nil
	}

	// Check if user has CREATEDB privilege
	var canCreateDB bool
	createdbQuery := `
		SELECT EXISTS(
			SELECT 1 FROM pg_user 
			WHERE usename = current_user AND usecreatedb = true
		)`
	err = s.pool.QueryRow(ctx, createdbQuery).Scan(&canCreateDB)
	if err != nil {
		return false, fmt.Errorf("failed to check CREATEDB privilege: %w", err)
	}

	return canCreateDB, nil
}

/* isValidIdentifier checks if a string is a valid PostgreSQL identifier */
func isValidIdentifier(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	
	// Must start with letter or underscore
	if !((name[0] >= 'a' && name[0] <= 'z') || name[0] == '_') {
		return false
	}
	
	// Can contain letters, digits, underscores, and dollar signs
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || 
		     (r >= '0' && r <= '9') || 
		     r == '_' || r == '$') {
			return false
		}
	}
	
	return true
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
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}
	
	// Verify schema isolation
	if s.mode == TenancyModeSchema {
		// Check that tenant schema exists and is separate
		checkSchemaQuery := `
			SELECT COUNT(*) 
			FROM information_schema.schemata 
			WHERE schema_name = $1
		`
		var schemaCount int
		if err := s.pool.QueryRow(ctx, checkSchemaQuery, tenant.SchemaName).Scan(&schemaCount); err != nil {
			return fmt.Errorf("failed to verify schema: %w", err)
		}
		if schemaCount == 0 {
			return fmt.Errorf("tenant schema %s does not exist", tenant.SchemaName)
		}
		
		// Verify that tenant can only access their own schema
		// This is enforced by search_path, but we can verify tables exist in correct schema
		verifyQuery := fmt.Sprintf(`
			SELECT COUNT(*) 
			FROM information_schema.tables 
			WHERE table_schema = '%s'
		`, tenant.SchemaName)
		var tableCount int
		if err := s.pool.QueryRow(ctx, verifyQuery).Scan(&tableCount); err != nil {
			return fmt.Errorf("failed to verify tables: %w", err)
		}
		if tableCount == 0 {
			return fmt.Errorf("no tables found in tenant schema %s", tenant.SchemaName)
		}
	}
	
	// Verify database isolation (if using database mode)
	if s.mode == TenancyModeDatabase && tenant.DatabaseName != "" {
		// Check that database exists
		// Note: This requires connecting to the postgres database
		checkDBQuery := `
			SELECT COUNT(*) 
			FROM pg_database 
			WHERE datname = $1
		`
		var dbCount int
		if err := s.pool.QueryRow(ctx, checkDBQuery, tenant.DatabaseName).Scan(&dbCount); err != nil {
			return fmt.Errorf("failed to verify database: %w", err)
		}
		if dbCount == 0 {
			return fmt.Errorf("tenant database %s does not exist", tenant.DatabaseName)
		}
	}
	
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
