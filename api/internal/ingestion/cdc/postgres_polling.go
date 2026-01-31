package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* PostgresPollingCDC implements CDC for PostgreSQL by polling neuronip.stream_events.
 * No replication connection or build tags required. Use for Postgres-native streaming.
 */
type PostgresPollingCDC struct {
	pool     *pgxpool.Pool
	connKey  string
	stopDone chan struct{}
}

/* NewPostgresPollingCDC creates a new polling-based Postgres CDC (no replication privilege needed) */
func NewPostgresPollingCDC(pool *pgxpool.Pool, connectorKey string) *PostgresPollingCDC {
	if connectorKey == "" {
		connectorKey = "postgres_polling_default"
	}
	return &PostgresPollingCDC{pool: pool, connKey: connectorKey}
}

/* StartCDC ensures stream_events table exists and starts polling readiness */
func (p *PostgresPollingCDC) StartCDC(ctx context.Context, config map[string]interface{}) error {
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS neuronip.stream_events (
			id BIGSERIAL PRIMARY KEY,
			event_type TEXT NOT NULL,
			source_table TEXT,
			payload JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("ensure stream_events table: %w", err)
	}
	if key, ok := config["connector_key"].(string); ok && key != "" {
		p.connKey = key
	}
	return nil
}

/* StopCDC stops the CDC process */
func (p *PostgresPollingCDC) StopCDC(ctx context.Context) error {
	if p.stopDone != nil {
		close(p.stopDone)
		p.stopDone = nil
	}
	return nil
}

/* GetChanges returns change events from stream_events after lastPosition (as last event id string) */
func (p *PostgresPollingCDC) GetChanges(ctx context.Context, lastPosition interface{}) ([]ChangeEvent, error) {
	lastID := int64(0)
	switch v := lastPosition.(type) {
	case string:
		if v != "" {
			lastID, _ = strconv.ParseInt(v, 10, 64)
		}
	case int64:
		lastID = v
	case float64:
		lastID = int64(v)
	}

	query := `SELECT id, event_type, source_table, payload, created_at
			  FROM neuronip.stream_events WHERE id > $1 ORDER BY id LIMIT 500`
	rows, err := p.pool.Query(ctx, query, lastID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ChangeEvent
	for rows.Next() {
		var id int64
		var eventType, sourceTable string
		var payload json.RawMessage
		var createdAt time.Time
		if err := rows.Scan(&id, &eventType, &sourceTable, &payload, &createdAt); err != nil {
			continue
		}
		table := sourceTable
		if table == "" {
			table = "stream_events"
		}
		var newData map[string]interface{}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &newData)
		}
		if newData == nil {
			newData = make(map[string]interface{})
		}
		newData["_event_id"] = id
		newData["_event_type"] = eventType
		newData["_created_at"] = createdAt
		changes = append(changes, ChangeEvent{
			Table:     table,
			Operation: "insert",
			LSN:       strconv.FormatInt(id, 10),
			Timestamp: createdAt,
			NewData:   newData,
		})
	}
	return changes, nil
}

/* SaveCheckpoint stores the last processed position in cdc_polling_state */
func (p *PostgresPollingCDC) SaveCheckpoint(ctx context.Context, dataSourceID string, tableName string, checkpoint map[string]interface{}) error {
	lastID := int64(0)
	if lsn, ok := checkpoint["lsn"].(string); ok && lsn != "" {
		lastID, _ = strconv.ParseInt(lsn, 10, 64)
	}
	if n, ok := checkpoint["lsn"].(float64); ok {
		lastID = int64(n)
	}
	key := p.connKey + ":" + dataSourceID + ":" + tableName
	query := `
		INSERT INTO neuronip.cdc_polling_state (connector_key, last_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (connector_key) DO UPDATE SET last_id = $2, updated_at = NOW()`
	_, err := p.pool.Exec(ctx, query, key, lastID)
	return err
}

/* GetCheckpoint returns the last checkpoint from cdc_polling_state */
func (p *PostgresPollingCDC) GetCheckpoint(ctx context.Context, dataSourceID string, tableName string) (map[string]interface{}, error) {
	key := p.connKey + ":" + dataSourceID + ":" + tableName
	var lastID int64
	query := `SELECT last_id FROM neuronip.cdc_polling_state WHERE connector_key = $1`
	err := p.pool.QueryRow(ctx, query, key).Scan(&lastID)
	if err != nil {
		return map[string]interface{}{"lsn": "0"}, nil
	}
	return map[string]interface{}{"lsn": strconv.FormatInt(lastID, 10)}, nil
}
