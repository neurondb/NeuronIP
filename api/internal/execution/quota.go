package execution

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* QuotaService provides resource quota functionality */
type QuotaService struct {
	pool *pgxpool.Pool
}

/* NewQuotaService creates a new quota service */
func NewQuotaService(pool *pgxpool.Pool) *QuotaService {
	return &QuotaService{pool: pool}
}

/* ResourceQuota represents a resource quota */
type ResourceQuota struct {
	ID           uuid.UUID  `json:"id"`
	WorkspaceID  *uuid.UUID `json:"workspace_id,omitempty"`
	UserID       *string    `json:"user_id,omitempty"`
	ResourceType string     `json:"resource_type"`
	MaxLimit     int64      `json:"max_limit"`
	CurrentUsage int64      `json:"current_usage"`
	Period       string     `json:"period"`
	ResetAt      time.Time  `json:"reset_at"`
	Enabled      bool       `json:"enabled"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

/* SetQuota sets a resource quota for a user or workspace */
func (s *QuotaService) SetQuota(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, maxLimit int64, period string) (*ResourceQuota, error) {
	id := uuid.New()
	now := time.Now()
	resetAt := s.calculateResetTime(period)

	query := `
		INSERT INTO neuronip.resource_quotas 
		(id, workspace_id, user_id, resource_type, max_limit, current_usage, period, reset_at, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, true, $8, $9)
		ON CONFLICT (workspace_id, user_id, resource_type, period) 
		DO UPDATE SET 
			max_limit = EXCLUDED.max_limit,
			updated_at = EXCLUDED.updated_at
		RETURNING id, workspace_id, user_id, resource_type, max_limit, current_usage, period, reset_at, enabled, created_at, updated_at`

	var quota ResourceQuota
	var workspaceIDVal sql.NullString
	var userIDVal sql.NullString

	var workspaceUUID *uuid.UUID
	var userIDStr *string
	if workspaceID != nil {
		workspaceUUID = workspaceID
	}
	if userID != nil {
		userIDStr = userID
	}

	err := s.pool.QueryRow(ctx, query, id, workspaceUUID, userIDStr, resourceType, maxLimit, period, resetAt, now, now).Scan(
		&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
		&quota.MaxLimit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
		&quota.Enabled, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set quota: %w", err)
	}

	if workspaceIDVal.Valid {
		wsID, _ := uuid.Parse(workspaceIDVal.String)
		quota.WorkspaceID = &wsID
	}
	if userIDVal.Valid {
		quota.UserID = &userIDVal.String
	}

	return &quota, nil
}

/* CheckQuota checks if a resource usage is within quota */
func (s *QuotaService) CheckQuota(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, requestedUsage int64) (bool, *ResourceQuota, error) {
	query := `
		SELECT id, workspace_id, user_id, resource_type, max_limit, current_usage, period, reset_at, enabled, created_at, updated_at
		FROM neuronip.resource_quotas
		WHERE resource_type = $1 AND enabled = true`

	args := []interface{}{resourceType}
	argIndex := 2

	if workspaceID != nil {
		query += fmt.Sprintf(" AND workspace_id = $%d", argIndex)
		args = append(args, *workspaceID)
		argIndex++
	} else {
		query += " AND workspace_id IS NULL"
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	} else {
		query += " AND user_id IS NULL"
	}

	// Select the most restrictive period (month > day > hour)
	query += " ORDER BY CASE period WHEN 'month' THEN 1 WHEN 'day' THEN 2 WHEN 'hour' THEN 3 END LIMIT 1"

	var quota ResourceQuota
	var workspaceIDVal sql.NullString
	var userIDVal sql.NullString

	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
		&quota.MaxLimit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
		&quota.Enabled, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err != nil {
		// No quota set, allow usage
		return true, nil, nil
	}

	// Check if reset is needed
	if time.Now().After(quota.ResetAt) {
		// Reset usage
		newResetAt := s.calculateResetTime(quota.Period)
		s.pool.Exec(ctx, `
			UPDATE neuronip.resource_quotas 
			SET current_usage = 0, reset_at = $1, updated_at = NOW()
			WHERE id = $2`, newResetAt, quota.ID)
		quota.CurrentUsage = 0
		quota.ResetAt = newResetAt
	}

	if workspaceIDVal.Valid {
		wsID, _ := uuid.Parse(workspaceIDVal.String)
		quota.WorkspaceID = &wsID
	}
	if userIDVal.Valid {
		quota.UserID = &userIDVal.String
	}

	// Check if usage would exceed quota
	allowed := quota.CurrentUsage+requestedUsage <= quota.MaxLimit
	return allowed, &quota, nil
}

/* RecordUsage records resource usage */
func (s *QuotaService) RecordUsage(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, usage int64) error {
	query := `
		UPDATE neuronip.resource_quotas 
		SET current_usage = current_usage + $1, updated_at = NOW()
		WHERE resource_type = $2 AND enabled = true`

	args := []interface{}{usage, resourceType}
	argIndex := 3

	if workspaceID != nil {
		query += fmt.Sprintf(" AND workspace_id = $%d", argIndex)
		args = append(args, *workspaceID)
		argIndex++
	} else {
		query += " AND workspace_id IS NULL"
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	} else {
		query += " AND user_id IS NULL"
	}

	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

/* GetQuotaUsage retrieves quota usage */
func (s *QuotaService) GetQuotaUsage(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType *string) ([]ResourceQuota, error) {
	query := `
		SELECT id, workspace_id, user_id, resource_type, max_limit, current_usage, period, reset_at, enabled, created_at, updated_at
		FROM neuronip.resource_quotas
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if workspaceID != nil {
		query += fmt.Sprintf(" AND workspace_id = $%d", argIndex)
		args = append(args, *workspaceID)
		argIndex++
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	if resourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *resourceType)
		argIndex++
	}

	query += " ORDER BY resource_type, period"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota usage: %w", err)
	}
	defer rows.Close()

	var quotas []ResourceQuota
	for rows.Next() {
		var quota ResourceQuota
		var workspaceIDVal sql.NullString
		var userIDVal sql.NullString

		err := rows.Scan(
			&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
			&quota.MaxLimit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
			&quota.Enabled, &quota.CreatedAt, &quota.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if workspaceIDVal.Valid {
			wsID, _ := uuid.Parse(workspaceIDVal.String)
			quota.WorkspaceID = &wsID
		}
		if userIDVal.Valid {
			quota.UserID = &userIDVal.String
		}

		quotas = append(quotas, quota)
	}

	return quotas, nil
}

/* calculateResetTime calculates the reset time based on period */
func (s *QuotaService) calculateResetTime(period string) time.Time {
	now := time.Now()
	switch period {
	case "hour":
		return now.Add(1 * time.Hour)
	case "day":
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case "month":
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	default:
		return now.Add(24 * time.Hour)
	}
}
