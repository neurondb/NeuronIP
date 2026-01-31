package observability

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/* EnhancedMetricsService provides enhanced metrics collection */
type EnhancedMetricsService struct {
	pool *pgxpool.Pool

	// Prometheus metrics
	requestDuration *prometheus.HistogramVec
	requestTotal    *prometheus.CounterVec
	activeRequests  prometheus.Gauge
	errorRate       *prometheus.CounterVec
	queryDuration   *prometheus.HistogramVec
	queryTotal      *prometheus.CounterVec
}

/* NewEnhancedMetricsService creates a new enhanced metrics service */
func NewEnhancedMetricsService(pool *pgxpool.Pool) *EnhancedMetricsService {
	ems := &EnhancedMetricsService{
		pool: pool,
	}

	// Initialize Prometheus metrics
	ems.requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "neuronip_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	ems.requestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "neuronip_requests_total",
			Help: "Total number of requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	ems.activeRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "neuronip_active_requests",
			Help: "Number of active requests",
		},
	)

	ems.errorRate = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "neuronip_errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "severity"},
	)

	ems.queryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "neuronip_query_duration_seconds",
			Help:    "Query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"query_type", "status"},
	)

	ems.queryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "neuronip_queries_total",
			Help: "Total number of queries",
		},
		[]string{"query_type", "status"},
	)

	return ems
}

/* RecordRequest records a request metric */
func (ems *EnhancedMetricsService) RecordRequest(method, endpoint, status string, duration time.Duration) {
	ems.requestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
	ems.requestTotal.WithLabelValues(method, endpoint, status).Inc()
}

/* IncrementActiveRequests increments active request count */
func (ems *EnhancedMetricsService) IncrementActiveRequests() {
	ems.activeRequests.Inc()
}

/* DecrementActiveRequests decrements active request count */
func (ems *EnhancedMetricsService) DecrementActiveRequests() {
	ems.activeRequests.Dec()
}

/* RecordError records an error */
func (ems *EnhancedMetricsService) RecordError(errorType, severity string) {
	ems.errorRate.WithLabelValues(errorType, severity).Inc()
}

/* RecordQuery records a query metric */
func (ems *EnhancedMetricsService) RecordQuery(queryType, status string, duration time.Duration) {
	ems.queryDuration.WithLabelValues(queryType, status).Observe(duration.Seconds())
	ems.queryTotal.WithLabelValues(queryType, status).Inc()
}

/* GetBusinessMetrics retrieves business metrics */
func (ems *EnhancedMetricsService) GetBusinessMetrics(ctx context.Context, startTime, endTime time.Time) (*BusinessMetrics, error) {
	metrics := &BusinessMetrics{
		PeriodStart: startTime,
		PeriodEnd:   endTime,
	}

	// Get queries per second
	query := `
		SELECT 
			COUNT(*)::float / EXTRACT(EPOCH FROM ($2 - $1)) as queries_per_second,
			AVG(execution_time_ms / 1000.0) as avg_latency_seconds,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY execution_time_ms / 1000.0) as p99_latency_seconds,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY execution_time_ms / 1000.0) as p95_latency_seconds,
			SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END)::float / COUNT(*)::float * 100 as error_rate_percent
		FROM neuronip.warehouse_queries
		WHERE executed_at >= $1 AND executed_at <= $2`

	err := ems.pool.QueryRow(ctx, query, startTime, endTime).Scan(
		&metrics.QueriesPerSecond,
		&metrics.AvgLatency,
		&metrics.P99Latency,
		&metrics.P95Latency,
		&metrics.ErrorRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get business metrics: %w", err)
	}

	// Get resource metrics
	resourceMetrics, err := ems.GetResourceMetrics(ctx)
	if err == nil {
		metrics.CPUUsage = resourceMetrics.CPUUsage
		metrics.MemoryUsage = resourceMetrics.MemoryUsage
		metrics.DiskUsage = resourceMetrics.DiskUsage
		metrics.NetworkIO = resourceMetrics.NetworkIO
	}

	return metrics, nil
}

/* GetResourceMetrics retrieves resource metrics using runtime stats when Prometheus is not configured */
func (ems *EnhancedMetricsService) GetResourceMetrics(ctx context.Context) (*ResourceMetrics, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	// Process memory: heap alloc as fraction of system memory (Sys); cap at 100%
	memoryUsage := 0.0
	if mem.Sys > 0 {
		memoryUsage = float64(mem.HeapAlloc) / float64(mem.Sys) * 100
		if memoryUsage > 100 {
			memoryUsage = 100
		}
	}
	// CPUUsage, DiskUsage, NetworkIO require Prometheus or OS-specific APIs; leave zero until integrated
	_ = runtime.NumCPU() // available CPUs; actual CPU % would need external metrics
	return &ResourceMetrics{
		CPUUsage:    0.0,
		MemoryUsage: memoryUsage,
		DiskUsage:   0.0,
		NetworkIO:   0.0,
		Timestamp:   time.Now(),
	}, nil
}

/* BusinessMetrics represents business metrics */
type BusinessMetrics struct {
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	QueriesPerSecond float64   `json:"queries_per_second"`
	AvgLatency       float64   `json:"avg_latency_seconds"`
	P99Latency       float64   `json:"p99_latency_seconds"`
	P95Latency       float64   `json:"p95_latency_seconds"`
	ErrorRate        float64   `json:"error_rate_percent"`
	CPUUsage         float64   `json:"cpu_usage_percent"`
	MemoryUsage      float64   `json:"memory_usage_percent"`
	DiskUsage        float64   `json:"disk_usage_percent"`
	NetworkIO        float64   `json:"network_io_mbps"`
}

/* ResourceMetrics represents resource metrics */
type ResourceMetrics struct {
	CPUUsage    float64   `json:"cpu_usage_percent"`
	MemoryUsage float64   `json:"memory_usage_percent"`
	DiskUsage   float64   `json:"disk_usage_percent"`
	NetworkIO   float64   `json:"network_io_mbps"`
	Timestamp   time.Time `json:"timestamp"`
}
