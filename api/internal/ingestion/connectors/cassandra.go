package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* CassandraConnector implements the Connector interface for Apache Cassandra */
type CassandraConnector struct {
	*ingestion.BaseConnector
	session *gocql.Session
}

/* NewCassandraConnector creates a new Cassandra connector */
func NewCassandraConnector() *CassandraConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "cassandra",
		Name:        "Apache Cassandra",
		Description: "Apache Cassandra NoSQL database connector",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("cassandra", metadata)

	return &CassandraConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Cassandra */
func (c *CassandraConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	hosts, _ := config["hosts"].([]interface{})
	keyspace, _ := config["keyspace"].(string)
	user, _ := config["user"].(string)
	password, _ := config["password"].(string)

	if len(hosts) == 0 {
		hosts = []interface{}{"localhost"}
	}

	hostsStr := make([]string, len(hosts))
	for i, host := range hosts {
		hostsStr[i] = host.(string)
	}

	cluster := gocql.NewCluster(hostsStr...)
	if keyspace != "" {
		cluster.Keyspace = keyspace
	}
	if user != "" && password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: user,
			Password: password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create Cassandra session: %w", err)
	}

	c.session = session
	c.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (c *CassandraConnector) Disconnect(ctx context.Context) error {
	if c.session != nil {
		c.session.Close()
	}
	c.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (c *CassandraConnector) TestConnection(ctx context.Context) error {
	if c.session == nil {
		return fmt.Errorf("not connected")
	}
	return c.session.Query("SELECT now() FROM system.local").Exec()
}

/* DiscoverSchema discovers Cassandra schema */
func (c *CassandraConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// Get all keyspaces
	iter := c.session.Query("SELECT keyspace_name FROM system_schema.keyspaces").Iter()
	var keyspaceName string
	for iter.Scan(&keyspaceName) {
		// Get tables in keyspace
		tableIter := c.session.Query(
			"SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?",
			keyspaceName,
		).Iter()

		var tableName string
		for tableIter.Scan(&tableName) {
			columns, err := c.getColumns(ctx, keyspaceName, tableName)
			if err != nil {
				continue
			}

			tables = append(tables, ingestion.TableSchema{
				Name:    fmt.Sprintf("%s.%s", keyspaceName, tableName),
				Columns: columns,
			})
		}
		tableIter.Close()
	}
	iter.Close()

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* getColumns gets columns for a table */
func (c *CassandraConnector) getColumns(ctx context.Context, keyspace, table string) ([]ingestion.ColumnSchema, error) {
	iter := c.session.Query(
		"SELECT column_name, type FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?",
		keyspace, table,
	).Iter()

	columns := []ingestion.ColumnSchema{}
	var colName, colType string
	for iter.Scan(&colName, &colType) {
		columns = append(columns, ingestion.ColumnSchema{
			Name:     colName,
			DataType: colType,
			Nullable: true, // Cassandra columns are generally nullable
		})
	}
	iter.Close()

	return columns, nil
}

/* Sync performs a sync operation */
func (c *CassandraConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if c.session == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := c.DiscoverSchema(ctx)
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
		// Parse keyspace.table
		var keyspace, table string
		for i := len(tableName) - 1; i >= 0; i-- {
			if tableName[i] == '.' {
				keyspace = tableName[:i]
				table = tableName[i+1:]
				break
			}
		}

		if keyspace == "" || table == "" {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: "invalid table name format (expected keyspace.table)",
			})
			continue
		}

		// Get row count (approximate)
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", keyspace, table)
		var count int64
		err = c.session.Query(countQuery).Scan(&count)
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
