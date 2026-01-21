package integrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* IntegrationsService provides unified integration management */
type IntegrationsService struct {
	pool           *pgxpool.Pool
	slackService   *SlackService
	teamsService   *TeamsService
	emailService   *EmailService
	webhookService *WebhookService
	helpdeskService *HelpdeskService
}

/* NewIntegrationsService creates a new integrations service */
func NewIntegrationsService(pool *pgxpool.Pool) *IntegrationsService {
	return &IntegrationsService{
		pool:           pool,
		slackService:   NewSlackService(pool),
		teamsService:   NewTeamsService(pool),
		emailService:   NewEmailService(pool),
		webhookService: NewWebhookService(pool),
		helpdeskService: NewHelpdeskService(pool),
	}
}

/* Integration represents an integration configuration */
type Integration struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	IntegrationType string               `json:"integration_type"`
	Config        map[string]interface{} `json:"config"`
	Enabled       bool                   `json:"enabled"`
	LastSyncAt    *time.Time             `json:"last_sync_at,omitempty"`
	Status        string                 `json:"status"`
	StatusMessage *string                `json:"status_message,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateIntegration creates a new integration */
func (s *IntegrationsService) CreateIntegration(ctx context.Context, integration Integration) (*Integration, error) {
	integration.ID = uuid.New()
	integration.CreatedAt = time.Now()
	integration.UpdatedAt = time.Now()
	if integration.Status == "" {
		integration.Status = "active"
	}

	configJSON, _ := json.Marshal(integration.Config)

	query := `
		INSERT INTO neuronip.integrations 
		(id, name, integration_type, config, enabled, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		integration.ID, integration.Name, integration.IntegrationType, configJSON,
		integration.Enabled, integration.Status, integration.CreatedAt, integration.UpdatedAt,
	).Scan(&integration.ID, &integration.CreatedAt, &integration.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	return &integration, nil
}

/* GetIntegration retrieves an integration */
func (s *IntegrationsService) GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error) {
	query := `
		SELECT id, name, integration_type, config, enabled, last_sync_at, created_at, updated_at
		FROM neuronip.integrations
		WHERE id = $1`

	var integration Integration
	var configJSON json.RawMessage
	var lastSyncAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&integration.ID, &integration.Name, &integration.IntegrationType, &configJSON,
		&integration.Enabled, &lastSyncAt, &integration.CreatedAt, &integration.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("integration not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	if configJSON != nil {
		json.Unmarshal(configJSON, &integration.Config)
	}
	if lastSyncAt.Valid {
		integration.LastSyncAt = &lastSyncAt.Time
	}

	// Check health status
	integration.Status, integration.StatusMessage = s.checkIntegrationHealth(ctx, &integration)

	return &integration, nil
}

/* ListIntegrations lists all integrations */
func (s *IntegrationsService) ListIntegrations(ctx context.Context, integrationType *string) ([]Integration, error) {
	query := `
		SELECT id, name, integration_type, config, enabled, last_sync_at, created_at, updated_at
		FROM neuronip.integrations
		WHERE ($1 IS NULL OR integration_type = $1)
		ORDER BY created_at DESC`

	args := []interface{}{integrationType}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	defer rows.Close()

	var integrations []Integration
	for rows.Next() {
		var integration Integration
		var configJSON json.RawMessage
		var lastSyncAt sql.NullTime

		err := rows.Scan(
			&integration.ID, &integration.Name, &integration.IntegrationType, &configJSON,
			&integration.Enabled, &lastSyncAt, &integration.CreatedAt, &integration.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configJSON != nil {
			json.Unmarshal(configJSON, &integration.Config)
		}
		if lastSyncAt.Valid {
			integration.LastSyncAt = &lastSyncAt.Time
		}

		// Check health status
		integration.Status, integration.StatusMessage = s.checkIntegrationHealth(ctx, &integration)

		integrations = append(integrations, integration)
	}

	return integrations, nil
}

/* UpdateIntegration updates an integration */
func (s *IntegrationsService) UpdateIntegration(ctx context.Context, id uuid.UUID, updates Integration) (*Integration, error) {
	configJSON, _ := json.Marshal(updates.Config)

	query := `
		UPDATE neuronip.integrations
		SET name = $1, config = $2, enabled = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, integration_type, config, enabled, last_sync_at, created_at, updated_at`

	var integration Integration
	var configJSONResult json.RawMessage
	var lastSyncAt sql.NullTime

	err := s.pool.QueryRow(ctx, query,
		updates.Name, configJSON, updates.Enabled, id,
	).Scan(
		&integration.ID, &integration.Name, &integration.IntegrationType, &configJSONResult,
		&integration.Enabled, &lastSyncAt, &integration.CreatedAt, &integration.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("integration not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	if configJSONResult != nil {
		json.Unmarshal(configJSONResult, &integration.Config)
	}
	if lastSyncAt.Valid {
		integration.LastSyncAt = &lastSyncAt.Time
	}

	integration.Status, integration.StatusMessage = s.checkIntegrationHealth(ctx, &integration)

	return &integration, nil
}

/* DeleteIntegration deletes an integration */
func (s *IntegrationsService) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.integrations WHERE id = $1`
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("integration not found")
	}

	return nil
}

/* TestIntegration tests an integration connection */
func (s *IntegrationsService) TestIntegration(ctx context.Context, id uuid.UUID) error {
	integration, err := s.GetIntegration(ctx, id)
	if err != nil {
		return err
	}

	switch integration.IntegrationType {
	case "slack":
		config := SlackConfig{}
		if webhookURL, ok := integration.Config["webhook_url"].(string); ok {
			config.WebhookURL = webhookURL
		}
		if botToken, ok := integration.Config["bot_token"].(string); ok {
			config.BotToken = botToken
		}
		if channel, ok := integration.Config["channel"].(string); ok {
			config.Channel = channel
		}
		return s.slackService.TestSlackConnection(ctx, config)

	case "teams":
		config := TeamsConfig{}
		if webhookURL, ok := integration.Config["webhook_url"].(string); ok {
			config.WebhookURL = webhookURL
		}
		return s.teamsService.TestTeamsConnection(ctx, config)

	case "email":
		config := EmailConfig{}
		if smtpHost, ok := integration.Config["smtp_host"].(string); ok {
			config.SMTPHost = smtpHost
		}
		if smtpPort, ok := integration.Config["smtp_port"].(float64); ok {
			config.SMTPPort = int(smtpPort)
		}
		if smtpUsername, ok := integration.Config["smtp_username"].(string); ok {
			config.SMTPUsername = smtpUsername
		}
		if smtpPassword, ok := integration.Config["smtp_password"].(string); ok {
			config.SMTPPassword = smtpPassword
		}
		if fromEmail, ok := integration.Config["from_email"].(string); ok {
			config.FromEmail = fromEmail
		}
		return s.emailService.TestEmailConnection(ctx, config)

	default:
		return fmt.Errorf("unsupported integration type: %s", integration.IntegrationType)
	}
}

/* checkIntegrationHealth checks the health status of an integration */
func (s *IntegrationsService) checkIntegrationHealth(ctx context.Context, integration *Integration) (string, *string) {
	if !integration.Enabled {
		msg := "Integration is disabled"
		return "disabled", &msg
	}

	// Basic health check - verify config is present
	if integration.Config == nil || len(integration.Config) == 0 {
		msg := "Configuration is missing"
		return "error", &msg
	}

	// Check last sync time for sync-based integrations
	if integration.LastSyncAt != nil {
		timeSinceSync := time.Since(*integration.LastSyncAt)
		if timeSinceSync > 24*time.Hour {
			msg := fmt.Sprintf("Last sync was %v ago", timeSinceSync)
			return "warning", &msg
		}
	}

	return "active", nil
}

/* GetIntegrationHealth retrieves health status for all integrations */
func (s *IntegrationsService) GetIntegrationHealth(ctx context.Context) (map[string]interface{}, error) {
	integrations, err := s.ListIntegrations(ctx, nil)
	if err != nil {
		return nil, err
	}

	health := map[string]interface{}{
		"total":        len(integrations),
		"active":       0,
		"error":        0,
		"warning":      0,
		"disabled":     0,
		"integrations": []map[string]interface{}{},
	}

	for _, integration := range integrations {
		status := integration.Status
		switch status {
		case "active":
			health["active"] = health["active"].(int) + 1
		case "error":
			health["error"] = health["error"].(int) + 1
		case "warning":
			health["warning"] = health["warning"].(int) + 1
		case "disabled":
			health["disabled"] = health["disabled"].(int) + 1
		}

		integrationMap := map[string]interface{}{
			"id":       integration.ID.String(),
			"name":     integration.Name,
			"type":     integration.IntegrationType,
			"status":   integration.Status,
			"enabled":  integration.Enabled,
		}
		if integration.StatusMessage != nil {
			integrationMap["status_message"] = *integration.StatusMessage
		}

		integrationsList := health["integrations"].([]map[string]interface{})
		health["integrations"] = append(integrationsList, integrationMap)
	}

	return health, nil
}
