package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* PoolMonitor monitors connection pool health and metrics */
type PoolMonitor struct {
	pool        *pgxpool.Pool
	metrics     *PoolMetrics
	alerts      []PoolAlert
	mu          sync.RWMutex
	alertThresholds *AlertThresholds
}

/* PoolMetrics tracks pool statistics */
type PoolMetrics struct {
	TotalConns        int32
	AcquiredConns     int32
	IdleConns         int32
	MaxConns          int32
	WaitingConns      int32
	AcquireCount      int64
	AcquireDuration   time.Duration
	AcquiredDuration  time.Duration
	CanceledAcquireCount int64
	ConstructingConns int32
	EmptyAcquireCount int64
	LastUpdated       time.Time
}

/* PoolAlert represents a pool health alert */
type PoolAlert struct {
	Type      AlertType
	Message   string
	Timestamp time.Time
	Severity  AlertSeverity
}

/* AlertType represents the type of alert */
type AlertType string

const (
	AlertTypeExhaustion AlertType = "exhaustion"
	AlertTypeLeak       AlertType = "leak"
	AlertTypeHighUsage  AlertType = "high_usage"
	AlertTypeSlowAcquire AlertType = "slow_acquire"
)

/* AlertSeverity represents alert severity */
type AlertSeverity string

const (
	SeverityWarning AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

/* AlertThresholds defines thresholds for alerts */
type AlertThresholds struct {
	ExhaustionThreshold float64 // Percentage of max connections
	HighUsageThreshold  float64 // Percentage of max connections
	SlowAcquireThreshold time.Duration // Duration threshold for slow acquires
	LeakDetectionWindow time.Duration // Time window for leak detection
}

/* DefaultAlertThresholds returns default alert thresholds */
func DefaultAlertThresholds() *AlertThresholds {
	return &AlertThresholds{
		ExhaustionThreshold: 0.9,  // 90% of max connections
		HighUsageThreshold:  0.75, // 75% of max connections
		SlowAcquireThreshold: 1 * time.Second,
		LeakDetectionWindow: 5 * time.Minute,
	}
}

/* NewPoolMonitor creates a new pool monitor */
func NewPoolMonitor(pool *pgxpool.Pool) *PoolMonitor {
	return &PoolMonitor{
		pool:            pool,
		metrics:         &PoolMetrics{},
		alerts:          make([]PoolAlert, 0),
		alertThresholds: DefaultAlertThresholds(),
	}
}

/* SetAlertThresholds sets custom alert thresholds */
func (pm *PoolMonitor) SetAlertThresholds(thresholds *AlertThresholds) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.alertThresholds = thresholds
}

/* UpdateMetrics updates pool metrics */
func (pm *PoolMonitor) UpdateMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.pool == nil {
		return
	}

	stats := pm.pool.Stat()
	pm.metrics = &PoolMetrics{
		TotalConns:          stats.TotalConns(),
		AcquiredConns:      stats.AcquiredConns(),
		IdleConns:          stats.IdleConns(),
		MaxConns:           stats.MaxConns(),
		AcquireCount:       stats.AcquireCount(),
		AcquireDuration:    stats.AcquireDuration(),
		// AcquiredDuration:   stats.AcquiredDuration(), // Not available in pgxpool.Stat
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:   stats.ConstructingConns(),
		EmptyAcquireCount:   stats.EmptyAcquireCount(),
		LastUpdated:         time.Now(),
	}

	// Check for alerts
	pm.checkAlerts()
}

/* GetMetrics returns current pool metrics */
func (pm *PoolMonitor) GetMetrics() *PoolMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.metrics
}

/* GetAlerts returns current alerts */
func (pm *PoolMonitor) GetAlerts() []PoolAlert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	alerts := make([]PoolAlert, len(pm.alerts))
	copy(alerts, pm.alerts)
	return alerts
}

/* ClearAlerts clears all alerts */
func (pm *PoolMonitor) ClearAlerts() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.alerts = make([]PoolAlert, 0)
}

/* checkAlerts checks for pool health issues */
func (pm *PoolMonitor) checkAlerts() {
	if pm.metrics == nil || pm.alertThresholds == nil {
		return
	}

	// Check for exhaustion
	exhaustionThreshold := float64(pm.metrics.MaxConns) * pm.alertThresholds.ExhaustionThreshold
	if float64(pm.metrics.AcquiredConns) >= exhaustionThreshold {
		pm.addAlert(AlertTypeExhaustion, SeverityCritical,
			fmt.Sprintf("Connection pool near exhaustion: %d/%d connections", 
				pm.metrics.AcquiredConns, pm.metrics.MaxConns))
	}

	// Check for high usage
	highUsageThreshold := float64(pm.metrics.MaxConns) * pm.alertThresholds.HighUsageThreshold
	if float64(pm.metrics.AcquiredConns) >= highUsageThreshold {
		pm.addAlert(AlertTypeHighUsage, SeverityWarning,
			fmt.Sprintf("High connection pool usage: %d/%d connections", 
				pm.metrics.AcquiredConns, pm.metrics.MaxConns))
	}

	// Check for slow acquires
	if pm.metrics.AcquireDuration > pm.alertThresholds.SlowAcquireThreshold {
		pm.addAlert(AlertTypeSlowAcquire, SeverityWarning,
			fmt.Sprintf("Slow connection acquisition: %v", pm.metrics.AcquireDuration))
	}

	// Check for potential leaks (high waiting connections)
	if pm.metrics.WaitingConns > 0 {
		pm.addAlert(AlertTypeLeak, SeverityWarning,
			fmt.Sprintf("Potential connection leak: %d waiting connections", 
				pm.metrics.WaitingConns))
	}
}

/* addAlert adds an alert */
func (pm *PoolMonitor) addAlert(alertType AlertType, severity AlertSeverity, message string) {
	alert := PoolAlert{
		Type:      alertType,
		Message:   message,
		Timestamp: time.Now(),
		Severity:  severity,
	}

	// Remove old alerts of the same type
	for i := len(pm.alerts) - 1; i >= 0; i-- {
		if pm.alerts[i].Type == alertType {
			pm.alerts = append(pm.alerts[:i], pm.alerts[i+1:]...)
		}
	}

	pm.alerts = append(pm.alerts, alert)
}

/* StartMonitoring starts periodic monitoring */
func (pm *PoolMonitor) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.UpdateMetrics()
		}
	}
}

/* DetectLeaks attempts to detect connection leaks */
func (pm *PoolMonitor) DetectLeaks(ctx context.Context) (bool, []string) {
	pm.UpdateMetrics()
	
	leaks := make([]string, 0)
	
	// Check if there are connections that have been acquired for too long
	// This is a simplified check - in production, you'd track individual connections
	if pm.metrics.AcquiredConns > 0 && pm.metrics.AcquiredDuration > pm.alertThresholds.LeakDetectionWindow {
		leaks = append(leaks, fmt.Sprintf("Potential leak: %d connections acquired for %v", 
			pm.metrics.AcquiredConns, pm.metrics.AcquiredDuration))
	}

	return len(leaks) > 0, leaks
}

/* GetPoolStats returns pool statistics for monitoring */
func (pm *PoolMonitor) GetPoolStats() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.metrics == nil {
		return nil
	}

	return map[string]interface{}{
		"total_conns":          pm.metrics.TotalConns,
		"acquired_conns":       pm.metrics.AcquiredConns,
		"idle_conns":          pm.metrics.IdleConns,
		"max_conns":           pm.metrics.MaxConns,
		"acquire_count":       pm.metrics.AcquireCount,
		"acquire_duration":    pm.metrics.AcquireDuration.String(),
		"acquired_duration":   pm.metrics.AcquiredDuration.String(),
		"canceled_acquire_count": pm.metrics.CanceledAcquireCount,
		"constructing_conns":  pm.metrics.ConstructingConns,
		"empty_acquire_count": pm.metrics.EmptyAcquireCount,
		"usage_percentage":    float64(pm.metrics.AcquiredConns) / float64(pm.metrics.MaxConns) * 100,
		"last_updated":        pm.metrics.LastUpdated,
	}
}
