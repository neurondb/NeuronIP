package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Optional driver - uncomment when available:
	// _ "github.com/trinodb/trino-go-client/trino"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* TrinoConnector implements the Connector interface for Trino */
type TrinoConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewTrinoConnector creates a new Trino connector */
func NewTrinoConnector() *TrinoConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "trino",
		Name:        "Trino",
		Description: "Trino distributed SQL query engine connector",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("trino", metadata)

	return &TrinoConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Trino */
func (t *TrinoConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	catalog, _ := config["catalog"].(string)
	schema, _ := config["schema"].(string)
	user, _ := config["user"].(string)

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if port == 0 {
		port = 8080
	}
	if catalog == "" {
		catalog = "hive"
	}
	if schema == "" {
		schema = "default"
	}
	if user == "" {
		user = "admin"
	}

	// Build Trino connection string
	// Format: http://host:port?catalog=catalog&schema=schema&user=user
	// Note: This requires the Trino Go client to be installed:
	// go get github.com/trinodb/trino-go-client/trino
	password, _ := config["password"].(string)
	dsn := fmt.Sprintf("http://%s:%.0f?catalog=%s&schema=%s&user=%s", host, port, catalog, schema, user)
	if password != "" {
		dsn += fmt.Sprintf("&password=%s", password)
	}
	
	// Attempt to open connection
	var db *sql.DB
	var err error
	db, err = sql.Open("trino", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Trino connection: Trino driver not available. "+
			"Please install: go get github.com/trinodb/trino-go-client/trino. "+
			"Error: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping Trino server at %s:%.0f: %w. "+
			"Please verify the server is running and accessible, and that the Trino driver is properly installed.", 
			host, port, err)
	}

	t.db = db
	t.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (t *TrinoConnector) Disconnect(ctx context.Context) error {
	if t.db != nil {
		t.db.Close()
	}
	t.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (t *TrinoConnector) TestConnection(ctx context.Context) error {
	if t.db == nil {
		return fmt.Errorf("not connected")
	}
	return t.db.PingContext(ctx)
}

/* DiscoverSchema discovers Trino schema */
func (t *TrinoConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if t.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := "SHOW TABLES"

	rows, err := t.db.QueryContext(ctx, query)
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

		columns, err := t.getColumns(ctx, tableName)
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
func (t *TrinoConnector) getColumns(ctx context.Context, tableName string) ([]ingestion.ColumnSchema, error) {
	query := fmt.Sprintf("DESCRIBE %s", tableName)
	rows, err := t.db.QueryContext(ctx, query)
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
func (t *TrinoConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
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

		var count int64
		err = t.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
