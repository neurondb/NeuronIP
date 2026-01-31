package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* AuditTrailService provides decision audit trail functionality for agents */
type AuditTrailService struct {
	pool *pgxpool.Pool
}

/* NewAuditTrailService creates a new audit trail service */
func NewAuditTrailService(pool *pgxpool.Pool) *AuditTrailService {
	return &AuditTrailService{pool: pool}
}

/* AuditEntry represents an audit trail entry */
type AuditEntry struct {
	ID            uuid.UUID              `json:"id"`
	TraceID       uuid.UUID              `json:"trace_id"`
	AgentID       string                 `json:"agent_id"`
	DecisionType  string                 `json:"decision_type"` // "tool_selection", "response_generation", "reasoning_step"
	Decision      map[string]interface{} `json:"decision"`
	Reasoning     string                 `json:"reasoning,omitempty"`
	Alternatives  []map[string]interface{} `json:"alternatives,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	UserID        *string                `json:"user_id,omitempty"`
}

/* LogDecision logs a decision to the audit trail */
func (ats *AuditTrailService) LogDecision(ctx context.Context, entry AuditEntry) error {
	entry.ID = uuid.New()
	entry.Timestamp = time.Now()
	
	decisionJSON, _ := json.Marshal(entry.Decision)
	alternativesJSON, _ := json.Marshal(entry.Alternatives)
	contextJSON, _ := json.Marshal(entry.Context)
	
	query := `
		INSERT INTO neuronip.agent_audit_trail 
		(id, trace_id, agent_id, decision_type, decision, reasoning, alternatives, context, timestamp, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err := ats.pool.Exec(ctx, query,
		entry.ID, entry.TraceID, entry.AgentID, entry.DecisionType, decisionJSON,
		entry.Reasoning, alternativesJSON, contextJSON, entry.Timestamp, entry.UserID,
	)
	return err
}

/* GetAuditTrail retrieves audit trail entries for a trace */
func (ats *AuditTrailService) GetAuditTrail(ctx context.Context, traceID uuid.UUID) ([]AuditEntry, error) {
	query := `
		SELECT id, trace_id, agent_id, decision_type, decision, reasoning, alternatives, context, timestamp, user_id
		FROM neuronip.agent_audit_trail
		WHERE trace_id = $1
		ORDER BY timestamp ASC
	`
	
	rows, err := ats.pool.Query(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}
	defer rows.Close()
	
	var entries []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var decisionJSON, alternativesJSON, contextJSON json.RawMessage
		var userID *string
		
		err := rows.Scan(
			&entry.ID, &entry.TraceID, &entry.AgentID, &entry.DecisionType,
			&decisionJSON, &entry.Reasoning, &alternativesJSON, &contextJSON,
			&entry.Timestamp, &userID,
		)
		if err != nil {
			continue
		}
		
		entry.UserID = userID
		if decisionJSON != nil {
			json.Unmarshal(decisionJSON, &entry.Decision)
		}
		if alternativesJSON != nil {
			json.Unmarshal(alternativesJSON, &entry.Alternatives)
		}
		if contextJSON != nil {
			json.Unmarshal(contextJSON, &entry.Context)
		}
		
		entries = append(entries, entry)
	}
	
	return entries, nil
}

/* GetAuditTrailForAgent retrieves audit trail entries for an agent */
func (ats *AuditTrailService) GetAuditTrailForAgent(ctx context.Context, agentID string, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	
	query := `
		SELECT id, trace_id, agent_id, decision_type, decision, reasoning, alternatives, context, timestamp, user_id
		FROM neuronip.agent_audit_trail
		WHERE agent_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	
	rows, err := ats.pool.Query(ctx, query, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit trail: %w", err)
	}
	defer rows.Close()
	
	var entries []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		var decisionJSON, alternativesJSON, contextJSON json.RawMessage
		var userID *string
		
		err := rows.Scan(
			&entry.ID, &entry.TraceID, &entry.AgentID, &entry.DecisionType,
			&decisionJSON, &entry.Reasoning, &alternativesJSON, &contextJSON,
			&entry.Timestamp, &userID,
		)
		if err != nil {
			continue
		}
		
		entry.UserID = userID
		if decisionJSON != nil {
			json.Unmarshal(decisionJSON, &entry.Decision)
		}
		if alternativesJSON != nil {
			json.Unmarshal(alternativesJSON, &entry.Alternatives)
		}
		if contextJSON != nil {
			json.Unmarshal(contextJSON, &entry.Context)
		}
		
		entries = append(entries, entry)
	}
	
	return entries, nil
}
