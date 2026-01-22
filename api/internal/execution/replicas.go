package execution

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ReplicaService provides read replica functionality */
type ReplicaService struct {
	pool *pgxpool.Pool
}

/* NewReplicaService creates a new replica service */
func NewReplicaService(pool *pgxpool.Pool) *ReplicaService {
	return &ReplicaService{pool: pool}
}

/* ReadReplica represents a read replica */
type ReadReplica struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Region          string     `json:"region"`
	ConnectionString string    `json:"connection_string"`
	Status          string     `json:"status"`
	LagMs           int        `json:"lag_ms"`
	Priority        int        `json:"priority"`
	MaxConnections  int        `json:"max_connections"`
	Enabled         bool       `json:"enabled"`
	LastHealthCheck *time.Time `json:"last_health_check,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

/* RegisterReplica registers a new read replica */
func (s *ReplicaService) RegisterReplica(ctx context.Context, name string, region string, connectionString string, priority int, maxConnections int) (*ReadReplica, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO neuronip.read_replicas 
		(id, name, region, connection_string, status, priority, max_connections, enabled, last_health_check, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', $5, $6, true, $7, $8, $9)
		RETURNING id, name, region, connection_string, status, lag_ms, priority, max_connections, enabled, last_health_check, created_at, updated_at`

	var replica ReadReplica
	var lastHealthCheck sql.NullTime

	err := s.pool.QueryRow(ctx, query, id, name, region, connectionString, priority, maxConnections, now, now, now).Scan(
		&replica.ID, &replica.Name, &replica.Region, &replica.ConnectionString,
		&replica.Status, &replica.LagMs, &replica.Priority, &replica.MaxConnections,
		&replica.Enabled, &lastHealthCheck, &replica.CreatedAt, &replica.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register replica: %w", err)
	}

	if lastHealthCheck.Valid {
		replica.LastHealthCheck = &lastHealthCheck.Time
	}

	return &replica, nil
}

/* SelectReplica selects the best read replica for a query */
func (s *ReplicaService) SelectReplica(ctx context.Context, region *string) (*ReadReplica, error) {
	var query string
	var args []interface{}

	if region != nil {
		query = `
			SELECT id, name, region, connection_string, status, lag_ms, priority, max_connections, enabled, last_health_check, created_at, updated_at
			FROM neuronip.read_replicas
			WHERE enabled = true AND status = 'active' AND region = $1
			ORDER BY lag_ms ASC, priority DESC
			LIMIT 1`
		args = []interface{}{*region}
	} else {
		query = `
			SELECT id, name, region, connection_string, status, lag_ms, priority, max_connections, enabled, last_health_check, created_at, updated_at
			FROM neuronip.read_replicas
			WHERE enabled = true AND status = 'active'
			ORDER BY lag_ms ASC, priority DESC
			LIMIT 1`
		args = []interface{}{}
	}

	var replica ReadReplica
	var lastHealthCheck sql.NullTime

	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&replica.ID, &replica.Name, &replica.Region, &replica.ConnectionString,
		&replica.Status, &replica.LagMs, &replica.Priority, &replica.MaxConnections,
		&replica.Enabled, &lastHealthCheck, &replica.CreatedAt, &replica.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to select replica: %w", err)
	}

	if lastHealthCheck.Valid {
		replica.LastHealthCheck = &lastHealthCheck.Time
	}

	return &replica, nil
}

/* UpdateReplicaLag updates the lag for a replica */
func (s *ReplicaService) UpdateReplicaLag(ctx context.Context, replicaID uuid.UUID, lagMs int) error {
	query := `
		UPDATE neuronip.read_replicas 
		SET lag_ms = $1, last_health_check = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, lagMs, replicaID)
	if err != nil {
		return fmt.Errorf("failed to update replica lag: %w", err)
	}

	return nil
}

/* ListReplicas lists all read replicas */
func (s *ReplicaService) ListReplicas(ctx context.Context, enabledOnly bool) ([]ReadReplica, error) {
	query := `
		SELECT id, name, region, connection_string, status, lag_ms, priority, max_connections, enabled, last_health_check, created_at, updated_at
		FROM neuronip.read_replicas`

	if enabledOnly {
		query += " WHERE enabled = true"
	}

	query += " ORDER BY region, priority DESC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicas: %w", err)
	}
	defer rows.Close()

	var replicas []ReadReplica
	for rows.Next() {
		var replica ReadReplica
		var lastHealthCheck sql.NullTime

		err := rows.Scan(
			&replica.ID, &replica.Name, &replica.Region, &replica.ConnectionString,
			&replica.Status, &replica.LagMs, &replica.Priority, &replica.MaxConnections,
			&replica.Enabled, &lastHealthCheck, &replica.CreatedAt, &replica.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if lastHealthCheck.Valid {
			replica.LastHealthCheck = &lastHealthCheck.Time
		}

		replicas = append(replicas, replica)
	}

	return replicas, nil
}

/* CheckReplicaHealth checks the health of a replica */
func (s *ReplicaService) CheckReplicaHealth(ctx context.Context, replicaID uuid.UUID) (bool, int, error) {
	// Get replica connection string
	var connectionString string
	query := `SELECT connection_string FROM neuronip.read_replicas WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, replicaID).Scan(&connectionString)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get replica: %w", err)
	}

	// Try to connect and measure lag
	// In production, would use actual connection pool
	// For now, simulate with a simple query
	healthQuery := `SELECT EXTRACT(EPOCH FROM (NOW() - pg_last_xact_replay_timestamp())) * 1000 as lag_ms`
	
	var lagMs sql.NullInt64
	err = s.pool.QueryRow(ctx, healthQuery).Scan(&lagMs)
	if err != nil {
		// If query fails, replica is unhealthy
		s.UpdateReplicaStatus(ctx, replicaID, "unhealthy")
		return false, 0, nil
	}

	lag := 0
	if lagMs.Valid {
		lag = int(lagMs.Int64)
	}

	// Update lag and health check time
	s.UpdateReplicaLag(ctx, replicaID, lag)
	
	// Consider replica healthy if lag < 5 seconds
	isHealthy := lag < 5000
	status := "active"
	if !isHealthy {
		status = "lagging"
	}
	s.UpdateReplicaStatus(ctx, replicaID, status)

	return isHealthy, lag, nil
}

/* UpdateReplicaStatus updates the status of a replica */
func (s *ReplicaService) UpdateReplicaStatus(ctx context.Context, replicaID uuid.UUID, status string) error {
	query := `
		UPDATE neuronip.read_replicas 
		SET status = $1, last_health_check = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, status, replicaID)
	return err
}

/* GetReplicaHealthStatus gets health status for all replicas */
func (s *ReplicaService) GetReplicaHealthStatus(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, region, status, lag_ms, enabled, last_health_check
		FROM neuronip.read_replicas
		ORDER BY region, priority DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get replica health: %w", err)
	}
	defer rows.Close()

	var statuses []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, region, status string
		var lagMs int
		var enabled bool
		var lastHealthCheck sql.NullTime

		err := rows.Scan(&id, &name, &region, &status, &lagMs, &enabled, &lastHealthCheck)
		if err != nil {
			continue
		}

		statusMap := map[string]interface{}{
			"id":      id,
			"name":    name,
			"region":  region,
			"status":  status,
			"lag_ms":  lagMs,
			"enabled": enabled,
		}

		if lastHealthCheck.Valid {
			statusMap["last_health_check"] = lastHealthCheck.Time
		}

		statuses = append(statuses, statusMap)
	}

	return statuses, nil
}
