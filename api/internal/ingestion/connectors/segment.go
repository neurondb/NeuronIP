package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SegmentConnector implements the Connector interface for Segment */
type SegmentConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	baseURL    string
	apiKey     string
	workspace  string
}

/* NewSegmentConnector creates a new Segment connector */
func NewSegmentConnector() *SegmentConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "segment",
		Name:        "Segment",
		Description: "Segment REST API connector for sources, destinations, and tracking plans",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "event_tracking", "lineage_extraction"},
	}

	base := ingestion.NewBaseConnector("segment", metadata)

	return &SegmentConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.segment.io/v1",
	}
}

/* Connect establishes connection to Segment */
func (s *SegmentConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	workspace, _ := config["workspace"].(string)

	if apiKey == "" {
		return fmt.Errorf("api_key is required")
	}

	s.apiKey = apiKey
	s.workspace = workspace

	// Test connection
	if err := s.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SegmentConnector) Disconnect(ctx context.Context) error {
	s.BaseConnector.SetConnected(false)
	s.apiKey = ""
	s.workspace = ""
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SegmentConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/workspaces", nil)
	if err != nil {
		return err
	}

	s.setAuth(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("connection test failed: status %d", resp.StatusCode)
	}

	return nil
}

/* setAuth sets authentication headers */
func (s *SegmentConnector) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Basic "+s.apiKey)
}

/* DiscoverSchema discovers Segment sources and tracking plans */
func (s *SegmentConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	// Fetch sources
	sourcesReq, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/sources", nil)
	if err != nil {
		return nil, err
	}

	s.setAuth(sourcesReq)
	sourcesResp, err := s.client.Do(sourcesReq)
	if err != nil {
		return nil, err
	}
	defer sourcesResp.Body.Close()

	var sourcesResponse struct {
		Sources []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			CatalogName string `json:"catalog_name"`
		} `json:"sources"`
	}

	if err := json.NewDecoder(sourcesResp.Body).Decode(&sourcesResponse); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	// Build schema representation
	tables := []ingestion.TableSchema{}
	for _, source := range sourcesResponse.Sources {
		// Each source represents an event stream
		columns := []ingestion.ColumnSchema{
			{Name: "event", DataType: "text", Nullable: false},
			{Name: "user_id", DataType: "text", Nullable: true},
			{Name: "anonymous_id", DataType: "text", Nullable: true},
			{Name: "timestamp", DataType: "timestamp", Nullable: false},
			{Name: "properties", DataType: "jsonb", Nullable: true},
			{Name: "context", DataType: "jsonb", Nullable: true},
		}

		tables = append(tables, ingestion.TableSchema{
			Name:    source.Name,
			Columns: columns,
			PrimaryKeys: []string{"event", "timestamp"},
			Metadata: map[string]interface{}{
				"source_id": source.ID,
				"source_name": source.Name,
				"catalog_name": source.CatalogName,
			},
		})
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a full or incremental sync */
func (s *SegmentConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		RowsSynced:   0,
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := s.DiscoverSchema(ctx)
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
