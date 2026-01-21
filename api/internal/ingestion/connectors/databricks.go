package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/databricks/databricks-sql-go"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* DatabricksConnector implements the Connector interface for Databricks */
type DatabricksConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewDatabricksConnector creates a new Databricks connector */
func NewDatabricksConnector() *DatabricksConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "databricks",
		Name:        "Databricks",
		Description: "Databricks SQL warehouse connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("databricks", metadata)

	return &DatabricksConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Databricks */
func (d *DatabricksConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	serverHostname, _ := config["server_hostname"].(string)
	httpPath, _ := config["http_path"].(string)
	accessToken, _ := config["access_token"].(string)

	if serverHostname == "" || httpPath == "" || accessToken == "" {
		return fmt.Errorf("server_hostname, http_path, and access_token are required")
	}

	dsn := fmt.Sprintf("token:%s@%s:443/%s", accessToken, serverHostname, httpPath)

	db, err := sql.Open("databricks", dsn)
	if err != nil {
		return fmt.Errorf("failed to open Databricks connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping Databricks: %w", err)
	}

	d.db = db
	d.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (d *DatabricksConnector) Disconnect(ctx context.Context) error {
	if d.db != nil {
		d.db.Close()
	}
	d.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (d *DatabricksConnector) TestConnection(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("not connected")
	}
	return d.db.PingContext(ctx)
}

/* DiscoverSchema discovers Databricks schema */
func (d *DatabricksConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all databases
	query := "SHOW DATABASES"
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	tables := []ingestion.TableSchema{}
	views := []ingestion.ViewSchema{}

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}

		// Get tables in database
		tableQuery := fmt.Sprintf("SHOW TABLES IN %s", dbName)
		tableRows, err := d.db.QueryContext(ctx, tableQuery)
		if err != nil {
			continue
		}

		for tableRows.Next() {
			var tableName, isTmp string
			if err := tableRows.Scan(&tableName, &isTmp); err != nil {
				continue
			}

			// Get columns
			columns, err := d.getColumns(ctx, dbName, tableName)
			if err != nil {
				continue
			}

			if isTmp == "true" {
				views = append(views, ingestion.ViewSchema{
					Name:    fmt.Sprintf("%s.%s", dbName, tableName),
					Columns: columns,
				})
			} else {
				tables = append(tables, ingestion.TableSchema{
					Name:    fmt.Sprintf("%s.%s", dbName, tableName),
					Columns: columns,
				})
			}
		}
		tableRows.Close()
	}

	return &ingestion.Schema{
		Tables:      tables,
		Views:       views,
		LastUpdated: time.Now(),
	}, nil
}

/* getColumns gets columns for a table */
func (d *DatabricksConnector) getColumns(ctx context.Context, dbName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := fmt.Sprintf("DESCRIBE %s.%s", dbName, tableName)
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := []ingestion.ColumnSchema{}
	for rows.Next() {
		var colName, dataType, comment sql.NullString
		if err := rows.Scan(&colName, &dataType, &comment); err != nil {
			continue
		}

		columns = append(columns, ingestion.ColumnSchema{
			Name:     colName.String,
			DataType: dataType.String,
			Nullable: true, // Databricks doesn't expose nullability in DESCRIBE
		})
	}

	return columns, nil
}

/* Sync performs a sync operation */
func (d *DatabricksConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
	schema, err := d.DiscoverSchema(ctx)
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
			err = d.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = d.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
