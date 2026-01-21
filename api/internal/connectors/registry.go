package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

/* ConnectorInterface defines the interface for connector implementations */
type ConnectorInterface interface {
	DiscoverSchema(ctx context.Context, connector *DataSourceConnector) (*Schema, error)
	TestConnection(ctx context.Context, connector *DataSourceConnector) error
}

/* ConnectorRegistry manages connector implementations */
type ConnectorRegistry struct {
	connectors map[ConnectorType]ConnectorInterface
}

/* NewConnectorRegistry creates a new connector registry */
func NewConnectorRegistry() *ConnectorRegistry {
	registry := &ConnectorRegistry{
		connectors: make(map[ConnectorType]ConnectorInterface),
	}

	// Register built-in connectors
	registry.Register(ConnectorPostgreSQL, &PostgreSQLConnector{})
	// Add more connectors as needed

	return registry
}

/* Register registers a connector implementation */
func (r *ConnectorRegistry) Register(connectorType ConnectorType, impl ConnectorInterface) {
	r.connectors[connectorType] = impl
}

/* GetConnector gets a connector implementation */
func (r *ConnectorRegistry) GetConnector(connectorType ConnectorType) (ConnectorInterface, error) {
	impl, exists := r.connectors[connectorType]
	if !exists {
		return nil, fmt.Errorf("connector type %s not supported", connectorType)
	}
	return impl, nil
}

/* PostgreSQLConnector implements PostgreSQL connector */
type PostgreSQLConnector struct{}

/* TestConnection tests PostgreSQL connection */
func (c *PostgreSQLConnector) TestConnection(ctx context.Context, connector *DataSourceConnector) error {
	connStr := c.buildConnectionString(connector)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

/* DiscoverSchema discovers PostgreSQL schema */
func (c *PostgreSQLConnector) DiscoverSchema(ctx context.Context, connector *DataSourceConnector) (*Schema, error) {
	connStr := c.buildConnectionString(connector)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	schema := &Schema{Tables: []Table{}}

	// Query to get all tables and views
	query := `
		SELECT 
			table_schema,
			table_name,
			table_type,
			COALESCE(obj_description(c.oid), '') as description
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY table_schema, table_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table Table
		var tableType string
		var description sql.NullString

		err := rows.Scan(&table.SchemaName, &table.TableName, &tableType, &description)
		if err != nil {
			continue
		}

		// Map PostgreSQL table types
		switch tableType {
		case "BASE TABLE":
			table.TableType = "table"
		case "VIEW":
			table.TableType = "view"
		case "MATERIALIZED VIEW":
			table.TableType = "materialized_view"
		default:
			table.TableType = "table"
		}

		if description.Valid {
			table.Description = &description.String
		}

		// Get table statistics
		c.getTableStats(ctx, db, &table)

		// Get columns
		columns, err := c.getColumns(ctx, db, table.SchemaName, table.TableName)
		if err == nil {
			table.Columns = columns
		}

		schema.Tables = append(schema.Tables, table)
	}

	return schema, nil
}

/* getTableStats gets table statistics */
func (c *PostgreSQLConnector) getTableStats(ctx context.Context, db *sql.DB, table *Table) {
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(n_live_tup, 0) as row_count,
			COALESCE(pg_total_relation_size('%s.%s'), 0) as size_bytes,
			COALESCE(relowner::regrole::text, '') as owner
		FROM pg_stat_user_tables
		WHERE schemaname = $1 AND relname = $2`,
		table.SchemaName, table.TableName)

	var rowCount, sizeBytes sql.NullInt64
	var owner sql.NullString
	err := db.QueryRowContext(ctx, query, table.SchemaName, table.TableName).Scan(
		&rowCount, &sizeBytes, &owner)
	if err == nil {
		if rowCount.Valid {
			table.RowCount = &rowCount.Int64
		}
		if sizeBytes.Valid {
			table.SizeBytes = &sizeBytes.Int64
		}
		if owner.Valid {
			table.Owner = &owner.String
		}
	}
}

/* getColumns gets columns for a table */
func (c *PostgreSQLConnector) getColumns(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]Column, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			ordinal_position,
			is_nullable,
			column_default,
			COALESCE(col_description(c.oid, a.attnum), '') as description
		FROM information_schema.columns c
		LEFT JOIN pg_class t ON t.relname = c.table_name
		LEFT JOIN pg_namespace n ON n.oid = t.relnamespace AND n.nspname = c.table_schema
		LEFT JOIN pg_attribute a ON a.attrelid = t.oid AND a.attname = c.column_name
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		var dataType string
		var isNullable string
		var defaultValue, description sql.NullString

		err := rows.Scan(
			&col.ColumnName, &dataType, &col.OrdinalPosition,
			&isNullable, &defaultValue, &description,
		)
		if err != nil {
			continue
		}

		col.ColumnType = dataType
		col.IsNullable = (isNullable == "YES")
		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}
		if description.Valid && description.String != "" {
			col.Description = &description.String
		}

		// Check if primary key
		col.IsPrimaryKey = c.isPrimaryKey(ctx, db, schemaName, tableName, col.ColumnName)
		// Check if foreign key
		col.IsForeignKey = c.isForeignKey(ctx, db, schemaName, tableName, col.ColumnName)

		columns = append(columns, col)
	}

	return columns, nil
}

/* isPrimaryKey checks if column is primary key */
func (c *PostgreSQLConnector) isPrimaryKey(ctx context.Context, db *sql.DB, schema, table, column string) bool {
	query := `
		SELECT COUNT(*) FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = $1 
			AND tc.table_name = $2
			AND kcu.column_name = $3
			AND tc.constraint_type = 'PRIMARY KEY'`

	var count int
	err := db.QueryRowContext(ctx, query, schema, table, column).Scan(&count)
	return err == nil && count > 0
}

/* isForeignKey checks if column is foreign key */
func (c *PostgreSQLConnector) isForeignKey(ctx context.Context, db *sql.DB, schema, table, column string) bool {
	query := `
		SELECT COUNT(*) FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = $1 
			AND tc.table_name = $2
			AND kcu.column_name = $3
			AND tc.constraint_type = 'FOREIGN KEY'`

	var count int
	err := db.QueryRowContext(ctx, query, schema, table, column).Scan(&count)
	return err == nil && count > 0
}

/* buildConnectionString builds PostgreSQL connection string */
func (c *PostgreSQLConnector) buildConnectionString(connector *DataSourceConnector) string {
	if connector.ConnectionString != nil {
		return *connector.ConnectionString
	}

	// Build from configuration
	host, _ := connector.Configuration["host"].(string)
	port, _ := connector.Configuration["port"].(string)
	user, _ := connector.Configuration["user"].(string)
	password, _ := connector.Configuration["password"].(string)
	database, _ := connector.Configuration["database"].(string)
	sslmode, _ := connector.Configuration["sslmode"].(string)

	if sslmode == "" {
		sslmode = "disable"
	}

	parts := []string{
		fmt.Sprintf("host=%s", host),
		fmt.Sprintf("port=%s", port),
		fmt.Sprintf("user=%s", user),
		fmt.Sprintf("password=%s", password),
		fmt.Sprintf("dbname=%s", database),
		fmt.Sprintf("sslmode=%s", sslmode),
	}

	return strings.Join(parts, " ")
}
