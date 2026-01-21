package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* ElasticsearchConnector implements the Connector interface for Elasticsearch */
type ElasticsearchConnector struct {
	*ingestion.BaseConnector
	client *elasticsearch.Client
}

/* NewElasticsearchConnector creates a new Elasticsearch connector */
func NewElasticsearchConnector() *ElasticsearchConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "elasticsearch",
		Name:        "Elasticsearch",
		Description: "Elasticsearch search engine connector for index discovery and data sync",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "full_sync"},
	}

	base := ingestion.NewBaseConnector("elasticsearch", metadata)

	return &ElasticsearchConnector{
		BaseConnector: base,
	}
}

/* Connect establishes connection to Elasticsearch */
func (e *ElasticsearchConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	addresses, _ := config["addresses"].([]interface{})
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)

	if len(addresses) == 0 {
		addresses = []interface{}{"http://localhost:9200"}
	}

	addressesStr := make([]string, len(addresses))
	for i, addr := range addresses {
		addressesStr[i] = addr.(string)
	}

	cfg := elasticsearch.Config{
		Addresses: addressesStr,
	}

	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	res.Body.Close()

	e.client = client
	e.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (e *ElasticsearchConnector) Disconnect(ctx context.Context) error {
	e.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (e *ElasticsearchConnector) TestConnection(ctx context.Context) error {
	if e.client == nil {
		return fmt.Errorf("not connected")
	}

	res, err := e.client.Info()
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

/* DiscoverSchema discovers Elasticsearch schema (indices) */
func (e *ElasticsearchConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if e.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Get all indices
	res, err := e.client.Indices.Get([]string{"_all"})
	if err != nil {
		return nil, fmt.Errorf("failed to get indices: %w", err)
	}
	defer res.Body.Close()

	var indices map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("failed to decode indices: %w", err)
	}

	tables := []ingestion.TableSchema{}

	for indexName, indexData := range indices {
		if strings.HasPrefix(indexName, ".") {
			continue // Skip system indices
		}

		indexMap, ok := indexData.(map[string]interface{})
		if !ok {
			continue
		}

		mappings, ok := indexMap["mappings"].(map[string]interface{})
		if !ok {
			continue
		}

		properties, ok := mappings["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		columns := []ingestion.ColumnSchema{}
		for fieldName, fieldData := range properties {
			fieldMap, ok := fieldData.(map[string]interface{})
			if !ok {
				continue
			}

			dataType := "text"
			if dt, ok := fieldMap["type"].(string); ok {
				dataType = dt
			}

			columns = append(columns, ingestion.ColumnSchema{
				Name:     fieldName,
				DataType: dataType,
				Nullable: true,
			})
		}

		tables = append(tables, ingestion.TableSchema{
			Name:    indexName,
			Columns: columns,
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (e *ElasticsearchConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if e.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	// Get schema to determine indices
	schema, err := e.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, indexName := range tables {
		// Get document count for index
		res, err := e.client.Count(e.client.Count.WithIndex(indexName))
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   indexName,
				Message: err.Error(),
			})
			continue
		}

		var countResp struct {
			Count int64 `json:"count"`
		}
		if err := json.NewDecoder(res.Body).Decode(&countResp); err != nil {
			res.Body.Close()
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   indexName,
				Message: err.Error(),
			})
			continue
		}
		res.Body.Close()

		result.RowsSynced += countResp.Count
		result.TablesSynced = append(result.TablesSynced, indexName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
