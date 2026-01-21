package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* AzureSQLConnector implements the Connector interface for Azure SQL Database */
type AzureSQLConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewAzureSQLConnector creates a new Azure SQL connector */
func NewAzureSQLConnector() *AzureSQLConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "azuresql",
		Name:        "Azure SQL Database",
		Description: "Azure SQL Database connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("azuresql", metadata)

	return &AzureSQLConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Azure SQL */
func (a *AzureSQLConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	server, _ := config["server"].(string)
	database, _ := config["database"].(string)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)

	if server == "" || database == "" || user == "" {
		return fmt.Errorf("server, database, and user are required")
	}

	dsn := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=true",
		server, user, password, database)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Azure SQL connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Azure SQL: %w", err)
	}

	a.db = db
	a.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (a *AzureSQLConnector) Disconnect(ctx context.Context) error {
	if a.db != nil {
		a.db.Close()
	}
	a.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (a *AzureSQLConnector) TestConnection(ctx context.Context) error {
	if a.db == nil {
		return fmt.Errorf("not connected")
	}
	return a.db.PingContext(ctx)
}

/* DiscoverSchema discovers Azure SQL schema */
func (a *AzureSQLConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if a.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE IN ('BASE TABLE', 'VIEW')
		ORDER BY TABLE_SCHEMA, TABLE_NAME`

	rows, err := a.db.QueryContext(ctx, query)
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

		columns, err := a.getColumns(ctx, schemaName, tableName)
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
func (a *AzureSQLConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = @p1 AND TABLE_NAME = @p2
		ORDER BY ORDINAL_POSITION`

	rows, err := a.db.QueryContext(ctx, query, sql.Named("p1", schemaName), sql.Named("p2", tableName))
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
func (a *AzureSQLConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if a.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := a.DiscoverSchema(ctx)
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
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM [%s] WHERE updated_at >= @p1", tableName)
		}

		var count int64
		if options.Since != nil {
			err = a.db.QueryRowContext(ctx, countQuery, sql.Named("p1", options.Since)).Scan(&count)
		} else {
			err = a.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
