package collaboration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* DecisionHistoryService provides decision history tracking functionality */
type DecisionHistoryService struct {
	pool *pgxpool.Pool
}

/* NewDecisionHistoryService creates a new decision history service */
func NewDecisionHistoryService(pool *pgxpool.Pool) *DecisionHistoryService {
	return &DecisionHistoryService{pool: pool}
}

/* Decision represents a decision in history */
type Decision struct {
	ID            uuid.UUID              `json:"id"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    uuid.UUID              `json:"resource_id"`
	DecisionType  string                 `json:"decision_type"` // "approval", "rejection", "change", "creation"
	Decision      string                 `json:"decision"`
	Reasoning     string                 `json:"reasoning,omitempty"`
	MadeBy        string                 `json:"made_by"`
	Participants  []string               `json:"participants,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* RecordDecision records a decision in history */
func (dhs *DecisionHistoryService) RecordDecision(ctx context.Context, decision Decision) error {
	decision.ID = uuid.New()
	decision.CreatedAt = time.Now()
	
	participantsJSON, _ := json.Marshal(decision.Participants)
	contextJSON, _ := json.Marshal(decision.Context)
	metadataJSON, _ := json.Marshal(decision.Metadata)
	
	query := `
		INSERT INTO neuronip.decision_history 
		(id, resource_type, resource_id, decision_type, decision, reasoning, made_by, participants, context, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err := dhs.pool.Exec(ctx, query,
		decision.ID, decision.ResourceType, decision.ResourceID, decision.DecisionType,
		decision.Decision, decision.Reasoning, decision.MadeBy, participantsJSON,
		contextJSON, metadataJSON, decision.CreatedAt,
	)
	return err
}

/* GetDecisionHistory retrieves decision history for a resource */
func (dhs *DecisionHistoryService) GetDecisionHistory(ctx context.Context, resourceType string, resourceID uuid.UUID, limit int) ([]Decision, error) {
	if limit <= 0 {
		limit = 100
	}
	
	query := `
		SELECT id, resource_type, resource_id, decision_type, decision, reasoning, made_by, participants, context, metadata, created_at
		FROM neuronip.decision_history
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`
	
	rows, err := dhs.pool.Query(ctx, query, resourceType, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision history: %w", err)
	}
	defer rows.Close()
	
	var decisions []Decision
	for rows.Next() {
		var decision Decision
		var participantsJSON, contextJSON, metadataJSON json.RawMessage
		
		err := rows.Scan(
			&decision.ID, &decision.ResourceType, &decision.ResourceID, &decision.DecisionType,
			&decision.Decision, &decision.Reasoning, &decision.MadeBy, &participantsJSON,
			&contextJSON, &metadataJSON, &decision.CreatedAt,
		)
		if err != nil {
			continue
		}
		
		if participantsJSON != nil {
			json.Unmarshal(participantsJSON, &decision.Participants)
		}
		if contextJSON != nil {
			json.Unmarshal(contextJSON, &decision.Context)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &decision.Metadata)
		}
		
		decisions = append(decisions, decision)
	}
	
	return decisions, nil
}
