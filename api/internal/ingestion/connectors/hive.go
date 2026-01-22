package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	// Optional drivers - uncomment when available:
	// _ "github.com/alexbrainman/odbc" // For Hive ODBC driver
	// _ "github.com/bippio/go-impala" // For Impala/Hive native driver
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* HiveConnector implements the Connector interface for Apache Hive */
type HiveConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewHiveConnector creates a new Hive connector */
func NewHiveConnector() *HiveConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "hive",
		Name:        "Apache Hive",
		Description: "Apache Hive data warehouse connector",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("hive", metadata)

	return &HiveConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Hive */
func (h *HiveConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	database, _ := config["database"].(string)
	user, _ := config["user"].(string)

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if port == 0 {
		port = 10000
	}
	if database == "" {
		database = "default"
	}
	if user == "" {
		user = "hive"
	}

	// Build Hive connection string
	// Hive uses JDBC-style connection strings
	// Format: jdbc:hive2://host:port/database or hive://host:port/database
	// Note: This requires a Hive JDBC driver or Go Hive client
	// Options:
	// 1. Use ODBC with Hive ODBC driver: go get github.com/alexbrainman/odbc
	// 2. Use Hive JDBC via Go JDBC bridge
	// 3. Use native Hive client if available
	dsn := fmt.Sprintf("hive2://%s:%.0f/%s?user=%s", host, port, database, user)
	
	// Attempt to open connection
	// Try different driver names depending on what's available
	var db *sql.DB
	var err error
	var lastErr error
	
	// Try hive driver first
	db, err = sql.Open("hive", dsn)
	if err == nil {
		pingErr := db.PingContext(ctx)
		if pingErr == nil {
			h.db = db
			h.BaseConnector.SetConnected(true)
			return nil
		}
		lastErr = fmt.Errorf("hive driver ping failed: %w", pingErr)
		db.Close()
	} else {
		lastErr = err
	}
	
	// Try ODBC with Hive ODBC driver
	password, _ := config["password"].(string)
	odbcDSN := fmt.Sprintf("DSN=Hive;Host=%s;Port=%.0f;Database=%s;UID=%s", host, port, database, user)
	if password != "" {
		odbcDSN += fmt.Sprintf(";PWD=%s", password)
	}
	
	db, err = sql.Open("odbc", odbcDSN)
	if err == nil {
		if pingErr := db.PingContext(ctx); pingErr == nil {
			h.db = db
			h.BaseConnector.SetConnected(true)
			return nil
		}
		lastErr = fmt.Errorf("odbc driver ping failed: %w", pingErr)
		db.Close()
	} else {
		lastErr = err
	}

	// If all drivers failed, return comprehensive error
	return fmt.Errorf("failed to connect to Hive: no suitable driver available. Tried 'hive' and 'odbc' drivers. "+
		"To use Hive connector, please install one of: "+
		"1) Hive ODBC driver and 'go get github.com/alexbrainman/odbc', "+
		"2) A Hive JDBC bridge, or "+
		"3) A native Hive Go client. "+
		"Last error: %w", lastErr)

	h.db = db
	h.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (h *HiveConnector) Disconnect(ctx context.Context) error {
	if h.db != nil {
		h.db.Close()
	}
	h.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (h *HiveConnector) TestConnection(ctx context.Context) error {
	if h.db == nil {
		return fmt.Errorf("not connected")
	}
	return h.db.PingContext(ctx)
}

/* DiscoverSchema discovers Hive schema */
func (h *HiveConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if h.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := "SHOW TABLES"

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	tables := []ingestion.TableSchema{}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		columns, err := h.getColumns(ctx, tableName)
		if err != nil {
			continue
		}

		tables = append(tables, ingestion.TableSchema{
			Name:    tableName,
			Columns: columns,
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* getColumns gets columns for a table */
func (h *HiveConnector) getColumns(ctx context.Context, tableName string) ([]ingestion.ColumnSchema, error) {
	query := fmt.Sprintf("DESCRIBE %s", tableName)
	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ingestion.ColumnSchema{}
	for rows.Next() {
		var colName, colType, comment sql.NullString
		if err := rows.Scan(&colName, &colType, &comment); err != nil {
			continue
		}

		columns = append(columns, ingestion.ColumnSchema{
			Name:     colName.String,
			DataType: colType.String,
			Nullable: true,
		})
	}

	return columns, nil
}

/* Sync performs a sync operation */
func (h *HiveConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if h.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := h.DiscoverSchema(ctx)
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

		var count int64
		err = h.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
