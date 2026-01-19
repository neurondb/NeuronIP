package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/ingestion"
)

/* SalesforceConnector implements the Connector interface for Salesforce */
type SalesforceConnector struct {
	*ingestion.BaseConnector
	client      *http.Client
	instanceURL string
	accessToken string
	apiVersion  string
}

/* NewSalesforceConnector creates a new Salesforce connector */
func NewSalesforceConnector() *SalesforceConnector {
	metadata := ingestion.ConnectorMetadata{
		Type:        "salesforce",
		Name:        "Salesforce",
		Description: "Salesforce REST API connector for objects and records",
		Version:     "1.0.0",
		Capabilities: []string{"incremental", "schema_discovery"},
	}
	
	base := ingestion.NewBaseConnector("salesforce", metadata)
	
	return &SalesforceConnector{
		BaseConnector: base,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiVersion: "v58.0",
	}
}

/* Connect establishes connection to Salesforce using OAuth */
func (s *SalesforceConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	// Support both OAuth and username/password authentication
	if accessToken, ok := config["access_token"].(string); ok {
		s.accessToken = accessToken
		if instanceURL, ok := config["instance_url"].(string); ok {
			s.instanceURL = instanceURL
		} else {
			return fmt.Errorf("instance_url is required when using access_token")
		}
	} else {
		// OAuth flow - in production, this would handle the full OAuth flow
		// For now, require access_token and instance_url
		return fmt.Errorf("access_token and instance_url are required")
	}
	
	// Test connection
	if err := s.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	
	s.BaseConnector.SetConnected(true)
	return nil
}

/* Disconnect closes the connection */
func (s *SalesforceConnector) Disconnect(ctx context.Context) error {
	s.BaseConnector.SetConnected(false)
	return nil
}

/* TestConnection tests if the connection is valid */
func (s *SalesforceConnector) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.instanceURL+"/services/data/"+s.apiVersion+"/", nil)
	if err != nil {
		return err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}
	
	return nil
}

/* DiscoverSchema discovers Salesforce schema */
func (s *SalesforceConnector) DiscoverSchema(ctx context.Context) (*ingestion.Schema, error) {
	// Get list of objects
	objects, err := s.listObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}
	
	tables := []ingestion.TableSchema{}
	
	// Get schema for each object
	for _, objName := range objects {
		objSchema, err := s.describeObject(ctx, objName)
		if err != nil {
			continue // Skip objects we can't describe
		}
		
		columns := []ingestion.ColumnSchema{}
		for _, field := range objSchema.Fields {
			columns = append(columns, ingestion.ColumnSchema{
				Name:     field.Name,
				DataType: s.mapSalesforceType(field.Type),
				Nullable: field.Nillable,
			})
		}
		
		tables = append(tables, ingestion.TableSchema{
			Name:    objName,
			Columns: columns,
		})
	}
	
	return &ingestion.Schema{
		Tables:      tables,
		LastUpdated: time.Now(),
	}, nil
}

/* Sync performs a sync operation */
func (s *SalesforceConnector) Sync(ctx context.Context, options ingestion.SyncOptions) (*ingestion.SyncResult, error) {
	startTime := time.Now()
	result := &ingestion.SyncResult{
		TablesSynced: []string{},
		Errors:       []ingestion.SyncError{},
	}
	
	tables := options.Tables
	if len(tables) == 0 {
		// Get all objects if none specified
		objects, err := s.listObjects(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		tables = objects
	}
	
	for _, table := range tables {
		rows, err := s.syncObject(ctx, table, options)
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

/* listObjects lists all available Salesforce objects */
func (s *SalesforceConnector) listObjects(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.instanceURL+"/services/data/"+s.apiVersion+"/sobjects/", nil)
	if err != nil {
		return nil, err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var data struct {
		SObjects []struct {
			Name string `json:"name"`
		} `json:"sobjects"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	
	objects := make([]string, 0, len(data.SObjects))
	for _, obj := range data.SObjects {
		objects = append(objects, obj.Name)
	}
	
	return objects, nil
}

/* describeObject describes a Salesforce object */
func (s *SalesforceConnector) describeObject(ctx context.Context, objectName string) (*ObjectDescription, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.instanceURL+"/services/data/"+s.apiVersion+"/sobjects/"+objectName+"/describe/", nil)
	if err != nil {
		return nil, err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var desc ObjectDescription
	if err := json.NewDecoder(resp.Body).Decode(&desc); err != nil {
		return nil, err
	}
	
	return &desc, nil
}

/* syncObject syncs a Salesforce object */
func (s *SalesforceConnector) syncObject(ctx context.Context, objectName string, options ingestion.SyncOptions) (int64, error) {
	soql := fmt.Sprintf("SELECT Id FROM %s", objectName)
	
	if options.Since != nil {
		// Add WHERE clause for incremental sync
		soql += fmt.Sprintf(" WHERE LastModifiedDate >= %s", options.Since.Format("2006-01-02T15:04:05Z"))
	}
	
	soql += " LIMIT 2000" // Salesforce limit
	
	req, err := http.NewRequestWithContext(ctx, "GET", s.instanceURL+"/services/data/"+s.apiVersion+"/query/?q="+url.QueryEscape(soql), nil)
	if err != nil {
		return 0, err
	}
	
	s.setAuth(req)
	
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed: status %d", resp.StatusCode)
	}
	
	var data struct {
		TotalSize int `json:"totalSize"`
		Records   []map[string]interface{} `json:"records"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	
	return int64(data.TotalSize), nil
}

/* mapSalesforceType maps Salesforce field types to SQL types */
func (s *SalesforceConnector) mapSalesforceType(sfType string) string {
	typeMap := map[string]string{
		"string":    "text",
		"textarea":  "text",
		"email":     "text",
		"url":       "text",
		"phone":     "text",
		"int":       "integer",
		"double":    "numeric",
		"currency":  "numeric",
		"percent":   "numeric",
		"date":      "date",
		"datetime":  "timestamp",
		"time":      "time",
		"boolean":   "boolean",
		"picklist":  "text",
		"multipicklist": "text[]",
		"reference": "bigint",
		"id":        "text",
	}
	
	if mapped, ok := typeMap[strings.ToLower(sfType)]; ok {
		return mapped
	}
	
	return "text" // Default
}

/* setAuth sets authentication headers */
func (s *SalesforceConnector) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Content-Type", "application/json")
}

/* ObjectDescription represents a Salesforce object description */
type ObjectDescription struct {
	Name   string `json:"name"`
	Fields []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Nillable bool   `json:"nillable"`
	} `json:"fields"`
}
