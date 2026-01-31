package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* RedshiftConnector implements the Connector interface for Amazon Redshift */
type RedshiftConnector struct {
	*ingestion.BaseConnector
	db     *sql.DB
	config map[string]interface{}
}

/* NewRedshiftConnector creates a new Redshift connector */
func NewRedshiftConnector() *RedshiftConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:         "redshift",
		Name:         "Amazon Redshift",
		Description:  "Amazon Redshift data warehouse connector",
		Version:      "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "query", "copy"},
	}

	base := ingestion.NewBaseConnector("redshift", metadata)

	return &RedshiftConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Redshift */
func (r *RedshiftConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	r.config = config

	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)

	if host == "" || user == "" || password == "" || database == "" {
		return fmt.Errorf("host, user, password, and database are required for Redshift")
	}

	if port == 0 {
		port = 5439 // Default Redshift port
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		user, password, host, int(port), database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to Redshift: %w", err)
	}

	r.db = db
	r.BaseConnector.SetConnected(true)

	// Test connection
	if err := r.TestConnection(ctx); err != nil {
		db.Close()
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

/* Disconnect closes the connection */
func (r *RedshiftConnector) Disconnect(ctx context.Context) error {
	if r.db != nil {
		err := r.db.Close()
		r.db = nil
		r.BaseConnector.SetConnected(false)
		return err
	}
	return nil
}

/* TestConnection tests if the connection is valid */
func (r *RedshiftConnector) TestConnection(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "SELECT 1")
	return err
}

/* DiscoverSchema discovers the schema of the Redshift database */
func (r *RedshiftConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if r.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	schema := &ingestion.Schema{
		Tables:      []ingestion.TableSchema{},
		LastUpdated: time.Now(),
	}

	// Query information schema for tables
	query := `
		SELECT schemaname, tablename
		FROM pg_tables
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY schemaname, tablename`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName string
		if err := rows.Scan(&schemaName, &tableName); err != nil {
			continue
		}

		// Get columns for this table
		columns, err := r.getColumns(ctx, schemaName, tableName)
		if err != nil {
			continue
		}

		table := ingestion.TableSchema{
			Name:    fmt.Sprintf("%s.%s", schemaName, tableName),
			Columns: columns,
			Metadata: map[string]interface{}{
				"schema": schemaName,
				"table":  tableName,
			},
		}

		schema.Tables = append(schema.Tables, table)
	}

	return schema, nil
}

/* getColumns gets columns for a table */
func (r *RedshiftConnector) getColumns(ctx context.Context, schema, table string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := r.db.QueryContext(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ingestion.ColumnSchema
	for rows.Next() {
		var col ingestion.ColumnSchema
		var nullable, maxLength sql.NullString
		var defaultVal sql.NullString

		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &defaultVal, &maxLength); err != nil {
			continue
		}

		col.Nullable = nullable.String == "YES"
		if defaultVal.Valid {
			col.DefaultValue = &defaultVal.String
		}

		columns = append(columns, col)
	}

	return columns, nil
}

/* Sync performs a full or incremental sync, counting rows for each table */
func (r *RedshiftConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if r.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := r.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, tableName := range tables {
		parts := strings.SplitN(tableName, ".", 2)
		var countQuery string
		if len(parts) == 2 {
			countQuery = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s"`, parts[0], parts[1])
			if options.Since != nil {
				countQuery = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"."%s" WHERE updated_at >= $1`, parts[0], parts[1])
			}
		} else {
			countQuery = fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
			if options.Since != nil {
				countQuery = fmt.Sprintf(`SELECT COUNT(*) FROM "%s" WHERE updated_at >= $1`, tableName)
			}
		}

		var count int64
		var scanErr error
		if options.Since != nil {
			scanErr = r.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			scanErr = r.db.QueryRowContext(ctx, countQuery).Scan(&count)
		}

		if scanErr != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: scanErr.Error(),
			})
			continue
		}

		result.RowsSynced += count
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

/* GetConnectorType returns the type identifier */
func (r *RedshiftConnector) GetConnectorType() string {
	return "redshift"
}

/* GetMetadata returns connector-specific metadata */
func (r *RedshiftConnector) GetMetadata() ingestion.ConnectorMetadata {
	return r.BaseConnector.GetMetadata()
}
