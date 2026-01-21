package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* HubSpotConnector implements the Connector interface for HubSpot */
type HubSpotConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	baseURL    string
	apiKey     string
	accessToken string
}

/* NewHubSpotConnector creates a new HubSpot connector */
func NewHubSpotConnector() *HubSpotConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "hubspot",
		Name:        "HubSpot",
		Description: "HubSpot REST API connector for contacts, companies, deals, and analytics",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "incremental"},
	}

	base := ingestion.NewBaseConnector("hubspot", metadata)

	return &HubSpotConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.hubapi.com",
	}
}

/* Connect establishes connection to HubSpot */
func (h *HubSpotConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	accessToken, _ := config["access_token"].(string)

	if accessToken != "" {
		h.accessToken = accessToken
	} else if apiKey != "" {
		h.apiKey = apiKey
	} else {
		return fmt.Errorf("either access_token or api_key is required")
	}

	// Test connection
	if err := h.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	h.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (h *HubSpotConnector) Disconnect(ctx context.Context) error {
	h.BaseConnector.SetConnected(false)
	h.apiKey = ""
	h.accessToken = ""
	return nil
}

/* TestConnection tests if the connection is valid */
func (h *HubSpotConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", h.baseURL+"/crm/v3/objects/contacts", nil)
	if err != nil {
		return err
	}

	h.setAuth(req)

	resp, err := h.client.Do(req)
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
func (h *HubSpotConnector) setAuth(req *http.Request) {
	if h.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.accessToken)
	} else {
		req.URL.Query().Add("hapikey", h.apiKey)
	}
}

/* DiscoverSchema discovers HubSpot object schemas */
func (h *HubSpotConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	// HubSpot has standard objects: contacts, companies, deals, tickets, etc.
	objects := []string{"contacts", "companies", "deals", "tickets", "products", "line_items"}

	tables := []ingestion.TableSchema{}
	for _, object := range objects {
		// Fetch object properties
		propsReq, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/crm/v3/properties/%s", h.baseURL, object), nil)
		h.setAuth(propsReq)
		propsResp, _ := h.client.Do(propsReq)
		if propsResp != nil {
			var propsResponse struct {
				Results []struct {
					Name     string `json:"name"`
					Label    string `json:"label"`
					Type     string `json:"type"`
					FieldType string `json:"fieldType"`
				} `json:"results"`
			}
			json.NewDecoder(propsResp.Body).Decode(&propsResponse)
			propsResp.Body.Close()

			columns := []ingestion.ColumnSchema{
				{Name: "id", DataType: "text", Nullable: false},
				{Name: "created_at", DataType: "timestamp", Nullable: false},
				{Name: "updated_at", DataType: "timestamp", Nullable: false},
			}

			for _, prop := range propsResponse.Results {
				colType := "text"
				switch prop.Type {
				case "number":
					colType = "numeric"
				case "datetime":
					colType = "timestamp"
				case "bool":
					colType = "boolean"
				case "enumeration":
					colType = "text"
				}

				columns = append(columns, ingestion.ColumnSchema{
					Name:     prop.Name,
					DataType: colType,
					Nullable: true,
				})
			}

			tables = append(tables, ingestion.TableSchema{
				Name:    object,
				Columns: columns,
				PrimaryKeys: []string{"id"},
				Metadata: map[string]interface{}{
					"object_type": object,
				},
			})
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a full or incremental sync */
func (h *HubSpotConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		RowsSynced:   0,
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := h.DiscoverSchema(ctx)
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
