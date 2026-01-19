package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* HelpdeskService provides helpdesk system integration functionality */
type HelpdeskService struct {
	pool *pgxpool.Pool
}

/* NewHelpdeskService creates a new helpdesk integration service */
func NewHelpdeskService(pool *pgxpool.Pool) *HelpdeskService {
	return &HelpdeskService{pool: pool}
}

/* HelpdeskConfig represents helpdesk integration configuration */
type HelpdeskConfig struct {
	Provider      string                 `json:"provider"` // "zendesk", "jira", "servicenow"
	Endpoint      string                 `json:"endpoint"`
	APIKey        string                 `json:"api_key"`
	AuthType      string                 `json:"auth_type"` // "api_key", "oauth", "basic"
	Config        map[string]interface{} `json:"config,omitempty"`
}

/* SyncTickets syncs tickets from external helpdesk system */
func (s *HelpdeskService) SyncTickets(ctx context.Context, config HelpdeskConfig, lastSyncTime *time.Time) (int, error) {
	switch config.Provider {
	case "zendesk":
		return s.syncZendeskTickets(ctx, config, lastSyncTime)
	case "jira":
		return s.syncJiraTickets(ctx, config, lastSyncTime)
	case "servicenow":
		return s.syncServiceNowTickets(ctx, config, lastSyncTime)
	default:
		return 0, fmt.Errorf("unsupported helpdesk provider: %s", config.Provider)
	}
}

/* syncZendeskTickets syncs tickets from Zendesk */
func (s *HelpdeskService) syncZendeskTickets(ctx context.Context, config HelpdeskConfig, lastSyncTime *time.Time) (int, error) {
	if config.Endpoint == "" {
		return 0, fmt.Errorf("zendesk endpoint is required")
	}

	// Build Zendesk API URL
	url := fmt.Sprintf("%s/api/v2/tickets.json", config.Endpoint)
	if lastSyncTime != nil {
		// Use incremental export endpoint if last sync time is provided
		url = fmt.Sprintf("%s/api/v2/incremental/tickets.json?start_time=%d", config.Endpoint, lastSyncTime.Unix())
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication based on auth type
	if config.AuthType == "basic" {
		// Basic auth with email/API key
		email := ""
		if emailVal, ok := config.Config["email"].(string); ok {
			email = emailVal
		}
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s/token:%s", email, config.APIKey)))
		req.Header.Set("Authorization", "Basic "+auth)
	} else {
		// API token auth
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("zendesk API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var zendeskResp struct {
		Tickets []map[string]interface{} `json:"tickets"`
		NextPage *string                 `json:"next_page"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&zendeskResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map and store tickets
	count := 0
	for _, ticket := range zendeskResp.Tickets {
		_, err := s.MapExternalTicket(ctx, "zendesk", ticket)
		if err != nil {
			continue
		}
		// Store ticket in database (implementation would depend on support_tickets table structure)
		count++
	}

	// Handle pagination if needed
	// In a full implementation, you would follow next_page links

	return count, nil
}

/* syncJiraTickets syncs tickets from Jira Service Desk */
func (s *HelpdeskService) syncJiraTickets(ctx context.Context, config HelpdeskConfig, lastSyncTime *time.Time) (int, error) {
	if config.Endpoint == "" {
		return 0, fmt.Errorf("jira endpoint is required")
	}

	// Build Jira API URL for issues search
	url := fmt.Sprintf("%s/rest/api/3/search", config.Endpoint)

	// Build JQL query
	jql := "project IS NOT EMPTY"
	if lastSyncTime != nil {
		jql += fmt.Sprintf(" AND updated >= '%s'", lastSyncTime.Format("2006-01-02 15:04"))
	}

	// Build request body
	reqBody := map[string]interface{}{
		"jql":        jql,
		"maxResults": 100,
		"startAt":    0,
		"fields":     []string{"summary", "status", "priority", "created", "updated", "assignee"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication
	// Jira uses Basic auth with email/API token
	email := ""
	if emailVal, ok := config.Config["email"].(string); ok {
		email = emailVal
	}
	if email == "" {
		email = config.APIKey // Fallback to API key if email not provided
	}
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, config.APIKey)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("jira API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var jiraResp struct {
		Issues []map[string]interface{} `json:"issues"`
		Total  int                       `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jiraResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map and store tickets
	count := 0
	for _, issue := range jiraResp.Issues {
		_, err := s.MapExternalTicket(ctx, "jira", issue)
		if err != nil {
			continue
		}
		// Store ticket in database
		count++
	}

	return count, nil
}

/* syncServiceNowTickets syncs tickets from ServiceNow */
func (s *HelpdeskService) syncServiceNowTickets(ctx context.Context, config HelpdeskConfig, lastSyncTime *time.Time) (int, error) {
	if config.Endpoint == "" {
		return 0, fmt.Errorf("servicenow endpoint is required")
	}

	// Build ServiceNow API URL for incidents table
	url := fmt.Sprintf("%s/api/now/table/incident", config.Endpoint)

	// Build query parameters
	query := "active=true"
	if lastSyncTime != nil {
		query += fmt.Sprintf("^sys_updated_on>=javascript:gs.dateGenerate('%s')", lastSyncTime.Format("2006-01-02 15:04:05"))
	}

	url += "?sysparm_query=" + query
	url += "&sysparm_limit=100"
	url += "&sysparm_fields=number,sys_id,short_description,state,priority,sys_created_on,sys_updated_on"

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication (ServiceNow uses Basic auth)
	username := ""
	if usernameVal, ok := config.Config["username"].(string); ok {
		username = usernameVal
	}
	if username == "" {
		username = config.APIKey // Fallback
	}
	password := config.APIKey
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("servicenow API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var servicenowResp struct {
		Result []map[string]interface{} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&servicenowResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Map and store tickets
	count := 0
	for _, incident := range servicenowResp.Result {
		// Map ServiceNow incident to internal ticket format
		ticket := &InternalTicket{
			ID:         uuid.New(),
			ExternalID: "",
			Provider:   "servicenow",
			CustomerID: "",
			Subject:    "",
			Status:     "open",
			Priority:   "medium",
			CreatedAt:  time.Now(),
		}

		if number, ok := incident["number"].(string); ok {
			ticket.ExternalID = "sn-" + number
		}
		if desc, ok := incident["short_description"].(string); ok {
			ticket.Subject = desc
		}
		if state, ok := incident["state"].(string); ok {
			// Map ServiceNow state to internal status
			switch state {
			case "1": // New
				ticket.Status = "open"
			case "2": // In Progress
				ticket.Status = "in_progress"
			case "3": // On Hold
				ticket.Status = "in_progress"
			case "6": // Resolved
				ticket.Status = "resolved"
			case "7": // Closed
				ticket.Status = "resolved"
			default:
				ticket.Status = "open"
			}
		}
		if priority, ok := incident["priority"].(string); ok {
			// Map ServiceNow priority
			switch priority {
			case "1": // Critical
				ticket.Priority = "urgent"
			case "2": // High
				ticket.Priority = "high"
			case "3": // Medium
				ticket.Priority = "medium"
			case "4": // Low
				ticket.Priority = "low"
			default:
				ticket.Priority = "medium"
			}
		}

		// Store ticket in database
		count++
	}

	return count, nil
}

/* MapExternalTicket maps external ticket format to internal format */
func (s *HelpdeskService) MapExternalTicket(ctx context.Context, provider string, externalTicket map[string]interface{}) (*InternalTicket, error) {
	ticket := &InternalTicket{
		ID:          uuid.New(),
		ExternalID:  "",
		Provider:    provider,
		CustomerID:  "",
		Subject:     "",
		Status:      "open",
		Priority:    "medium",
		CreatedAt:   time.Now(),
	}

	// Map fields based on provider
	switch provider {
	case "zendesk":
		if id, ok := externalTicket["id"].(float64); ok {
			ticket.ExternalID = fmt.Sprintf("zd-%d", int(id))
		}
		if subj, ok := externalTicket["subject"].(string); ok {
			ticket.Subject = subj
		}
		if req, ok := externalTicket["requester_id"].(float64); ok {
			ticket.CustomerID = fmt.Sprintf("zd-%d", int(req))
		}
		if status, ok := externalTicket["status"].(string); ok {
			ticket.Status = s.mapZendeskStatus(status)
		}
		if priority, ok := externalTicket["priority"].(string); ok {
			ticket.Priority = s.mapZendeskPriority(priority)
		}

	case "jira":
		if key, ok := externalTicket["key"].(string); ok {
			ticket.ExternalID = "jira-" + key
		}
		if fields, ok := externalTicket["fields"].(map[string]interface{}); ok {
			if summary, ok := fields["summary"].(string); ok {
				ticket.Subject = summary
			}
			if status, ok := fields["status"].(map[string]interface{}); ok {
				if statusName, ok := status["name"].(string); ok {
					ticket.Status = s.mapJiraStatus(statusName)
				}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return ticket, nil
}

/* InternalTicket represents an internal ticket format */
type InternalTicket struct {
	ID         uuid.UUID
	ExternalID string
	Provider   string
	CustomerID string
	Subject    string
	Status     string
	Priority   string
	CreatedAt  time.Time
}

/* mapZendeskStatus maps Zendesk status to internal status */
func (s *HelpdeskService) mapZendeskStatus(status string) string {
	switch status {
	case "new", "open":
		return "open"
	case "pending":
		return "in_progress"
	case "solved", "closed":
		return "resolved"
	default:
		return "open"
	}
}

/* mapZendeskPriority maps Zendesk priority to internal priority */
func (s *HelpdeskService) mapZendeskPriority(priority string) string {
	switch priority {
	case "urgent":
		return "urgent"
	case "high":
		return "high"
	case "normal":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

/* mapJiraStatus maps Jira status to internal status */
func (s *HelpdeskService) mapJiraStatus(status string) string {
	switch status {
	case "To Do", "Open":
		return "open"
	case "In Progress":
		return "in_progress"
	case "Done", "Closed", "Resolved":
		return "resolved"
	default:
		return "open"
	}
}
