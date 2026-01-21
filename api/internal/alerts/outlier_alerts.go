package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/profiling"
)

/* OutlierAlertService provides outlier-based alerting functionality */
type OutlierAlertService struct {
	pool           *pgxpool.Pool
	outlierService *profiling.OutlierService
}

/* NewOutlierAlertService creates a new outlier alert service */
func NewOutlierAlertService(pool *pgxpool.Pool, outlierService *profiling.OutlierService) *OutlierAlertService {
	return &OutlierAlertService{
		pool:           pool,
		outlierService: outlierService,
	}
}

/* OutlierAlertRule represents a rule for outlier-based alerts */
type OutlierAlertRule struct {
	ID               uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	ConnectorID       uuid.UUID              `json:"connector_id"`
	SchemaName        string                 `json:"schema_name"`
	TableName         string                 `json:"table_name"`
	ColumnName        string                 `json:"column_name"`
	OutlierType       string                 `json:"outlier_type"`
	SeverityThreshold string                 `json:"severity_threshold"` // "low", "medium", "high", "critical"
	MinDeviation      float64                `json:"min_deviation"`
	Enabled           bool                   `json:"enabled"`
	NotificationChannels []string             `json:"notification_channels,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

/* CheckOutlierAlerts checks for outliers and generates alerts */
func (s *OutlierAlertService) CheckOutlierAlerts(ctx context.Context, ruleID *uuid.UUID) ([]Alert, error) {
	var rules []OutlierAlertRule

	if ruleID != nil {
		// Get specific rule
		rule, err := s.getRule(ctx, *ruleID)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	} else {
		// Get all enabled rules
		allRules, err := s.getAllEnabledRules(ctx)
		if err != nil {
			return nil, err
		}
		rules = allRules
	}

	var alerts []Alert

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Detect outliers
		outliers, err := s.outlierService.DetectOutliers(ctx,
			rule.ConnectorID, rule.SchemaName, rule.TableName, rule.ColumnName,
			rule.OutlierType)
		if err != nil {
			continue
		}

		// Generate alerts for outliers matching severity threshold
		for _, outlier := range outliers {
			if s.shouldAlert(outlier, rule) {
				alert := s.createAlertFromOutlier(outlier, rule)
				alerts = append(alerts, alert)

				// Save alert
				s.saveAlert(ctx, alert)
			}
		}
	}

	return alerts, nil
}

/* getRule retrieves a specific alert rule */
func (s *OutlierAlertService) getRule(ctx context.Context, ruleID uuid.UUID) (*OutlierAlertRule, error) {
	var rule OutlierAlertRule
	var channelsJSON, metadataJSON []byte

	query := `
		SELECT id, name, description, connector_id, schema_name, table_name,
		       column_name, outlier_type, severity_threshold, min_deviation,
		       enabled, notification_channels, created_at, updated_at, metadata
		FROM neuronip.outlier_alert_rules
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, ruleID).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.ConnectorID,
		&rule.SchemaName, &rule.TableName, &rule.ColumnName, &rule.OutlierType,
		&rule.SeverityThreshold, &rule.MinDeviation, &rule.Enabled,
		&channelsJSON, &rule.CreatedAt, &rule.UpdatedAt, &metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	json.Unmarshal(channelsJSON, &rule.NotificationChannels)
	json.Unmarshal(metadataJSON, &rule.Metadata)

	return &rule, nil
}

/* getAllEnabledRules retrieves all enabled alert rules */
func (s *OutlierAlertService) getAllEnabledRules(ctx context.Context) ([]OutlierAlertRule, error) {
	query := `
		SELECT id, name, description, connector_id, schema_name, table_name,
		       column_name, outlier_type, severity_threshold, min_deviation,
		       enabled, notification_channels, created_at, updated_at, metadata
		FROM neuronip.outlier_alert_rules
		WHERE enabled = true
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}
	defer rows.Close()

	var rules []OutlierAlertRule
	for rows.Next() {
		var rule OutlierAlertRule
		var channelsJSON, metadataJSON []byte

		err := rows.Scan(&rule.ID, &rule.Name, &rule.Description, &rule.ConnectorID,
			&rule.SchemaName, &rule.TableName, &rule.ColumnName, &rule.OutlierType,
			&rule.SeverityThreshold, &rule.MinDeviation, &rule.Enabled,
			&channelsJSON, &rule.CreatedAt, &rule.UpdatedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(channelsJSON, &rule.NotificationChannels)
		json.Unmarshal(metadataJSON, &rule.Metadata)
		rules = append(rules, rule)
	}

	return rules, nil
}

/* shouldAlert determines if an outlier should trigger an alert */
func (s *OutlierAlertService) shouldAlert(outlier profiling.OutlierDetection, rule OutlierAlertRule) bool {
	// Check severity threshold
	severityOrder := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	outlierSeverity := severityOrder[outlier.Severity]
	thresholdSeverity := severityOrder[rule.SeverityThreshold]

	if outlierSeverity < thresholdSeverity {
		return false
	}

	// Check minimum deviation
	if outlier.Deviation < rule.MinDeviation {
		return false
	}

	return true
}

/* createAlertFromOutlier creates an alert from an outlier detection */
func (s *OutlierAlertService) createAlertFromOutlier(outlier profiling.OutlierDetection, rule OutlierAlertRule) Alert {
	alertID := uuid.New()
	now := time.Now()

	message := fmt.Sprintf("Outlier detected in %s.%s.%s: %s (Deviation: %.2f)",
		outlier.SchemaName, outlier.TableName, outlier.ColumnName,
		outlier.OutlierValue, outlier.Deviation)

	details := map[string]interface{}{
		"outlier_id":        outlier.ID.String(),
		"outlier_type":      outlier.OutlierType,
		"outlier_value":     outlier.OutlierValue,
		"expected_value":    outlier.ExpectedValue,
		"deviation":         outlier.Deviation,
		"connector_id":      outlier.ConnectorID.String(),
		"schema_name":       outlier.SchemaName,
		"table_name":        outlier.TableName,
		"column_name":       outlier.ColumnName,
	}

	return Alert{
		ID:        alertID,
		RuleID:    rule.ID,
		Severity:  outlier.Severity,
		Message:   message,
		Details:   details,
		Status:    "active",
		CreatedAt: now,
	}
}

/* saveAlert saves an alert to the database */
func (s *OutlierAlertService) saveAlert(ctx context.Context, alert Alert) error {
	detailsJSON, _ := json.Marshal(alert.Details)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.alerts
		(id, rule_id, severity, message, details, status, created_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		alert.ID, alert.RuleID, alert.Severity, alert.Message,
		detailsJSON, alert.Status, alert.CreatedAt, alert.ResolvedAt,
	)

	return err
}

/* CreateOutlierAlertRule creates a new outlier alert rule */
func (s *OutlierAlertService) CreateOutlierAlertRule(ctx context.Context, rule OutlierAlertRule) (*OutlierAlertRule, error) {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	channelsJSON, _ := json.Marshal(rule.NotificationChannels)
	metadataJSON, _ := json.Marshal(rule.Metadata)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO neuronip.outlier_alert_rules
		(id, name, description, connector_id, schema_name, table_name,
		 column_name, outlier_type, severity_threshold, min_deviation,
		 enabled, notification_channels, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		rule.ID, rule.Name, rule.Description, rule.ConnectorID,
		rule.SchemaName, rule.TableName, rule.ColumnName, rule.OutlierType,
		rule.SeverityThreshold, rule.MinDeviation, rule.Enabled,
		channelsJSON, rule.CreatedAt, rule.UpdatedAt, metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return &rule, nil
}
