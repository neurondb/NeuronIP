package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AutoScaler manages auto-scaling policies and events */
type AutoScaler struct {
	pool *pgxpool.Pool
}

/* NewAutoScaler creates a new auto-scaler */
func NewAutoScaler(pool *pgxpool.Pool) *AutoScaler {
	return &AutoScaler{pool: pool}
}

/* CreatePolicy creates a new auto-scaling policy */
func (as *AutoScaler) CreatePolicy(ctx context.Context, policy AutoScalingPolicy) (*AutoScalingPolicy, error) {
	policy.ID = uuid.New()
	metadataJSON, _ := json.Marshal(policy.Metadata)

	query := `
		INSERT INTO neuronip.auto_scaling_policies 
		(id, policy_name, resource_type, min_instances, max_instances, target_metric, target_value,
		 scale_up_threshold, scale_down_threshold, cooldown_seconds, enabled, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING id, policy_name, resource_type, min_instances, max_instances, target_metric, target_value,
		          scale_up_threshold, scale_down_threshold, cooldown_seconds, enabled, metadata, created_at, updated_at`

	var metadataJSONRaw json.RawMessage
	err := as.pool.QueryRow(ctx, query,
		policy.ID, policy.PolicyName, policy.ResourceType, policy.MinInstances, policy.MaxInstances,
		policy.TargetMetric, policy.TargetValue, policy.ScaleUpThreshold, policy.ScaleDownThreshold,
		policy.CooldownSeconds, policy.Enabled, metadataJSON,
	).Scan(
		&policy.ID, &policy.PolicyName, &policy.ResourceType, &policy.MinInstances, &policy.MaxInstances,
		&policy.TargetMetric, &policy.TargetValue, &policy.ScaleUpThreshold, &policy.ScaleDownThreshold,
		&policy.CooldownSeconds, &policy.Enabled, &metadataJSONRaw, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	if metadataJSONRaw != nil {
		json.Unmarshal(metadataJSONRaw, &policy.Metadata)
	}

	return &policy, nil
}

/* EvaluatePolicy evaluates a policy and determines if scaling is needed */
func (as *AutoScaler) EvaluatePolicy(ctx context.Context, policyID uuid.UUID) (*ScalingDecision, error) {
	// Get policy
	policy, err := as.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, err
	}

	if !policy.Enabled {
		return &ScalingDecision{
			Action: "no_action",
			Reason: "policy disabled",
		}, nil
	}

	// Check if we're in cooldown period
	lastEvent, err := as.getLastScalingEvent(ctx, policyID)
	if err == nil && lastEvent != nil {
		cooldownEnd := lastEvent.CreatedAt.Add(time.Duration(policy.CooldownSeconds) * time.Second)
		if time.Now().Before(cooldownEnd) {
			return &ScalingDecision{
				Action: "no_action",
				Reason: fmt.Sprintf("in cooldown period until %v", cooldownEnd),
			}, nil
		}
	}

	// Get current metric value
	currentValue, currentInstances, err := as.getCurrentMetric(ctx, policy.ResourceType, policy.TargetMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to get current metric: %w", err)
	}

	decision := &ScalingDecision{
		PolicyID:          policyID,
		CurrentInstances:  currentInstances,
		CurrentMetricValue: currentValue,
		TargetMetric:     policy.TargetMetric,
	}

	// Evaluate scaling conditions
	if currentValue >= policy.ScaleUpThreshold && currentInstances < policy.MaxInstances {
		decision.Action = "scale_up"
		decision.TargetInstances = currentInstances + 1
		decision.Reason = fmt.Sprintf("metric %s (%.2f) >= threshold (%.2f)", policy.TargetMetric, currentValue, policy.ScaleUpThreshold)
	} else if currentValue <= policy.ScaleDownThreshold && currentInstances > policy.MinInstances {
		decision.Action = "scale_down"
		decision.TargetInstances = currentInstances - 1
		decision.Reason = fmt.Sprintf("metric %s (%.2f) <= threshold (%.2f)", policy.TargetMetric, currentValue, policy.ScaleDownThreshold)
	} else {
		decision.Action = "no_action"
		decision.Reason = "metric within acceptable range"
	}

	return decision, nil
}

/* RecordScalingEvent records a scaling event */
func (as *AutoScaler) RecordScalingEvent(ctx context.Context, event ScalingEvent) (*ScalingEvent, error) {
	event.ID = uuid.New()
	query := `
		INSERT INTO neuronip.auto_scaling_events 
		(id, policy_id, action, current_instances, target_instances, metric_value, reason, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, policy_id, action, current_instances, target_instances, metric_value, reason, status, created_at, completed_at`

	var completedAt *time.Time
	err := as.pool.QueryRow(ctx, query,
		event.ID, event.PolicyID, event.Action, event.CurrentInstances, event.TargetInstances,
		event.MetricValue, event.Reason, event.Status,
	).Scan(
		&event.ID, &event.PolicyID, &event.Action, &event.CurrentInstances, &event.TargetInstances,
		&event.MetricValue, &event.Reason, &event.Status, &event.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record scaling event: %w", err)
	}

	if completedAt != nil {
		event.CompletedAt = completedAt
	}

	return &event, nil
}

/* GetPolicy retrieves a policy */
func (as *AutoScaler) GetPolicy(ctx context.Context, policyID uuid.UUID) (*AutoScalingPolicy, error) {
	query := `
		SELECT id, policy_name, resource_type, min_instances, max_instances, target_metric, target_value,
		       scale_up_threshold, scale_down_threshold, cooldown_seconds, enabled, metadata, created_at, updated_at
		FROM neuronip.auto_scaling_policies
		WHERE id = $1`

	var policy AutoScalingPolicy
	var metadataJSON json.RawMessage
	err := as.pool.QueryRow(ctx, query, policyID).Scan(
		&policy.ID, &policy.PolicyName, &policy.ResourceType, &policy.MinInstances, &policy.MaxInstances,
		&policy.TargetMetric, &policy.TargetValue, &policy.ScaleUpThreshold, &policy.ScaleDownThreshold,
		&policy.CooldownSeconds, &policy.Enabled, &metadataJSON, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &policy.Metadata)
	}

	return &policy, nil
}

/* getCurrentMetric gets the current metric value for a resource type */
func (as *AutoScaler) getCurrentMetric(ctx context.Context, resourceType string, metricName string) (float64, int, error) {
	// Get current instance count
	instanceQuery := `
		SELECT COUNT(*)
		FROM neuronip.cluster_nodes
		WHERE node_type = $1 AND status = 'active' AND last_heartbeat > NOW() - INTERVAL '30 seconds'`

	var instanceCount int
	err := as.pool.QueryRow(ctx, instanceQuery, resourceType).Scan(&instanceCount)
	if err != nil {
		return 0, 0, err
	}

	// Get latest metric value
	metricQuery := `
		SELECT metric_value
		FROM neuronip.cluster_metrics
		WHERE metric_name = $1
			AND timestamp > NOW() - INTERVAL '5 minutes'
		ORDER BY timestamp DESC
		LIMIT 1`

	var metricValue float64
	err = as.pool.QueryRow(ctx, metricQuery, metricName).Scan(&metricValue)
	if err != nil {
		// If no metric found, return 0
		metricValue = 0
	}

	return metricValue, instanceCount, nil
}

/* getLastScalingEvent gets the last scaling event for a policy */
func (as *AutoScaler) getLastScalingEvent(ctx context.Context, policyID uuid.UUID) (*ScalingEvent, error) {
	query := `
		SELECT id, policy_id, action, current_instances, target_instances, metric_value, reason, status, created_at, completed_at
		FROM neuronip.auto_scaling_events
		WHERE policy_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var event ScalingEvent
	var completedAt *time.Time
	err := as.pool.QueryRow(ctx, query, policyID).Scan(
		&event.ID, &event.PolicyID, &event.Action, &event.CurrentInstances, &event.TargetInstances,
		&event.MetricValue, &event.Reason, &event.Status, &event.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	if completedAt != nil {
		event.CompletedAt = completedAt
	}

	return &event, nil
}

/* AutoScalingPolicy represents an auto-scaling policy */
type AutoScalingPolicy struct {
	ID                 uuid.UUID              `json:"id"`
	PolicyName         string                 `json:"policy_name"`
	ResourceType       string                 `json:"resource_type"`
	MinInstances       int                    `json:"min_instances"`
	MaxInstances       int                    `json:"max_instances"`
	TargetMetric       string                 `json:"target_metric"`
	TargetValue        float64                `json:"target_value"`
	ScaleUpThreshold   float64                `json:"scale_up_threshold"`
	ScaleDownThreshold float64                `json:"scale_down_threshold"`
	CooldownSeconds    int                    `json:"cooldown_seconds"`
	Enabled            bool                   `json:"enabled"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

/* ScalingEvent represents a scaling event */
type ScalingEvent struct {
	ID               uuid.UUID  `json:"id"`
	PolicyID         uuid.UUID  `json:"policy_id"`
	Action           string     `json:"action"`
	CurrentInstances int        `json:"current_instances"`
	TargetInstances  int        `json:"target_instances"`
	MetricValue      float64    `json:"metric_value"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

/* ScalingDecision represents a scaling decision */
type ScalingDecision struct {
	PolicyID          uuid.UUID `json:"policy_id"`
	Action            string    `json:"action"` // "scale_up", "scale_down", "no_action"
	CurrentInstances  int       `json:"current_instances"`
	TargetInstances   int       `json:"target_instances,omitempty"`
	CurrentMetricValue float64   `json:"current_metric_value"`
	TargetMetric      string    `json:"target_metric"`
	Reason            string    `json:"reason"`
}
