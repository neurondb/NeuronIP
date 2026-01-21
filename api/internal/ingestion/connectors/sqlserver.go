package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SQLServerConnector implements the Connector interface for SQL Server */
type SQLServerConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewSQLServerConnector creates a new SQL Server connector */
func NewSQLServerConnector() *SQLServerConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "sqlserver",
		Name:        "SQL Server",
		Description: "Microsoft SQL Server database connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("sqlserver", metadata)

	return &SQLServerConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to SQL Server */
func (s *SQLServerConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	if port == 0 {
		port = 1433
	}
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)

	if host == "" || user == "" || database == "" {
		return fmt.Errorf("host, user, and database are required")
	}

	dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;port=%.0f",
		host, user, password, database, port)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQL Server connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping SQL Server: %w", err)
	}

	s.db = db
	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SQLServerConnector) Disconnect(ctx context.Context) error {
	if s.db != nil {
		s.db.Close()
	}
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SQLServerConnector) TestConnection(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("not connected")
	}
	return s.db.PingContext(ctx)
}

/* DiscoverSchema discovers SQL Server schema */
func (s *SQLServerConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
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

		// Get primary keys
		primaryKeys, err := s.getPrimaryKeys(ctx, schemaName, tableName)
		if err == nil {
			// primaryKeys handled
		}

		if tableType == "VIEW" {
			views = append(views, ingestion.ViewSchema{
				Name:    tableName,
				Columns: columns,
			})
		} else {
			tables = append(tables, ingestion.TableSchema{
				Name:        tableName,
				Columns:     columns,
				PrimaryKeys: primaryKeys,
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
func (s *SQLServerConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2
		ORDER BY ORDINAL_POSITION`

	rows, err := s.db.QueryContext(ctx, query, sql.Named("p1", schemaName), sql.Named("p2", tableName))
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

/* getPrimaryKeys gets primary keys for a table */
func (s *SQLServerConnector) getPrimaryKeys(ctx context.Context, schemaName, tableName string) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2 AND CONSTRAINT_NAME LIKE 'PK_%'
		ORDER BY ORDINAL_POSITION`

	rows, err := s.db.QueryContext(ctx, query, sql.Named("p1", schemaName), sql.Named("p2", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []string{}
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			continue
		}
		keys = append(keys, colName)
	}

	return keys, nil
}

/* Sync performs a sync operation */
func (s *SQLServerConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
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
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM [%s]", tableName)
		if options.Since != nil {
			// Try to find a timestamp column for incremental sync
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM [%s] WHERE updated_at >= @p1", tableName)
		}

		var count int64
		if options.Since != nil {
			err = s.db.QueryRowContext(ctx, countQuery, sql.Named("p1", options.Since)).Scan(&count)
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
