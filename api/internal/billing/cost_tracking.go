package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* CostTrackingService provides cost tracking functionality */
type CostTrackingService struct {
	pool *pgxpool.Pool
}

/* NewCostTrackingService creates a new cost tracking service */
func NewCostTrackingService(pool *pgxpool.Pool) *CostTrackingService {
	return &CostTrackingService{pool: pool}
}

/* RecordCost records a cost event */
func (cts *CostTrackingService) RecordCost(ctx context.Context, cost CostRecord) error {
	cost.ID = uuid.New()
	cost.Timestamp = time.Now()

	metadataJSON, _ := json.Marshal(cost.Metadata)

	query := `
		INSERT INTO neuronip.cost_records 
		(id, resource_type, resource_id, user_id, workspace_id, cost_amount, currency, 
		 category, metadata, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`

	_, err := cts.pool.Exec(ctx, query,
		cost.ID, cost.ResourceType, cost.ResourceID, cost.UserID, cost.WorkspaceID,
		cost.CostAmount, cost.Currency, cost.Category, metadataJSON, cost.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to record cost: %w", err)
	}

	return nil
}

/* GetCostSummary gets cost summary for a time period */
func (cts *CostTrackingService) GetCostSummary(ctx context.Context, userID *string, startTime, endTime time.Time) (*CostSummary, error) {
	baseQuery := `
		SELECT 
			SUM(cost_amount) as total_cost,
			COUNT(*) as record_count,
			resource_type,
			category
		FROM neuronip.cost_records
		WHERE timestamp >= $1 AND timestamp <= $2`
	
	args := []interface{}{startTime, endTime}
	argCount := 3

	if userID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *userID)
		argCount++
	}

	baseQuery += " GROUP BY resource_type, category ORDER BY total_cost DESC"

	rows, err := cts.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost summary: %w", err)
	}
	defer rows.Close()

	summary := &CostSummary{
		StartTime:   startTime,
		EndTime:     endTime,
		Breakdown:   []CostBreakdown{},
		TotalCost:  0,
		RecordCount: 0,
	}

	for rows.Next() {
		var breakdown CostBreakdown
		var totalCost float64
		var recordCount int64

		err := rows.Scan(&totalCost, &recordCount, &breakdown.ResourceType, &breakdown.Category)
		if err != nil {
			continue
		}

		breakdown.TotalCost = totalCost
		breakdown.RecordCount = recordCount
		summary.Breakdown = append(summary.Breakdown, breakdown)
		summary.TotalCost += totalCost
		summary.RecordCount += recordCount
	}

	return summary, nil
}

/* CostRecord represents a cost record */
type CostRecord struct {
	ID          uuid.UUID              `json:"id"`
	ResourceType string                `json:"resource_type"` // "query", "workflow", "model_inference", "storage", "compute"
	ResourceID   *uuid.UUID            `json:"resource_id,omitempty"`
	UserID       *string                `json:"user_id,omitempty"`
	WorkspaceID  *uuid.UUID            `json:"workspace_id,omitempty"`
	CostAmount   float64               `json:"cost_amount"`
	Currency     string                `json:"currency"` // "USD", "EUR", etc.
	Category     string                `json:"category"` // "compute", "storage", "network", "api_calls"
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time             `json:"timestamp"`
}

/* CostSummary represents a cost summary */
type CostSummary struct {
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	TotalCost   float64         `json:"total_cost"`
	RecordCount int64           `json:"record_count"`
	Breakdown   []CostBreakdown `json:"breakdown"`
}

/* CostBreakdown represents cost breakdown by resource type and category */
type CostBreakdown struct {
	ResourceType string  `json:"resource_type"`
	Category     string  `json:"category"`
	TotalCost    float64 `json:"total_cost"`
	RecordCount  int64   `json:"record_count"`
}
