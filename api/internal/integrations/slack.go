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

/* SlackService provides Slack integration functionality */
type SlackService struct {
	pool *pgxpool.Pool
}

/* NewSlackService creates a new Slack integration service */
func NewSlackService(pool *pgxpool.Pool) *SlackService {
	return &SlackService{pool: pool}
}

/* SlackConfig represents Slack integration configuration */
type SlackConfig struct {
	WebhookURL string                 `json:"webhook_url"`
	BotToken   string                 `json:"bot_token,omitempty"`
	Channel    string                 `json:"channel,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

/* SendSlackMessage sends a message to Slack */
func (s *SlackService) SendSlackMessage(ctx context.Context, config SlackConfig, message string, attachments []map[string]interface{}) error {
	if config.WebhookURL == "" && config.BotToken == "" {
		return fmt.Errorf("either webhook_url or bot_token is required")
	}

	if config.WebhookURL != "" {
		return s.sendViaWebhook(ctx, config, message, attachments)
	}

	return s.sendViaAPI(ctx, config, message, attachments)
}

/* sendViaWebhook sends message via Slack webhook */
func (s *SlackService) sendViaWebhook(ctx context.Context, config SlackConfig, message string, attachments []map[string]interface{}) error {
	payload := map[string]interface{}{
		"text": message,
	}

	if config.Channel != "" {
		payload["channel"] = config.Channel
	}

	if len(attachments) > 0 {
		payload["attachments"] = attachments
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

/* sendViaAPI sends message via Slack API */
func (s *SlackService) sendViaAPI(ctx context.Context, config SlackConfig, message string, attachments []map[string]interface{}) error {
	payload := map[string]interface{}{
		"text": message,
	}

	if config.Channel != "" {
		payload["channel"] = config.Channel
	} else {
		payload["channel"] = "#general"
	}

	if len(attachments) > 0 {
		payload["attachments"] = attachments
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := "https://slack.com/api/chat.postMessage"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.BotToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if ok, _ := result["ok"].(bool); !ok {
		return fmt.Errorf("slack API error: %v", result["error"])
	}

	return nil
}

/* TestSlackConnection tests Slack connection */
func (s *SlackService) TestSlackConnection(ctx context.Context, config SlackConfig) error {
	return s.SendSlackMessage(ctx, config, "Test message from NeuronIP", nil)
}

/* GetSlackChannels retrieves available Slack channels (requires bot token) */
func (s *SlackService) GetSlackChannels(ctx context.Context, botToken string) ([]map[string]interface{}, error) {
	if botToken == "" {
		return nil, fmt.Errorf("bot_token is required")
	}

	url := "https://slack.com/api/conversations.list"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+botToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK       bool                     `json:"ok"`
		Channels []map[string]interface{} `json:"channels"`
		Error    string                   `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("slack API error: %s", result.Error)
	}

	return result.Channels, nil
}
