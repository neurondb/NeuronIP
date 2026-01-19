package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* BillingService provides billing and usage tracking functionality */
type BillingService struct {
	pool *pgxpool.Pool
}

/* NewBillingService creates a new billing service */
func NewBillingService(pool *pgxpool.Pool) *BillingService {
	return &BillingService{pool: pool}
}

/* UsageMetric represents a usage metric */
type UsageMetric struct {
	ID         uuid.UUID              `json:"id"`
	MetricType string                 `json:"metric_type"`
	MetricName string                 `json:"metric_name"`
	Count      int                    `json:"count"`
	UserID     *string                `json:"user_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

/* BillingRecord represents a billing record */
type BillingRecord struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *string    `json:"user_id,omitempty"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Seats       int       `json:"seats"`
	APICalls    int       `json:"api_calls"`
	Queries     int       `json:"queries"`
	Embeddings  int       `json:"embeddings"`
	Cost        float64   `json:"cost"`
	CreatedAt   time.Time `json:"created_at"`
}

/* TrackUsage tracks a usage event */
func (s *BillingService) TrackUsage(ctx context.Context, metric UsageMetric) error {
	metadataJSON, _ := json.Marshal(metric.Metadata)

	query := `
		INSERT INTO neuronip.usage_metrics (id, metric_type, metric_name, count, user_id, timestamp, metadata)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6)`

	_, err := s.pool.Exec(ctx, query,
		metric.MetricType, metric.MetricName, metric.Count, metric.UserID, metric.Timestamp, metadataJSON,
	)
	return err
}

/* GetUsageMetrics retrieves usage metrics */
func (s *BillingService) GetUsageMetrics(ctx context.Context, metricType, userID string, limit int) ([]UsageMetric, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, metric_type, metric_name, count, user_id, timestamp, metadata
		FROM neuronip.usage_metrics
		WHERE ($1 = '' OR metric_type = $1) AND ($2 = '' OR user_id = $2)
		ORDER BY timestamp DESC
		LIMIT $3`

	rows, err := s.pool.Query(ctx, query, metricType, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage metrics: %w", err)
	}
	defer rows.Close()

	var metrics []UsageMetric
	for rows.Next() {
		var m UsageMetric
		var metadataJSON json.RawMessage

		err := rows.Scan(&m.ID, &m.MetricType, &m.MetricName, &m.Count, &m.UserID, &m.Timestamp, &metadataJSON)
		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &m.Metadata)
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

/* GetDetailedMetrics retrieves detailed metrics summary */
func (s *BillingService) GetDetailedMetrics(ctx context.Context, userID string, periodStart, periodEnd time.Time) (map[string]interface{}, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN metric_type = 'api_call' THEN count ELSE 0 END), 0) as api_calls,
			COALESCE(SUM(CASE WHEN metric_type = 'query' THEN count ELSE 0 END), 0) as queries,
			COALESCE(SUM(CASE WHEN metric_type = 'embedding' THEN count ELSE 0 END), 0) as embeddings,
			COALESCE(SUM(CASE WHEN metric_type = 'seat' THEN count ELSE 0 END), 0) as seats
		FROM neuronip.usage_metrics
		WHERE ($1 = '' OR user_id = $1) AND timestamp BETWEEN $2 AND $3`

	var apiCalls, queries, embeddings, seats int
	err := s.pool.QueryRow(ctx, query, userID, periodStart, periodEnd).Scan(&apiCalls, &queries, &embeddings, &seats)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed metrics: %w", err)
	}

	return map[string]interface{}{
		"seats":      seats,
		"api_calls":  apiCalls,
		"queries":    queries,
		"embeddings": embeddings,
	}, nil
}

/* GetDashboardData retrieves monetization dashboard data */
func (s *BillingService) GetDashboardData(ctx context.Context) (map[string]interface{}, error) {
	// Get total usage
	totalQuery := `
		SELECT 
			COUNT(DISTINCT user_id) as active_users,
			SUM(CASE WHEN metric_type = 'api_call' THEN count ELSE 0 END) as total_api_calls,
			SUM(CASE WHEN metric_type = 'query' THEN count ELSE 0 END) as total_queries,
			SUM(CASE WHEN metric_type = 'embedding' THEN count ELSE 0 END) as total_embeddings
		FROM neuronip.usage_metrics
		WHERE timestamp > NOW() - INTERVAL '30 days'`

	var activeUsers int
	var totalAPICalls, totalQueries, totalEmbeddings *int

	err := s.pool.QueryRow(ctx, totalQuery).Scan(&activeUsers, &totalAPICalls, &totalQueries, &totalEmbeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard data: %w", err)
	}

	result := map[string]interface{}{
		"active_users": activeUsers,
	}

	if totalAPICalls != nil {
		result["total_api_calls"] = *totalAPICalls
	}
	if totalQueries != nil {
		result["total_queries"] = *totalQueries
	}
	if totalEmbeddings != nil {
		result["total_embeddings"] = *totalEmbeddings
	}

	// Get recent billing records
	billingQuery := `
		SELECT id, user_id, period_start, period_end, seats, api_calls, queries, embeddings, cost, created_at
		FROM neuronip.billing_records
		ORDER BY created_at DESC
		LIMIT 10`

	billingRows, err := s.pool.Query(ctx, billingQuery)
	if err == nil {
		defer billingRows.Close()

		var records []BillingRecord
		for billingRows.Next() {
			var record BillingRecord
			billingRows.Scan(
				&record.ID, &record.UserID, &record.PeriodStart, &record.PeriodEnd,
				&record.Seats, &record.APICalls, &record.Queries, &record.Embeddings,
				&record.Cost, &record.CreatedAt,
			)
			records = append(records, record)
		}

		result["recent_billing"] = records
	}

	return result, nil
}
