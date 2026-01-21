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

/* LookerConnector implements the Connector interface for Looker */
type LookerConnector struct {
	*ingestion.BaseConnector
	client      *http.Client
	baseURL     string
	accessToken string
	clientID    string
	clientSecret string
}

/* NewLookerConnector creates a new Looker connector */
func NewLookerConnector() *LookerConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "looker",
		Name:        "Looker",
		Description: "Looker API connector for explores, dashboards, and looks",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "lineage_extraction"},
	}

	base := ingestion.NewBaseConnector("looker", metadata)

	return &LookerConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* Connect establishes connection to Looker */
func (l *LookerConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	baseURL, _ := config["base_url"].(string)
	clientID, _ := config["client_id"].(string)
	clientSecret, _ := config["client_secret"].(string)
	accessToken, _ := config["access_token"].(string)

	if baseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	l.baseURL = baseURL
	l.clientID = clientID
	l.clientSecret = clientSecret

	if accessToken != "" {
		l.accessToken = accessToken
	} else if clientID != "" && clientSecret != "" {
		// Authenticate to get access token
		authURL := fmt.Sprintf("%s/api/4.0/login", baseURL)
		payload := map[string]string{
			"client_id":     clientID,
			"client_secret": clientSecret,
		}

		jsonData, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create auth request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := l.client.Do(req)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		defer resp.Body.Close()

		var authResponse struct {
			AccessToken string `json:"access_token"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
			return fmt.Errorf("failed to decode auth response: %w", err)
		}

		l.accessToken = authResponse.AccessToken
	} else {
		return fmt.Errorf("either access_token or client_id/client_secret is required")
	}

	l.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (l *LookerConnector) Disconnect(ctx context.Context) error {
	if l.accessToken != "" && l.baseURL != "" {
		logoutURL := fmt.Sprintf("%s/api/4.0/logout", l.baseURL)
		req, _ := http.NewRequestWithContext(ctx, "DELETE", logoutURL, nil)
		req.Header.Set("Authorization", "token "+l.accessToken)
		l.client.Do(req)
	}

	l.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (l *LookerConnector) TestConnection(ctx context.Context) error {
	if l.accessToken == "" {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/4.0/user", l.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+l.accessToken)

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: %d", resp.StatusCode)
	}

	return nil
}

/* DiscoverSchema discovers Looker explores and dashboards */
func (l *LookerConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if l.accessToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	tables := []ingestion.TableSchema{}

	// Get explores
	exploresURL := fmt.Sprintf("%s/api/4.0/lookml_models", l.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", exploresURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+l.accessToken)

	resp, err := l.client.Do(req)
	if err == nil {
		defer resp.Body.Close()

		var exploresResponse []struct {
			Name     string `json:"name"`
			Project  string `json:"project_name"`
			Explores []struct {
				Name string `json:"name"`
			} `json:"explores"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&exploresResponse); err == nil {
			for _, model := range exploresResponse {
				for _, explore := range model.Explores {
					tables = append(tables, ingestion.TableSchema{
						Name: explore.Name,
						Columns: []ingestion.ColumnSchema{
							{Name: "name", DataType: "string", Nullable: false},
							{Name: "model", DataType: "string", Nullable: false},
						},
						Metadata: map[string]interface{}{
							"model_name": model.Name,
							"project": model.Project,
							"resource_type": "explore",
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

/* Sync performs a sync operation */
func (l *LookerConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if l.accessToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := l.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	tables := options.Tables
	if len(tables) == 0 {
		for _, table := range schema.Tables {
			tables = append(tables, table.Name)
		}
	}

	for _, exploreName := range tables {
		result.TablesSynced = append(result.TablesSynced, exploreName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
