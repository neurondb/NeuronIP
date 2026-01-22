package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PostgresCDC implements CDC for PostgreSQL using logical replication */
type PostgresCDC struct {
	pool           *pgxpool.Pool
	slotName       string
	publicationName string
	replConn       *pgx.Conn // Separate replication connection (requires replication privilege)
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
	
	// Attempt to establish replication connection
	// This requires REPLICATION privilege on the database user
	if err := p.establishReplicationConnection(ctx, config); err != nil {
		// Log warning but continue with polling fallback
		// Replication connection requires special privileges
	}
	
	return nil
}

/* establishReplicationConnection establishes a replication connection */
func (p *PostgresCDC) establishReplicationConnection(ctx context.Context, config map[string]interface{}) error {
	// Get connection config from pool
	connConfig := p.pool.Config()
	
	// Create a new config for replication connection
	replConfig := connConfig.Copy()
	replConfig.RuntimeParams["replication"] = "database"
	
	// Create replication connection
	conn, err := pgx.ConnectConfig(ctx, replConfig)
	if err != nil {
		return fmt.Errorf("failed to create replication connection (requires REPLICATION privilege): %w", err)
	}
	
	p.replConn = conn
	return nil
}

/* StopCDC stops the CDC replication process */
func (p *PostgresCDC) StopCDC(ctx context.Context) error {
	// Close replication connection
	if p.replConn != nil {
		p.replConn.Close(ctx)
		p.replConn = nil
	}
	
	return nil
}

/* GetChanges retrieves changes from the replication stream */
func (p *PostgresCDC) GetChanges(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error) {
	lastLSN, ok := lastPosition.(string)
	if !ok {
		return nil, fmt.Errorf("lastPosition must be a string for PostgreSQL CDC")
	}

	// Try to use logical replication if replication connection is available
	if p.replConn != nil {
		changes, err := p.getChangesFromReplication(ctx, lastLSN)
		if err == nil && len(changes) > 0 {
			return changes, nil
		}
		// If replication fails, fall back to polling
	}
	
	// Fallback: use polling approach (works without replication privileges)
	return p.getChangesFromPolling(ctx, lastLSN)
}

/* getChangesFromReplication retrieves changes using logical replication */
func (p *PostgresCDC) getChangesFromReplication(ctx context.Context, lastLSN string) ([]ChangeEvent, error) {
	// Start replication stream
	// Note: This requires pgxreplication package or manual WAL message parsing
	// For now, we'll use a query-based approach that reads from replication slot
	
	// Query the replication slot to get changes
	// This is a simplified approach - full implementation would parse WAL messages directly
	query := `
		SELECT 
			pg_current_wal_lsn() as current_lsn,
			confirmed_flush_lsn
		FROM pg_replication_slots
		WHERE slot_name = $1
	`
	
	var currentLSN, confirmedLSN string
	err := p.pool.QueryRow(ctx, query, p.slotName).Scan(&currentLSN, &confirmedLSN)
	if err != nil {
		return nil, fmt.Errorf("failed to get replication slot info: %w", err)
	}
	
	// For proper logical replication, we would:
	// 1. Use pg_logical_slot_get_changes or pg_logical_slot_peek_changes
	// 2. Parse the WAL messages (INSERT, UPDATE, DELETE)
	// 3. Convert to ChangeEvent format
	
	// Simplified: use pg_logical_slot_get_changes if available
	slotQuery := `
		SELECT 
			lsn,
			data
		FROM pg_logical_slot_get_changes($1, NULL, NULL)
		WHERE lsn > $2::pg_lsn
		LIMIT 1000
	`
	
	rows, err := p.pool.Query(ctx, slotQuery, p.slotName, lastLSN)
	if err != nil {
		// If function not available or permission denied, fall back to polling
		return nil, fmt.Errorf("logical replication not available, using polling: %w", err)
	}
	defer rows.Close()
	
	var changes []ChangeEvent
	for rows.Next() {
		var lsn, walData string
		if err := rows.Scan(&lsn, &walData); err != nil {
			continue
		}
		
		// Parse WAL data (simplified - would need proper pgoutput decoder)
		// For now, return basic change event
		change := ChangeEvent{
			LSN:       lsn,
			Timestamp: time.Now(),
			Operation: "change", // Would be parsed from WAL data
		}
		changes = append(changes, change)
	}
	
	return changes, nil
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
