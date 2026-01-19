package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* TeamsConnector implements the Connector interface for Microsoft Teams */
type TeamsConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	accessToken string
	baseURL    string
}

/* NewTeamsConnector creates a new Teams connector */
func NewTeamsConnector() *TeamsConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "teams",
		Name:        "Microsoft Teams",
		Description: "Microsoft Teams Graph API connector for messages, channels, and users",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery"},
	}
	
	base := ingestion.NewBaseConnector("teams", metadata)
	
	return &TeamsConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://graph.microsoft.com/v1.0",
	}
}

/* Connect establishes connection to Microsoft Teams */
func (t *TeamsConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	accessToken, ok := config["access_token"].(string)
	if !ok {
		return fmt.Errorf("access_token is required (OAuth token)")
	}
	
	t.accessToken = accessToken
	
	// Test connection
	if err := t.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	t.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (t *TeamsConnector) Disconnect(ctx context.Context) error {
	t.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (t *TeamsConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/me", nil)
	if err != nil {
		return err
	}
	
	t.setAuth(req)
	
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}
	
	return nil
}

/* DiscoverSchema discovers Teams schema */
func (t *TeamsConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	schema := &ingestion.Schema{
		Tables: []ingestion.TableSchema{
			{
				Name: "messages",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "text", Nullable: false},
					{Name: "chat_id", DataType: "text", Nullable: true},
					{Name: "channel_id", DataType: "text", Nullable: true},
					{Name: "from", DataType: "jsonb", Nullable: true},
					{Name: "body", DataType: "jsonb", Nullable: true},
					{Name: "subject", DataType: "text", Nullable: true},
					{Name: "importance", DataType: "text", Nullable: true},
					{Name: "created_datetime", DataType: "timestamp", Nullable: false},
					{Name: "last_modified_datetime", DataType: "timestamp", Nullable: false},
				},
				PrimaryKeys: []string{"id"},
			},
			{
				Name: "channels",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "text", Nullable: false},
					{Name: "display_name", DataType: "text", Nullable: false},
					{Name: "description", DataType: "text", Nullable: true},
					{Name: "created_datetime", DataType: "timestamp", Nullable: false},
					{Name: "is_favorite_by_default", DataType: "boolean", Nullable: true},
					{Name: "email", DataType: "text", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
			{
				Name: "teams",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "text", Nullable: false},
					{Name: "display_name", DataType: "text", Nullable: false},
					{Name: "description", DataType: "text", Nullable: true},
					{Name: "created_datetime", DataType: "timestamp", Nullable: false},
				},
				PrimaryKeys: []string{"id"},
			},
		},
		LastUpdated: time.Now(),
	}
	
	return schema, nil
}

/* Sync performs a sync operation */
func (t *TeamsConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}
	
	tables := options.Tables
	if len(tables) == 0 {
		tables = []string{"messages", "channels", "teams"}
	}
	
	for _, table := range tables {
		rows, err := t.syncTable(ctx, table, options)
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
func (t *TeamsConnector) syncTable(ctx context.Context, table string, options ingestion.SyncOptions) (int64, error) {
	switch table {
	case "messages":
		return t.syncMessages(ctx, options)
	case "channels":
		return t.syncChannels(ctx, options)
	case "teams":
		return t.syncTeams(ctx, options)
	default:
		return 0, fmt.Errorf("unknown table: %s", table)
	}
}

/* syncMessages syncs messages from Teams */
func (t *TeamsConnector) syncMessages(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	// Get teams first
	teams, err := t.listTeams(ctx)
	if err != nil {
		return 0, err
	}
	
	var totalRows int64
	for _, teamID := range teams {
		// Get channels in team
		channels, err := t.listChannels(ctx, teamID)
		if err != nil {
			continue
		}
		
		for _, channelID := range channels {
			req, err := http.NewRequestWithContext(ctx, "GET", 
				t.baseURL+"/teams/"+teamID+"/channels/"+channelID+"/messages", nil)
			if err != nil {
				continue
			}
			
			if options.Since != nil {
				q := req.URL.Query()
				q.Set("$filter", fmt.Sprintf("lastModifiedDateTime ge %s", options.Since.Format(time.RFC3339)))
				req.URL.RawQuery = q.Encode()
			}
			
			t.setAuth(req)
			
			resp, err := t.client.Do(req)
			if err != nil {
				continue
			}
			
			var data struct {
				Value []map[string]interface{} `json:"value"`
			}
			
			json.NewDecoder(resp.Body).Decode(&data)
			resp.Body.Close()
			
			totalRows += int64(len(data.Value))
		}
	}
	
	return totalRows, nil
}

/* syncChannels syncs channels from Teams */
func (t *TeamsConnector) syncChannels(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	teams, err := t.listTeams(ctx)
	if err != nil {
		return 0, err
	}
	
	var totalRows int64
	for _, teamID := range teams {
		channels, err := t.listChannels(ctx, teamID)
		if err != nil {
			continue
		}
		totalRows += int64(len(channels))
	}
	
	return totalRows, nil
}

/* syncTeams syncs teams from Teams */
func (t *TeamsConnector) syncTeams(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	teams, err := t.listTeams(ctx)
	if err != nil {
		return 0, err
	}
	
	return int64(len(teams)), nil
}

/* listTeams lists all teams */
func (t *TeamsConnector) listTeams(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/me/joinedTeams", nil)
	if err != nil {
		return nil, err
	}
	
	t.setAuth(req)
	
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var data struct {
		Value []map[string]interface{} `json:"value"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	teamIDs := make([]string, 0, len(data.Value))
	for _, team := range data.Value {
		if id, ok := team["id"].(string); ok {
			teamIDs = append(teamIDs, id)
		}
	}
	
	return teamIDs, nil
}

/* listChannels lists channels in a team */
func (t *TeamsConnector) listChannels(ctx context.Context, teamID string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL+"/teams/"+teamID+"/channels", nil)
	if err != nil {
		return nil, err
	}
	
	t.setAuth(req)
	
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var data struct {
		Value []map[string]interface{} `json:"value"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	channelIDs := make([]string, 0, len(data.Value))
	for _, channel := range data.Value {
		if id, ok := channel["id"].(string); ok {
			channelIDs = append(channelIDs, id)
		}
	}
	
	return channelIDs, nil
}

/* setAuth sets authentication headers */
func (t *TeamsConnector) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+t.accessToken)
	req.Header.Set("Content-Type", "application/json")
}
