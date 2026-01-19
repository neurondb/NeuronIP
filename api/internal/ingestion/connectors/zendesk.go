package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* ZendeskConnector implements the Connector interface for Zendesk */
type ZendeskConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	baseURL    string
	apiToken   string
	apiEmail   string
	subdomain  string
}

/* NewZendeskConnector creates a new Zendesk connector */
func NewZendeskConnector() *ZendeskConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "zendesk",
		Name:        "Zendesk",
		Description: "Zendesk API connector for tickets, users, and organizations",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery"},
	}
	
	base := ingestion.NewBaseConnector("zendesk", metadata)
	
	return &ZendeskConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* Connect establishes connection to Zendesk */
func (z *ZendeskConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	subdomain, ok := config["subdomain"].(string)
	if !ok {
		return fmt.Errorf("subdomain is required")
	}
	
	apiEmail, ok := config["api_email"].(string)
	if !ok {
		return fmt.Errorf("api_email is required")
	}
	
	apiToken, ok := config["api_token"].(string)
	if !ok {
		return fmt.Errorf("api_token is required")
	}
	
	z.subdomain = subdomain
	z.apiEmail = apiEmail
	z.apiToken = apiToken
	z.baseURL = fmt.Sprintf("https://%s.zendesk.com/api/v2", subdomain)
	
	// Test connection
	if err := z.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	z.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (z *ZendeskConnector) Disconnect(ctx context.Context) error {
	z.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (z *ZendeskConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", z.baseURL+"/users/me.json", nil)
	if err != nil {
		return err
	}
	
	z.setAuth(req)
	
	resp, err := z.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}
	
	return nil
}

/* DiscoverSchema discovers Zendesk schema */
func (z *ZendeskConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	schema := &ingestion.Schema{
		Tables: []ingestion.TableSchema{
			{
				Name: "tickets",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "bigint", Nullable: false},
					{Name: "url", DataType: "text", Nullable: true},
					{Name: "external_id", DataType: "text", Nullable: true},
					{Name: "created_at", DataType: "timestamp", Nullable: false},
					{Name: "updated_at", DataType: "timestamp", Nullable: false},
					{Name: "type", DataType: "text", Nullable: true},
					{Name: "subject", DataType: "text", Nullable: true},
					{Name: "description", DataType: "text", Nullable: true},
					{Name: "priority", DataType: "text", Nullable: true},
					{Name: "status", DataType: "text", Nullable: false},
					{Name: "requester_id", DataType: "bigint", Nullable: true},
					{Name: "submitter_id", DataType: "bigint", Nullable: true},
					{Name: "assignee_id", DataType: "bigint", Nullable: true},
					{Name: "organization_id", DataType: "bigint", Nullable: true},
					{Name: "group_id", DataType: "bigint", Nullable: true},
					{Name: "tags", DataType: "text[]", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
			{
				Name: "users",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "bigint", Nullable: false},
					{Name: "url", DataType: "text", Nullable: true},
					{Name: "external_id", DataType: "text", Nullable: true},
					{Name: "name", DataType: "text", Nullable: false},
					{Name: "email", DataType: "text", Nullable: true},
					{Name: "created_at", DataType: "timestamp", Nullable: false},
					{Name: "updated_at", DataType: "timestamp", Nullable: false},
					{Name: "time_zone", DataType: "text", Nullable: true},
					{Name: "phone", DataType: "text", Nullable: true},
					{Name: "role", DataType: "text", Nullable: false},
					{Name: "organization_id", DataType: "bigint", Nullable: true},
					{Name: "tags", DataType: "text[]", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
			{
				Name: "organizations",
				Columns: []ingestion.ColumnSchema{
					{Name: "id", DataType: "bigint", Nullable: false},
					{Name: "url", DataType: "text", Nullable: true},
					{Name: "external_id", DataType: "text", Nullable: true},
					{Name: "name", DataType: "text", Nullable: false},
					{Name: "created_at", DataType: "timestamp", Nullable: false},
					{Name: "updated_at", DataType: "timestamp", Nullable: false},
					{Name: "domain_names", DataType: "text[]", Nullable: true},
					{Name: "tags", DataType: "text[]", Nullable: true},
				},
				PrimaryKeys: []string{"id"},
			},
		},
		LastUpdated: time.Now(),
	}
	
	return schema, nil
}

/* Sync performs a sync operation */
func (z *ZendeskConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}
	
	tables := options.Tables
	if len(tables) == 0 {
		tables = []string{"tickets", "users", "organizations"}
	}
	
	for _, table := range tables {
		rows, err := z.syncTable(ctx, table, options)
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
func (z *ZendeskConnector) syncTable(ctx context.Context, table string, options ingestion.SyncOptions) (int64, error) {
	var rows int64
	
	switch table {
	case "tickets":
		return z.syncTickets(ctx, options)
	case "users":
		return z.syncUsers(ctx, options)
	case "organizations":
		return z.syncOrganizations(ctx, options)
	default:
		return 0, fmt.Errorf("unknown table: %s", table)
	}
	
	return rows, nil
}

/* syncTickets syncs tickets from Zendesk */
func (z *ZendeskConnector) syncTickets(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	url := z.baseURL + "/tickets.json"
	if options.Since != nil {
		url += fmt.Sprintf("?start_time=%d", options.Since.Unix())
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	z.setAuth(req)
	
	resp, err := z.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var data struct {
		Tickets []map[string]interface{} `json:"tickets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	
	return int64(len(data.Tickets)), nil
}

/* syncUsers syncs users from Zendesk */
func (z *ZendeskConnector) syncUsers(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	url := z.baseURL + "/users.json"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	z.setAuth(req)
	
	resp, err := z.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var data struct {
		Users []map[string]interface{} `json:"users"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	
	return int64(len(data.Users)), nil
}

/* syncOrganizations syncs organizations from Zendesk */
func (z *ZendeskConnector) syncOrganizations(ctx context.Context, options ingestion.SyncOptions) (int64, error) {
	url := z.baseURL + "/organizations.json"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	z.setAuth(req)
	
	resp, err := z.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var data struct {
		Organizations []map[string]interface{} `json:"organizations"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	
	return int64(len(data.Organizations)), nil
}

/* setAuth sets authentication headers */
func (z *ZendeskConnector) setAuth(req *http.Request) {
	req.SetBasicAuth(z.apiEmail+"/token", z.apiToken)
	req.Header.Set("Content-Type", "application/json")
}
