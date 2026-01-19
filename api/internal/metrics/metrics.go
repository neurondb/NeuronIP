package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	// Database connection pool metrics
	dbPoolMaxConns = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_max_conns",
			Help: "Maximum number of database connections in pool",
		},
	)

	dbPoolAcquiredConns = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_acquired_conns",
			Help: "Number of currently acquired database connections",
		},
	)

	dbPoolIdleConns = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_pool_idle_conns",
			Help: "Number of idle database connections in pool",
		},
	)

	// Business metrics
	semanticSearchesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "semantic_searches_total",
			Help: "Total number of semantic searches performed",
		},
	)

	documentsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "documents_created_total",
			Help: "Total number of documents created",
		},
	)

	// Retrieval quality metrics
	semanticSearchQuality = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "semantic_search_quality_score",
			Help:    "Semantic search result quality scores (similarity)",
			Buckets: []float64{0.5, 0.6, 0.7, 0.8, 0.9, 0.95, 1.0},
		},
	)

	// Warehouse query metrics
	warehouseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "warehouse_queries_total",
			Help: "Total number of warehouse queries",
		},
		[]string{"status"},
	)

	warehouseQueryDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "warehouse_query_duration_seconds",
			Help:    "Warehouse query execution duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
	)

	// Agent execution metrics
	agentExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agent_executions_total",
			Help: "Total number of agent executions",
		},
		[]string{"agent_id", "status"},
	)

	agentExecutionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "agent_execution_duration_seconds",
			Help:    "Agent execution duration in seconds",
			Buckets: []float64{1.0, 5.0, 10.0, 30.0, 60.0, 120.0},
		},
	)

	// Workflow metrics
	workflowExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_executions_total",
			Help: "Total number of workflow executions",
		},
		[]string{"workflow_id", "status"},
	)

	workflowExecutionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "workflow_execution_duration_seconds",
			Help:    "Workflow execution duration in seconds",
			Buckets: []float64{5.0, 10.0, 30.0, 60.0, 300.0, 600.0},
		},
	)

	// Compliance metrics
	complianceChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "compliance_checks_total",
			Help: "Total number of compliance checks",
		},
		[]string{"policy_type", "match_status"},
	)

	// Support metrics
	supportTicketsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "support_tickets_total",
			Help: "Total number of support tickets",
		},
		[]string{"status", "priority"},
	)

	supportTicketResolutionTime = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "support_ticket_resolution_time_hours",
			Help:    "Support ticket resolution time in hours",
			Buckets: []float64{1, 6, 12, 24, 48, 72, 168},
		},
	)
)

/* UpdateDBPoolMetrics updates database pool metrics */
func UpdateDBPoolMetrics(maxConns, acquiredConns, idleConns int32) {
	dbPoolMaxConns.Set(float64(maxConns))
	dbPoolAcquiredConns.Set(float64(acquiredConns))
	dbPoolIdleConns.Set(float64(idleConns))
}

/* RecordHTTPRequest records HTTP request metrics */
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	status := http.StatusText(statusCode)
	if status == "" {
		status = "unknown"
	}

	httpRequestsTotal.WithLabelValues(method, path, status).Inc()
	httpRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

/* IncrementSemanticSearches increments the semantic searches counter */
func IncrementSemanticSearches() {
	semanticSearchesTotal.Inc()
}

/* IncrementDocumentsCreated increments the documents created counter */
func IncrementDocumentsCreated() {
	documentsCreatedTotal.Inc()
}

/* RecordSemanticSearchQuality records semantic search quality score */
func RecordSemanticSearchQuality(score float64) {
	semanticSearchQuality.Observe(score)
}

/* IncrementWarehouseQuery increments warehouse query counter */
func IncrementWarehouseQuery(status string) {
	warehouseQueriesTotal.WithLabelValues(status).Inc()
}

/* RecordWarehouseQueryDuration records warehouse query execution duration */
func RecordWarehouseQueryDuration(duration time.Duration) {
	warehouseQueryDuration.Observe(duration.Seconds())
}

/* IncrementAgentExecution increments agent execution counter */
func IncrementAgentExecution(agentID, status string) {
	agentExecutionsTotal.WithLabelValues(agentID, status).Inc()
}

/* RecordAgentExecutionDuration records agent execution duration */
func RecordAgentExecutionDuration(duration time.Duration) {
	agentExecutionDuration.Observe(duration.Seconds())
}

/* IncrementWorkflowExecution increments workflow execution counter */
func IncrementWorkflowExecution(workflowID, status string) {
	workflowExecutionsTotal.WithLabelValues(workflowID, status).Inc()
}

/* RecordWorkflowExecutionDuration records workflow execution duration */
func RecordWorkflowExecutionDuration(duration time.Duration) {
	workflowExecutionDuration.Observe(duration.Seconds())
}

/* IncrementComplianceCheck increments compliance check counter */
func IncrementComplianceCheck(policyType, matchStatus string) {
	complianceChecksTotal.WithLabelValues(policyType, matchStatus).Inc()
}

/* IncrementSupportTicket increments support ticket counter */
func IncrementSupportTicket(status, priority string) {
	supportTicketsTotal.WithLabelValues(status, priority).Inc()
}

/* RecordSupportTicketResolutionTime records support ticket resolution time */
func RecordSupportTicketResolutionTime(duration time.Duration) {
	supportTicketResolutionTime.Observe(duration.Hours())
}

/* Handler returns the Prometheus metrics handler */
func Handler() http.Handler {
	return promhttp.Handler()
}
