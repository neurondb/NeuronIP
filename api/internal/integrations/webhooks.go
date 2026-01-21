package integrations

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* WebhookService provides webhook functionality */
type WebhookService struct {
	pool   *pgxpool.Pool
	client *http.Client
}

/* NewWebhookService creates a new webhook service */
func NewWebhookService(pool *pgxpool.Pool) *WebhookService {
	return &WebhookService{
		pool: pool,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

/* Webhook represents a webhook registration */
type Webhook struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	URL            string                 `json:"url"`
	Events         []string               `json:"events"`
	Secret         *string                `json:"secret,omitempty"`
	Headers        map[string]string     `json:"headers,omitempty"`
	Enabled        bool                   `json:"enabled"`
	RetryConfig    map[string]interface{} `json:"retry_config,omitempty"`
	LastTriggeredAt *time.Time            `json:"last_triggered_at,omitempty"`
	TriggerCount   int                    `json:"trigger_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* CreateWebhook creates a new webhook */
func (s *WebhookService) CreateWebhook(ctx context.Context, webhook Webhook) (*Webhook, error) {
	webhook.ID = uuid.New()
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()
	
	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)
	retryConfigJSON, _ := json.Marshal(webhook.RetryConfig)
	
	query := `
		INSERT INTO neuronip.webhooks 
		(id, name, url, events, secret, headers, enabled, retry_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`
	
	err := s.pool.QueryRow(ctx, query,
		webhook.ID, webhook.Name, webhook.URL, eventsJSON, webhook.Secret, headersJSON,
		webhook.Enabled, retryConfigJSON, webhook.CreatedAt, webhook.UpdatedAt,
	).Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	
	return &webhook, nil
}

/* TriggerWebhook triggers a webhook */
func (s *WebhookService) TriggerWebhook(ctx context.Context, webhookID uuid.UUID, event string, payload map[string]interface{}) error {
	webhook, err := s.GetWebhook(ctx, webhookID)
	if err != nil {
		return err
	}
	
	if !webhook.Enabled {
		return fmt.Errorf("webhook is disabled")
	}
	
	// Check if webhook subscribes to this event
	subscribes := false
	for _, e := range webhook.Events {
		if e == event || e == "*" {
			subscribes = true
			break
		}
	}
	
	if !subscribes {
		return fmt.Errorf("webhook does not subscribe to event: %s", event)
	}
	
	// Send webhook with retry logic
	return s.sendWebhookWithRetry(ctx, webhook, event, payload)
}

/* sendWebhookWithRetry sends webhook with retry logic */
func (s *WebhookService) sendWebhookWithRetry(ctx context.Context, webhook *Webhook, event string, payload map[string]interface{}) error {
	maxRetries := 3
	if retryConfig, ok := webhook.RetryConfig["max_retries"].(float64); ok {
		maxRetries = int(retryConfig)
	}
	
	payloadJSON, _ := json.Marshal(payload)
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return err
		}
		
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Event", event)
		
		// Add custom headers
		for k, v := range webhook.Headers {
			req.Header.Set(k, v)
		}
		
		// Add signature if secret is set
		if webhook.Secret != nil && *webhook.Secret != "" {
			// Compute HMAC-SHA256 signature
			mac := hmac.New(sha256.New, []byte(*webhook.Secret))
			
			// Create payload for signing (body + timestamp if available)
			var payload []byte
			if req.Body != nil {
				bodyBytes, _ := io.ReadAll(req.Body)
				payload = bodyBytes
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Reset body for request
			}
			
			mac.Write(payload)
			signature := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-Webhook-Signature", fmt.Sprintf("sha256=%s", signature))
		}
		
		resp, err := s.client.Do(req)
		if err != nil {
			if attempt < maxRetries-1 {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success - update webhook stats
			now := time.Now()
			s.pool.Exec(ctx, `
				UPDATE neuronip.webhooks 
				SET last_triggered_at = $1, trigger_count = trigger_count + 1, updated_at = $1
				WHERE id = $2`, now, webhook.ID)
			return nil
		}
		
		// Retry on server errors
		if resp.StatusCode >= 500 && attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	return fmt.Errorf("webhook delivery failed after %d attempts", maxRetries)
}

/* GetWebhook retrieves a webhook */
func (s *WebhookService) GetWebhook(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	query := `
		SELECT id, name, url, events, secret, headers, enabled, retry_config,
		       last_triggered_at, trigger_count, created_at, updated_at
		FROM neuronip.webhooks
		WHERE id = $1`
	
	var webhook Webhook
	var eventsJSON, headersJSON, retryConfigJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID, &webhook.Name, &webhook.URL, &eventsJSON, &webhook.Secret, &headersJSON,
		&webhook.Enabled, &retryConfigJSON, &webhook.LastTriggeredAt, &webhook.TriggerCount,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	
	if eventsJSON != nil {
		json.Unmarshal(eventsJSON, &webhook.Events)
	}
	if headersJSON != nil {
		json.Unmarshal(headersJSON, &webhook.Headers)
	}
	if retryConfigJSON != nil {
		json.Unmarshal(retryConfigJSON, &webhook.RetryConfig)
	}
	
	return &webhook, nil
}

/* ListWebhooks lists all webhooks */
func (s *WebhookService) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	query := `
		SELECT id, name, url, events, secret, headers, enabled, retry_config,
		       last_triggered_at, trigger_count, created_at, updated_at
		FROM neuronip.webhooks
		ORDER BY created_at DESC`
	
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()
	
	webhooks := make([]Webhook, 0)
	for rows.Next() {
		var webhook Webhook
		var eventsJSON, headersJSON, retryConfigJSON json.RawMessage
		
		err := rows.Scan(
			&webhook.ID, &webhook.Name, &webhook.URL, &eventsJSON, &webhook.Secret, &headersJSON,
			&webhook.Enabled, &retryConfigJSON, &webhook.LastTriggeredAt, &webhook.TriggerCount,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if eventsJSON != nil {
			json.Unmarshal(eventsJSON, &webhook.Events)
		}
		if headersJSON != nil {
			json.Unmarshal(headersJSON, &webhook.Headers)
		}
		if retryConfigJSON != nil {
			json.Unmarshal(retryConfigJSON, &webhook.RetryConfig)
		}
		
		webhooks = append(webhooks, webhook)
	}
	
	return webhooks, nil
}
