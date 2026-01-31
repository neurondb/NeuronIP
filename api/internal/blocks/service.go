package blocks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* Service handles block operations */
type Service struct {
	db *pgxpool.Pool
}

/* NewService creates a new blocks service */
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

/* Block represents a content block */
type Block struct {
	ID        uuid.UUID              `json:"id"`
	PageID    uuid.UUID              `json:"page_id"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	Order     int                    `json:"order"`
	ParentID  *uuid.UUID             `json:"parent_id,omitempty"`
	CreatedBy *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

/* CreateBlockRequest represents a request to create a block */
type CreateBlockRequest struct {
	PageID   uuid.UUID              `json:"page_id"`
	Type     string                 `json:"type"`
	Content  map[string]interface{} `json:"content"`
	Order    *int                   `json:"order,omitempty"`
	ParentID *uuid.UUID             `json:"parent_id,omitempty"`
}

/* UpdateBlockRequest represents a request to update a block */
type UpdateBlockRequest struct {
	Content *map[string]interface{} `json:"content,omitempty"`
	Order   *int                    `json:"order,omitempty"`
}

/* ReorderBlocksRequest represents a request to reorder blocks */
type ReorderBlocksRequest struct {
	PageID   uuid.UUID   `json:"page_id"`
	BlockIDs []uuid.UUID `json:"block_ids"`
}

/* GetBlocks retrieves all blocks for a page */
func (s *Service) GetBlocks(ctx context.Context, pageID uuid.UUID) ([]Block, error) {
	query := `
		SELECT id, page_id, type, content, order_index, parent_id, created_by, created_at, updated_at, metadata
		FROM neuronip.blocks
		WHERE page_id = $1 AND deleted_at IS NULL
		ORDER BY order_index ASC, created_at ASC
	`

	rows, err := s.db.Query(ctx, query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}
	defer rows.Close()

	var blocks []Block
	for rows.Next() {
		var block Block
		var contentJSON, metadataJSON []byte
		var parentID, createdBy *uuid.UUID

		err := rows.Scan(
			&block.ID, &block.PageID, &block.Type, &contentJSON,
			&block.Order, &parentID, &createdBy, &block.CreatedAt,
			&block.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}

		if err := json.Unmarshal(contentJSON, &block.Content); err != nil {
			block.Content = make(map[string]interface{})
		}
		if err := json.Unmarshal(metadataJSON, &block.Metadata); err != nil {
			block.Metadata = make(map[string]interface{})
		}

		block.ParentID = parentID
		block.CreatedBy = createdBy
		blocks = append(blocks, block)
	}

	return blocks, nil
}

/* CreateBlock creates a new block */
func (s *Service) CreateBlock(ctx context.Context, req CreateBlockRequest, userID *uuid.UUID) (*Block, error) {
	// Get max order for the page
	var maxOrder int
	orderQuery := `SELECT COALESCE(MAX(order_index), -1) + 1 FROM neuronip.blocks WHERE page_id = $1 AND deleted_at IS NULL`
	err := s.db.QueryRow(ctx, orderQuery, req.PageID).Scan(&maxOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to get max order: %w", err)
	}

	order := maxOrder
	if req.Order != nil {
		order = *req.Order
	}

	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, errors.BadRequest("Invalid content format")
	}

	blockID := uuid.New()
	query := `
		INSERT INTO neuronip.blocks (id, page_id, type, content, order_index, parent_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, page_id, type, content, order_index, parent_id, created_by, created_at, updated_at, metadata
	`

	var block Block
	var contentJSONOut, metadataJSON []byte
	var parentID, createdBy *uuid.UUID

	err = s.db.QueryRow(ctx, query, blockID, req.PageID, req.Type, contentJSON, order, req.ParentID, userID).Scan(
		&block.ID, &block.PageID, &block.Type, &contentJSONOut,
		&block.Order, &parentID, &createdBy, &block.CreatedAt,
		&block.UpdatedAt, &metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	if err := json.Unmarshal(contentJSONOut, &block.Content); err != nil {
		block.Content = req.Content
	}
	if err := json.Unmarshal(metadataJSON, &block.Metadata); err != nil {
		block.Metadata = make(map[string]interface{})
	}

	block.ParentID = parentID
	block.CreatedBy = createdBy

	return &block, nil
}

/* UpdateBlock updates an existing block */
func (s *Service) UpdateBlock(ctx context.Context, blockID uuid.UUID, req UpdateBlockRequest) (*Block, error) {
	updates := []string{}
	args := []interface{}{blockID}
	argIndex := 2

	if req.Content != nil {
		contentJSON, err := json.Marshal(*req.Content)
		if err != nil {
			return nil, errors.BadRequest("Invalid content format")
		}
		updates = append(updates, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, contentJSON)
		argIndex++
	}

	if req.Order != nil {
		updates = append(updates, fmt.Sprintf("order_index = $%d", argIndex))
		args = append(args, *req.Order)
		argIndex++
	}

	if len(updates) == 0 {
		return s.GetBlock(ctx, blockID)
	}

	updateClause := updates[0]
	for i := 1; i < len(updates); i++ {
		updateClause = updateClause + ", " + updates[i]
	}

	query := fmt.Sprintf(`
		UPDATE neuronip.blocks
		SET %s, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, page_id, type, content, order_index, parent_id, created_by, created_at, updated_at, metadata
	`, updateClause)

	var block Block
	var contentJSON, metadataJSON []byte
	var parentID, createdBy *uuid.UUID

	err := s.db.QueryRow(ctx, query, args...).Scan(
		&block.ID, &block.PageID, &block.Type, &contentJSON,
		&block.Order, &parentID, &createdBy, &block.CreatedAt,
		&block.UpdatedAt, &metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update block: %w", err)
	}

	if err := json.Unmarshal(contentJSON, &block.Content); err != nil {
		block.Content = make(map[string]interface{})
	}
	if err := json.Unmarshal(metadataJSON, &block.Metadata); err != nil {
		block.Metadata = make(map[string]interface{})
	}

	block.ParentID = parentID
	block.CreatedBy = createdBy

	return &block, nil
}

/* GetBlock retrieves a single block */
func (s *Service) GetBlock(ctx context.Context, blockID uuid.UUID) (*Block, error) {
	query := `
		SELECT id, page_id, type, content, order_index, parent_id, created_by, created_at, updated_at, metadata
		FROM neuronip.blocks
		WHERE id = $1 AND deleted_at IS NULL
	`

	var block Block
	var contentJSON, metadataJSON []byte
	var parentID, createdBy *uuid.UUID

	err := s.db.QueryRow(ctx, query, blockID).Scan(
		&block.ID, &block.PageID, &block.Type, &contentJSON,
		&block.Order, &parentID, &createdBy, &block.CreatedAt,
		&block.UpdatedAt, &metadataJSON,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("Block not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	if err := json.Unmarshal(contentJSON, &block.Content); err != nil {
		block.Content = make(map[string]interface{})
	}
	if err := json.Unmarshal(metadataJSON, &block.Metadata); err != nil {
		block.Metadata = make(map[string]interface{})
	}

	block.ParentID = parentID
	block.CreatedBy = createdBy

	return &block, nil
}

/* DeleteBlock soft deletes a block */
func (s *Service) DeleteBlock(ctx context.Context, blockID uuid.UUID) error {
	query := `UPDATE neuronip.blocks SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := s.db.Exec(ctx, query, blockID)
	if err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("Block not found")
	}

	return nil
}

/* ReorderBlocks reorders blocks for a page */
func (s *Service) ReorderBlocks(ctx context.Context, req ReorderBlocksRequest) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, blockID := range req.BlockIDs {
		updateQuery := `UPDATE neuronip.blocks SET order_index = $1 WHERE id = $2 AND page_id = $3 AND deleted_at IS NULL`
		_, err := tx.Exec(ctx, updateQuery, i, blockID, req.PageID)
		if err != nil {
			return fmt.Errorf("failed to reorder blocks: %w", err)
		}
	}

	return tx.Commit(ctx)
}
