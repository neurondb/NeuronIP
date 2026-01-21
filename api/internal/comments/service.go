package comments

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides comments functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new comments service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* Comment represents a comment */
type Comment struct {
	ID            uuid.UUID              `json:"id"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    uuid.UUID              `json:"resource_id"`
	UserID        string                 `json:"user_id"`
	CommentText   string                 `json:"comment_text"`
	ParentCommentID *uuid.UUID            `json:"parent_comment_id,omitempty"`
	IsResolved    bool                   `json:"is_resolved"`
	ResolvedBy    *string                `json:"resolved_by,omitempty"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateComment creates a new comment */
func (s *Service) CreateComment(ctx context.Context, comment Comment) (*Comment, error) {
	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(comment.Metadata)
	var parentID sql.NullString
	if comment.ParentCommentID != nil {
		parentID = sql.NullString{String: comment.ParentCommentID.String(), Valid: true}
	}

	query := `
		INSERT INTO neuronip.comments
		(id, resource_type, resource_id, user_id, comment_text, parent_comment_id,
		 is_resolved, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		comment.ID, comment.ResourceType, comment.ResourceID, comment.UserID,
		comment.CommentText, parentID, comment.IsResolved, metadataJSON,
		comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return &comment, nil
}

/* GetComment retrieves a comment by ID */
func (s *Service) GetComment(ctx context.Context, id uuid.UUID) (*Comment, error) {
	var comment Comment
	var parentID sql.NullString
	var resolvedBy sql.NullString
	var resolvedAt sql.NullTime
	var metadataJSON []byte

	query := `
		SELECT id, resource_type, resource_id, user_id, comment_text, parent_comment_id,
		       is_resolved, resolved_by, resolved_at, metadata, created_at, updated_at
		FROM neuronip.comments
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.ResourceType, &comment.ResourceID, &comment.UserID,
		&comment.CommentText, &parentID, &comment.IsResolved, &resolvedBy, &resolvedAt,
		&metadataJSON, &comment.CreatedAt, &comment.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	if parentID.Valid {
		parentUUID, _ := uuid.Parse(parentID.String)
		comment.ParentCommentID = &parentUUID
	}
	if resolvedBy.Valid {
		comment.ResolvedBy = &resolvedBy.String
	}
	if resolvedAt.Valid {
		comment.ResolvedAt = &resolvedAt.Time
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &comment.Metadata)
	}

	return &comment, nil
}

/* ListComments lists comments for a resource */
func (s *Service) ListComments(ctx context.Context, resourceType string, resourceID uuid.UUID, includeResolved bool) ([]Comment, error) {
	query := `
		SELECT id, resource_type, resource_id, user_id, comment_text, parent_comment_id,
		       is_resolved, resolved_by, resolved_at, metadata, created_at, updated_at
		FROM neuronip.comments
		WHERE resource_type = $1 AND resource_id = $2`
	
	if !includeResolved {
		query += " AND is_resolved = false"
	}
	query += " ORDER BY created_at ASC"

	rows, err := s.pool.Query(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var parentID sql.NullString
		var resolvedBy sql.NullString
		var resolvedAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&comment.ID, &comment.ResourceType, &comment.ResourceID, &comment.UserID,
			&comment.CommentText, &parentID, &comment.IsResolved, &resolvedBy, &resolvedAt,
			&metadataJSON, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			parentUUID, _ := uuid.Parse(parentID.String)
			comment.ParentCommentID = &parentUUID
		}
		if resolvedBy.Valid {
			comment.ResolvedBy = &resolvedBy.String
		}
		if resolvedAt.Valid {
			comment.ResolvedAt = &resolvedAt.Time
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &comment.Metadata)
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

/* ResolveComment resolves a comment */
func (s *Service) ResolveComment(ctx context.Context, id uuid.UUID, resolvedBy string) error {
	query := `
		UPDATE neuronip.comments
		SET is_resolved = true, resolved_by = $1, resolved_at = NOW(), updated_at = NOW()
		WHERE id = $2`

	_, err := s.pool.Exec(ctx, query, resolvedBy, id)
	if err != nil {
		return fmt.Errorf("failed to resolve comment: %w", err)
	}

	return nil
}

/* DeleteComment deletes a comment */
func (s *Service) DeleteComment(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM neuronip.comments WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
