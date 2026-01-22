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

/* QuotaService provides resource quota management */
type QuotaService struct {
	pool *pgxpool.Pool
}

/* NewQuotaService creates a new quota service */
func NewQuotaService(pool *pgxpool.Pool) *QuotaService {
	return &QuotaService{pool: pool}
}

/* ResourceQuota represents a resource quota */
type ResourceQuota struct {
	ID                uuid.UUID              `json:"id"`
	WorkspaceID       *uuid.UUID             `json:"workspace_id,omitempty"`
	UserID            *string                `json:"user_id,omitempty"`
	ResourceType      string                 `json:"resource_type"` // query, agent, storage, etc.
	Limit             int64                  `json:"limit"`
	CurrentUsage      int64                  `json:"current_usage"`
	Period            string                 `json:"period"` // daily, weekly, monthly
	ResetAt           time.Time              `json:"reset_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

/* SetQuota sets a resource quota for a workspace or user */
func (s *QuotaService) SetQuota(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, limit int64, period string) (*ResourceQuota, error) {
	id := uuid.New()
	now := time.Now()
	
	// Calculate reset time based on period
	var resetAt time.Time
	switch period {
	case "daily":
		resetAt = now.Add(24 * time.Hour)
		resetAt = time.Date(resetAt.Year(), resetAt.Month(), resetAt.Day(), 0, 0, 0, 0, resetAt.Location())
	case "weekly":
		resetAt = now.Add(7 * 24 * time.Hour)
		// Reset on Monday
		daysUntilMonday := (8 - int(resetAt.Weekday())) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		resetAt = resetAt.Add(time.Duration(daysUntilMonday) * 24 * time.Hour)
		resetAt = time.Date(resetAt.Year(), resetAt.Month(), resetAt.Day(), 0, 0, 0, 0, resetAt.Location())
	case "monthly":
		resetAt = now.AddDate(0, 1, 0)
		resetAt = time.Date(resetAt.Year(), resetAt.Month(), 1, 0, 0, 0, 0, resetAt.Location())
	default:
		resetAt = now.Add(24 * time.Hour)
	}

	query := `
		INSERT INTO neuronip.resource_quotas 
		(id, workspace_id, user_id, resource_type, limit_value, current_usage, period, reset_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, $9)
		ON CONFLICT (COALESCE(workspace_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(user_id, ''), resource_type, period)
		DO UPDATE SET 
			limit_value = EXCLUDED.limit_value,
			reset_at = EXCLUDED.reset_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, workspace_id, user_id, resource_type, limit_value, current_usage, period, reset_at, metadata, created_at, updated_at`

	var quota ResourceQuota
	var workspaceIDVal, userIDVal sql.NullString
	var metadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, id, workspaceID, userID, resourceType, limit, period, resetAt, now, now).Scan(
		&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
		&quota.Limit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
		&metadataJSON, &quota.CreatedAt, &quota.UpdatedAt,
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
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &quota.Metadata)
	}

	return &quota, nil
}

/* CheckQuota checks if a resource usage is within quota */
func (s *QuotaService) CheckQuota(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, requestedAmount int64) (bool, *ResourceQuota, error) {
	query := `
		SELECT id, workspace_id, user_id, resource_type, limit_value, current_usage, period, reset_at, metadata, created_at, updated_at
		FROM neuronip.resource_quotas
		WHERE resource_type = $1
		  AND (workspace_id = $2 OR ($2 IS NULL AND workspace_id IS NULL))
		  AND (user_id = $3 OR ($3 IS NULL AND user_id IS NULL))
		  AND reset_at > NOW()`

	var quota ResourceQuota
	var workspaceIDVal, userIDVal sql.NullString
	var metadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, resourceType, workspaceID, userID).Scan(
		&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
		&quota.Limit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
		&metadataJSON, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// No quota set, allow
		return true, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("failed to check quota: %w", err)
	}

	if workspaceIDVal.Valid {
		wsID, _ := uuid.Parse(workspaceIDVal.String)
		quota.WorkspaceID = &wsID
	}
	if userIDVal.Valid {
		quota.UserID = &userIDVal.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &quota.Metadata)
	}

	// Check if usage would exceed limit
	withinQuota := quota.CurrentUsage+requestedAmount <= quota.Limit
	return withinQuota, &quota, nil
}

/* IncrementUsage increments resource usage */
func (s *QuotaService) IncrementUsage(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string, amount int64) error {
	query := `
		UPDATE neuronip.resource_quotas 
		SET current_usage = current_usage + $1, updated_at = NOW()
		WHERE resource_type = $2
		  AND (workspace_id = $3 OR ($3 IS NULL AND workspace_id IS NULL))
		  AND (user_id = $4 OR ($4 IS NULL AND user_id IS NULL))
		  AND reset_at > NOW()`

	result, err := s.pool.Exec(ctx, query, amount, resourceType, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	if result.RowsAffected() == 0 {
		// No quota exists, that's fine
		return nil
	}

	return nil
}

/* ResetQuotas resets quotas that have passed their reset time */
func (s *QuotaService) ResetQuotas(ctx context.Context) error {
	query := `
		UPDATE neuronip.resource_quotas 
		SET current_usage = 0, 
		    reset_at = CASE 
		        WHEN period = 'daily' THEN reset_at + INTERVAL '1 day'
		        WHEN period = 'weekly' THEN reset_at + INTERVAL '1 week'
		        WHEN period = 'monthly' THEN reset_at + INTERVAL '1 month'
		        ELSE reset_at + INTERVAL '1 day'
		    END,
		    updated_at = NOW()
		WHERE reset_at <= NOW()`

	_, err := s.pool.Exec(ctx, query)
	return err
}

/* GetQuota retrieves quota for a workspace/user */
func (s *QuotaService) GetQuota(ctx context.Context, workspaceID *uuid.UUID, userID *string, resourceType string) (*ResourceQuota, error) {
	query := `
		SELECT id, workspace_id, user_id, resource_type, limit_value, current_usage, period, reset_at, metadata, created_at, updated_at
		FROM neuronip.resource_quotas
		WHERE resource_type = $1
		  AND (workspace_id = $2 OR ($2 IS NULL AND workspace_id IS NULL))
		  AND (user_id = $3 OR ($3 IS NULL AND user_id IS NULL))
		ORDER BY reset_at DESC
		LIMIT 1`

	var quota ResourceQuota
	var workspaceIDVal, userIDVal sql.NullString
	var metadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, resourceType, workspaceID, userID).Scan(
		&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
		&quota.Limit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
		&metadataJSON, &quota.CreatedAt, &quota.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	if workspaceIDVal.Valid {
		wsID, _ := uuid.Parse(workspaceIDVal.String)
		quota.WorkspaceID = &wsID
	}
	if userIDVal.Valid {
		quota.UserID = &userIDVal.String
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &quota.Metadata)
	}

	return &quota, nil
}

/* ListQuotas lists all quotas */
func (s *QuotaService) ListQuotas(ctx context.Context, workspaceID *uuid.UUID, userID *string) ([]ResourceQuota, error) {
	query := `
		SELECT id, workspace_id, user_id, resource_type, limit_value, current_usage, period, reset_at, metadata, created_at, updated_at
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

	query += " ORDER BY resource_type, created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}
	defer rows.Close()

	var quotas []ResourceQuota
	for rows.Next() {
		var quota ResourceQuota
		var workspaceIDVal, userIDVal sql.NullString
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&quota.ID, &workspaceIDVal, &userIDVal, &quota.ResourceType,
			&quota.Limit, &quota.CurrentUsage, &quota.Period, &quota.ResetAt,
			&metadataJSON, &quota.CreatedAt, &quota.UpdatedAt,
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
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &quota.Metadata)
		}

		quotas = append(quotas, quota)
	}

	return quotas, nil
}
