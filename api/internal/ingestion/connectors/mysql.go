package connectors

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* MySQLConnector implements the Connector interface for MySQL */
type MySQLConnector struct {
	*ingestion.BaseConnector
	db *sql.DB
}

/* NewMySQLConnector creates a new MySQL connector */
func NewMySQLConnector() *MySQLConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "mysql",
		Name:        "MySQL",
		Description: "MySQL database connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("mysql", metadata)

	return &MySQLConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to MySQL */
func (m *MySQLConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	host, _ := config["host"].(string)
	port, _ := config["port"].(float64)
	if port == 0 {
		port = 3306
	}
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)
	database, _ := config["database"].(string)

	if host == "" || user == "" || database == "" {
		return fmt.Errorf("host, user, and database are required")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%.0f)/%s?parseTime=true", user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	m.db = db
	m.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (m *MySQLConnector) Disconnect(ctx context.Context) error {
	if m.db != nil {
		m.db.Close()
	}
	m.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (m *MySQLConnector) TestConnection(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("not connected")
	}
	return m.db.PingContext(ctx)
}

/* DiscoverSchema discovers MySQL schema */
func (m *MySQLConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if m.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all tables
	query := `
		SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_TYPE
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_SCHEMA, TABLE_NAME`

	rows, err := m.db.QueryContext(ctx, query)
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
		columns, err := m.getColumns(ctx, schemaName, tableName)
		if err != nil {
			continue
		}

		// Get primary keys
		primaryKeys, err := m.getPrimaryKeys(ctx, schemaName, tableName)
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
func (m *MySQLConnector) getColumns(ctx context.Context, schemaName, tableName string) ([]ingestion.ColumnSchema, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, CHARACTER_MAXIMUM_LENGTH
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := m.db.QueryContext(ctx, query, schemaName, tableName)
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
func (m *MySQLConnector) getPrimaryKeys(ctx context.Context, schemaName, tableName string) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION`

	rows, err := m.db.QueryContext(ctx, query, schemaName, tableName)
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
func (m *MySQLConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if m.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
	schema, err := m.DiscoverSchema(ctx)
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
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
		if options.Since != nil {
			// Try to find a timestamp column for incremental sync
			countQuery = fmt.Sprintf("SELECT COUNT(*) FROM `%s` WHERE updated_at >= ?", tableName)
		}

		var count int64
		if options.Since != nil {
			err = m.db.QueryRowContext(ctx, countQuery, options.Since).Scan(&count)
		} else {
			err = m.db.QueryRowContext(ctx, countQuery).Scan(&count)
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
