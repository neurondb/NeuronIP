package tenancy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Region represents a deployment region */
type Region struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Code          string                 `json:"code"` // e.g., "us-east-1", "eu-west-1"
	Primary       bool                   `json:"primary"`
	Active        bool                   `json:"active"`
	Endpoint      string                 `json:"endpoint"`
	DatabaseHost  string                 `json:"database_host,omitempty"`
	DatabasePort  int                    `json:"database_port,omitempty"`
	ReplicaOf     *uuid.UUID             `json:"replica_of,omitempty"`
	LastSyncAt    *time.Time             `json:"last_sync_at,omitempty"`
	HealthStatus  string                 `json:"health_status"` // "healthy", "degraded", "down"
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* RegionService provides multi-region deployment functionality */
type RegionService struct {
	pool *pgxpool.Pool
}

/* NewRegionService creates a new region service */
func NewRegionService(pool *pgxpool.Pool) *RegionService {
	return &RegionService{pool: pool}
}

/* CreateRegion creates a new region */
func (s *RegionService) CreateRegion(ctx context.Context, region Region) (*Region, error) {
	region.ID = uuid.New()
	region.CreatedAt = time.Now()
	region.UpdatedAt = time.Now()
	region.HealthStatus = "healthy"

	if region.HealthStatus == "" {
		region.HealthStatus = "healthy"
	}

	metadataJSON, _ := json.Marshal(region.Metadata)

	var replicaOf sql.NullString
	if region.ReplicaOf != nil {
		replicaOf = sql.NullString{String: region.ReplicaOf.String(), Valid: true}
	}

	var lastSyncAt sql.NullTime
	if region.LastSyncAt != nil {
		lastSyncAt = sql.NullTime{Time: *region.LastSyncAt, Valid: true}
	}

	query := `
		INSERT INTO neuronip.regions
		(id, name, code, primary_region, active, endpoint, database_host, database_port,
		 replica_of, last_sync_at, health_status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		region.ID, region.Name, region.Code, region.Primary, region.Active,
		region.Endpoint, region.DatabaseHost, region.DatabasePort,
		replicaOf, lastSyncAt, region.HealthStatus, metadataJSON,
		region.CreatedAt, region.UpdatedAt,
	).Scan(&region.ID, &region.CreatedAt, &region.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}

	return &region, nil
}

/* GetRegion gets a region by ID */
func (s *RegionService) GetRegion(ctx context.Context, regionID uuid.UUID) (*Region, error) {
	var region Region
	var replicaOf sql.NullString
	var lastSyncAt sql.NullTime
	var metadataJSON []byte

	query := `
		SELECT id, name, code, primary_region, active, endpoint, database_host, database_port,
		       replica_of, last_sync_at, health_status, metadata, created_at, updated_at
		FROM neuronip.regions
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, regionID).Scan(
		&region.ID, &region.Name, &region.Code, &region.Primary, &region.Active,
		&region.Endpoint, &region.DatabaseHost, &region.DatabasePort,
		&replicaOf, &lastSyncAt, &region.HealthStatus, &metadataJSON,
		&region.CreatedAt, &region.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}

	if replicaOf.Valid {
		replicaUUID, _ := uuid.Parse(replicaOf.String)
		region.ReplicaOf = &replicaUUID
	}
	if lastSyncAt.Valid {
		region.LastSyncAt = &lastSyncAt.Time
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &region.Metadata)
	}

	return &region, nil
}

/* ListRegions lists all regions */
func (s *RegionService) ListRegions(ctx context.Context, activeOnly bool) ([]Region, error) {
	query := `
		SELECT id, name, code, primary_region, active, endpoint, database_host, database_port,
		       replica_of, last_sync_at, health_status, metadata, created_at, updated_at
		FROM neuronip.regions`

	if activeOnly {
		query += " WHERE active = true"
	}

	query += " ORDER BY primary_region DESC, name ASC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list regions: %w", err)
	}
	defer rows.Close()

	regions := []Region{}
	for rows.Next() {
		var region Region
		var replicaOf sql.NullString
		var lastSyncAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&region.ID, &region.Name, &region.Code, &region.Primary, &region.Active,
			&region.Endpoint, &region.DatabaseHost, &region.DatabasePort,
			&replicaOf, &lastSyncAt, &region.HealthStatus, &metadataJSON,
			&region.CreatedAt, &region.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if replicaOf.Valid {
			replicaUUID, _ := uuid.Parse(replicaOf.String)
			region.ReplicaOf = &replicaUUID
		}
		if lastSyncAt.Valid {
			region.LastSyncAt = &lastSyncAt.Time
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &region.Metadata)
		}

		regions = append(regions, region)
	}

	return regions, nil
}

/* GetPrimaryRegion gets the primary region */
func (s *RegionService) GetPrimaryRegion(ctx context.Context) (*Region, error) {
	var region Region
	var replicaOf sql.NullString
	var lastSyncAt sql.NullTime
	var metadataJSON []byte

	query := `
		SELECT id, name, code, primary_region, active, endpoint, database_host, database_port,
		       replica_of, last_sync_at, health_status, metadata, created_at, updated_at
		FROM neuronip.regions
		WHERE primary_region = true AND active = true
		LIMIT 1`

	err := s.pool.QueryRow(ctx, query).Scan(
		&region.ID, &region.Name, &region.Code, &region.Primary, &region.Active,
		&region.Endpoint, &region.DatabaseHost, &region.DatabasePort,
		&replicaOf, &lastSyncAt, &region.HealthStatus, &metadataJSON,
		&region.CreatedAt, &region.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("primary region not found: %w", err)
	}

	if replicaOf.Valid {
		replicaUUID, _ := uuid.Parse(replicaOf.String)
		region.ReplicaOf = &replicaUUID
	}
	if lastSyncAt.Valid {
		region.LastSyncAt = &lastSyncAt.Time
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &region.Metadata)
	}

	return &region, nil
}

/* UpdateRegionHealth updates region health status */
func (s *RegionService) UpdateRegionHealth(ctx context.Context, regionID uuid.UUID, status string) error {
	query := `
		UPDATE neuronip.regions
		SET health_status = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, status, regionID)
	return err
}

/* CheckRegionHealth checks the health of a region */
func (s *RegionService) CheckRegionHealth(ctx context.Context, regionID uuid.UUID) (string, error) {
	region, err := s.GetRegion(ctx, regionID)
	if err != nil {
		return "down", err
	}

	// Simple health check - ping the database
	if region.DatabaseHost != "" {
		// In production, this would actually ping the database
		// For now, we'll assume healthy if active
		if region.Active {
			return "healthy", nil
		}
		return "down", nil
	}

	// If no database host, check endpoint
	if region.Endpoint != "" {
		// In production, this would make an HTTP request to health endpoint
		// For now, assume healthy if active
		if region.Active {
			return "healthy", nil
		}
		return "down", nil
	}

	return "unknown", nil
}

/* FailoverToRegion initiates failover to a secondary region */
func (s *RegionService) FailoverToRegion(ctx context.Context, targetRegionID uuid.UUID) error {
	// Get target region
	targetRegion, err := s.GetRegion(ctx, targetRegionID)
	if err != nil {
		return fmt.Errorf("target region not found: %w", err)
	}

	if !targetRegion.Active {
		return fmt.Errorf("target region is not active")
	}

	// Get current primary
	primaryRegion, err := s.GetPrimaryRegion(ctx)
	if err != nil {
		return fmt.Errorf("primary region not found: %w", err)
	}

	// Update primary region to secondary
	updatePrimaryQuery := `
		UPDATE neuronip.regions
		SET primary_region = false, updated_at = NOW()
		WHERE id = $1`
	_, err = s.pool.Exec(ctx, updatePrimaryQuery, primaryRegion.ID)
	if err != nil {
		return fmt.Errorf("failed to demote primary region: %w", err)
	}

	// Promote target region to primary
	updateTargetQuery := `
		UPDATE neuronip.regions
		SET primary_region = true, updated_at = NOW()
		WHERE id = $1`
	_, err = s.pool.Exec(ctx, updateTargetQuery, targetRegionID)
	if err != nil {
		return fmt.Errorf("failed to promote target region: %w", err)
	}

	return nil
}

/* UpdateReplicationSync updates replication sync timestamp */
func (s *RegionService) UpdateReplicationSync(ctx context.Context, regionID uuid.UUID) error {
	query := `
		UPDATE neuronip.regions
		SET last_sync_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, regionID)
	return err
}
