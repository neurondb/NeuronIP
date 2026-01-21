package connectors

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* BigQueryConnector implements the Connector interface for BigQuery */
type BigQueryConnector struct {
	*ingestion.BaseConnector
	client *bigquery.Client
	projectID string
}

/* NewBigQueryConnector creates a new BigQuery connector */
func NewBigQueryConnector() *BigQueryConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "bigquery",
		Name:        "BigQuery",
		Description: "Google BigQuery data warehouse connector for schema discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "full_sync", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("bigquery", metadata)

	return &BigQueryConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to BigQuery */
func (b *BigQueryConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	projectID, _ := config["project_id"].(string)
	credentialsJSON, _ := config["credentials_json"].(string)

	if projectID == "" {
		return fmt.Errorf("project_id is required")
	}

	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return fmt.Errorf("failed to create BigQuery client: %w", err)
	}

	b.client = client
	b.projectID = projectID
	b.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (b *BigQueryConnector) Disconnect(ctx context.Context) error {
	if b.client != nil {
		b.client.Close()
	}
	b.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (b *BigQueryConnector) TestConnection(ctx context.Context) error {
	if b.client == nil {
		return fmt.Errorf("not connected")
	}

	// Try to list datasets
	it := b.client.Datasets(ctx)
	_, err := it.Next()
	return err
}

/* DiscoverSchema discovers BigQuery schema */
func (b *BigQueryConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if b.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}
	views := []ingestion.ViewSchema{}

	// List all datasets
	datasets := b.client.Datasets(ctx)
	for {
		dataset, err := datasets.Next()
		if err != nil {
			break
		}

		// List tables in dataset
		tablesIter := dataset.Tables(ctx)
		for {
			table, err := tablesIter.Next()
			if err != nil {
				break
			}

			// Get table metadata
			meta, err := table.Metadata(ctx)
			if err != nil {
				continue
			}

			columns := []ingestion.ColumnSchema{}
			for _, field := range meta.Schema {
				columns = append(columns, ingestion.ColumnSchema{
					Name:     field.Name,
					DataType: string(field.Type),
					Nullable: !field.Required,
				})
			}

			if meta.Type == bigquery.ViewTable {
				views = append(views, ingestion.ViewSchema{
					Name:    table.TableID,
					Columns: columns,
				})
			} else {
				tables = append(tables, ingestion.TableSchema{
					Name:    table.TableID,
					Columns: columns,
				})
			}
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		Views:       views,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (b *BigQueryConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if b.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine tables
	schema, err := b.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	// For BigQuery, we'll use COUNT queries
	for _, tableName := range tables {
		// Find the dataset for this table
		datasets := b.client.Datasets(ctx)
		var tableRef *bigquery.Table
		for {
			dataset, err := datasets.Next()
			if err != nil {
				break
			}
			table := dataset.Table(tableName)
			if _, err := table.Metadata(ctx); err == nil {
				tableRef = table
				break
			}
		}

		if tableRef == nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: "table not found",
			})
			continue
		}

		// Get row count
		meta, err := tableRef.Metadata(ctx)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: err.Error(),
			})
			continue
		}

		result.RowsSynced += int64(meta.NumRows)
		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
