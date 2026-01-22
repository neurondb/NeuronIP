package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides webhook functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new webhook service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* Webhook represents a webhook configuration */
type Webhook struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	URL           string                 `json:"url"`
	Events        []string               `json:"events"`
	Enabled       bool                   `json:"enabled"`
	Secret        *string                `json:"secret,omitempty"`
	Headers       map[string]string      `json:"headers,omitempty"`
	RetryCount    int                    `json:"retry_count"`
	TimeoutSeconds int                   `json:"timeout_seconds"`
	CreatedBy     *string                `json:"created_by,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateWebhook creates a new webhook */
func (s *Service) CreateWebhook(ctx context.Context, webhook Webhook) (*Webhook, error) {
	webhook.ID = uuid.New()
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	headersJSON, _ := json.Marshal(webhook.Headers)
	var secret sql.NullString
	if webhook.Secret != nil {
		secret = sql.NullString{String: *webhook.Secret, Valid: true}
	}
	var createdBy sql.NullString
	if webhook.CreatedBy != nil {
		createdBy = sql.NullString{String: *webhook.CreatedBy, Valid: true}
	}

	query := `
		INSERT INTO neuronip.webhooks
		(id, name, url, events, enabled, secret, headers, retry_count, timeout_seconds,
		 created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		webhook.ID, webhook.Name, webhook.URL, webhook.Events, webhook.Enabled,
		secret, headersJSON, webhook.RetryCount, webhook.TimeoutSeconds,
		createdBy, webhook.CreatedAt, webhook.UpdatedAt,
	).Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &webhook, nil
}

/* GetWebhook retrieves a webhook by ID */
func (s *Service) GetWebhook(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	var webhook Webhook
	var secret sql.NullString
	var createdBy sql.NullString
	var headersJSON []byte

	query := `
		SELECT id, name, url, events, enabled, secret, headers, retry_count,
		       timeout_seconds, created_by, created_at, updated_at
		FROM neuronip.webhooks
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&webhook.ID, &webhook.Name, &webhook.URL, &webhook.Events, &webhook.Enabled,
		&secret, &headersJSON, &webhook.RetryCount, &webhook.TimeoutSeconds,
		&createdBy, &webhook.CreatedAt, &webhook.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	if secret.Valid {
		webhook.Secret = &secret.String
	}
	if createdBy.Valid {
		webhook.CreatedBy = &createdBy.String
	}
	if headersJSON != nil {
		json.Unmarshal(headersJSON, &webhook.Headers)
	}

	return &webhook, nil
}

/* ListWebhooks lists all webhooks */
func (s *Service) ListWebhooks(ctx context.Context, enabledOnly bool) ([]Webhook, error) {
	query := `
		SELECT id, name, url, events, enabled, secret, headers, retry_count,
		       timeout_seconds, created_by, created_at, updated_at
		FROM neuronip.webhooks`
	
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var webhook Webhook
		var secret sql.NullString
		var createdBy sql.NullString
		var headersJSON []byte

		err := rows.Scan(
			&webhook.ID, &webhook.Name, &webhook.URL, &webhook.Events, &webhook.Enabled,
			&secret, &headersJSON, &webhook.RetryCount, &webhook.TimeoutSeconds,
			&createdBy, &webhook.CreatedAt, &webhook.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if secret.Valid {
			webhook.Secret = &secret.String
		}
		if createdBy.Valid {
			webhook.CreatedBy = &createdBy.String
		}
		if headersJSON != nil {
			json.Unmarshal(headersJSON, &webhook.Headers)
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

/* DeliverWebhook delivers a webhook event */
func (s *Service) DeliverWebhook(ctx context.Context, webhookID uuid.UUID, eventType string, payload map[string]interface{}) error {
	webhook, err := s.GetWebhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if !webhook.Enabled {
		return fmt.Errorf("webhook is disabled")
	}

	// Check if webhook subscribes to this event
	subscribes := false
	for _, event := range webhook.Events {
		if event == eventType || event == "*" {
			subscribes = true
			break
		}
	}

	if !subscribes {
		return fmt.Errorf("webhook does not subscribe to event: %s", eventType)
	}

	// Create delivery record
	deliveryID := uuid.New()
	payloadJSON, _ := json.Marshal(payload)

	query := `
		INSERT INTO neuronip.webhook_deliveries
		(id, webhook_id, event_type, payload, status, attempt_number, created_at)
		VALUES ($1, $2, $3, $4, 'pending', 1, NOW())
		RETURNING id`

	err = s.pool.QueryRow(ctx, query, deliveryID, webhookID, eventType, payloadJSON).Scan(&deliveryID)
	if err != nil {
		return fmt.Errorf("failed to create delivery record: %w", err)
	}

	// Deliver webhook (async)
	go s.deliverWebhookAsync(ctx, webhook, deliveryID, eventType, payload)

	return nil
}

/* deliverWebhookAsync delivers webhook asynchronously with retry */
func (s *Service) deliverWebhookAsync(ctx context.Context, webhook *Webhook, deliveryID uuid.UUID, eventType string, payload map[string]interface{}) {
	payloadJSON, _ := json.Marshal(payload)
	maxAttempts := webhook.RetryCount + 1
	timeout := time.Duration(webhook.TimeoutSeconds) * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Update attempt number
		s.pool.Exec(ctx, `
			UPDATE neuronip.webhook_deliveries
			SET attempt_number = $1, status = 'retrying'
			WHERE id = $2`, attempt, deliveryID)

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payloadJSON))
		if err != nil {
			s.updateDeliveryStatus(ctx, deliveryID, "failed", nil, nil, fmt.Sprintf("Failed to create request: %v", err))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Event", eventType)
		req.Header.Set("X-Webhook-ID", webhook.ID.String())

		// Add custom headers
		for k, v := range webhook.Headers {
			req.Header.Set(k, v)
		}

		// Add signature if secret is set
		if webhook.Secret != nil {
			signature := computeHMAC(payloadJSON, *webhook.Secret)
			req.Header.Set("X-Webhook-Signature", signature)
			req.Header.Set("X-Webhook-Signature-Algorithm", "sha256")
		}

		// Make request with timeout
		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		
		if err != nil {
			if attempt < maxAttempts {
				// Retry with exponential backoff
				backoff := time.Duration(attempt) * time.Second
				time.Sleep(backoff)
				continue
			}
			s.updateDeliveryStatus(ctx, deliveryID, "failed", nil, nil, err.Error())
			return
		}
		defer resp.Body.Close()

		// Read response body
		var responseBody string
		if resp.Body != nil {
			buf := make([]byte, 1024)
			n, _ := resp.Body.Read(buf)
			if n > 0 {
				responseBody = string(buf[:n])
			}
		}

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.updateDeliveryStatus(ctx, deliveryID, "success", &resp.StatusCode, &responseBody, "")
			return
		}

		// Retry on failure
		if attempt < maxAttempts {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
			continue
		}

		// Final failure
		s.updateDeliveryStatus(ctx, deliveryID, "failed", &resp.StatusCode, &responseBody, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return
	}
}

/* updateDeliveryStatus updates webhook delivery status */
func (s *Service) updateDeliveryStatus(ctx context.Context, deliveryID uuid.UUID, status string, statusCode *int, responseBody *string, errorMsg string) {
	var statusCodeVal sql.NullInt32
	if statusCode != nil {
		statusCodeVal = sql.NullInt32{Int32: int32(*statusCode), Valid: true}
	}
	var responseBodyVal sql.NullString
	if responseBody != nil {
		responseBodyVal = sql.NullString{String: *responseBody, Valid: true}
	}
	var errorMsgVal sql.NullString
	if errorMsg != "" {
		errorMsgVal = sql.NullString{String: errorMsg, Valid: true}
	}

	query := `
		UPDATE neuronip.webhook_deliveries
		SET status = $1, status_code = $2, response_body = $3, error_message = $4, delivered_at = NOW()
		WHERE id = $5`

	s.pool.Exec(ctx, query, status, statusCodeVal, responseBodyVal, errorMsgVal, deliveryID)
}

/* TriggerEvent triggers webhook events for all matching webhooks */
func (s *Service) TriggerEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	// Get all enabled webhooks that subscribe to this event
	webhooks, err := s.ListWebhooks(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	for _, webhook := range webhooks {
		// Check if webhook subscribes to this event
		subscribes := false
		for _, event := range webhook.Events {
			if event == eventType || event == "*" {
				subscribes = true
				break
			}
		}

		if subscribes {
			// Deliver webhook (non-blocking)
			go s.DeliverWebhook(context.Background(), webhook.ID, eventType, payload)
		}
	}

	return nil
}

/* computeHMAC computes HMAC-SHA256 signature for webhook payload */
func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
