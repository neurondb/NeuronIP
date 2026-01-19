package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SlackConnector implements the Connector interface for Slack */
type SlackConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	apiToken   string
	baseURL    string
}

/* NewSlackConnector creates a new Slack connector */
func NewSlackConnector() *SlackConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "slack",
		Name:        "Slack",
		Description: "Slack API connector for messages, channels, and users",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery"},
	}
	
	base := ingestion.NewBaseConnector("slack", metadata)
	
	return &SlackConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://slack.com/api",
	}
}

/* Connect establishes connection to Slack */
func (s *SlackConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	apiToken, ok := config["api_token"].(string)
	if !ok {
		return fmt.Errorf("api_token is required")
	}
	
	s.apiToken = apiToken
	
	// Test connection
	if err := s.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SlackConnector) Disconnect(ctx context.Context) error {
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SlackConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/auth.test", nil)
	if err != nil {
		return err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	var result struct {
		OK bool `json:"ok"`
		Error string `json:"error,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	
	if !result.OK {
		return fmt.Errorf("authentication failed: %s", result.Error)
	}
	
	return nil
}

/* DiscoverSchema discovers Slack schema */
func (s *SlackConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	schema := &ingestion.Schema{
		Tables: []ingestion.TableSchema{
			{
				Name: "messages",
				Columns: []ingestion.ColumnSchema{
					{Name: "ts", DataType: "text", Nullable: false},
					{Name: "channel", DataType: "text", Nullable: false},
					{Name: "user", DataType: "text", Nullable: true},
					{Name: "text", DataType: "text", Nullable: true},
					{Name: "type", DataType: "text", Nullable: false},
					{Name: "subtype", DataType: "text", Nullable: true},
					{Name: "thread_ts", DataType: "text", Nullable: true},
					{Name: "created_at", DataType: "timestamp", Nullable: false},
				},
				PrimaryKeys: []string{"ts", "channel"},
			},
			{
				Name: "channels",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "text", Nullable: false},
					{Name: "name", DataType: "text", Nullable: false},
					{Name: "created", DataType: "bigint", Nullable: false},
					{Name: "is_archived", DataType: "boolean", Nullable: false},
					{Name: "is_private", DataType: "boolean", Nullable: false},
					{Name: "topic", DataType: "text", Nullable: true},
					{Name: "purpose", DataType: "text", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
			{
				Name: "users",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "text", Nullable: false},
					{Name: "name", DataType: "text", Nullable: false},
					{Name: "real_name", DataType: "text", Nullable: true},
					{Name: "email", DataType: "text", Nullable: true},
					{Name: "is_bot", DataType: "boolean", Nullable: false},
					{Name: "is_deleted", DataType: "boolean", Nullable: false},
					{Name: "tz", DataType: "text", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
		},
		LastUpdated: time.Now(),
	}
	
	return schema, nil
}

/* Sync performs a sync operation */
func (s *SlackConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}
	
	tables := options.Tables
	if len(tables) == 0 {
		tables = []string{"messages", "channels", "users"}
	}
	
	for _, table := range tables {
		rows, err := s.syncTable(ctx, table, options)
		if err != nil {
			result.Errors = append(result.Errors, ingestion.SyncError{
				Table:   table,
				Message: err.Error(),
			})
			continue
		}
		
		result.RowsSynced += rows
		result.TablesSynced = append(result.TablesSynced, table)
	}
	
	result.Duration = time.Since(startTime)
	return result, nil
}

/* syncTable syncs a specific table */
func (s *SlackConnector) syncTable(ctx context.Context, table string, options ingestion.SyncOptions) (int64, error) {
	switch table {
	case "messages":
		return s.syncMessages(ctx, options)
	case "channels":
		return s.syncChannels(ctx, options)
	case "users":
		return s.syncUsers(ctx, options)
	default:
		return 0, fmt.Errorf("unknown table: %s", table)
	}
}

/* syncMessages syncs messages from Slack */
func (s *SlackConnector) syncMessages(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	// Get channels first
	channels, err := s.listChannels(ctx)
	if err != nil {
		return 0, err
	}
	
	var totalRows int64
	for _, channelID := range channels {
		req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/conversations.history", nil)
		if err != nil {
			continue
		}
		
		q := req.URL.Query()
		q.Set("channel", channelID)
		if options.Since != nil {
			q.Set("oldest", fmt.Sprintf("%d", options.Since.Unix()))
		}
		req.URL.RawQuery = q.Encode()
		
		s.setAuth(req)
		
		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		
		var data struct {
			OK       bool                     `json:"ok"`
			Messages []map[string]interface{} `json:"messages"`
		}
		
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		
		if data.OK {
			totalRows += int64(len(data.Messages))
		}
	}
	
	return totalRows, nil
}

/* syncChannels syncs channels from Slack */
func (s *SlackConnector) syncChannels(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	channels, err := s.listChannels(ctx)
	if err != nil {
		return 0, err
	}
	
	return int64(len(channels)), nil
}

/* syncUsers syncs users from Slack */
func (s *SlackConnector) syncUsers(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/users.list", nil)
	if err != nil {
		return 0, err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	var data struct {
		OK    bool                     `json:"ok"`
		Users []map[string]interface{} `json:"members"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	
	if !data.OK {
		return 0, fmt.Errorf("API request failed")
	}
	
	return int64(len(data.Users)), nil
}

/* listChannels lists all channels */
func (s *SlackConnector) listChannels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/conversations.list", nil)
	if err != nil {
		return nil, err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var data struct {
		OK        bool                     `json:"ok"`
		Channels  []map[string]interface{} `json:"channels"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	if !data.OK {
		return nil, fmt.Errorf("API request failed")
	}
	
	channelIDs := make([]string, 0, len(data.Channels))
	for _, ch := range data.Channels {
		if id, ok := ch["id"].(string); ok {
			channelIDs = append(channelIDs, id)
		}
	}
	
	return channelIDs, nil
}

/* setAuth sets authentication headers */
func (s *SlackConnector) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.apiToken)
}
