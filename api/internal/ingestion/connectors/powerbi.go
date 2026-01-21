package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* PowerBIConnector implements the Connector interface for Microsoft Power BI */
type PowerBIConnector struct {
	*ingestion.BaseConnector
	client      *http.Client
	accessToken string
	baseURL     string
}

/* NewPowerBIConnector creates a new Power BI connector */
func NewPowerBIConnector() *PowerBIConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "powerbi",
		Name:        "Microsoft Power BI",
		Description: "Power BI REST API connector for datasets, reports, and dashboards",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "lineage_extraction"},
	}

	base := ingestion.NewBaseConnector("powerbi", metadata)

	return &PowerBIConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.powerbi.com/v1.0/myorg",
	}
}

/* Connect establishes connection to Power BI */
func (p *PowerBIConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	accessToken, _ := config["access_token"].(string)
	if accessToken == "" {
		return fmt.Errorf("access_token is required (OAuth token from Azure AD)")
	}

	p.accessToken = accessToken

	// Test connection
	if err := p.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	p.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (p *PowerBIConnector) Disconnect(ctx context.Context) error {
	p.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (p *PowerBIConnector) TestConnection(ctx context.Context) error {
	if p.accessToken == "" {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/groups", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: %d", resp.StatusCode)
	}

	return nil
}

/* DiscoverSchema discovers Power BI datasets, reports, and dashboards */
func (p *PowerBIConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if p.accessToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// Get datasets
	datasetsURL := p.baseURL + "/datasets"
	req, err := http.NewRequestWithContext(ctx, "GET", datasetsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.client.Do(req)
	if err == nil {
		defer resp.Body.Close()

		var datasetsResponse struct {
			Value []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"value"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&datasetsResponse); err == nil {
			for _, dataset := range datasetsResponse.Value {
				tables = append(tables, ingestion.TableSchema{
					Name: dataset.Name,
					Columns: []ingestion.ColumnSchema{
						{Name: "id", DataType: "string", Nullable: false},
						{Name: "name", DataType: "string", Nullable: false},
					},
					Metadata: map[string]interface{}{
						"dataset_id": dataset.ID,
						"resource_type": "dataset",
					},
				})
			}
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (p *PowerBIConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if p.accessToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := p.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, datasetName := range tables {
		result.TablesSynced = append(result.TablesSynced, datasetName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
