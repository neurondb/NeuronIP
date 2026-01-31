package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* RegionManager manages multi-region support */
type RegionManager struct {
	pool *pgxpool.Pool
}

/* NewRegionManager creates a new region manager */
func NewRegionManager(pool *pgxpool.Pool) *RegionManager {
	return &RegionManager{pool: pool}
}

/* RegisterRegion registers a region */
func (rm *RegionManager) RegisterRegion(ctx context.Context, region Region) error {
	query := `
		INSERT INTO neuronip.regions (region_code, region_name, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (region_code) DO UPDATE SET
			region_name = $2,
			status = $3,
			metadata = $4,
			updated_at = NOW()`

	_, err := rm.pool.Exec(ctx, query, region.RegionCode, region.RegionName, region.Status, region.Metadata)
	return err
}

/* GetRegion retrieves a region */
func (rm *RegionManager) GetRegion(ctx context.Context, regionCode string) (*Region, error) {
	query := `
		SELECT region_code, region_name, status, metadata, created_at, updated_at
		FROM neuronip.regions
		WHERE region_code = $1`

	var region Region
	err := rm.pool.QueryRow(ctx, query, regionCode).Scan(
		&region.RegionCode, &region.RegionName, &region.Status,
		&region.Metadata, &region.CreatedAt, &region.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	return &region, nil
}

/* GetRegions retrieves all regions */
func (rm *RegionManager) GetRegions(ctx context.Context) ([]Region, error) {
	query := `
		SELECT region_code, region_name, status, metadata, created_at, updated_at
		FROM neuronip.regions
		WHERE status = 'active'
		ORDER BY region_code`

	rows, err := rm.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []Region
	for rows.Next() {
		var region Region
		err := rows.Scan(
			&region.RegionCode, &region.RegionName, &region.Status,
			&region.Metadata, &region.CreatedAt, &region.UpdatedAt,
		)
		if err != nil {
			continue
		}
		regions = append(regions, region)
	}

	return regions, nil
}

/* RouteToRegion routes a request to the appropriate region */
func (rm *RegionManager) RouteToRegion(ctx context.Context, routeKey string, strategy string) (*string, error) {
	// Get region based on routing strategy
	var query string
	var args []interface{}

	switch strategy {
	case "hash":
		// Hash-based routing
		query = `
			SELECT region_code
			FROM neuronip.regions
			WHERE status = 'active'
			ORDER BY MOD(hashtext($1), (SELECT COUNT(*) FROM neuronip.regions WHERE status = 'active'))
			LIMIT 1`
		args = []interface{}{routeKey}
	case "locality":
		// Route to same region as user/data
		query = `
			SELECT region_code
			FROM neuronip.regions
			WHERE status = 'active'
			ORDER BY region_code
			LIMIT 1`
		args = []interface{}{}
	default:
		// Default: round-robin or least connections
		query = `
			SELECT region_code
			FROM neuronip.regions
			WHERE status = 'active'
			ORDER BY region_code
			LIMIT 1`
		args = []interface{}{}
	}

	var regionCode string
	err := rm.pool.QueryRow(ctx, query, args...).Scan(&regionCode)
	if err != nil {
		return nil, fmt.Errorf("failed to route to region: %w", err)
	}

	return &regionCode, nil
}

/* ReplicateToRegion replicates data to a region */
func (rm *RegionManager) ReplicateToRegion(ctx context.Context, sourceRegion string, targetRegion string, dataType string) error {
	query := `
		INSERT INTO neuronip.region_replication (source_region, target_region, data_type, status, created_at)
		VALUES ($1, $2, $3, 'pending', NOW())`

	_, err := rm.pool.Exec(ctx, query, sourceRegion, targetRegion, dataType)
	return err
}

/* GetReplicationStatus gets replication status */
func (rm *RegionManager) GetReplicationStatus(ctx context.Context, sourceRegion string, targetRegion string) (*ReplicationStatus, error) {
	query := `
		SELECT source_region, target_region, data_type, status, last_sync_at, created_at
		FROM neuronip.region_replication
		WHERE source_region = $1 AND target_region = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var status ReplicationStatus
	var lastSyncAt *time.Time
	err := rm.pool.QueryRow(ctx, query, sourceRegion, targetRegion).Scan(
		&status.SourceRegion, &status.TargetRegion, &status.DataType,
		&status.Status, &lastSyncAt, &status.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get replication status: %w", err)
	}

	if lastSyncAt != nil {
		status.LastSyncAt = lastSyncAt
	}

	return &status, nil
}

/* Region represents a region */
type Region struct {
	RegionCode string                 `json:"region_code"`
	RegionName string                 `json:"region_name"`
	Status     string                 `json:"status"` // "active", "inactive"
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

/* ReplicationStatus represents replication status */
type ReplicationStatus struct {
	SourceRegion string     `json:"source_region"`
	TargetRegion string     `json:"target_region"`
	DataType     string     `json:"data_type"`
	Status       string     `json:"status"` // "pending", "syncing", "completed", "failed"
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
