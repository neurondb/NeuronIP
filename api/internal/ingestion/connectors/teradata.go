package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Optional drivers - uncomment when available:
	// _ "github.com/Teradata/teradata-driver" // Official Teradata driver
	// _ "github.com/alexbrainman/odbc" // For Teradata ODBC driver
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* TeradataConnector implements the Connector interface for Teradata */
type TeradataConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewTeradataConnector creates a new Teradata connector */
func NewTeradataConnector() *TeradataConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "teradata",
		Name:        "Teradata",
		Description: "Teradata data warehouse connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("teradata", metadata)

	return &TeradataConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Teradata */
func (t *TeradataConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	database, _ := config["database"].(string)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	port, _ := config["port"].(float64)

	if host == "" || user == "" || password == "" {
		return fmt.Errorf("host, user, and password are required")
	}

	// Default port for Teradata
	if port == 0 {
		port = 1025
	}

	// Build connection string
	// Teradata connection string format: host:port/database,user,password
	// Note: This requires the Teradata Go driver to be installed:
	// go get github.com/Teradata/teradata-driver
	// Or use ODBC driver with: github.com/alexbrainman/odbc
	dsn := fmt.Sprintf("host=%s;port=%.0f;user=%s;password=%s", host, port, user, password)
	if database != "" {
		dsn += fmt.Sprintf(";database=%s", database)
	}

	// Attempt to open connection
	// The driver name depends on which Teradata driver is installed
	// Common options: "teradata", "odbc" (with Teradata ODBC driver)
	var db *sql.DB
	var err error
	var lastErr error
	
	// Try teradata driver first
	db, err = sql.Open("teradata", dsn)
	if err == nil {
		if pingErr := db.PingContext(ctx); pingErr == nil {
			t.db = db
			t.BaseConnector.SetConnected(true)
			return nil
		}
		lastErr = fmt.Errorf("teradata driver ping failed: %w", pingErr)
		db.Close()
	} else {
		lastErr = err
	}
	
	// If teradata driver not available, try ODBC
	// ODBC DSN format may differ
	odbcDSN := fmt.Sprintf("DSN=Teradata;Host=%s;Port=%.0f;UID=%s;PWD=%s", host, port, user, password)
	if database != "" {
		odbcDSN += fmt.Sprintf(";Database=%s", database)
	}
	
	db, err = sql.Open("odbc", odbcDSN)
	if err == nil {
		if pingErr := db.PingContext(ctx); pingErr == nil {
			t.db = db
			t.BaseConnector.SetConnected(true)
			return nil
		}
		lastErr = fmt.Errorf("odbc driver ping failed: %w", pingErr)
		db.Close()
	} else if lastErr == nil {
		lastErr = err
	}

	// If all drivers failed, return comprehensive error
	return fmt.Errorf("failed to connect to Teradata: no suitable driver available. Tried 'teradata' and 'odbc' drivers. "+
		"To use Teradata connector, please install one of: "+
		"1) Official Teradata driver: go get github.com/Teradata/teradata-driver, or "+
		"2) Configure Teradata ODBC driver and: go get github.com/alexbrainman/odbc. "+
		"Last error: %w", lastErr)

	t.db = db
	t.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (t *TeradataConnector) Disconnect(ctx context.Context) error {
	if t.db != nil {
		t.db.Close()
	}
	t.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (t *TeradataConnector) TestConnection(ctx context.Context) error {
	if t.db == nil {
		return fmt.Errorf("not connected")
	}
	return t.db.PingContext(ctx)
}

/* DiscoverSchema discovers Teradata schema */
func (t *TeradataConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if t.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT DatabaseName, TableName, TableKind
		FROM DBC.TablesV
		WHERE TableKind IN ('T', 'V')
		ORDER BY DatabaseName, TableName`

	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	tables := []ingestion.TableSchema{}
	views := []ingestion.ViewSchema{}

	for rows.Next() {
		var dbName, tableName, tableKind string
		if err := rows.Scan(&dbName, &tableName, &tableKind); err != nil {
			continue
		}

		columns, err := t.getColumns(ctx, dbName, tableName)
		if err != nil {
			continue
		}

		if tableKind == "V" {
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
func (t *TeradataConnector) getColumns(ctx context.Context, dbName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT ColumnName, ColumnType, Nullable, DefaultValue, ColumnLength
		FROM DBC.ColumnsV
		WHERE DatabaseName = ? AND TableName = ?
		ORDER BY ColumnId`

	rows, err := t.db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ingestion.ColumnSchema{}
	for rows.Next() {
		var colName, colType string
		var nullable sql.NullString
		var defaultValue sql.NullString
		var colLength sql.NullInt64

		if err := rows.Scan(&colName, &colType, &nullable, &defaultValue, &colLength); err != nil {
			continue
		}

		var maxLen *int
		if colLength.Valid {
			ml := int(colLength.Int64)
			maxLen = &ml
		}

		var defVal *string
		if defaultValue.Valid {
			defVal = &defaultValue.String
		}

		isNullable := nullable.String == "Y"

		columns = append(columns, ingestion.ColumnSchema{
			Name:         colName,
			DataType:     colType,
			Nullable:     isNullable,
			DefaultValue: defVal,
			MaxLength:    maxLen,
		})
	}

	return columns, nil
}

/* Sync performs a sync operation */
func (t *TeradataConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if t.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := t.DiscoverSchema(ctx)
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
		if options.Since != nil {
			err = t.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = t.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
