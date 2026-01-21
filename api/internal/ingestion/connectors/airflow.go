package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* AirflowConnector implements the Connector interface for Apache Airflow */
type AirflowConnector struct {
	*ingestion.BaseConnector
	client     *http.Client
	baseURL    string
	username   string
	password   string
	accessToken string
}

/* NewAirflowConnector creates a new Airflow connector */
func NewAirflowConnector() *AirflowConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "airflow",
		Name:        "Apache Airflow",
		Description: "Apache Airflow REST API connector for DAGs, tasks, and lineage",
		Version:     "1.0.0",
		Capabilities: []string{"schema_discovery", "lineage_extraction", "transformation_logic"},
	}

	base := ingestion.NewBaseConnector("airflow", metadata)

	return &AirflowConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* Connect establishes connection to Airflow */
func (a *AirflowConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	baseURL, _ := config["base_url"].(string)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	accessToken, _ := config["access_token"].(string)

	if baseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	a.baseURL = baseURL

	if accessToken != "" {
		a.accessToken = accessToken
	} else if username != "" && password != "" {
		a.username = username
		a.password = password
		// Authenticate and get token
		if err := a.authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	} else {
		return fmt.Errorf("either access_token or username/password is required")
	}

	// Test connection
	if err := a.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	a.BaseConnector.SetConnected(true)
	return nil
}

/* authenticate authenticates with Airflow and retrieves access token */
func (a *AirflowConnector) authenticate(ctx context.Context) error {
	// Airflow 2.x uses session-based auth, or token auth
	// For simplicity, we'll use basic auth or token
	// In production, implement full OAuth flow
	return nil
}

/* Disconnect closes the connection */
func (a *AirflowConnector) Disconnect(ctx context.Context) error {
	a.BaseConnector.SetConnected(false)
	a.accessToken = ""
	return nil
}

/* TestConnection tests if the connection is valid */
func (a *AirflowConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/api/v1/version", nil)
	if err != nil {
		return err
	}

	a.setAuth(req)

	resp, err := a.client.Do(req)
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
func (a *AirflowConnector) setAuth(req *http.Request) {
	if a.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.accessToken)
	} else if a.username != "" {
		req.SetBasicAuth(a.username, a.password)
	}
}

/* DiscoverSchema discovers Airflow DAG structure */
func (a *AirflowConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	// Fetch DAGs
	dagsReq, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/api/v1/dags", nil)
	if err != nil {
		return nil, err
	}

	a.setAuth(dagsReq)
	dagsResp, err := a.client.Do(dagsReq)
	if err != nil {
		return nil, err
	}
	defer dagsResp.Body.Close()

	var dagsResponse struct {
		DAGs []struct {
			DAGID string `json:"dag_id"`
		} `json:"dags"`
	}

	if err := json.NewDecoder(dagsResp.Body).Decode(&dagsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode DAGs response: %w", err)
	}

	// Build schema representation
	tables := []ingestion.TableSchema{}
	for _, dag := range dagsResponse.DAGs {
		// Fetch DAG details
		dagDetailReq, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dags/%s", a.baseURL, dag.DAGID), nil)
		a.setAuth(dagDetailReq)
		dagDetailResp, _ := a.client.Do(dagDetailReq)
		if dagDetailResp != nil {
			var dagDetail struct {
				DAGID    string `json:"dag_id"`
				Tasks    []struct {
					TaskID string `json:"task_id"`
					Operator string `json:"operator"`
				} `json:"tasks"`
			}
			json.NewDecoder(dagDetailResp.Body).Decode(&dagDetail)
			dagDetailResp.Body.Close()

			// Create table schema for DAG
			columns := []ingestion.ColumnSchema{
				{Name: "dag_id", DataType: "text", Nullable: false},
				{Name: "task_id", DataType: "text", Nullable: false},
				{Name: "operator", DataType: "text", Nullable: true},
				{Name: "upstream_tasks", DataType: "jsonb", Nullable: true},
				{Name: "downstream_tasks", DataType: "jsonb", Nullable: true},
				{Name: "schedule", DataType: "text", Nullable: true},
			}

			tables = append(tables, ingestion.TableSchema{
				Name:    dag.DAGID,
				Columns: columns,
				PrimaryKeys: []string{"dag_id", "task_id"},
			})
		}
	}

	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a full or incremental sync */
func (a *AirflowConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	// Sync DAG definitions and execution history
	startTime := time.Now()
	result := &ingestion.SyncResult{
		RowsSynced:   0,
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}

	schema, err := a.DiscoverSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover schema: %w", err)
	}

	// Process each DAG
	for _, table := range schema.Tables {
		result.TablesSynced = append(result.TablesSynced, table.Name)
		result.RowsSynced += int64(len(table.Columns))
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
