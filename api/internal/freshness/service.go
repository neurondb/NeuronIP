package freshness

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides data freshness monitoring functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new freshness service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* FreshnessMonitor represents a freshness monitor */
type FreshnessMonitor struct {
	ID              uuid.UUID              `json:"id"`
	ConnectorID     *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName      string                 `json:"schema_name"`
	TableName       string                 `json:"table_name"`
	TimestampColumn string                 `json:"timestamp_column"`
	ExpectedIntervalMinutes int            `json:"expected_interval_minutes"`
	AlertThreshold  int                    `json:"alert_threshold"` // minutes past expected
	Enabled         bool                   `json:"enabled"`
	LastCheckAt     *time.Time             `json:"last_check_at,omitempty"`
	LastUpdateAt    *time.Time             `json:"last_update_at,omitempty"`
	Status          string                 `json:"status"` // "fresh", "stale", "critical"
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

/* CheckFreshness checks data freshness for a monitor */
func (s *Service) CheckFreshness(ctx context.Context, monitorID uuid.UUID) (*FreshnessMonitor, error) {
	monitor, err := s.GetMonitor(ctx, monitorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	if !monitor.Enabled {
		return nil, fmt.Errorf("monitor is disabled")
	}

	now := time.Now()
	checkQuery := fmt.Sprintf(`
		SELECT MAX(%s) as last_update
		FROM %s.%s`,
		monitor.TimestampColumn, monitor.SchemaName, monitor.TableName)

	var lastUpdate sql.NullTime
	err = s.pool.QueryRow(ctx, checkQuery).Scan(&lastUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to check freshness: %w", err)
	}

	monitor.LastCheckAt = &now

	if lastUpdate.Valid {
		monitor.LastUpdateAt = &lastUpdate.Time
		
		// Calculate age
		ageMinutes := int(now.Sub(lastUpdate.Time).Minutes())
		expectedAge := monitor.ExpectedIntervalMinutes
		alertThreshold := monitor.AlertThreshold
		
		// Determine status
		if ageMinutes <= expectedAge {
			monitor.Status = "fresh"
		} else if ageMinutes <= expectedAge+alertThreshold {
			monitor.Status = "stale"
		} else {
			monitor.Status = "critical"
		}

		// Update monitor
		monitor.UpdatedAt = now
		metadataJSON, _ := json.Marshal(monitor.Metadata)
		
		updateQuery := `
			UPDATE neuronip.freshness_monitors
			SET last_check_at = $1, last_update_at = $2, status = $3, updated_at = $4, metadata = $5
			WHERE id = $6`

		_, err = s.pool.Exec(ctx, updateQuery,
			monitor.LastCheckAt, monitor.LastUpdateAt, monitor.Status,
			monitor.UpdatedAt, metadataJSON, monitor.ID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update monitor: %w", err)
		}
	} else {
		// No data found
		monitor.Status = "critical"
		monitor.UpdatedAt = now
		metadataJSON, _ := json.Marshal(monitor.Metadata)
		
		updateQuery := `
			UPDATE neuronip.freshness_monitors
			SET last_check_at = $1, status = $2, updated_at = $3, metadata = $4
			WHERE id = $5`

		_, err = s.pool.Exec(ctx, updateQuery,
			monitor.LastCheckAt, monitor.Status, monitor.UpdatedAt,
			metadataJSON, monitor.ID,
		)
	}

	return monitor, nil
}

/* CreateMonitor creates a new freshness monitor */
func (s *Service) CreateMonitor(ctx context.Context, monitor FreshnessMonitor) (*FreshnessMonitor, error) {
	monitor.ID = uuid.New()
	monitor.CreatedAt = time.Now()
	monitor.UpdatedAt = time.Now()
	
	if monitor.Status == "" {
		monitor.Status = "fresh"
	}

	metadataJSON, _ := json.Marshal(monitor.Metadata)

	query := `
		INSERT INTO neuronip.freshness_monitors
		(id, connector_id, schema_name, table_name, timestamp_column,
		 expected_interval_minutes, alert_threshold, enabled, status,
		 created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		monitor.ID, monitor.ConnectorID, monitor.SchemaName, monitor.TableName,
		monitor.TimestampColumn, monitor.ExpectedIntervalMinutes,
		monitor.AlertThreshold, monitor.Enabled, monitor.Status,
		monitor.CreatedAt, monitor.UpdatedAt, metadataJSON,
	).Scan(&monitor.ID, &monitor.CreatedAt, &monitor.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create monitor: %w", err)
	}

	return &monitor, nil
}

/* GetMonitor retrieves a freshness monitor by ID */
func (s *Service) GetMonitor(ctx context.Context, id uuid.UUID) (*FreshnessMonitor, error) {
	var monitor FreshnessMonitor
	var connectorID sql.NullString
	var lastCheckAt, lastUpdateAt sql.NullTime
	var metadataJSON []byte

	query := `
		SELECT id, connector_id, schema_name, table_name, timestamp_column,
		       expected_interval_minutes, alert_threshold, enabled,
		       last_check_at, last_update_at, status, created_at, updated_at, metadata
		FROM neuronip.freshness_monitors
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&monitor.ID, &connectorID, &monitor.SchemaName, &monitor.TableName,
		&monitor.TimestampColumn, &monitor.ExpectedIntervalMinutes,
		&monitor.AlertThreshold, &monitor.Enabled,
		&lastCheckAt, &lastUpdateAt, &monitor.Status,
		&monitor.CreatedAt, &monitor.UpdatedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("monitor not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	if connectorID.Valid {
		id, _ := uuid.Parse(connectorID.String)
		monitor.ConnectorID = &id
	}
	if lastCheckAt.Valid {
		monitor.LastCheckAt = &lastCheckAt.Time
	}
	if lastUpdateAt.Valid {
		monitor.LastUpdateAt = &lastUpdateAt.Time
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &monitor.Metadata)
	}

	return &monitor, nil
}

/* ListMonitors lists freshness monitors */
func (s *Service) ListMonitors(ctx context.Context, enabled *bool, status *string) ([]FreshnessMonitor, error) {
	query := `
		SELECT id, connector_id, schema_name, table_name, timestamp_column,
		       expected_interval_minutes, alert_threshold, enabled,
		       last_check_at, last_update_at, status, created_at, updated_at, metadata
		FROM neuronip.freshness_monitors
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if enabled != nil {
		query += fmt.Sprintf(" AND enabled = $%d", argIdx)
		args = append(args, *enabled)
		argIdx++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list monitors: %w", err)
	}
	defer rows.Close()

	var monitors []FreshnessMonitor
	for rows.Next() {
		var monitor FreshnessMonitor
		var connectorID sql.NullString
		var lastCheckAt, lastUpdateAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&monitor.ID, &connectorID, &monitor.SchemaName, &monitor.TableName,
			&monitor.TimestampColumn, &monitor.ExpectedIntervalMinutes,
			&monitor.AlertThreshold, &monitor.Enabled,
			&lastCheckAt, &lastUpdateAt, &monitor.Status,
			&monitor.CreatedAt, &monitor.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			monitor.ConnectorID = &id
		}
		if lastCheckAt.Valid {
			monitor.LastCheckAt = &lastCheckAt.Time
		}
		if lastUpdateAt.Valid {
			monitor.LastUpdateAt = &lastUpdateAt.Time
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &monitor.Metadata)
		}

		monitors = append(monitors, monitor)
	}

	return monitors, nil
}

/* GetFreshnessDashboard retrieves freshness dashboard data */
func (s *Service) GetFreshnessDashboard(ctx context.Context) (map[string]interface{}, error) {
	// Get freshness summary
	var totalMonitors, freshCount, staleCount, criticalCount int

	err := s.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'fresh') as fresh,
			COUNT(*) FILTER (WHERE status = 'stale') as stale,
			COUNT(*) FILTER (WHERE status = 'critical') as critical
		FROM neuronip.freshness_monitors
		WHERE enabled = true`).Scan(&totalMonitors, &freshCount, &staleCount, &criticalCount)

	if err != nil {
		return nil, fmt.Errorf("failed to get freshness summary: %w", err)
	}

	// Get freshness trends (last 7 days)
	trendQuery := `
		SELECT DATE(last_check_at) as date, status, COUNT(*) as count
		FROM neuronip.freshness_monitors
		WHERE last_check_at >= NOW() - INTERVAL '7 days'
		AND enabled = true
		GROUP BY DATE(last_check_at), status
		ORDER BY date ASC`

	rows, err := s.pool.Query(ctx, trendQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get trends: %w", err)
	}
	defer rows.Close()

	trends := []map[string]interface{}{}
	for rows.Next() {
		var date time.Time
		var status string
		var count int

		err := rows.Scan(&date, &status, &count)
		if err != nil {
			continue
		}

		trends = append(trends, map[string]interface{}{
			"date":   date,
			"status": status,
			"count":  count,
		})
	}

	// Get critical monitors
	criticalStatus := "critical"
	criticalMonitors, _ := s.ListMonitors(ctx, nil, &criticalStatus)

	dashboard := map[string]interface{}{
		"summary": map[string]interface{}{
			"total":    totalMonitors,
			"fresh":    freshCount,
			"stale":    staleCount,
			"critical": criticalCount,
		},
		"trends":          trends,
		"critical_monitors": criticalMonitors,
		"freshness_rate":   float64(freshCount) / float64(totalMonitors) * 100,
	}

	return dashboard, nil
}

/* GetFreshnessMetrics retrieves freshness metrics for a monitor */
func (s *Service) GetFreshnessMetrics(ctx context.Context, monitorID uuid.UUID, days int) ([]map[string]interface{}, error) {
	if days == 0 {
		days = 7
	}

	// Get freshness check history
	query := `
		SELECT last_check_at, last_update_at, status,
		       EXTRACT(EPOCH FROM (last_check_at - last_update_at)) / 60 as age_minutes
		FROM neuronip.freshness_monitors
		WHERE id = $1
		AND last_check_at >= NOW() - INTERVAL '%d days'
		ORDER BY last_check_at ASC`

	query = fmt.Sprintf(query, days)

	rows, err := s.pool.Query(ctx, query, monitorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	metrics := []map[string]interface{}{}
	for rows.Next() {
		var checkAt, updateAt *time.Time
		var status string
		var ageMinutes *float64

		err := rows.Scan(&checkAt, &updateAt, &status, &ageMinutes)
		if err != nil {
			continue
		}

		metric := map[string]interface{}{
			"check_at": checkAt,
			"status":   status,
		}

		if updateAt != nil {
			metric["update_at"] = updateAt
		}
		if ageMinutes != nil {
			metric["age_minutes"] = *ageMinutes
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}
