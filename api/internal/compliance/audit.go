package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AuditService provides audit logging functionality */
type AuditService struct {
	pool *pgxpool.Pool
}

/* NewAuditService creates a new audit service */
func NewAuditService(pool *pgxpool.Pool) *AuditService {
	return &AuditService{pool: pool}
}

/* AuditEvent represents an audit event */
type AuditEvent struct {
	EventType  string                 `json:"event_type"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	UserID     *string                `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  *net.IP                `json:"ip_address,omitempty"`
	UserAgent  *string                `json:"user_agent,omitempty"`
}

/* LogEvent logs an audit event */
func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error {
	detailsJSON, _ := json.Marshal(event.Details)

	query := `
		INSERT INTO neuronip.audit_events 
		(event_type, entity_type, entity_id, user_id, action, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var ipAddr interface{}
	if event.IPAddress != nil {
		ipAddr = *event.IPAddress
	}

	_, err := s.pool.Exec(ctx, query,
		event.EventType,
		event.EntityType,
		event.EntityID,
		event.UserID,
		event.Action,
		detailsJSON,
		ipAddr,
		event.UserAgent,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

/* LogDocumentEvent is a helper to log document-related audit events */
func (s *AuditService) LogDocumentEvent(ctx context.Context, action string, documentID uuid.UUID, userID *string, details map[string]interface{}) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:  "document",
		EntityType: "knowledge_document",
		EntityID:   documentID.String(),
		UserID:     userID,
		Action:     action,
		Details:    details,
	})
}

/* LogSearchEvent is a helper to log search-related audit events */
func (s *AuditService) LogSearchEvent(ctx context.Context, query string, resultCount int, userID *string) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:  "search",
		EntityType: "semantic_search",
		EntityID:   uuid.New().String(), // Generate unique ID for search
		UserID:     userID,
		Action:     "search",
		Details: map[string]interface{}{
			"query":        query,
			"result_count": resultCount,
		},
	})
}

/* LogQueryEvent is a helper to log query-related audit events */
func (s *AuditService) LogQueryEvent(ctx context.Context, queryID uuid.UUID, action string, userID *string, details map[string]interface{}) error {
	return s.LogEvent(ctx, AuditEvent{
		EventType:  "query",
		EntityType: "warehouse_query",
		EntityID:   queryID.String(),
		UserID:     userID,
		Action:     action,
		Details:    details,
	})
}

/* GetAuditEvents retrieves audit events with filtering */
func (s *AuditService) GetAuditEvents(ctx context.Context, eventType, entityType, userID string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT event_type, entity_type, entity_id, user_id, action, details, ip_address, user_agent, created_at
		FROM neuronip.audit_events
		WHERE ($1 = '' OR event_type = $1) 
		  AND ($2 = '' OR entity_type = $2)
		  AND ($3 = '' OR user_id = $3)
		ORDER BY created_at DESC
		LIMIT $4`

	rows, err := s.pool.Query(ctx, query, eventType, entityType, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var eventType, entityType, entityID, action string
		var userID, userAgent *string
		var detailsJSON json.RawMessage
		var ipAddress *net.IP
		var createdAt time.Time

		err := rows.Scan(&eventType, &entityType, &entityID, &userID, &action, &detailsJSON, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			continue
		}

		event := map[string]interface{}{
			"event_type":  eventType,
			"entity_type": entityType,
			"entity_id":   entityID,
			"action":      action,
			"created_at":  createdAt,
		}

		if userID != nil {
			event["user_id"] = *userID
		}
		if userAgent != nil {
			event["user_agent"] = *userAgent
		}
		if ipAddress != nil {
			event["ip_address"] = ipAddress.String()
		}
		if detailsJSON != nil {
			var details map[string]interface{}
			json.Unmarshal(detailsJSON, &details)
			event["details"] = details
		}

		events = append(events, event)
	}

	return events, nil
}

/* GetActivityTimeline retrieves activity timeline for users and agents */
func (s *AuditService) GetActivityTimeline(ctx context.Context, userID string, limit int) ([]map[string]interface{}, error) {
	return s.GetAuditEvents(ctx, "", "", userID, limit)
}

/* GetComplianceTrail retrieves compliance audit trail */
func (s *AuditService) GetComplianceTrail(ctx context.Context, entityType, entityID string, limit int) ([]map[string]interface{}, error) {
	return s.GetAuditEvents(ctx, "compliance", entityType, "", limit)
}

/* SearchAuditEvents searches audit events by query */
func (s *AuditService) SearchAuditEvents(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}

	searchQuery := `
		SELECT event_type, entity_type, entity_id, user_id, action, details, ip_address, user_agent, created_at
		FROM neuronip.audit_events
		WHERE action ILIKE $1 OR details::text ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2`

	pattern := "%" + query + "%"
	rows, err := s.pool.Query(ctx, searchQuery, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search audit events: %w", err)
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var eventType, entityType, entityID, action string
		var userID, userAgent *string
		var detailsJSON json.RawMessage
		var ipAddress *net.IP
		var createdAt time.Time

		err := rows.Scan(&eventType, &entityType, &entityID, &userID, &action, &detailsJSON, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			continue
		}

		event := map[string]interface{}{
			"event_type":  eventType,
			"entity_type": entityType,
			"entity_id":   entityID,
			"action":      action,
			"created_at":  createdAt,
		}

		if userID != nil {
			event["user_id"] = *userID
		}
		if userAgent != nil {
			event["user_agent"] = *userAgent
		}
		if ipAddress != nil {
			event["ip_address"] = ipAddress.String()
		}
		if detailsJSON != nil {
			var details map[string]interface{}
			json.Unmarshal(detailsJSON, &details)
			event["details"] = details
		}

		events = append(events, event)
	}

	return events, nil
}
