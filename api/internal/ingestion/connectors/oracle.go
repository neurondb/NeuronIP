package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/godror/godror"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* OracleConnector implements the Connector interface for Oracle Database */
type OracleConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewOracleConnector creates a new Oracle connector */
func NewOracleConnector() *OracleConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "oracle",
		Name:        "Oracle Database",
		Description: "Oracle Database connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("oracle", metadata)

	return &OracleConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Oracle */
func (o *OracleConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	if port == 0 {
		port = 1521
	}
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	serviceName, _ := config["service_name"].(string)
	sid, _ := config["sid"].(string)

	if host == "" || user == "" {
		return fmt.Errorf("host and user are required")
	}

	var dsn string
	if serviceName != "" {
		dsn = fmt.Sprintf("%s/%s@%s:%.0f/%s", user, password, host, port, serviceName)
	} else if sid != "" {
		dsn = fmt.Sprintf("%s/%s@%s:%.0f:%s", user, password, host, port, sid)
	} else {
		return fmt.Errorf("service_name or sid is required")
	}

	db, err := sql.Open("godror", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Oracle connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Oracle: %w", err)
	}

	o.db = db
	o.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (o *OracleConnector) Disconnect(ctx context.Context) error {
	if o.db != nil {
		o.db.Close()
	}
	o.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (o *OracleConnector) TestConnection(ctx context.Context) error {
	if o.db == nil {
		return fmt.Errorf("not connected")
	}
	return o.db.PingContext(ctx)
}

/* DiscoverSchema discovers Oracle schema */
func (o *OracleConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if o.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all tables and views
	query := `
		SELECT OWNER, TABLE_NAME, 'TABLE' as TABLE_TYPE
		FROM ALL_TABLES
		WHERE OWNER = USER
		UNION ALL
		SELECT OWNER, VIEW_NAME as TABLE_NAME, 'VIEW' as TABLE_TYPE
		FROM ALL_VIEWS
		WHERE OWNER = USER
		ORDER BY OWNER, TABLE_NAME`

	rows, err := o.db.QueryContext(ctx, query)
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
		columns, err := o.getColumns(ctx, schemaName, tableName)
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
func (o *OracleConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, NULLABLE, DATA_DEFAULT, DATA_LENGTH
		FROM ALL_TAB_COLUMNS
		WHERE OWNER = :1 AND TABLE_NAME = :2
		ORDER BY COLUMN_ID`

	rows, err := o.db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ingestion.ColumnSchema{}
	for rows.Next() {
		var colName, dataType, nullable, defaultValue sql.NullString
		var dataLength sql.NullInt64

		if err := rows.Scan(&colName, &dataType, &nullable, &defaultValue, &dataLength); err != nil {
			continue
		}

		var maxLen *int
		if dataLength.Valid {
			ml := int(dataLength.Int64)
			maxLen = &ml
		}

		var defVal *string
		if defaultValue.Valid {
			defVal = &defaultValue.String
		}

		columns = append(columns, ingestion.ColumnSchema{
			Name:         colName.String,
			DataType:     dataType.String,
			Nullable:     nullable.String == "Y",
			DefaultValue: defVal,
			MaxLength:    maxLen,
		})
	}

	return columns, nil
}

/* Sync performs a sync operation */
func (o *OracleConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if o.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
	schema, err := o.DiscoverSchema(ctx)
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
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= :1", tableName)
		}

		var count int64
		if options.Since != nil {
			err = o.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = o.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
