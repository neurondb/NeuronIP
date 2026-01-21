package metrics

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* MetricsCollector collects and aggregates metrics */
type MetricsCollector struct {
	pool *pgxpool.Pool
}

/* NewMetricsCollector creates a new metrics collector */
func NewMetricsCollector(pool *pgxpool.Pool) *MetricsCollector {
	return &MetricsCollector{pool: pool}
}

/* LatencyMetrics represents latency percentiles */
type LatencyMetrics struct {
	P50  float64 `json:"p50"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	P999 float64 `json:"p999"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Avg  float64 `json:"avg"`
}

/* CalculatePercentiles calculates percentiles from a slice of values */
func CalculatePercentiles(values []float64) LatencyMetrics {
	if len(values) == 0 {
		return LatencyMetrics{}
	}

	sort.Float64s(values)
	n := len(values)

	return LatencyMetrics{
		P50:  values[int(float64(n)*0.50)],
		P95:  values[int(float64(n)*0.95)],
		P99:  values[int(float64(n)*0.99)],
		P999: values[int(float64(n)*0.999)],
		Min:  values[0],
		Max:  values[n-1],
		Avg:  calculateAverage(values),
	}
}

/* calculateAverage calculates average of values */
func calculateAverage(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

/* RecordLatency records latency for an endpoint */
func (c *MetricsCollector) RecordLatency(ctx context.Context, endpoint string, latencyMs float64) error {
	query := `
		INSERT INTO neuronip.latency_metrics (endpoint, latency_ms, recorded_at)
		VALUES ($1, $2, NOW())
	`
	_, err := c.pool.Exec(ctx, query, endpoint, latencyMs)
	return err
}

/* GetLatencyMetrics gets latency metrics for an endpoint */
func (c *MetricsCollector) GetLatencyMetrics(ctx context.Context, endpoint string, startTime, endTime time.Time) (*LatencyMetrics, error) {
	query := `
		SELECT latency_ms
		FROM neuronip.latency_metrics
		WHERE endpoint = $1 AND recorded_at BETWEEN $2 AND $3
		ORDER BY recorded_at DESC
	`
	rows, err := c.pool.Query(ctx, query, endpoint, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var latencies []float64
	for rows.Next() {
		var latency float64
		if err := rows.Scan(&latency); err == nil {
			latencies = append(latencies, latency)
		}
	}

	metrics := CalculatePercentiles(latencies)
	return &metrics, nil
}

/* RecordError records an error occurrence */
func (c *MetricsCollector) RecordError(ctx context.Context, endpoint string, statusCode int, errorType string) error {
	query := `
		INSERT INTO neuronip.error_metrics (endpoint, status_code, error_type, recorded_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := c.pool.Exec(ctx, query, endpoint, statusCode, errorType)
	return err
}

/* GetErrorRate gets error rate for an endpoint */
func (c *MetricsCollector) GetErrorRate(ctx context.Context, endpoint string, startTime, endTime time.Time) (float64, error) {
	query := `
		SELECT
			COUNT(CASE WHEN status_code >= 400 THEN 1 END)::float / NULLIF(COUNT(*), 0) as error_rate
		FROM neuronip.error_metrics
		WHERE endpoint = $1 AND recorded_at BETWEEN $2 AND $3
	`
	var errorRate *float64
	err := c.pool.QueryRow(ctx, query, endpoint, startTime, endTime).Scan(&errorRate)
	if err != nil || errorRate == nil {
		return 0.0, err
	}
	return *errorRate, nil
}

/* RecordTokenUsage records token usage */
func (c *MetricsCollector) RecordTokenUsage(ctx context.Context, endpoint, userID string, tokens int, cost float64) error {
	query := `
		INSERT INTO neuronip.token_usage (endpoint, user_id, tokens, cost_usd, recorded_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := c.pool.Exec(ctx, query, endpoint, userID, tokens, cost)
	return err
}

/* GetTokenUsage gets token usage metrics */
func (c *MetricsCollector) GetTokenUsage(ctx context.Context, userID *string, startTime, endTime time.Time) (map[string]interface{}, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT
				SUM(tokens) as total_tokens,
				SUM(cost_usd) as total_cost,
				AVG(tokens) as avg_tokens_per_request
			FROM neuronip.token_usage
			WHERE user_id = $1 AND recorded_at BETWEEN $2 AND $3
		`
		args = []interface{}{*userID, startTime, endTime}
	} else {
		query = `
			SELECT
				SUM(tokens) as total_tokens,
				SUM(cost_usd) as total_cost,
				AVG(tokens) as avg_tokens_per_request
			FROM neuronip.token_usage
			WHERE recorded_at BETWEEN $1 AND $2
		`
		args = []interface{}{startTime, endTime}
	}

	var totalTokens, totalCost *int
	var avgTokens *float64
	err := c.pool.QueryRow(ctx, query, args...).Scan(&totalTokens, &totalCost, &avgTokens)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	if totalTokens != nil {
		result["total_tokens"] = *totalTokens
	}
	if totalCost != nil {
		result["total_cost_usd"] = *totalCost
	}
	if avgTokens != nil {
		result["avg_tokens_per_request"] = *avgTokens
	}

	return result, nil
}

/* RecordEmbeddingCost records embedding generation cost */
func (c *MetricsCollector) RecordEmbeddingCost(ctx context.Context, model string, tokens int, cost float64) error {
	query := `
		INSERT INTO neuronip.embedding_costs (model_name, tokens, cost_usd, recorded_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := c.pool.Exec(ctx, query, model, tokens, cost)
	return err
}

/* GetEmbeddingCost gets embedding cost metrics */
func (c *MetricsCollector) GetEmbeddingCost(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			model_name,
			SUM(tokens) as total_tokens,
			SUM(cost_usd) as total_cost,
			COUNT(*) as request_count
		FROM neuronip.embedding_costs
		WHERE recorded_at BETWEEN $1 AND $2
		GROUP BY model_name
	`
	rows, err := c.pool.Query(ctx, query, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]interface{})
	for rows.Next() {
		var model string
		var totalTokens, requestCount int
		var totalCost float64
		if err := rows.Scan(&model, &totalTokens, &totalCost, &requestCount); err == nil {
			results[model] = map[string]interface{}{
				"total_tokens":   totalTokens,
				"total_cost_usd": totalCost,
				"request_count":  requestCount,
			}
		}
	}

	return results, nil
}
