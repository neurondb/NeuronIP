package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* RedshiftConnector implements the Connector interface for Amazon Redshift */
type RedshiftConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewRedshiftConnector creates a new Redshift connector */
func NewRedshiftConnector() *RedshiftConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "redshift",
		Name:        "Amazon Redshift",
		Description: "Amazon Redshift data warehouse connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("redshift", metadata)

	return &RedshiftConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Redshift */
func (r *RedshiftConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	if port == 0 {
		port = 5439
	}
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)

	if host == "" || user == "" || database == "" {
		return fmt.Errorf("host, user, and database are required")
	}

	dsn := fmt.Sprintf("host=%s port=%.0f user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Redshift connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Redshift: %w", err)
	}

	r.db = db
	r.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (r *RedshiftConnector) Disconnect(ctx context.Context) error {
	if r.db != nil {
		r.db.Close()
	}
	r.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (r *RedshiftConnector) TestConnection(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("not connected")
	}
	return r.db.PingContext(ctx)
}

/* DiscoverSchema discovers Redshift schema */
func (r *RedshiftConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if r.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all tables and views
	query := `
		SELECT schemaname, tablename, 'BASE TABLE' as table_type
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		UNION ALL
		SELECT schemaname, viewname as tablename, 'VIEW' as table_type
		FROM pg_views
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY schemaname, tablename`

	rows, err := r.db.QueryContext(ctx, query)
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
		columns, err := r.getColumns(ctx, schemaName, tableName)
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
func (r *RedshiftConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := r.db.QueryContext(ctx, query, schemaName, tableName)
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
func (r *RedshiftConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if r.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
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
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		if options.Since != nil {
			// Try to find a timestamp column for incremental sync
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= $1", tableName)
		}

		var count int64
		if options.Since != nil {
			err = r.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = r.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
