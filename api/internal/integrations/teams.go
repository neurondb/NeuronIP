package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* TeamsService provides Microsoft Teams integration functionality */
type TeamsService struct {
	pool *pgxpool.Pool
}

/* NewTeamsService creates a new Teams integration service */
func NewTeamsService(pool *pgxpool.Pool) *TeamsService {
	return &TeamsService{pool: pool}
}

/* TeamsConfig represents Teams integration configuration */
type TeamsConfig struct {
	WebhookURL string                 `json:"webhook_url"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	ClientID   string                 `json:"client_id,omitempty"`
	ClientSecret string               `json:"client_secret,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

/* SendTeamsMessage sends a message to Microsoft Teams */
func (s *TeamsService) SendTeamsMessage(ctx context.Context, config TeamsConfig, title string, message string, themeColor string) error {
	if config.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "https://schema.org/extensions",
		"summary":    title,
		"themeColor": themeColor,
		"title":      title,
		"text":       message,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("teams webhook returned status %d", resp.StatusCode)
	}

	return nil
}

/* SendTeamsCard sends a rich card to Microsoft Teams */
func (s *TeamsService) SendTeamsCard(ctx context.Context, config TeamsConfig, card map[string]interface{}) error {
	if config.WebhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	payloadJSON, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("teams webhook returned status %d", resp.StatusCode)
	}

	return nil
}

/* TestTeamsConnection tests Teams connection */
func (s *TeamsService) TestTeamsConnection(ctx context.Context, config TeamsConfig) error {
	return s.SendTeamsMessage(ctx, config, "Test Message", "Test message from NeuronIP", "0078D4")
}
