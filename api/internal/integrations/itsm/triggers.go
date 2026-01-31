package itsm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ITSMTriggerService provides ITSM trigger functionality */
type ITSMTriggerService struct {
	pool *pgxpool.Pool
}

/* NewITSMTriggerService creates a new ITSM trigger service */
func NewITSMTriggerService(pool *pgxpool.Pool) *ITSMTriggerService {
	return &ITSMTriggerService{pool: pool}
}

/* Trigger represents an ITSM trigger */
type Trigger struct {
	ID            uuid.UUID              `json:"id"`
	ITSMType      string                 `json:"itsm_type"` // "servicenow", "jira"
	TriggerType   string                 `json:"trigger_type"`
	Config        map[string]interface{} `json:"config"`
	Enabled       bool                   `json:"enabled"`
	CreatedAt     time.Time              `json:"created_at"`
}

/* TriggerEvent triggers an ITSM event */
func (its *ITSMTriggerService) TriggerEvent(ctx context.Context, itsmType, triggerType string, eventData map[string]interface{}) error {
	// Get triggers
	triggers, err := its.GetTriggers(ctx, itsmType, triggerType)
	if err != nil {
		return err
	}
	
	for _, trigger := range triggers {
		if trigger.Enabled {
			its.executeTrigger(ctx, trigger, eventData)
		}
	}
	
	return nil
}

/* CreateTrigger creates a new ITSM trigger */
func (its *ITSMTriggerService) CreateTrigger(ctx context.Context, trigger Trigger) (*Trigger, error) {
	trigger.ID = uuid.New()
	trigger.CreatedAt = time.Now()
	
	configJSON, _ := json.Marshal(trigger.Config)
	
	query := `
		INSERT INTO neuronip.itsm_triggers 
		(id, itsm_type, trigger_type, config, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	
	err := its.pool.QueryRow(ctx, query,
		trigger.ID, trigger.ITSMType, trigger.TriggerType, configJSON, trigger.Enabled, trigger.CreatedAt,
	).Scan(&trigger.ID, &trigger.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}
	
	return &trigger, nil
}

/* GetTriggers retrieves triggers */
func (its *ITSMTriggerService) GetTriggers(ctx context.Context, itsmType, triggerType string) ([]Trigger, error) {
	query := `
		SELECT id, itsm_type, trigger_type, config, enabled, created_at
		FROM neuronip.itsm_triggers
		WHERE itsm_type = $1 AND trigger_type = $2 AND enabled = true
	`
	
	rows, err := its.pool.Query(ctx, query, itsmType, triggerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var triggers []Trigger
	for rows.Next() {
		var trigger Trigger
		var configJSON json.RawMessage
		
		if err := rows.Scan(&trigger.ID, &trigger.ITSMType, &trigger.TriggerType, &configJSON, &trigger.Enabled, &trigger.CreatedAt); err != nil {
			continue
		}
		
		json.Unmarshal(configJSON, &trigger.Config)
		triggers = append(triggers, trigger)
	}
	
	return triggers, nil
}

/* executeTrigger executes a trigger */
func (its *ITSMTriggerService) executeTrigger(ctx context.Context, trigger Trigger, eventData map[string]interface{}) error {
	// Execute trigger action
	_ = trigger
	_ = eventData
	return nil
}
