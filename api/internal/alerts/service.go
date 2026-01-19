package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
)

/* Service provides alerting functionality */
type Service struct {
	pool          *pgxpool.Pool
	anomalyService *compliance.AnomalyService
}

/* NewService creates a new alerts service */
func NewService(pool *pgxpool.Pool, anomalyService *compliance.AnomalyService) *Service {
	return &Service{
		pool:          pool,
		anomalyService: anomalyService,
	}
}

/* AlertRule represents an alert rule */
type AlertRule struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	RuleType    string                 `json:"rule_type"` // "threshold", "anomaly", "data_drift"
	Threshold   float64                `json:"threshold,omitempty"`
	Metric      string                 `json:"metric"`
	Condition   string                 `json:"condition"` // "gt", "lt", "eq", "gte", "lte"
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

/* Alert represents an alert */
type Alert struct {
	ID          uuid.UUID              `json:"id"`
	RuleID      uuid.UUID              `json:"rule_id"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Status      string                 `json:"status"` // "active", "acknowledged", "resolved"
	CreatedAt   time.Time              `json:"created_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

/* CheckAlerts checks all alert rules and creates alerts for violations */
func (s *Service) CheckAlerts(ctx context.Context) ([]Alert, error) {
	// Get all enabled alert rules
	rules, err := s.GetEnabledAlertRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}

	var newAlerts []Alert

	for _, rule := range rules {
		alert, err := s.checkRule(ctx, rule)
		if err != nil {
			continue // Skip rules that fail to check
		}

		if alert != nil {
			newAlerts = append(newAlerts, *alert)
		}
	}

	return newAlerts, nil
}

/* checkRule checks a single alert rule */
func (s *Service) checkRule(ctx context.Context, rule AlertRule) (*Alert, error) {
	switch rule.RuleType {
	case "threshold":
		return s.checkThresholdRule(ctx, rule)
	case "anomaly":
		return s.checkAnomalyRule(ctx, rule)
	case "data_drift":
		return s.checkDataDriftRule(ctx, rule)
	default:
		return nil, fmt.Errorf("unknown rule type: %s", rule.RuleType)
	}
}

/* checkThresholdRule checks a threshold-based alert rule */
func (s *Service) checkThresholdRule(ctx context.Context, rule AlertRule) (*Alert, error) {
	// Get current metric value (placeholder - would query actual metrics)
	metricValue, err := s.getMetricValue(ctx, rule.Metric)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric value: %w", err)
	}

	// Check threshold condition
	triggered := false
	switch rule.Condition {
	case "gt":
		triggered = metricValue > rule.Threshold
	case "gte":
		triggered = metricValue >= rule.Threshold
	case "lt":
		triggered = metricValue < rule.Threshold
	case "lte":
		triggered = metricValue <= rule.Threshold
	case "eq":
		triggered = metricValue == rule.Threshold
	}

	if !triggered {
		return nil, nil // No alert
	}

	// Create alert
	alert := Alert{
		ID:       uuid.New(),
		RuleID:   rule.ID,
		Severity: "medium",
		Message:  fmt.Sprintf("Metric %s = %.2f violates threshold (%.2f %s)", rule.Metric, metricValue, rule.Threshold, rule.Condition),
		Details: map[string]interface{}{
			"metric":      rule.Metric,
			"value":       metricValue,
			"threshold":   rule.Threshold,
			"condition":   rule.Condition,
		},
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// Store alert
	err = s.storeAlert(ctx, alert)
	if err != nil {
		return nil, fmt.Errorf("failed to store alert: %w", err)
	}

	return &alert, nil
}

/* checkAnomalyRule checks an anomaly-based alert rule */
func (s *Service) checkAnomalyRule(ctx context.Context, rule AlertRule) (*Alert, error) {
	// Use anomaly detection service to check for recent anomalies
	entityType := "metric"
	if entityTypeVal, ok := rule.Config["entity_type"].(string); ok {
		entityType = entityTypeVal
	}

	entityID := rule.Metric

	// Get recent anomalies for this metric
	detections, err := s.anomalyService.GetAnomalyDetections(ctx, entityType, entityID, "detected", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to check anomalies: %w", err)
	}

	if len(detections) == 0 {
		return nil, nil // No anomalies detected
	}

	// Check if any detection is above threshold
	threshold := 0.7
	if thresh, ok := rule.Config["threshold"].(float64); ok {
		threshold = thresh
	}

	for _, detection := range detections {
		if detection.AnomalyScore >= threshold {
			// Create alert for high-score anomaly
			alert := Alert{
				ID:       uuid.New(),
				RuleID:   rule.ID,
				Severity: s.getSeverityFromScore(detection.AnomalyScore),
				Message:  fmt.Sprintf("Anomaly detected for %s: score %.2f", rule.Metric, detection.AnomalyScore),
				Details: map[string]interface{}{
					"anomaly_id":   detection.ID.String(),
					"anomaly_score": detection.AnomalyScore,
					"entity_type":   entityType,
					"entity_id":     entityID,
					"detection_details": detection.Details,
				},
				Status:    "active",
				CreatedAt: time.Now(),
			}

			err = s.storeAlert(ctx, alert)
			if err != nil {
				return nil, fmt.Errorf("failed to store alert: %w", err)
			}

			return &alert, nil
		}
	}

	return nil, nil
}

/* checkDataDriftRule checks a data drift alert rule */
func (s *Service) checkDataDriftRule(ctx context.Context, rule AlertRule) (*Alert, error) {
	// For data drift, we compare current data distribution with baseline
	threshold := rule.Threshold
	if threshold == 0 {
		threshold = 0.3 // Default drift threshold
	}

	// Get baseline and current data distribution from config
	baselineTable := ""
	if table, ok := rule.Config["baseline_table"].(string); ok {
		baselineTable = table
	}
	if baselineTable == "" {
		return nil, fmt.Errorf("baseline_table not specified in rule config")
	}

	currentTable := baselineTable
	if table, ok := rule.Config["current_table"].(string); ok {
		currentTable = table
	}

	column := rule.Metric
	if col, ok := rule.Config["column"].(string); ok {
		column = col
	}

	// Use statistical method to detect drift
	// Calculate mean and std deviation of baseline vs current
	baselineQuery := fmt.Sprintf(`
		SELECT AVG(%s::float) as mean, STDDEV(%s::float) as stddev, COUNT(*) as count
		FROM %s WHERE %s IS NOT NULL`, column, column, baselineTable, column)

	var baselineMean, baselineStddev float64
	var baselineCount int
	err := s.pool.QueryRow(ctx, baselineQuery).Scan(&baselineMean, &baselineStddev, &baselineCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate baseline: %w", err)
	}

	currentQuery := fmt.Sprintf(`
		SELECT AVG(%s::float) as mean, STDDEV(%s::float) as stddev, COUNT(*) as count
		FROM %s WHERE %s IS NOT NULL`, column, column, currentTable, column)

	var currentMean, currentStddev float64
	var currentCount int
	err = s.pool.QueryRow(ctx, currentQuery).Scan(&currentMean, &currentStddev, &currentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate current: %w", err)
	}

	// Calculate drift score using statistical distance
	// Using coefficient of variation difference as a simple drift metric
	var driftScore float64
	if baselineStddev > 0 && baselineMean != 0 {
		baselineCV := baselineStddev / baselineMean
		currentCV := currentStddev / currentMean
		driftScore = abs(currentCV - baselineCV)
	} else if baselineMean != 0 {
		// If no variance in baseline, check mean shift
		driftScore = abs((currentMean - baselineMean) / baselineMean)
	} else {
		driftScore = abs(currentMean - baselineMean)
	}

	if driftScore >= threshold {
		// Create alert for data drift
		alert := Alert{
			ID:       uuid.New(),
			RuleID:   rule.ID,
			Severity: "high",
			Message:  fmt.Sprintf("Data drift detected for %s: drift score %.3f exceeds threshold %.3f", rule.Metric, driftScore, threshold),
			Details: map[string]interface{}{
				"metric":        rule.Metric,
				"drift_score":   driftScore,
				"threshold":     threshold,
				"baseline_mean": baselineMean,
				"baseline_stddev": baselineStddev,
				"current_mean":  currentMean,
				"current_stddev": currentStddev,
				"baseline_count": baselineCount,
				"current_count": currentCount,
			},
			Status:    "active",
			CreatedAt: time.Now(),
		}

		err = s.storeAlert(ctx, alert)
		if err != nil {
			return nil, fmt.Errorf("failed to store alert: %w", err)
		}

		return &alert, nil
	}

	return nil, nil
}

/* abs returns absolute value of float64 */
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

/* getSeverityFromScore determines alert severity based on anomaly score */
func (s *Service) getSeverityFromScore(score float64) string {
	if score >= 0.9 {
		return "critical"
	} else if score >= 0.8 {
		return "high"
	} else if score >= 0.7 {
		return "medium"
	}
	return "low"
}

/* getMetricValue gets current value for a metric */
func (s *Service) getMetricValue(ctx context.Context, metric string) (float64, error) {
	// Create metric_values table if it doesn't exist for caching metric values
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS neuronip.metric_values (
			metric_name TEXT PRIMARY KEY,
			metric_value FLOAT8 NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	s.pool.Exec(ctx, createTableQuery)

	// Query cached metric value (in production, this would query Prometheus or metric storage)
	query := `SELECT metric_value FROM neuronip.metric_values WHERE metric_name = $1`
	var value float64
	err := s.pool.QueryRow(ctx, query, metric).Scan(&value)
	if err != nil {
		// If metric not found, return 0 (could also query Prometheus directly here)
		// In production, you would query Prometheus metrics endpoint
		return 0.0, nil
	}

	return value, nil
}

/* GetEnabledAlertRules retrieves all enabled alert rules */
func (s *Service) GetEnabledAlertRules(ctx context.Context) ([]AlertRule, error) {
	// Create alert_rules table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS neuronip.alert_rules (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			rule_type TEXT NOT NULL CHECK (rule_type IN ('threshold', 'anomaly', 'data_drift')),
			threshold FLOAT8,
			metric TEXT NOT NULL,
			condition TEXT NOT NULL CHECK (condition IN ('gt', 'gte', 'lt', 'lte', 'eq')),
			enabled BOOLEAN NOT NULL DEFAULT true,
			config JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	s.pool.Exec(ctx, createTableQuery)

	query := `
		SELECT id, name, rule_type, threshold, metric, condition, enabled, config
		FROM neuronip.alert_rules
		WHERE enabled = true
		ORDER BY created_at ASC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rules: %w", err)
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var rule AlertRule
		var configJSON json.RawMessage

		err := rows.Scan(&rule.ID, &rule.Name, &rule.RuleType, &rule.Threshold,
			&rule.Metric, &rule.Condition, &rule.Enabled, &configJSON)
		if err != nil {
			continue
		}

		if configJSON != nil {
			json.Unmarshal(configJSON, &rule.Config)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

/* storeAlert stores an alert */
func (s *Service) storeAlert(ctx context.Context, alert Alert) error {
	// Create alerts table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS neuronip.alerts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			rule_id UUID NOT NULL REFERENCES neuronip.alert_rules(id) ON DELETE CASCADE,
			severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
			message TEXT NOT NULL,
			details JSONB DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ
		)`
	s.pool.Exec(ctx, createTableQuery)

	detailsJSON, _ := json.Marshal(alert.Details)

	query := `
		INSERT INTO neuronip.alerts (id, rule_id, severity, message, details, status, created_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, resolved_at = EXCLUDED.resolved_at`

	_, err := s.pool.Exec(ctx, query,
		alert.ID, alert.RuleID, alert.Severity, alert.Message, detailsJSON,
		alert.Status, alert.CreatedAt, alert.ResolvedAt)
	
	if err != nil {
		return err
	}

	// Trigger workflows associated with this alert rule (if configured)
	go s.triggerWorkflowsForAlert(context.Background(), alert)
	
	return nil
}

/* triggerWorkflowsForAlert triggers workflows configured for an alert rule */
func (s *Service) triggerWorkflowsForAlert(ctx context.Context, alert Alert) {
	// Get alert rule to check for workflow_id in config
	var configJSON json.RawMessage
	query := `SELECT config FROM neuronip.alert_rules WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, alert.RuleID).Scan(&configJSON)
	if err != nil {
		// Rule not found or no config - skip workflow triggering
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return
	}

	// Check if workflow_id is configured
	workflowIDStr, ok := config["workflow_id"].(string)
	if !ok {
		// No workflow_id configured - skip
		return
	}

	// Parse workflow ID
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return
	}

	// Prepare workflow input with alert details
	input := map[string]interface{}{
		"alert_id":   alert.ID.String(),
		"rule_id":    alert.RuleID.String(),
		"severity":   alert.Severity,
		"message":    alert.Message,
		"details":    alert.Details,
		"created_at": alert.CreatedAt.Format(time.RFC3339),
	}

	// Execute workflow by inserting into workflow_executions directly
	// This avoids circular dependency while still triggering the workflow
	executionID := uuid.New()
	inputJSON, _ := json.Marshal(input)
	now := time.Now()

	// Insert execution record with status 'pending' - workflow service will pick it up
	execQuery := `
		INSERT INTO neuronip.workflow_executions 
		(id, workflow_id, status, input_data, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`

	s.pool.Exec(ctx, execQuery, executionID, workflowID, "pending", inputJSON, now)
}

/* GetAlerts retrieves active alerts */
func (s *Service) GetAlerts(ctx context.Context, status string, limit int) ([]Alert, error) {
	if limit <= 0 {
		limit = 100
	}

	whereClause := ""
	if status != "" {
		whereClause = fmt.Sprintf("WHERE status = '%s'", status)
	}

	query := fmt.Sprintf(`
		SELECT id, rule_id, severity, message, details, status, created_at, resolved_at
		FROM neuronip.alerts
		%s
		ORDER BY created_at DESC
		LIMIT $1`, whereClause)

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		var detailsJSON json.RawMessage

		err := rows.Scan(&alert.ID, &alert.RuleID, &alert.Severity, &alert.Message,
			&detailsJSON, &alert.Status, &alert.CreatedAt, &alert.ResolvedAt)
		if err != nil {
			continue
		}

		if detailsJSON != nil {
			json.Unmarshal(detailsJSON, &alert.Details)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

/* ResolveAlert resolves an alert */
func (s *Service) ResolveAlert(ctx context.Context, alertID uuid.UUID, resolution string) error {
	now := time.Now()
	query := `
		UPDATE neuronip.alerts 
		SET status = 'resolved', resolved_at = $1
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, now, alertID)
	return err
}

/* CreateAlertRule creates a new alert rule */
func (s *Service) CreateAlertRule(ctx context.Context, rule AlertRule) error {
	rule.ID = uuid.New()
	configJSON, _ := json.Marshal(rule.Config)
	
	query := `
		INSERT INTO neuronip.alert_rules (id, name, rule_type, threshold, metric, condition, enabled, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (name) DO UPDATE SET
			rule_type = EXCLUDED.rule_type,
			threshold = EXCLUDED.threshold,
			metric = EXCLUDED.metric,
			condition = EXCLUDED.condition,
			enabled = EXCLUDED.enabled,
			config = EXCLUDED.config,
			updated_at = NOW()`
	
	_, err := s.pool.Exec(ctx, query, rule.ID, rule.Name, rule.RuleType, rule.Threshold,
		rule.Metric, rule.Condition, rule.Enabled, configJSON)
	return err
}

/* GetAlertRule retrieves an alert rule by ID */
func (s *Service) GetAlertRule(ctx context.Context, ruleID uuid.UUID) (*AlertRule, error) {
	query := `
		SELECT id, name, rule_type, threshold, metric, condition, enabled, config
		FROM neuronip.alert_rules
		WHERE id = $1`
	
	var rule AlertRule
	var configJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, ruleID).Scan(&rule.ID, &rule.Name, &rule.RuleType,
		&rule.Threshold, &rule.Metric, &rule.Condition, &rule.Enabled, &configJSON)
	if err != nil {
		return nil, fmt.Errorf("alert rule not found: %w", err)
	}
	
	if configJSON != nil {
		json.Unmarshal(configJSON, &rule.Config)
	}
	
	return &rule, nil
}

/* UpdateAlertRule updates an alert rule */
func (s *Service) UpdateAlertRule(ctx context.Context, rule AlertRule) error {
	configJSON, _ := json.Marshal(rule.Config)
	
	query := `
		UPDATE neuronip.alert_rules
		SET name = $2, rule_type = $3, threshold = $4, metric = $5, condition = $6, enabled = $7, config = $8, updated_at = NOW()
		WHERE id = $1`
	
	_, err := s.pool.Exec(ctx, query, rule.ID, rule.Name, rule.RuleType, rule.Threshold,
		rule.Metric, rule.Condition, rule.Enabled, configJSON)
	return err
}

/* DeleteAlertRule deletes an alert rule */
func (s *Service) DeleteAlertRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `DELETE FROM neuronip.alert_rules WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, ruleID)
	return err
}
