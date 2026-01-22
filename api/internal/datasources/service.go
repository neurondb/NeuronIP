package datasources

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DataSourceService provides data source management functionality */
type DataSourceService struct {
	pool *pgxpool.Pool
}

/* NewDataSourceService creates a new data source service */
func NewDataSourceService(pool *pgxpool.Pool) *DataSourceService {
	return &DataSourceService{pool: pool}
}

/* DataSource represents a data source */
type DataSource struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	SourceType     string                 `json:"source_type"`
	ConnectionString *string              `json:"connection_string,omitempty"`
	Config         map[string]interface{} `json:"config"`
	Enabled        bool                   `json:"enabled"`
	SyncSchedule   *string                `json:"sync_schedule,omitempty"`
	SyncStatus     *string                `json:"sync_status,omitempty"`
	LastAccessedAt *time.Time             `json:"last_accessed_at,omitempty"`
	LastSyncAt     *time.Time             `json:"last_sync_at,omitempty"`
	SyncError      *string                `json:"sync_error,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* CreateDataSource creates a new data source */
func (s *DataSourceService) CreateDataSource(ctx context.Context, ds DataSource) (*DataSource, error) {
	id := uuid.New()
	configJSON, _ := json.Marshal(ds.Config)
	now := time.Now()

	query := `
		INSERT INTO neuronip.data_sources 
		(id, name, source_type, connection_string, config, enabled, sync_schedule, sync_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, source_type, connection_string, config, enabled, sync_schedule, sync_status, 
		          last_accessed_at, last_sync_at, sync_error, created_at, updated_at`

	var result DataSource
	var configJSONRaw json.RawMessage
	syncStatus := "idle"
	if ds.SyncStatus != nil {
		syncStatus = *ds.SyncStatus
	}

	err := s.pool.QueryRow(ctx, query,
		id, ds.Name, ds.SourceType, ds.ConnectionString, configJSON, ds.Enabled,
		ds.SyncSchedule, syncStatus, now, now,
	).Scan(
		&result.ID, &result.Name, &result.SourceType, &result.ConnectionString,
		&configJSONRaw, &result.Enabled, &result.SyncSchedule, &result.SyncStatus,
		&result.LastAccessedAt, &result.LastSyncAt, &result.SyncError,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	return &result, nil
}

/* GetDataSource retrieves a data source by ID */
func (s *DataSourceService) GetDataSource(ctx context.Context, id uuid.UUID) (*DataSource, error) {
	// Map existing table columns to service fields
	query := `
		SELECT id, name, type as source_type, NULL as connection_string, config, 
		       (status = 'active') as enabled, NULL as sync_schedule, NULL as sync_status,
		       NULL as last_accessed_at, NULL as last_sync_at, NULL as sync_error, 
		       created_at, updated_at
		FROM neuronip.data_sources
		WHERE id = $1`

	var result DataSource
	var configJSONRaw json.RawMessage
	var enabled sql.NullBool
	var syncSchedule, syncStatus, syncError sql.NullString
	var lastAccessedAt, lastSyncAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.Name, &result.SourceType, &result.ConnectionString,
		&configJSONRaw, &enabled, &syncSchedule, &syncStatus,
		&lastAccessedAt, &lastSyncAt, &syncError,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get data source: %w", err)
	}

	if enabled.Valid {
		result.Enabled = enabled.Bool
	} else {
		result.Enabled = true
	}
	if syncSchedule.Valid {
		result.SyncSchedule = &syncSchedule.String
	}
	if syncStatus.Valid {
		result.SyncStatus = &syncStatus.String
	}
	if lastAccessedAt.Valid {
		result.LastAccessedAt = &lastAccessedAt.Time
	}
	if lastSyncAt.Valid {
		result.LastSyncAt = &lastSyncAt.Time
	}
	if syncError.Valid {
		result.SyncError = &syncError.String
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	return &result, nil
}

/* ListDataSources lists all data sources */
func (s *DataSourceService) ListDataSources(ctx context.Context) ([]DataSource, error) {
	// Map existing table columns (type, status) to service fields (source_type, enabled)
	query := `
		SELECT id, name, type as source_type, NULL as connection_string, config, 
		       (status = 'active') as enabled, NULL as sync_schedule, NULL as sync_status,
		       NULL as last_accessed_at, NULL as last_sync_at, NULL as sync_error, 
		       created_at, updated_at
		FROM neuronip.data_sources
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}
	defer rows.Close()

	var sources []DataSource
	for rows.Next() {
		var ds DataSource
		var configJSONRaw json.RawMessage
		var enabled sql.NullBool
		var syncSchedule, syncStatus, syncError sql.NullString
		var lastAccessedAt, lastSyncAt sql.NullTime

		err := rows.Scan(
			&ds.ID, &ds.Name, &ds.SourceType, &ds.ConnectionString,
			&configJSONRaw, &enabled, &syncSchedule, &syncStatus,
			&lastAccessedAt, &lastSyncAt, &syncError,
			&ds.CreatedAt, &ds.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if enabled.Valid {
			ds.Enabled = enabled.Bool
		} else {
			ds.Enabled = true // Default to enabled
		}
		if syncSchedule.Valid {
			ds.SyncSchedule = &syncSchedule.String
		}
		if syncStatus.Valid {
			ds.SyncStatus = &syncStatus.String
		}
		if lastAccessedAt.Valid {
			ds.LastAccessedAt = &lastAccessedAt.Time
		}
		if lastSyncAt.Valid {
			ds.LastSyncAt = &lastSyncAt.Time
		}
		if syncError.Valid {
			ds.SyncError = &syncError.String
		}

		json.Unmarshal(configJSONRaw, &ds.Config)
		sources = append(sources, ds)
	}

	return sources, nil
}

/* UpdateDataSource updates a data source */
func (s *DataSourceService) UpdateDataSource(ctx context.Context, id uuid.UUID, ds DataSource) (*DataSource, error) {
	configJSON, _ := json.Marshal(ds.Config)

	query := `
		UPDATE neuronip.data_sources
		SET name = $1, source_type = $2, connection_string = $3, config = $4, enabled = $5,
		    sync_schedule = $6, sync_status = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING id, name, source_type, connection_string, config, enabled, sync_schedule, sync_status,
		          last_accessed_at, last_sync_at, sync_error, created_at, updated_at`

	var result DataSource
	var configJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query,
		ds.Name, ds.SourceType, ds.ConnectionString, configJSON, ds.Enabled,
		ds.SyncSchedule, ds.SyncStatus, id,
	).Scan(
		&result.ID, &result.Name, &result.SourceType, &result.ConnectionString,
		&configJSONRaw, &result.Enabled, &result.SyncSchedule, &result.SyncStatus,
		&result.LastAccessedAt, &result.LastSyncAt, &result.SyncError,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update data source: %w", err)
	}

	json.Unmarshal(configJSONRaw, &result.Config)
	return &result, nil
}

/* DeleteDataSource deletes a data source */
func (s *DataSourceService) DeleteDataSource(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.data_sources WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

/* TriggerSync triggers a sync for a data source */
func (s *DataSourceService) TriggerSync(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE neuronip.data_sources
		SET sync_status = 'syncing', last_sync_at = NOW(), sync_error = NULL, updated_at = NOW()
		WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	return err
}

/* GetSyncStatus gets the sync status for a data source */
func (s *DataSourceService) GetSyncStatus(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT sync_status, last_sync_at, sync_error, last_accessed_at
		FROM neuronip.data_sources
		WHERE id = $1`

	status := map[string]interface{}{}
	var syncStatus, syncError *string
	var lastSyncAt, lastAccessedAt *time.Time

	err := s.pool.QueryRow(ctx, query, id).Scan(&syncStatus, &lastSyncAt, &syncError, &lastAccessedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	if syncStatus != nil {
		status["sync_status"] = *syncStatus
	}
	if lastSyncAt != nil {
		status["last_sync_at"] = *lastSyncAt
	}
	if syncError != nil {
		status["sync_error"] = *syncError
	}
	if lastAccessedAt != nil {
		status["last_accessed_at"] = *lastAccessedAt
	}

	return status, nil
}
