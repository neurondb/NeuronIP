package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* TableauConnector implements the Connector interface for Tableau */
type TableauConnector struct {
	*ingestion.BaseConnector
	client      *http.Client
	serverURL   string
	siteID      string
	token       string
}

/* NewTableauConnector creates a new Tableau connector */
func NewTableauConnector() *TableauConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "tableau",
		Name:        "Tableau",
		Description: "Tableau Server REST API connector for workbooks, views, and data sources",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "lineage_extraction"},
	}

	base := ingestion.NewBaseConnector("tableau", metadata)

	return &TableauConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* Connect establishes connection to Tableau Server */
func (t *TableauConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	serverURL, _ := config["server_url"].(string)
	siteID, _ := config["site_id"].(string)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	token, _ := config["token"].(string)

	if serverURL == "" {
		return fmt.Errorf("server_url is required")
	}

	t.serverURL = serverURL
	t.siteID = siteID

	// If token provided, use it; otherwise authenticate with username/password
	if token != "" {
		t.token = token
	} else if username != "" && password != "" {
		// Sign in to get token
		signInURL := fmt.Sprintf("%s/api/3.21/auth/signin", serverURL)
		payload := map[string]interface{}{
			"credentials": map[string]interface{}{
				"name":     username,
				"password": password,
				"site": map[string]interface{}{
					"contentUrl": siteID,
				},
			},
		}

		jsonData, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, "POST", signInURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create sign-in request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := t.client.Do(req)
		if err != nil {
			return fmt.Errorf("sign-in failed: %w", err)
		}
		defer resp.Body.Close()

		var signInResponse struct {
			Credentials struct {
				Token string `json:"token"`
			} `json:"credentials"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&signInResponse); err != nil {
			return fmt.Errorf("failed to decode sign-in response: %w", err)
		}

		t.token = signInResponse.Credentials.Token
	} else {
		return fmt.Errorf("either token or username/password is required")
	}

	t.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (t *TableauConnector) Disconnect(ctx context.Context) error {
	// Sign out
	if t.token != "" && t.serverURL != "" {
		signOutURL := fmt.Sprintf("%s/api/3.21/auth/signout", t.serverURL)
		req, _ := http.NewRequestWithContext(ctx, "POST", signOutURL, nil)
		req.Header.Set("X-Tableau-Auth", t.token)
		t.client.Do(req)
	}

	t.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (t *TableauConnector) TestConnection(ctx context.Context) error {
	if t.token == "" {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/3.21/sites", t.serverURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Tableau-Auth", t.token)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: %d", resp.StatusCode)
	}

	return nil
}

/* DiscoverSchema discovers Tableau workbooks, views, and data sources */
func (t *TableauConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if t.token == "" {
		return nil, fmt.Errorf("not connected")
	}

	sitePath := ""
	if t.siteID != "" {
		sitePath = fmt.Sprintf("/sites/%s", t.siteID)
	}

	// Get workbooks
	workbooksURL := fmt.Sprintf("%s/api/3.21%s/workbooks", t.serverURL, sitePath)
	req, err := http.NewRequestWithContext(ctx, "GET", workbooksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-Tableau-Auth", t.token)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get workbooks: %w", err)
	}
	defer resp.Body.Close()

	var workbooksResponse struct {
		Workbooks struct {
			Workbook []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"workbook"`
		} `json:"workbooks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&workbooksResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tables := []ingestion.TableSchema{}
	for _, workbook := range workbooksResponse.Workbooks.Workbook {
		// Each workbook becomes a "table"
		tables = append(tables, ingestion.TableSchema{
			Name: workbook.Name,
			Columns: []ingestion.ColumnSchema{
				{Name: "id", DataType: "string", Nullable: false},
				{Name: "name", DataType: "string", Nullable: false},
			},
			Metadata: map[string]interface{}{
				"workbook_id": workbook.ID,
				"resource_type": "workbook",
			},
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (t *TableauConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if t.token == "" {
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

	for _, workbookName := range tables {
		result.TablesSynced = append(result.TablesSynced, workbookName)
		// Metadata is synced, not actual data
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
