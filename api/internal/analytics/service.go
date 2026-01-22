package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides analytics aggregation functionality */
type Service struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
}

/* NewService creates a new analytics service */
func NewService(pool *pgxpool.Pool, neurondbClient *neurondb.Client, mcpClient *mcp.Client) *Service {
	return &Service{
		pool:           pool,
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
	}
}

/* AnalyticsQuery represents an analytics query */
type AnalyticsQuery struct {
	StartDate   *time.Time
	EndDate     *time.Time
	EntityType  string
	UserID      *string
	GroupBy     []string
}

/* TimeSeriesData represents time-series data point */
type TimeSeriesData struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Labels    map[string]interface{} `json:"labels,omitempty"`
}

/* AnalyticsResult represents aggregated analytics data */
type AnalyticsResult struct {
	TimeSeries   []TimeSeriesData       `json:"time_series,omitempty"`
	Aggregates   map[string]interface{} `json:"aggregates,omitempty"`
	Breakdowns   []map[string]interface{} `json:"breakdowns,omitempty"`
}

/* GetSearchAnalytics returns analytics for semantic search */
func (s *Service) GetSearchAnalytics(ctx context.Context, query AnalyticsQuery) (*AnalyticsResult, error) {
	var startDate, endDate time.Time
	if query.StartDate != nil {
		startDate = *query.StartDate
	} else {
		startDate = time.Now().AddDate(0, 0, -30) // Default: last 30 days
	}
	if query.EndDate != nil {
		endDate = *query.EndDate
	} else {
		endDate = time.Now()
	}

	// Count searches over time
	timeSeriesQuery := `
		SELECT DATE_TRUNC('day', created_at) as day, COUNT(*) as count
		FROM neuronip.search_history
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY day
		ORDER BY day`

	rows, err := s.pool.Query(ctx, timeSeriesQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query search analytics: %w", err)
	}
	defer rows.Close()

	var timeSeries []TimeSeriesData
	for rows.Next() {
		var day time.Time
		var count int64

		err := rows.Scan(&day, &count)
		if err != nil {
			continue
		}

		timeSeries = append(timeSeries, TimeSeriesData{
			Timestamp: day,
			Value:     float64(count),
		})
	}

	// Get aggregates
	totalQuery := `
		SELECT COUNT(*) as total, AVG(result_count) as avg_results
		FROM neuronip.search_history
		WHERE created_at >= $1 AND created_at <= $2`

	var total int64
	var avgResults sql.NullFloat64
	s.pool.QueryRow(ctx, totalQuery, startDate, endDate).Scan(&total, &avgResults)

	aggregates := map[string]interface{}{
		"total_searches": total,
	}
	if avgResults.Valid {
		aggregates["avg_results_per_search"] = avgResults.Float64
	}

	return &AnalyticsResult{
		TimeSeries: timeSeries,
		Aggregates: aggregates,
	}, nil
}

/* GetWarehouseAnalytics returns analytics for warehouse queries */
func (s *Service) GetWarehouseAnalytics(ctx context.Context, query AnalyticsQuery) (*AnalyticsResult, error) {
	var startDate, endDate time.Time
	if query.StartDate != nil {
		startDate = *query.StartDate
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if query.EndDate != nil {
		endDate = *query.EndDate
	} else {
		endDate = time.Now()
	}

	// Query execution statistics
	statsQuery := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			AVG(EXTRACT(EPOCH FROM (executed_at - created_at))) as avg_duration_seconds
		FROM neuronip.warehouse_queries
		WHERE created_at >= $1 AND created_at <= $2`

	var total, completed, failed int64
	var avgDuration sql.NullFloat64
	s.pool.QueryRow(ctx, statsQuery, startDate, endDate).Scan(&total, &completed, &failed, &avgDuration)

	aggregates := map[string]interface{}{
		"total_queries":   total,
		"completed":       completed,
		"failed":          failed,
		"success_rate":    0.0,
	}
	if total > 0 {
		aggregates["success_rate"] = float64(completed) / float64(total)
	}
	if avgDuration.Valid {
		aggregates["avg_duration_seconds"] = avgDuration.Float64
	}

	return &AnalyticsResult{
		Aggregates: aggregates,
	}, nil
}

/* GetWorkflowAnalytics returns analytics for workflow executions */
func (s *Service) GetWorkflowAnalytics(ctx context.Context, query AnalyticsQuery) (*AnalyticsResult, error) {
	var startDate, endDate time.Time
	if query.StartDate != nil {
		startDate = *query.StartDate
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if query.EndDate != nil {
		endDate = *query.EndDate
	} else {
		endDate = time.Now()
	}

	statsQuery := `
		SELECT 
			w.name as workflow_name,
			we.status,
			COUNT(*) as count,
			AVG(EXTRACT(EPOCH FROM (we.completed_at - we.started_at))) as avg_duration_seconds
		FROM neuronip.workflow_executions we
		JOIN neuronip.workflows w ON w.id = we.workflow_id
		WHERE we.created_at >= $1 AND we.created_at <= $2
		GROUP BY w.name, we.status`

	rows, err := s.pool.Query(ctx, statsQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow analytics: %w", err)
	}
	defer rows.Close()

	var breakdowns []map[string]interface{}
	for rows.Next() {
		var workflowName, status string
		var count int64
		var avgDuration sql.NullFloat64

		err := rows.Scan(&workflowName, &status, &count, &avgDuration)
		if err != nil {
			continue
		}

		breakdown := map[string]interface{}{
			"workflow_name": workflowName,
			"status":        status,
			"count":         count,
		}
		if avgDuration.Valid {
			breakdown["avg_duration_seconds"] = avgDuration.Float64
		}

		breakdowns = append(breakdowns, breakdown)
	}

	return &AnalyticsResult{
		Breakdowns: breakdowns,
	}, nil
}

/* GetComplianceAnalytics returns analytics for compliance */
func (s *Service) GetComplianceAnalytics(ctx context.Context, query AnalyticsQuery) (*AnalyticsResult, error) {
	var startDate, endDate time.Time
	if query.StartDate != nil {
		startDate = *query.StartDate
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if query.EndDate != nil {
		endDate = *query.EndDate
	} else {
		endDate = time.Now()
	}

	statsQuery := `
		SELECT 
			cp.policy_type,
			cm.status,
			COUNT(*) as count,
			AVG(cm.match_score) as avg_match_score
		FROM neuronip.compliance_matches cm
		JOIN neuronip.compliance_policies cp ON cp.id = cm.policy_id
		WHERE cm.created_at >= $1 AND cm.created_at <= $2
		GROUP BY cp.policy_type, cm.status`

	rows, err := s.pool.Query(ctx, statsQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance analytics: %w", err)
	}
	defer rows.Close()

	var breakdowns []map[string]interface{}
	for rows.Next() {
		var policyType, status string
		var count int64
		var avgScore sql.NullFloat64

		err := rows.Scan(&policyType, &status, &count, &avgScore)
		if err != nil {
			continue
		}

		breakdown := map[string]interface{}{
			"policy_type": policyType,
			"status":      status,
			"count":       count,
		}
		if avgScore.Valid {
			breakdown["avg_match_score"] = avgScore.Float64
		}

		breakdowns = append(breakdowns, breakdown)
	}

	return &AnalyticsResult{
		Breakdowns: breakdowns,
	}, nil
}

/* GetRetrievalQualityMetrics calculates retrieval quality metrics */
func (s *Service) GetRetrievalQualityMetrics(ctx context.Context, query AnalyticsQuery) (map[string]interface{}, error) {
	// Query semantic search results to calculate quality metrics
	qualityQuery := `
		SELECT 
			AVG(
				CASE 
					WHEN result_count > 0 THEN 1.0 
					ELSE 0.0 
				END
			) as recall_rate,
			AVG(result_count) as avg_results_per_query
		FROM neuronip.search_history
		WHERE created_at >= $1 AND created_at <= $2`

	var startDate, endDate time.Time
	if query.StartDate != nil {
		startDate = *query.StartDate
	} else {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if query.EndDate != nil {
		endDate = *query.EndDate
	} else {
		endDate = time.Now()
	}

	var recallRate, avgResults sql.NullFloat64
	err := s.pool.QueryRow(ctx, qualityQuery, startDate, endDate).Scan(&recallRate, &avgResults)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate retrieval quality: %w", err)
	}

	result := map[string]interface{}{}
	if recallRate.Valid {
		result["recall_rate"] = recallRate.Float64
	}
	if avgResults.Valid {
		result["avg_results_per_query"] = avgResults.Float64
	}

	return result, nil
}

/* ClusterCustomers performs customer segmentation using NeuronDB clustering or MCP tools */
func (s *Service) ClusterCustomers(ctx context.Context, tableName string, featureColumns []string, numClusters int, options map[string]interface{}) ([]map[string]interface{}, error) {
	if numClusters <= 0 {
		numClusters = 5
	}
	if options == nil {
		options = make(map[string]interface{})
	}
	algorithm := "kmeans"
	if alg, ok := options["algorithm"].(string); ok {
		algorithm = alg
	}

	// Try MCP tool first if available
	if s.mcpClient != nil {
		result, err := s.mcpClient.ClusterData(ctx, algorithm, tableName, featureColumns, numClusters, options)
		if err == nil {
			if clusters, ok := result["clusters"].([]interface{}); ok {
				results := make([]map[string]interface{}, 0, len(clusters))
				for _, c := range clusters {
					if clusterMap, ok := c.(map[string]interface{}); ok {
						results = append(results, clusterMap)
					}
				}
				return results, nil
			}
			// Return result as-is if format is different
			if len(result) > 0 {
				return []map[string]interface{}{result}, nil
			}
		}
	}

	// Fallback to NeuronDB client
	if s.neurondbClient == nil {
		return nil, fmt.Errorf("neuronDB client not available")
	}
	return s.neurondbClient.ClusterData(ctx, algorithm, tableName, featureColumns, numClusters, options)
}

/* AnalyzeTimeSeries performs time series analysis using NeuronDB */
func (s *Service) AnalyzeTimeSeries(ctx context.Context, tableName string, timeColumn string, valueColumn string, method string, options map[string]interface{}) (map[string]interface{}, error) {
	if s.neurondbClient == nil {
		return nil, fmt.Errorf("neuronDB client not available")
	}
	if method == "" {
		method = "trend"
	}
	if options == nil {
		options = make(map[string]interface{})
	}
	return s.neurondbClient.TimeSeriesAnalysis(ctx, tableName, timeColumn, valueColumn, method, options)
}

/* DiscoverTopics performs topic discovery on text data using MCP */
func (s *Service) DiscoverTopics(ctx context.Context, tableName string, textColumn string, numTopics int, options map[string]interface{}) (map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}
	if numTopics <= 0 {
		numTopics = 10
	}
	if options == nil {
		options = make(map[string]interface{})
	}

	result, err := s.mcpClient.TopicDiscovery(ctx, tableName, textColumn, numTopics, options)
	if err != nil {
		return nil, fmt.Errorf("failed to discover topics: %w", err)
	}

	return result, nil
}

/* AnalyzeData performs comprehensive data analysis using MCP */
func (s *Service) AnalyzeData(ctx context.Context, tableName string, columns []string, analysisType string) (map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}
	if analysisType == "" {
		analysisType = "comprehensive"
	}

	result, err := s.mcpClient.AnalyzeData(ctx, tableName, columns, analysisType)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze data: %w", err)
	}

	return result, nil
}

/* DetectDataDrift detects data drift using MCP */
func (s *Service) DetectDataDrift(ctx context.Context, tableName string, referenceTable string, featureColumns []string) (map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}

	// MCP DetectDrift tool (assuming it exists in the client)
	// Note: This may need to be added to the MCP client if not already present
	result, err := s.mcpClient.ExecuteTool(ctx, "detect_drift", map[string]interface{}{
		"table":            tableName,
		"reference_table": referenceTable,
		"feature_columns": featureColumns,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to detect data drift: %w", err)
	}

	return result, nil
}
