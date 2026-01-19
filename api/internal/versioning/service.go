package versioning

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* VersioningService provides versioning functionality */
type VersioningService struct {
	pool *pgxpool.Pool
}

/* NewVersioningService creates a new versioning service */
func NewVersioningService(pool *pgxpool.Pool) *VersioningService {
	return &VersioningService{pool: pool}
}

/* Version represents a version */
type Version struct {
	ID             uuid.UUID              `json:"id"`
	ResourceType   string                 `json:"resource_type"`
	ResourceID     uuid.UUID              `json:"resource_id"`
	VersionNumber  string                 `json:"version_number"`
	VersionData    map[string]interface{} `json:"version_data"`
	CreatedBy      *string                `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	IsCurrent      bool                   `json:"is_current"`
}

/* VersionHistory represents version history */
type VersionHistory struct {
	ID           uuid.UUID              `json:"id"`
	VersionID    uuid.UUID              `json:"version_id"`
	Action       string                 `json:"action"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	RollbackData map[string]interface{} `json:"rollback_data,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

/* ListVersions lists versions for a resource */
func (s *VersioningService) ListVersions(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]Version, error) {
	query := `
		SELECT id, resource_type, resource_id, version_number, version_data, created_by, created_at, is_current
		FROM neuronip.versions
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	var versions []Version
	for rows.Next() {
		var v Version
		var versionJSON json.RawMessage

		err := rows.Scan(&v.ID, &v.ResourceType, &v.ResourceID, &v.VersionNumber, &versionJSON, &v.CreatedBy, &v.CreatedAt, &v.IsCurrent)
		if err != nil {
			continue
		}

		if versionJSON != nil {
			json.Unmarshal(versionJSON, &v.VersionData)
		}

		versions = append(versions, v)
	}

	return versions, nil
}

/* CreateVersion creates a new version */
func (s *VersioningService) CreateVersion(ctx context.Context, v Version) (*Version, error) {
	id := uuid.New()
	versionJSON, _ := json.Marshal(v.VersionData)

	// If this is marked as current, unset current on other versions
	if v.IsCurrent {
		_, err := s.pool.Exec(ctx, `
			UPDATE neuronip.versions
			SET is_current = false
			WHERE resource_type = $1 AND resource_id = $2`,
			v.ResourceType, v.ResourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset current version: %w", err)
		}
	}

	query := `
		INSERT INTO neuronip.versions (id, resource_type, resource_id, version_number, version_data, created_by, created_at, is_current)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, resource_type, resource_id, version_number, version_data, created_by, created_at, is_current`

	var result Version
	var versionJSONRaw json.RawMessage
	now := time.Now()

	err := s.pool.QueryRow(ctx, query,
		id, v.ResourceType, v.ResourceID, v.VersionNumber, versionJSON, v.CreatedBy, now, v.IsCurrent,
	).Scan(
		&result.ID, &result.ResourceType, &result.ResourceID, &result.VersionNumber,
		&versionJSONRaw, &result.CreatedBy, &result.CreatedAt, &result.IsCurrent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	if versionJSONRaw != nil {
		json.Unmarshal(versionJSONRaw, &result.VersionData)
	}

	// Record history
	s.recordHistory(ctx, result.ID, "create", map[string]interface{}{}, nil)

	return &result, nil
}

/* GetVersion retrieves a version by ID */
func (s *VersioningService) GetVersion(ctx context.Context, id uuid.UUID) (*Version, error) {
	query := `
		SELECT id, resource_type, resource_id, version_number, version_data, created_by, created_at, is_current
		FROM neuronip.versions
		WHERE id = $1`

	var result Version
	var versionJSONRaw json.RawMessage

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&result.ID, &result.ResourceType, &result.ResourceID, &result.VersionNumber,
		&versionJSONRaw, &result.CreatedBy, &result.CreatedAt, &result.IsCurrent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	if versionJSONRaw != nil {
		json.Unmarshal(versionJSONRaw, &result.VersionData)
	}

	return &result, nil
}

/* RollbackVersion rolls back to a specific version */
func (s *VersioningService) RollbackVersion(ctx context.Context, versionID uuid.UUID) error {
	// Get the version
	version, err := s.GetVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	// Get current version for rollback data
	currentVersions, err := s.ListVersions(ctx, version.ResourceType, version.ResourceID)
	if err == nil && len(currentVersions) > 0 {
		currentVersion := currentVersions[0]
		rollbackData := map[string]interface{}{
			"previous_version": currentVersion.ID,
			"previous_data":    currentVersion.VersionData,
		}

		// Record history with rollback data
		s.recordHistory(ctx, versionID, "rollback", map[string]interface{}{}, rollbackData)
	}

	// Mark this version as current and unset others
	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.versions
		SET is_current = false
		WHERE resource_type = $1 AND resource_id = $2`,
		version.ResourceType, version.ResourceID)
	if err != nil {
		return fmt.Errorf("failed to unset current versions: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE neuronip.versions
		SET is_current = true
		WHERE id = $1`,
		versionID)
	if err != nil {
		return fmt.Errorf("failed to set current version: %w", err)
	}

	return nil
}

/* GetVersionHistory retrieves version history */
func (s *VersioningService) GetVersionHistory(ctx context.Context, versionID uuid.UUID) ([]VersionHistory, error) {
	query := `
		SELECT id, version_id, action, changes, rollback_data, created_at
		FROM neuronip.version_history
		WHERE version_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version history: %w", err)
	}
	defer rows.Close()

	var history []VersionHistory
	for rows.Next() {
		var h VersionHistory
		var changesJSON, rollbackJSON json.RawMessage

		err := rows.Scan(&h.ID, &h.VersionID, &h.Action, &changesJSON, &rollbackJSON, &h.CreatedAt)
		if err != nil {
			continue
		}

		if changesJSON != nil {
			json.Unmarshal(changesJSON, &h.Changes)
		}
		if rollbackJSON != nil {
			json.Unmarshal(rollbackJSON, &h.RollbackData)
		}

		history = append(history, h)
	}

	return history, nil
}

/* recordHistory records a version history entry */
func (s *VersioningService) recordHistory(ctx context.Context, versionID uuid.UUID, action string, changes, rollbackData map[string]interface{}) {
	changesJSON, _ := json.Marshal(changes)
	rollbackJSON, _ := json.Marshal(rollbackData)

	query := `
		INSERT INTO neuronip.version_history (id, version_id, action, changes, rollback_data, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`

	s.pool.Exec(ctx, query, versionID, action, changesJSON, rollbackJSON)
}
