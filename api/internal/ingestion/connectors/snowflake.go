package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/snowflakedb/gosnowflake"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SnowflakeConnector implements the Connector interface for Snowflake */
type SnowflakeConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewSnowflakeConnector creates a new Snowflake connector */
func NewSnowflakeConnector() *SnowflakeConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "snowflake",
		Name:        "Snowflake",
		Description: "Snowflake data warehouse connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("snowflake", metadata)

	return &SnowflakeConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Snowflake */
func (s *SnowflakeConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	account, _ := config["account"].(string)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)
	warehouse, _ := config["warehouse"].(string)
	schema, _ := config["schema"].(string)
	role, _ := config["role"].(string)

	if account == "" || user == "" || password == "" || database == "" {
		return fmt.Errorf("account, user, password, and database are required")
	}

	dsn := fmt.Sprintf("%s:%s@%s/%s", user, password, account, database)
	if warehouse != "" {
		dsn += fmt.Sprintf("?warehouse=%s", warehouse)
	}
	if schema != "" {
		if warehouse != "" {
			dsn += fmt.Sprintf("&schema=%s", schema)
		} else {
			dsn += fmt.Sprintf("?schema=%s", schema)
		}
	}
	if role != "" {
		if warehouse != "" || schema != "" {
			dsn += fmt.Sprintf("&role=%s", role)
		} else {
			dsn += fmt.Sprintf("?role=%s", role)
		}
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Snowflake connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Snowflake: %w", err)
	}

	s.db = db
	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SnowflakeConnector) Disconnect(ctx context.Context) error {
	if s.db != nil {
		s.db.Close()
	}
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SnowflakeConnector) TestConnection(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("not connected")
	}
	return s.db.PingContext(ctx)
}

/* DiscoverSchema discovers Snowflake schema */
func (s *SnowflakeConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if s.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all tables and views
	query := `
		SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE IN ('BASE TABLE', 'VIEW')
		ORDER BY TABLE_SCHEMA, TABLE_NAME`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	tables := []ingestion.TableSchema{}
	views := []ingestion.ViewSchema{}

	for rows.Next() {
		var schemaName, tableName, tableType string
		if err := rows.Scan(&schemaName, &tableName, &tableType); err != nil {
			continue
		}

		// Get columns for this table
		columns, err := s.getColumns(ctx, schemaName, tableName)
		if err != nil {
			continue
		}

		if tableType == "VIEW" {
			views = append(views, ingestion.ViewSchema{
				Name:    tableName,
				Columns: columns,
			})
		} else {
			tables = append(tables, ingestion.TableSchema{
				Name:    tableName,
				Columns: columns,
			})
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		Views:       views,
		LastUpdated: time.Now(),
	}, nil
}

/* getColumns gets columns for a table */
func (s *SnowflakeConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := s.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ingestion.ColumnSchema{}
	for rows.Next() {
		var colName, dataType, isNullable, defaultValue sql.NullString
		var maxLength sql.NullInt64

		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue, &maxLength); err != nil {
			continue
		}

		var maxLen *int
		if maxLength.Valid {
			ml := int(maxLength.Int64)
			maxLen = &ml
		}

		var defVal *string
		if defaultValue.Valid {
			defVal = &defaultValue.String
		}

		columns = append(columns, ingestion.ColumnSchema{
			Name:         colName.String,
			DataType:     dataType.String,
			Nullable:     isNullable.String == "YES",
			DefaultValue: defVal,
			MaxLength:    maxLen,
		})
	}

	return columns, nil
}

/* Sync performs a sync operation */
func (s *SnowflakeConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
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
			// Try to find a timestamp column for incremental sync
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= ?", tableName)
		}

		var count int64
		if options.Since != nil {
			err = s.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx, countQuery).Scan(&count)
		}

		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: err.Error(),
			})
			continue
		}

		result.RowsSynced += count
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
