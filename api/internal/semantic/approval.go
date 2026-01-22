package semantic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ApprovalService provides metric approval workflow functionality */
type ApprovalService struct {
	pool *pgxpool.Pool
}

/* NewApprovalService creates a new approval service */
func NewApprovalService(pool *pgxpool.Pool) *ApprovalService {
	return &ApprovalService{pool: pool}
}

/* MetricApproval represents a metric approval */
type MetricApproval struct {
	ID         uuid.UUID  `json:"id"`
	MetricID   uuid.UUID  `json:"metric_id"`
	ApproverID string     `json:"approver_id"`
	Status     string     `json:"status"` // pending, approved, rejected, changes_requested
	Comments   *string    `json:"comments,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

/* CreateApproval creates a new metric approval request */
func (s *ApprovalService) CreateApproval(ctx context.Context, metricID uuid.UUID, approverID string, comments *string) (*MetricApproval, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO neuronip.metric_approvals 
		(id, metric_id, approver_id, status, comments, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6)
		RETURNING id, metric_id, approver_id, status, comments, approved_at, created_at, updated_at`

	var approval MetricApproval
	var commentsVal sql.NullString
	if comments != nil {
		commentsVal = sql.NullString{String: *comments, Valid: true}
	}
	var approvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, id, metricID, approverID, commentsVal, now, now).Scan(
		&approval.ID, &approval.MetricID, &approval.ApproverID, &approval.Status,
		&commentsVal, &approvedAt, &approval.CreatedAt, &approval.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	if commentsVal.Valid {
		approval.Comments = &commentsVal.String
	}
	if approvedAt.Valid {
		approval.ApprovedAt = &approvedAt.Time
	}

	// Update metric status to pending_approval
	updateMetricQuery := `
		UPDATE neuronip.business_metrics 
		SET approval_status = 'pending_approval', status = 'pending_approval', updated_at = NOW()
		WHERE id = $1`
	s.pool.Exec(ctx, updateMetricQuery, metricID)

	return &approval, nil
}

/* ApproveMetric approves a metric */
func (s *ApprovalService) ApproveMetric(ctx context.Context, approvalID uuid.UUID, approverID string, comments *string) error {
	now := time.Now()

	query := `
		UPDATE neuronip.metric_approvals 
		SET status = 'approved', 
		    approved_at = $1,
		    comments = COALESCE($2, comments),
		    updated_at = $1
		WHERE id = $3 AND approver_id = $4
		RETURNING metric_id`

	var metricID uuid.UUID
	var commentsVal sql.NullString
	if comments != nil {
		commentsVal = sql.NullString{String: *comments, Valid: true}
	}

	err := s.pool.QueryRow(ctx, query, now, commentsVal, approvalID, approverID).Scan(&metricID)
	if err != nil {
		return fmt.Errorf("failed to approve metric: %w", err)
	}

	// Update metric status to approved
	updateMetricQuery := `
		UPDATE neuronip.business_metrics 
		SET approval_status = 'approved', status = 'approved', updated_at = NOW()
		WHERE id = $1`
	s.pool.Exec(ctx, updateMetricQuery, metricID)

	return nil
}

/* RejectMetric rejects a metric */
func (s *ApprovalService) RejectMetric(ctx context.Context, approvalID uuid.UUID, approverID string, comments string) error {
	query := `
		UPDATE neuronip.metric_approvals 
		SET status = 'rejected', 
		    comments = $1,
		    updated_at = NOW()
		WHERE id = $2 AND approver_id = $3
		RETURNING metric_id`

	var metricID uuid.UUID
	err := s.pool.QueryRow(ctx, query, comments, approvalID, approverID).Scan(&metricID)
	if err != nil {
		return fmt.Errorf("failed to reject metric: %w", err)
	}

	// Update metric status to draft
	updateMetricQuery := `
		UPDATE neuronip.business_metrics 
		SET approval_status = 'draft', status = 'draft', updated_at = NOW()
		WHERE id = $1`
	s.pool.Exec(ctx, updateMetricQuery, metricID)

	return nil
}

/* RequestChanges requests changes to a metric */
func (s *ApprovalService) RequestChanges(ctx context.Context, approvalID uuid.UUID, approverID string, comments string) error {
	query := `
		UPDATE neuronip.metric_approvals 
		SET status = 'changes_requested', 
		    comments = $1,
		    updated_at = NOW()
		WHERE id = $2 AND approver_id = $3
		RETURNING metric_id`

	var metricID uuid.UUID
	err := s.pool.QueryRow(ctx, query, comments, approvalID, approverID).Scan(&metricID)
	if err != nil {
		return fmt.Errorf("failed to request changes: %w", err)
	}

	// Update metric status to draft
	updateMetricQuery := `
		UPDATE neuronip.business_metrics 
		SET approval_status = 'draft', status = 'draft', updated_at = NOW()
		WHERE id = $1`
	s.pool.Exec(ctx, updateMetricQuery, metricID)

	return nil
}

/* GetApproval retrieves an approval by ID */
func (s *ApprovalService) GetApproval(ctx context.Context, approvalID uuid.UUID) (*MetricApproval, error) {
	query := `
		SELECT id, metric_id, approver_id, status, comments, approved_at, created_at, updated_at
		FROM neuronip.metric_approvals
		WHERE id = $1`

	var approval MetricApproval
	var commentsVal sql.NullString
	var approvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query, approvalID).Scan(
		&approval.ID, &approval.MetricID, &approval.ApproverID, &approval.Status,
		&commentsVal, &approvedAt, &approval.CreatedAt, &approval.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval: %w", err)
	}

	if commentsVal.Valid {
		approval.Comments = &commentsVal.String
	}
	if approvedAt.Valid {
		approval.ApprovedAt = &approvedAt.Time
	}

	return &approval, nil
}

/* GetMetricApprovals retrieves all approvals for a metric */
func (s *ApprovalService) GetMetricApprovals(ctx context.Context, metricID uuid.UUID) ([]MetricApproval, error) {
	query := `
		SELECT id, metric_id, approver_id, status, comments, approved_at, created_at, updated_at
		FROM neuronip.metric_approvals
		WHERE metric_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric approvals: %w", err)
	}
	defer rows.Close()

	var approvals []MetricApproval
	for rows.Next() {
		var approval MetricApproval
		var commentsVal sql.NullString
		var approvedAt sql.NullTime

		err := rows.Scan(
			&approval.ID, &approval.MetricID, &approval.ApproverID, &approval.Status,
			&commentsVal, &approvedAt, &approval.CreatedAt, &approval.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if commentsVal.Valid {
			approval.Comments = &commentsVal.String
		}
		if approvedAt.Valid {
			approval.ApprovedAt = &approvedAt.Time
		}

		approvals = append(approvals, approval)
	}

	return approvals, nil
}

/* MetricOwnershipService provides metric ownership management */
type MetricOwnershipService struct {
	pool *pgxpool.Pool
}

/* NewMetricOwnershipService creates a new ownership service */
func NewMetricOwnershipService(pool *pgxpool.Pool) *MetricOwnershipService {
	return &MetricOwnershipService{pool: pool}
}

/* UpdateMetricOwner updates the owner of a metric */
func (s *MetricOwnershipService) UpdateMetricOwner(ctx context.Context, metricID uuid.UUID, ownerID string) error {
	query := `
		UPDATE neuronip.business_metrics 
		SET owner_id = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := s.pool.Exec(ctx, query, ownerID, metricID)
	if err != nil {
		return fmt.Errorf("failed to update metric owner: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("metric not found")
	}

	return nil
}

/* GetMetricsByOwner retrieves all metrics owned by a user */
func (s *MetricOwnershipService) GetMetricsByOwner(ctx context.Context, ownerID string) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, display_name, description, metric_type, owner_id, status, version
		FROM neuronip.business_metrics
		WHERE owner_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics by owner: %w", err)
	}
	defer rows.Close()

	var metrics []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, displayName, metricType, status, version string
		var description, ownerID sql.NullString

		err := rows.Scan(&id, &name, &displayName, &description, &metricType, &ownerID, &status, &version)
		if err != nil {
			continue
		}

		metric := map[string]interface{}{
			"id":          id,
			"name":        name,
			"display_name": displayName,
			"metric_type": metricType,
			"status":      status,
			"version":     version,
		}

		if description.Valid {
			metric["description"] = description.String
		}
		if ownerID.Valid {
			metric["owner_id"] = ownerID.String
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

/* TransferOwnership transfers ownership of a metric */
func (s *MetricOwnershipService) TransferOwnership(ctx context.Context, metricID uuid.UUID, newOwnerID string, currentOwnerID string) error {
	// Verify current ownership
	var currentOwner sql.NullString
	checkQuery := `SELECT owner_id FROM neuronip.business_metrics WHERE id = $1`
	err := s.pool.QueryRow(ctx, checkQuery, metricID).Scan(&currentOwner)
	if err != nil {
		return fmt.Errorf("failed to verify ownership: %w", err)
	}

	if !currentOwner.Valid || currentOwner.String != currentOwnerID {
		return fmt.Errorf("current user is not the owner")
	}

	// Transfer ownership
	return s.UpdateMetricOwner(ctx, metricID, newOwnerID)
}

/* GetApprovalQueue retrieves pending approvals */
func (s *ApprovalService) GetApprovalQueue(ctx context.Context, approverID *string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT 
			approval_id,
			metric_id,
			metric_name,
			metric_display_name,
			approver_id,
			approval_status,
			comments,
			requested_at,
			updated_at,
			metric_status,
			primary_owner_id
		FROM neuronip.metric_approval_queue
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if approverID != nil {
		query += fmt.Sprintf(" AND approver_id = $%d", argIndex)
		args = append(args, *approverID)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY requested_at ASC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval queue: %w", err)
	}
	defer rows.Close()

	var queue []map[string]interface{}
	for rows.Next() {
		var approvalID, metricID uuid.UUID
		var metricName, metricDisplayName, approverID, approvalStatus, metricStatus string
		var comments, primaryOwnerID sql.NullString
		var requestedAt, updatedAt time.Time

		err := rows.Scan(
			&approvalID, &metricID, &metricName, &metricDisplayName,
			&approverID, &approvalStatus, &comments, &requestedAt,
			&updatedAt, &metricStatus, &primaryOwnerID,
		)
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"approval_id":          approvalID,
			"metric_id":            metricID,
			"metric_name":          metricName,
			"metric_display_name":  metricDisplayName,
			"approver_id":          approverID,
			"approval_status":      approvalStatus,
			"requested_at":         requestedAt,
			"updated_at":           updatedAt,
			"metric_status":        metricStatus,
		}

		if comments.Valid {
			item["comments"] = comments.String
		}
		if primaryOwnerID.Valid {
			item["primary_owner_id"] = primaryOwnerID.String
		}

		queue = append(queue, item)
	}

	return queue, nil
}

/* RequestMetricApproval requests approval for a metric */
func (s *ApprovalService) RequestMetricApproval(ctx context.Context, metricID uuid.UUID, approverID string, comments *string) (*MetricApproval, error) {
	return s.CreateApproval(ctx, metricID, approverID, comments)
}

/* AddMetricOwner adds an owner to a metric (multi-owner support) */
func (s *MetricOwnershipService) AddMetricOwner(ctx context.Context, metricID uuid.UUID, ownerID string, ownerType string) error {
	if ownerType == "" {
		ownerType = "secondary"
	}

	query := `
		INSERT INTO neuronip.metric_owners (metric_id, owner_id, owner_type, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (metric_id, owner_id) 
		DO UPDATE SET owner_type = EXCLUDED.owner_type, updated_at = NOW()`

	_, err := s.pool.Exec(ctx, query, metricID, ownerID, ownerType)
	if err != nil {
		return fmt.Errorf("failed to add metric owner: %w", err)
	}

	return nil
}

/* GetMetricOwners retrieves all owners of a metric */
func (s *MetricOwnershipService) GetMetricOwners(ctx context.Context, metricID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT owner_id, owner_type, created_at, updated_at
		FROM neuronip.metric_owners
		WHERE metric_id = $1
		ORDER BY owner_type, created_at`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric owners: %w", err)
	}
	defer rows.Close()

	var owners []map[string]interface{}
	for rows.Next() {
		var ownerID, ownerType string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&ownerID, &ownerType, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		owners = append(owners, map[string]interface{}{
			"owner_id":   ownerID,
			"owner_type": ownerType,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	return owners, nil
}

/* RemoveMetricOwner removes an owner from a metric */
func (s *MetricOwnershipService) RemoveMetricOwner(ctx context.Context, metricID uuid.UUID, ownerID string) error {
	query := `DELETE FROM neuronip.metric_owners WHERE metric_id = $1 AND owner_id = $2`

	result, err := s.pool.Exec(ctx, query, metricID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to remove metric owner: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("owner not found")
	}

	return nil
}
