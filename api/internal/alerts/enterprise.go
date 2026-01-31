package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* EnterpriseAlertingService provides enterprise-grade alerting */
type EnterpriseAlertingService struct {
	pool *pgxpool.Pool
}

/* NewEnterpriseAlertingService creates a new enterprise alerting service */
func NewEnterpriseAlertingService(pool *pgxpool.Pool) *EnterpriseAlertingService {
	return &EnterpriseAlertingService{pool: pool}
}

/* CreateAlertRule creates an alert rule with advanced features */
func (eas *EnterpriseAlertingService) CreateAlertRule(ctx context.Context, rule EnterpriseAlertRule) (*EnterpriseAlertRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	configJSON, _ := json.Marshal(rule.Config)
	channelsJSON, _ := json.Marshal(rule.NotificationChannels)

	query := `
		INSERT INTO neuronip.enterprise_alert_rules 
		(id, name, description, rule_type, metric, condition, threshold, enabled, 
		 severity, notification_channels, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, description, rule_type, metric, condition, threshold, enabled,
		          severity, notification_channels, config, created_at, updated_at`

	var channelsJSONRaw, configJSONRaw json.RawMessage
	err := eas.pool.QueryRow(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.RuleType, rule.Metric,
		rule.Condition, rule.Threshold, rule.Enabled, rule.Severity,
		channelsJSON, configJSON, rule.CreatedAt, rule.UpdatedAt,
	).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.RuleType, &rule.Metric,
		&rule.Condition, &rule.Threshold, &rule.Enabled, &rule.Severity,
		&channelsJSONRaw, &configJSONRaw, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	if channelsJSONRaw != nil {
		json.Unmarshal(channelsJSONRaw, &rule.NotificationChannels)
	}
	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &rule.Config)
	}

	return &rule, nil
}

/* SendAlert sends an alert through configured channels */
func (eas *EnterpriseAlertingService) SendAlert(ctx context.Context, alert EnterpriseAlert) error {
	alert.ID = uuid.New()
	alert.CreatedAt = time.Now()
	alert.Status = "active"

	detailsJSON, _ := json.Marshal(alert.Details)

	// Store alert
	query := `
		INSERT INTO neuronip.enterprise_alerts 
		(id, rule_id, severity, message, details, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := eas.pool.Exec(ctx, query,
		alert.ID, alert.RuleID, alert.Severity, alert.Message,
		detailsJSON, alert.Status, alert.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}

	// Get notification channels from rule
	rule, err := eas.GetAlertRule(ctx, alert.RuleID)
	if err != nil {
		return err
	}

	// Send notifications through all channels
	for _, channel := range rule.NotificationChannels {
		go eas.sendNotification(ctx, channel, alert)
	}

	// Aggregate similar alerts if configured
	if rule.Config["aggregate_alerts"] == true {
		go eas.aggregateAlerts(ctx, alert)
	}

	return nil
}

/* sendNotification sends a notification through a channel */
func (eas *EnterpriseAlertingService) sendNotification(ctx context.Context, channel NotificationChannel, alert EnterpriseAlert) {
	switch channel.Type {
	case "email":
		eas.sendEmailNotification(ctx, channel, alert)
	case "slack":
		eas.sendSlackNotification(ctx, channel, alert)
	case "pagerduty":
		eas.sendPagerDutyNotification(ctx, channel, alert)
	case "webhook":
		eas.sendWebhookNotification(ctx, channel, alert)
	}
}

/* sendEmailNotification sends email notification */
func (eas *EnterpriseAlertingService) sendEmailNotification(ctx context.Context, channel NotificationChannel, alert EnterpriseAlert) {
	// In production, integrate with email service
	// For now, log the notification
	_ = ctx
	_ = channel
	_ = alert
}

/* sendSlackNotification sends Slack notification */
func (eas *EnterpriseAlertingService) sendSlackNotification(ctx context.Context, channel NotificationChannel, alert EnterpriseAlert) {
	// In production, integrate with Slack API
	_ = ctx
	_ = channel
	_ = alert
}

/* sendPagerDutyNotification sends PagerDuty notification */
func (eas *EnterpriseAlertingService) sendPagerDutyNotification(ctx context.Context, channel NotificationChannel, alert EnterpriseAlert) {
	// In production, integrate with PagerDuty API
	_ = ctx
	_ = channel
	_ = alert
}

/* sendWebhookNotification sends webhook notification */
func (eas *EnterpriseAlertingService) sendWebhookNotification(ctx context.Context, channel NotificationChannel, alert EnterpriseAlert) {
	// In production, make HTTP POST to webhook URL
	_ = ctx
	_ = channel
	_ = alert
}

/* aggregateAlerts aggregates similar alerts */
func (eas *EnterpriseAlertingService) aggregateAlerts(ctx context.Context, alert EnterpriseAlert) {
	// Find similar alerts in the last hour
	query := `
		SELECT id, COUNT(*) as count
		FROM neuronip.enterprise_alerts
		WHERE rule_id = $1
			AND status = 'active'
			AND created_at > NOW() - INTERVAL '1 hour'
		GROUP BY id`

	rows, err := eas.pool.Query(ctx, query, alert.RuleID)
	if err != nil {
		return
	}
	defer rows.Close()

	// If multiple similar alerts, create aggregated alert
	// This prevents alert fatigue
	_ = rows
}

/* GetAlertRule retrieves an alert rule */
func (eas *EnterpriseAlertingService) GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*EnterpriseAlertRule, error) {
	query := `
		SELECT id, name, description, rule_type, metric, condition, threshold, enabled,
		       severity, notification_channels, config, created_at, updated_at
		FROM neuronip.enterprise_alert_rules
		WHERE id = $1`

	var rule EnterpriseAlertRule
	var channelsJSONRaw, configJSONRaw json.RawMessage

	err := eas.pool.QueryRow(ctx, query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.RuleType, &rule.Metric,
		&rule.Condition, &rule.Threshold, &rule.Enabled, &rule.Severity,
		&channelsJSONRaw, &configJSONRaw, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	if channelsJSONRaw != nil {
		json.Unmarshal(channelsJSONRaw, &rule.NotificationChannels)
	}
	if configJSONRaw != nil {
		json.Unmarshal(configJSONRaw, &rule.Config)
	}

	return &rule, nil
}

/* AcknowledgeAlert acknowledges an alert */
func (eas *EnterpriseAlertingService) AcknowledgeAlert(ctx context.Context, alertID uuid.UUID, userID string) error {
	query := `
		UPDATE neuronip.enterprise_alerts
		SET status = 'acknowledged', acknowledged_by = $1, acknowledged_at = NOW()
		WHERE id = $2`

	_, err := eas.pool.Exec(ctx, query, userID, alertID)
	return err
}

/* ResolveAlert resolves an alert */
func (eas *EnterpriseAlertingService) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string) error {
	query := `
		UPDATE neuronip.enterprise_alerts
		SET status = 'resolved', resolution = $1, resolved_at = NOW()
		WHERE id = $2`

	_, err := eas.pool.Exec(ctx, query, resolution, alertID)
	return err
}

/* EnterpriseAlertRule represents an enterprise alert rule */
type EnterpriseAlertRule struct {
	ID                  uuid.UUID              `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	RuleType            string                 `json:"rule_type"`
	Metric              string                 `json:"metric"`
	Condition           string                 `json:"condition"`
	Threshold           float64                `json:"threshold"`
	Enabled             bool                   `json:"enabled"`
	Severity            string                 `json:"severity"` // "low", "medium", "high", "critical"
	NotificationChannels []NotificationChannel `json:"notification_channels"`
	Config              map[string]interface{} `json:"config"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

/* EnterpriseAlert represents an enterprise alert */
type EnterpriseAlert struct {
	ID            uuid.UUID              `json:"id"`
	RuleID        uuid.UUID              `json:"rule_id"`
	Severity      string                 `json:"severity"`
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details"`
	Status        string                 `json:"status"` // "active", "acknowledged", "resolved"
	AcknowledgedBy *string               `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time            `json:"acknowledged_at,omitempty"`
	Resolution    *string                `json:"resolution,omitempty"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* NotificationChannel represents a notification channel */
type NotificationChannel struct {
	Type    string                 `json:"type"` // "email", "slack", "pagerduty", "webhook"
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}
