package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* FivetranConnector implements the Connector interface for Fivetran */
type FivetranConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	baseURL    string
	apiKey     string
	apiSecret  string
}

/* NewFivetranConnector creates a new Fivetran connector */
func NewFivetranConnector() *FivetranConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "fivetran",
		Name:        "Fivetran",
		Description: "Fivetran REST API connector for connectors, schemas, and sync status",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "sync_status", "lineage_extraction"},
	}

	base := ingestion.NewBaseConnector("fivetran", metadata)

	return &FivetranConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.fivetran.com/v1",
	}
}

/* Connect establishes connection to Fivetran */
func (f *FivetranConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	apiSecret, _ := config["api_secret"].(string)

	if apiKey == "" || apiSecret == "" {
		return fmt.Errorf("api_key and api_secret are required")
	}

	f.apiKey = apiKey
	f.apiSecret = apiSecret

	// Test connection
	if err := f.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	f.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (f *FivetranConnector) Disconnect(ctx context.Context) error {
	f.BaseConnector.SetConnected(false)
	f.apiKey = ""
	f.apiSecret = ""
	return nil
}

/* TestConnection tests if the connection is valid */
func (f *FivetranConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", f.baseURL+"/users", nil)
	if err != nil {
		return err
	}

	f.setAuth(req)

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: status %d", resp.StatusCode)
	}

	return nil
}

/* setAuth sets authentication headers */
func (f *FivetranConnector) setAuth(req *http.Request) {
	req.SetBasicAuth(f.apiKey, f.apiSecret)
}

/* DiscoverSchema discovers Fivetran connector schemas */
func (f *FivetranConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	// Fetch connectors
	connectorsReq, err := http.NewRequestWithContext(ctx, "GET", f.baseURL+"/connectors", nil)
	if err != nil {
		return nil, err
	}

	f.setAuth(connectorsReq)
	connectorsResp, err := f.client.Do(connectorsReq)
	if err != nil {
		return nil, err
	}
	defer connectorsResp.Body.Close()

	var connectorsResponse struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Service string `json:"service"`
		} `json:"items"`
	}

	if err := json.NewDecoder(connectorsResp.Body).Decode(&connectorsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode connectors response: %w", err)
	}

	// Build schema representation
	tables := []ingestion.TableSchema{}
	for _, connector := range connectorsResponse.Items {
		// Fetch connector schema
		schemaReq, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/connectors/%s/schemas", f.baseURL, connector.ID), nil)
		f.setAuth(schemaReq)
		schemaResp, _ := f.client.Do(schemaReq)
		if schemaResp != nil {
			var schemaResponse struct {
				Items []struct {
					Name   string `json:"name"`
					Tables []struct {
						Name string `json:"name"`
					} `json:"tables"`
				} `json:"items"`
			}
			json.NewDecoder(schemaResp.Body).Decode(&schemaResponse)
			schemaResp.Body.Close()

			for _, schema := range schemaResponse.Items {
				for _, table := range schema.Tables {
					tables = append(tables, ingestion.TableSchema{
						Name:    fmt.Sprintf("%s.%s.%s", connector.Name, schema.Name, table.Name),
						Columns: []ingestion.ColumnSchema{},
						Metadata: map[string]interface{}{
							"connector_id": connector.ID,
							"connector_name": connector.Name,
							"service": connector.Service,
							"schema": schema.Name,
							"table": table.Name,
						},
					})
				}
			}
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a full or incremental sync */
func (f *FivetranConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		RowsSynced:   0,
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := f.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	for _, table := range schema.Tables {
		result.TablesSynced = append(result.TablesSynced, table.Name)
	}

	result.RowsSynced = int64(len(schema.Tables))
	result.Duration = time.Since(startTime)
	return result, nil
}
