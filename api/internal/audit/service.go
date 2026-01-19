package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AuditService provides comprehensive audit logging */
type AuditService struct {
	pool *pgxpool.Pool
}

/* NewAuditService creates a new audit service */
func NewAuditService(pool *pgxpool.Pool) *AuditService {
	return &AuditService{pool: pool}
}

/* AuditLog represents an audit log entry */
type AuditLog struct {
	ID           uuid.UUID              `json:"id"`
	UserID       *string                `json:"user_id,omitempty"`
	ActionType   string                 `json:"action_type"`
	ResourceType *string                `json:"resource_type,omitempty"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	Action       string                 `json:"action"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	Status       string                 `json:"status"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	DurationMs   *int                   `json:"duration_ms,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

/* LogAction logs an action to the audit trail */
func (s *AuditService) LogAction(ctx context.Context, log AuditLog) error {
	log.ID = uuid.New()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	
	detailsJSON, _ := json.Marshal(log.Details)
	
	query := `
		INSERT INTO neuronip.audit_logs 
		(id, user_id, action_type, resource_type, resource_id, action, details,
		 ip_address, user_agent, status, error_message, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	_, err := s.pool.Exec(ctx, query,
		log.ID, log.UserID, log.ActionType, log.ResourceType, log.ResourceID, log.Action,
		detailsJSON, log.IPAddress, log.UserAgent, log.Status, log.ErrorMessage, log.DurationMs, log.CreatedAt,
	)
	
	return err
}

/* GetAuditLogs retrieves audit logs with filtering */
func (s *AuditService) GetAuditLogs(ctx context.Context, filters AuditFilters, limit int) ([]AuditLog, error) {
	query := `
		SELECT id, user_id, action_type, resource_type, resource_id, action, details,
		       ip_address, user_agent, status, error_message, duration_ms, created_at
		FROM neuronip.audit_logs
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}
	
	if filters.ActionType != nil {
		query += fmt.Sprintf(" AND action_type = $%d", argIndex)
		args = append(args, *filters.ActionType)
		argIndex++
	}
	
	if filters.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, *filters.ResourceType)
		argIndex++
	}
	
	if filters.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *filters.ResourceID)
		argIndex++
	}
	
	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}
	
	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}
	
	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()
	
	logs := make([]AuditLog, 0)
	for rows.Next() {
		var log AuditLog
		var detailsJSON json.RawMessage
		
		err := rows.Scan(
			&log.ID, &log.UserID, &log.ActionType, &log.ResourceType, &log.ResourceID, &log.Action,
			&detailsJSON, &log.IPAddress, &log.UserAgent, &log.Status, &log.ErrorMessage, &log.DurationMs, &log.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		if detailsJSON != nil {
			json.Unmarshal(detailsJSON, &log.Details)
		}
		
		logs = append(logs, log)
	}
	
	return logs, nil
}

/* AuditFilters filters for audit log queries */
type AuditFilters struct {
	UserID       *string
	ActionType   *string
	ResourceType *string
	ResourceID   *string
	Status       *string
	StartTime    *time.Time
	EndTime      *time.Time
}

/* LogQueryAction logs a query action */
func (s *AuditService) LogQueryAction(ctx context.Context, userID string, queryID uuid.UUID, query string, status string, durationMs int, errorMsg *string) error {
	return s.LogAction(ctx, AuditLog{
		UserID:       &userID,
		ActionType:   "query",
		ResourceType: stringPtr("query"),
		ResourceID:   stringPtr(queryID.String()),
		Action:       "execute_query",
		Details: map[string]interface{}{
			"query": query,
		},
		Status:       status,
		ErrorMessage: errorMsg,
		DurationMs:   &durationMs,
	})
}

/* LogAgentAction logs an agent action */
func (s *AuditService) LogAgentAction(ctx context.Context, userID string, agentID uuid.UUID, action string, status string, details map[string]interface{}) error {
	return s.LogAction(ctx, AuditLog{
		UserID:       &userID,
		ActionType:   "agent_execution",
		ResourceType: stringPtr("agent"),
		ResourceID:   stringPtr(agentID.String()),
		Action:       action,
		Details:      details,
		Status:       status,
	})
}

/* LogAIAction logs an AI-related action */
func (s *AuditService) LogAIAction(ctx context.Context, userID string, action string, resourceType string, resourceID string, details map[string]interface{}) error {
	return s.LogAction(ctx, AuditLog{
		UserID:       &userID,
		ActionType:   "ai_action",
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Action:       action,
		Details:      details,
		Status:       "success",
	})
}

func stringPtr(s string) *string {
	return &s
}
