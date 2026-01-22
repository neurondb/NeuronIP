package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Error rate metrics
	errorRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "error_rate_percentage",
			Help: "Error rate as percentage of total requests",
		},
		[]string{"endpoint", "error_type"},
	)

	// Query per second metrics
	queriesPerSecond = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queries_per_second",
			Help: "Number of queries per second",
		},
		[]string{"query_type"},
	)

	// Alert thresholds
	alertThresholds = make(map[string]float64)
)

/* RecordErrorRate records error rate for an endpoint */
func RecordErrorRate(endpoint, errorType string, rate float64) {
	errorRate.WithLabelValues(endpoint, errorType).Set(rate)
}

/* RecordQueriesPerSecond records queries per second */
func RecordQueriesPerSecond(queryType string, qps float64) {
	queriesPerSecond.WithLabelValues(queryType).Set(qps)
}

/* SetAlertThreshold sets an alert threshold */
func SetAlertThreshold(metricName string, threshold float64) {
	alertThresholds[metricName] = threshold
}

/* GetAlertThreshold gets an alert threshold */
func GetAlertThreshold(metricName string) (float64, bool) {
	threshold, exists := alertThresholds[metricName]
	return threshold, exists
}

/* CheckAlert checks if a metric value exceeds its threshold */
func CheckAlert(metricName string, value float64) bool {
	threshold, exists := alertThresholds[metricName]
	if !exists {
		return false
	}
	return value > threshold
}
