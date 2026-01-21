package ownership

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides resource ownership functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new ownership service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* Ownership represents resource ownership */
type Ownership struct {
	ID           uuid.UUID              `json:"id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   uuid.UUID              `json:"resource_id"`
	OwnerID      string                 `json:"owner_id"`
	OwnerType    string                 `json:"owner_type"` // user, team, organization
	AssignedBy   *string                `json:"assigned_by,omitempty"`
	AssignedAt   time.Time              `json:"assigned_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* AssignOwnership assigns ownership to a resource */
func (s *Service) AssignOwnership(ctx context.Context, ownership Ownership) (*Ownership, error) {
	ownership.ID = uuid.New()
	ownership.AssignedAt = time.Now()

	metadataJSON, _ := json.Marshal(ownership.Metadata)
	var assignedBy sql.NullString
	if ownership.AssignedBy != nil {
		assignedBy = sql.NullString{String: *ownership.AssignedBy, Valid: true}
	}

	query := `
		INSERT INTO neuronip.resource_ownership
		(id, resource_type, resource_id, owner_id, owner_type, assigned_by, assigned_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (resource_type, resource_id)
		DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			owner_type = EXCLUDED.owner_type,
			assigned_by = EXCLUDED.assigned_by,
			assigned_at = EXCLUDED.assigned_at,
			metadata = EXCLUDED.metadata
		RETURNING id, assigned_at`

	err := s.pool.QueryRow(ctx, query,
		ownership.ID, ownership.ResourceType, ownership.ResourceID,
		ownership.OwnerID, ownership.OwnerType, assignedBy,
		ownership.AssignedAt, metadataJSON,
	).Scan(&ownership.ID, &ownership.AssignedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to assign ownership: %w", err)
	}

	return &ownership, nil
}

/* GetOwnership retrieves ownership for a resource */
func (s *Service) GetOwnership(ctx context.Context, resourceType string, resourceID uuid.UUID) (*Ownership, error) {
	var ownership Ownership
	var assignedBy sql.NullString
	var metadataJSON []byte

	query := `
		SELECT id, resource_type, resource_id, owner_id, owner_type,
		       assigned_by, assigned_at, metadata
		FROM neuronip.resource_ownership
		WHERE resource_type = $1 AND resource_id = $2`

	err := s.pool.QueryRow(ctx, query, resourceType, resourceID).Scan(
		&ownership.ID, &ownership.ResourceType, &ownership.ResourceID,
		&ownership.OwnerID, &ownership.OwnerType, &assignedBy,
		&ownership.AssignedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No ownership assigned
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ownership: %w", err)
	}

	if assignedBy.Valid {
		ownership.AssignedBy = &assignedBy.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &ownership.Metadata)
	}

	return &ownership, nil
}

/* ListOwnershipByOwner lists resources owned by a user/team/org */
func (s *Service) ListOwnershipByOwner(ctx context.Context, ownerID string, ownerType string) ([]Ownership, error) {
	query := `
		SELECT id, resource_type, resource_id, owner_id, owner_type,
		       assigned_by, assigned_at, metadata
		FROM neuronip.resource_ownership
		WHERE owner_id = $1 AND owner_type = $2
		ORDER BY assigned_at DESC`

	rows, err := s.pool.Query(ctx, query, ownerID, ownerType)
	if err != nil {
		return nil, fmt.Errorf("failed to list ownership: %w", err)
	}
	defer rows.Close()

	var ownerships []Ownership
	for rows.Next() {
		var ownership Ownership
		var assignedBy sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&ownership.ID, &ownership.ResourceType, &ownership.ResourceID,
			&ownership.OwnerID, &ownership.OwnerType, &assignedBy,
			&ownership.AssignedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if assignedBy.Valid {
			ownership.AssignedBy = &assignedBy.String
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &ownership.Metadata)
		}

		ownerships = append(ownerships, ownership)
	}

	return ownerships, nil
}

/* RemoveOwnership removes ownership from a resource */
func (s *Service) RemoveOwnership(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	query := `
		DELETE FROM neuronip.resource_ownership
		WHERE resource_type = $1 AND resource_id = $2`

	_, err := s.pool.Exec(ctx, query, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("failed to remove ownership: %w", err)
	}
	return nil
}
