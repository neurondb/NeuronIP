package neuronip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

/* Client provides NeuronIP API client */
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

/* NewClient creates a new NeuronIP client */
func NewClient(baseURL string, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

/* request makes an HTTP request */
func (c *Client) request(method string, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

/* HealthCheck checks API health */
func (c *Client) HealthCheck() (map[string]interface{}, error) {
	data, err := c.request("GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result, nil
}

/* SemanticSearch performs semantic search */
func (c *Client) SemanticSearch(query string, limit int) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	data, err := c.request("POST", "/semantic/search", body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result, nil
}

/* WarehouseQuery executes a warehouse query */
func (c *Client) WarehouseQuery(query string, schemaID *string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"query": query,
	}
	if schemaID != nil {
		body["schema_id"] = *schemaID
	}

	data, err := c.request("POST", "/warehouse/query", body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result, nil
}

/* CreateIngestionJob creates an ingestion job */
func (c *Client) CreateIngestionJob(dataSourceID string, jobType string, config map[string]interface{}) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"data_source_id": dataSourceID,
		"job_type":       jobType,
		"config":         config,
	}

	data, err := c.request("POST", "/ingestion/jobs", body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result, nil
}
