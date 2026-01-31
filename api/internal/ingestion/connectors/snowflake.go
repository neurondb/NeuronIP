package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	_ "github.com/snowflakedb/gosnowflake" // Optional dependency
)

/* SnowflakeConnector implements the Connector interface for Snowflake */
type SnowflakeConnector struct {
	*ingestion.BaseConnector
	db     *sql.DB
	config map[string]interface{}
}

/* NewSnowflakeConnector creates a new Snowflake connector */
func NewSnowflakeConnector() *SnowflakeConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:         "snowflake",
		Name:         "Snowflake",
		Description:  "Snowflake data warehouse connector",
		Version:      "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "query", "bulk_load"},
	}

	base := ingestion.NewBaseConnector("snowflake", metadata)

	return &SnowflakeConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Snowflake */
func (s *SnowflakeConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	s.config = config

	account, _ := config["account"].(string)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)
	schema, _ := config["schema"].(string)
	warehouse, _ := config["warehouse"].(string)
	role, _ := config["role"].(string)

	if account == "" || user == "" || password == "" {
		return fmt.Errorf("account, user, and password are required for Snowflake")
	}

	dsn := fmt.Sprintf("%s:%s@%s/%s", user, password, account, database)
	if schema != "" {
		dsn += fmt.Sprintf("/%s", schema)
	}
	if warehouse != "" {
		dsn += fmt.Sprintf("?warehouse=%s", warehouse)
	}
	if role != "" {
		if warehouse != "" {
			dsn += fmt.Sprintf("&role=%s", role)
		} else {
			dsn += fmt.Sprintf("?role=%s", role)
		}
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to Snowflake: %w", err)
	}

	s.db = db
	s.BaseConnector.SetConnected(true)

	// Test connection
	if err := s.TestConnection(ctx); err != nil {
		s.db.Close()
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

/* Disconnect closes the connection */
func (s *SnowflakeConnector) Disconnect(ctx context.Context) error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		s.BaseConnector.SetConnected(false)
		return err
	}
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SnowflakeConnector) TestConnection(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, "SELECT 1")
	return err
}

/* DiscoverSchema discovers the schema of the Snowflake database */
func (s *SnowflakeConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if s.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	schema := &ingestion.Schema{
		Tables:      []ingestion.TableSchema{},
		LastUpdated: time.Now(),
	}

	// Query information schema for tables
	query := `
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('INFORMATION_SCHEMA')
		ORDER BY table_schema, table_name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableSchema, tableName, tableType string
		if err := rows.Scan(&tableSchema, &tableName, &tableType); err != nil {
			continue
		}

		// Get columns for this table
		columns, err := s.getColumns(ctx, tableSchema, tableName)
		if err != nil {
			continue
		}

		table := ingestion.TableSchema{
			Name:    fmt.Sprintf("%s.%s", tableSchema, tableName),
			Columns: columns,
			Metadata: map[string]interface{}{
				"schema": tableSchema,
				"table":  tableName,
				"type":   tableType,
			},
		}

		schema.Tables = append(schema.Tables, table)
	}

	return schema, nil
}

/* getColumns gets columns for a table */
func (s *SnowflakeConnector) getColumns(ctx context.Context, schema, table string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := s.db.QueryContext(ctx, query, schema, table)
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
func (s *SnowflakeConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := s.DiscoverSchema(ctx)
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
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		if options.Since != nil {
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= ?", tableName)
		}

		var count int64
		var scanErr error
		if options.Since != nil {
			scanErr = s.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			scanErr = s.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
func (s *SnowflakeConnector) GetConnectorType() string {
	return "snowflake"
}

/* GetMetadata returns connector-specific metadata */
func (s *SnowflakeConnector) GetMetadata() ingestion.ConnectorMetadata {
	return s.BaseConnector.GetMetadata()
}
