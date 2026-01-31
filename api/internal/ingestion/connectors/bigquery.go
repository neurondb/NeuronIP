package connectors

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	"google.golang.org/api/option"
)

/* BigQueryConnector implements the Connector interface for Google BigQuery */
type BigQueryConnector struct {
	*ingestion.BaseConnector
	client *bigquery.Client
	config map[string]interface{}
}

/* NewBigQueryConnector creates a new BigQuery connector */
func NewBigQueryConnector() *BigQueryConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:         "bigquery",
		Name:         "Google BigQuery",
		Description:  "Google BigQuery data warehouse connector",
		Version:      "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery", "query", "streaming_insert"},
	}

	base := ingestion.NewBaseConnector("bigquery", metadata)

	return &BigQueryConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to BigQuery */
func (b *BigQueryConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	b.config = config

	projectID, _ := config["project_id"].(string)
	credentialsJSON, _ := config["credentials_json"].(string)

	if projectID == "" {
		return fmt.Errorf("project_id is required for BigQuery")
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
	b.BaseConnector.SetConnected(true)

	// Test connection
	if err := b.TestConnection(ctx); err != nil {
		client.Close()
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

/* Disconnect closes the connection */
func (b *BigQueryConnector) Disconnect(ctx context.Context) error {
	if b.client != nil {
		err := b.client.Close()
		b.client = nil
		b.BaseConnector.SetConnected(false)
		return err
	}
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

/* DiscoverSchema discovers the schema of the BigQuery project */
func (b *BigQueryConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if b.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	schema := &ingestion.Schema{
		Tables:      []ingestion.TableSchema{},
		LastUpdated: time.Now(),
	}

	// List datasets
	datasets := b.client.Datasets(ctx)
	for {
		dataset, err := datasets.Next()
		if err != nil {
			break
		}

		// List tables in dataset
		tables := dataset.Tables(ctx)
		for {
			table, err := tables.Next()
			if err != nil {
				break
			}

			// Get table metadata
			meta, err := table.Metadata(ctx)
			if err != nil {
				continue
			}

			// Convert BigQuery schema to our schema
			columns := make([]ingestion.ColumnSchema, 0, len(meta.Schema))
			for _, field := range meta.Schema {
				col := ingestion.ColumnSchema{
					Name:     field.Name,
					DataType: string(field.Type),
					Nullable: !field.Required,
				}
				columns = append(columns, col)
			}

			// Extract table reference from table and dataset
			tableID := table.TableID
			datasetID := dataset.DatasetID
			projectID := dataset.ProjectID
			tableSchema := ingestion.TableSchema{
				Name:    fmt.Sprintf("%s.%s.%s", projectID, datasetID, tableID),
				Columns: columns,
				Metadata: map[string]interface{}{
					"dataset_id": datasetID,
					"table_id":   tableID,
					"project_id": projectID,
					"location":   meta.Location,
				},
			}

			schema.Tables = append(schema.Tables, tableSchema)
		}
	}

	return schema, nil
}

/* Sync performs a full or incremental sync, counting rows for each table */
func (b *BigQueryConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if b.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

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

	for _, tableName := range tables {
		countQuery := fmt.Sprintf("SELECT COUNT(*) as cnt FROM `%s`", tableName)
		if options.Since != nil {
			countQuery = fmt.Sprintf("SELECT COUNT(*) as cnt FROM `%s` WHERE updated_at >= TIMESTAMP('%s')", tableName, options.Since.Format(time.RFC3339))
		}

		q := b.client.Query(countQuery)
		it, err := q.Read(ctx)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   tableName,
				Message: err.Error(),
			})
			continue
		}

		var row []bigquery.Value
		if err := it.Next(&row); err == nil && len(row) > 0 {
			if count, ok := row[0].(int64); ok {
				result.RowsSynced += count
			}
		}

		result.TablesSynced = append(result.TablesSynced, tableName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

/* GetConnectorType returns the type identifier */
func (b *BigQueryConnector) GetConnectorType() string {
	return "bigquery"
}

/* GetMetadata returns connector-specific metadata */
func (b *BigQueryConnector) GetMetadata() ingestion.ConnectorMetadata {
	return b.BaseConnector.GetMetadata()
}
