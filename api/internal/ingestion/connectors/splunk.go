package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SplunkConnector implements the Connector interface for Splunk */
type SplunkConnector struct {
	*ingestion.BaseConnector
	client      *http.Client
	baseURL     string
	sessionKey  string
}

/* NewSplunkConnector creates a new Splunk connector */
func NewSplunkConnector() *SplunkConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "splunk",
		Name:        "Splunk",
		Description: "Splunk Enterprise connector for indexes and searches",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "query_log_analysis"},
	}

	base := ingestion.NewBaseConnector("splunk", metadata)

	return &SplunkConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* Connect establishes connection to Splunk */
func (s *SplunkConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	baseURL, _ := config["base_url"].(string)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)

	if baseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	s.baseURL = baseURL

	// Login to get session key
	loginURL := fmt.Sprintf("%s/services/auth/login", baseURL)
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.SetBasicAuth(username, password)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Extract session key from response
	// In production, parse XML response to get sessionKey
	s.sessionKey = resp.Header.Get("X-Splunk-Session-Key")
	if s.sessionKey == "" {
		// Fallback: use basic auth header
		s.sessionKey = fmt.Sprintf("Basic %s:%s", username, password)
	}

	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SplunkConnector) Disconnect(ctx context.Context) error {
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SplunkConnector) TestConnection(ctx context.Context) error {
	if s.baseURL == "" {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/services/server/info", s.baseURL), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Splunk "+s.sessionKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: %d", resp.StatusCode)
	}

	return nil
}

/* DiscoverSchema discovers Splunk indexes */
func (s *SplunkConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	if s.baseURL == "" {
		return nil, fmt.Errorf("not connected")
	}

	// List indexes
	indexesURL := fmt.Sprintf("%s/services/data/indexes", s.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", indexesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Splunk "+s.sessionKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	defer resp.Body.Close()

	var indexesResponse struct {
		Entry []struct {
			Name string `json:"name"`
		} `json:"entry"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&indexesResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	tables := []ingestion.TableSchema{}
	for _, entry := range indexesResponse.Entry {
		tables = append(tables, ingestion.TableSchema{
			Name: entry.Name,
			Columns: []ingestion.ColumnSchema{
				{Name: "_time", DataType: "timestamp", Nullable: false},
				{Name: "_raw", DataType: "string", Nullable: true},
				{Name: "host", DataType: "string", Nullable: true},
				{Name: "source", DataType: "string", Nullable: true},
				{Name: "sourcetype", DataType: "string", Nullable: true},
			},
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (s *SplunkConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	if s.baseURL == "" {
		return nil, fmt.Errorf("not connected")
	}

	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := s.DiscoverSchema(ctx)
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
		// Run search to count events
		searchQuery := fmt.Sprintf("search index=%s | stats count", indexName)
		searchURL := fmt.Sprintf("%s/services/search/jobs/export", s.baseURL)

		data := url.Values{}
		data.Set("search", searchQuery)
		data.Set("output_mode", "json")

		req, err := http.NewRequestWithContext(ctx, "POST", searchURL, nil)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   indexName,
				Message: err.Error(),
			})
			continue
		}
		req.Header.Set("Authorization", "Splunk "+s.sessionKey)
		req.URL.RawQuery = data.Encode()

		resp, err := s.client.Do(req)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   indexName,
				Message: err.Error(),
			})
			continue
		}
		defer resp.Body.Close()

		var searchResult struct {
			Result struct {
				Count string `json:"count"`
			} `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&searchResult); err == nil {
			// Parse count if available
			// For now, just mark as synced
		}

		result.TablesSynced = append(result.TablesSynced, indexName)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
