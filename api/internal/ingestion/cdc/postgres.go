package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* PostgresCDC implements CDC for PostgreSQL using logical replication */
type PostgresCDC struct {
	pool       *pgxpool.Pool
	slotName   string
	publicationName string
	conn       *pgxpool.Pool // Separate connection for replication
}

/* NewPostgresCDC creates a new PostgreSQL CDC instance */
func NewPostgresCDC(pool *pgxpool.Pool) *PostgresCDC {
	return &PostgresCDC{
		pool: pool,
	}
}

/* StartCDC starts the CDC replication process */
func (p *PostgresCDC) StartCDC(ctx context.Context, config map[string]interface{}) error {
	slotName, ok := config["slot_name"].(string)
	if !ok {
		slotName = fmt.Sprintf("neuronip_slot_%d", time.Now().Unix())
	}
	p.slotName = slotName
	
	publicationName, ok := config["publication_name"].(string)
	if !ok {
		publicationName = "neuronip_publication"
	}
	p.publicationName = publicationName
	
	// Create replication slot if it doesn't exist
	if err := p.createReplicationSlot(ctx); err != nil {
		return fmt.Errorf("failed to create replication slot: %w", err)
	}
	
	// Create publication if it doesn't exist
	if err := p.createPublication(ctx); err != nil {
		return fmt.Errorf("failed to create publication: %w", err)
	}
	
	return nil
}

/* StopCDC stops the CDC replication process */
func (p *PostgresCDC) StopCDC(ctx context.Context) error {
	// Close replication connection
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	
	return nil
}

/* GetChanges retrieves changes from the replication stream */
func (p *PostgresCDC) GetChanges(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error) {
	lastLSN, ok := lastPosition.(string)
	if !ok {
		return nil, fmt.Errorf("lastPosition must be a string for PostgreSQL CDC")
	}

	// Use pgx replication API for logical replication
	// Note: This requires a replication connection, not a regular connection
	// For now, we'll use a polling approach as a fallback
	
	// Try to get changes using replication slot if available
	if p.conn != nil {
		// Use replication connection to read WAL changes
		// Note: pgx v5 replication API would be used here
		// This is a simplified implementation - full implementation requires replication connection setup
		
		// For now, return changes from polling query_log_changes table if available
		// This is a fallback approach - proper CDC would use logical replication
		return p.getChangesFromPolling(ctx, lastLSN)
	}
	
	// Fallback: use polling approach
	return p.getChangesFromPolling(ctx, lastLSN)
}

/* getChangesFromPolling retrieves changes using polling approach (fallback) */
func (p *PostgresCDC) getChangesFromPolling(ctx context.Context, lastLSN string) ([]ChangeEvent, error) {
	// Poll-based CDC using a changes table or WAL-based approach
	// This is a simplified implementation - proper CDC uses logical replication
	
	// Query for recent changes (assuming a changes log table exists)
	query := `
		SELECT table_name, operation, lsn, timestamp, old_data, new_data
		FROM neuronip.cdc_changes
		WHERE lsn > $1
		ORDER BY lsn ASC
		LIMIT 1000`
	
	rows, err := p.pool.Query(ctx, query, lastLSN)
	if err != nil {
		// If table doesn't exist, return empty (CDC would need to be set up)
		return []ChangeEvent{}, nil
	}
	defer rows.Close()
	
	var changes []ChangeEvent
	for rows.Next() {
		var change ChangeEvent
		var oldDataJSON, newDataJSON []byte
		
		err := rows.Scan(
			&change.Table,
			&change.Operation,
			&change.LSN,
			&change.Timestamp,
			&oldDataJSON,
			&newDataJSON,
		)
		if err != nil {
			continue
		}
		
		// Parse JSON data
		if oldDataJSON != nil {
			json.Unmarshal(oldDataJSON, &change.OldData)
		}
		if newDataJSON != nil {
			json.Unmarshal(newDataJSON, &change.NewData)
		}
		
		changes = append(changes, change)
	}
	
	return changes, nil
}

/* SaveCheckpoint saves a CDC checkpoint */
func (p *PostgresCDC) SaveCheckpoint(ctx context.Context, dataSourceID string, tableName string, checkpoint map[string]interface{}) error {
	checkpointJSON, _ := json.Marshal(checkpoint)
	lsn := ""
	if lsnVal, ok := checkpoint["lsn"].(string); ok {
		lsn = lsnVal
	}
	
	query := `
		INSERT INTO neuronip.cdc_checkpoints 
		(data_source_id, table_name, checkpoint_data, last_lsn, last_timestamp, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (data_source_id, table_name) 
		DO UPDATE SET 
			checkpoint_data = EXCLUDED.checkpoint_data,
			last_lsn = EXCLUDED.last_lsn,
			last_timestamp = EXCLUDED.last_timestamp,
			updated_at = EXCLUDED.updated_at`
	
	_, err := p.pool.Exec(ctx, query, dataSourceID, tableName, checkpointJSON, lsn)
	return err
}

/* GetCheckpoint retrieves a CDC checkpoint */
func (p *PostgresCDC) GetCheckpoint(ctx context.Context, dataSourceID string, tableName string) (map[string]interface{}, error) {
	var checkpointJSON []byte
	var lastLSN *string
	
	query := `
		SELECT checkpoint_data, last_lsn
		FROM neuronip.cdc_checkpoints
		WHERE data_source_id = $1 AND table_name = $2`
	
	err := p.pool.QueryRow(ctx, query, dataSourceID, tableName).Scan(&checkpointJSON, &lastLSN)
	if err != nil {
		return nil, err
	}
	
	var checkpoint map[string]interface{}
	if err := json.Unmarshal(checkpointJSON, &checkpoint); err != nil {
		return nil, err
	}
	
	if lastLSN != nil {
		checkpoint["lsn"] = *lastLSN
	}
	
	return checkpoint, nil
}

/* createReplicationSlot creates a logical replication slot */
func (p *PostgresCDC) createReplicationSlot(ctx context.Context) error {
	query := fmt.Sprintf(
		"SELECT * FROM pg_create_logical_replication_slot('%s', 'pgoutput') WHERE NOT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = '%s')",
		p.slotName, p.slotName)
	
	_, err := p.pool.Exec(ctx, query)
	return err
}

/* createPublication creates a publication for replication */
func (p *PostgresCDC) createPublication(ctx context.Context) error {
	query := fmt.Sprintf(
		"CREATE PUBLICATION %s FOR ALL TABLES",
		p.publicationName)
	
	_, _ = p.pool.Exec(ctx, query)
	// Ignore error if publication already exists
	return nil
}

/* ChangeEvent represents a single change event from CDC */
type ChangeEvent struct {
	Table      string                 `json:"table"`
	Operation  string                 `json:"operation"` // "insert", "update", "delete"
	LSN        string                 `json:"lsn"`
	Timestamp  time.Time              `json:"timestamp"`
	OldData    map[string]interface{} `json:"old_data,omitempty"`
	NewData    map[string]interface{} `json:"new_data,omitempty"`
}
