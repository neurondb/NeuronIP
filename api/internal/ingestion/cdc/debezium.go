package cdc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* DebeziumCDC provides Debezium-based universal CDC */
type DebeziumCDC struct {
	pool *pgxpool.Pool
}

/* NewDebeziumCDC creates a new Debezium CDC instance */
func NewDebeziumCDC(pool *pgxpool.Pool) *DebeziumCDC {
	return &DebeziumCDC{pool: pool}
}

/* StartCDC starts CDC using Debezium */
func (d *DebeziumCDC) StartCDC(ctx context.Context, config map[string]interface{}) error {
	// In production, this would:
	// 1. Configure Debezium connector
	// 2. Start Kafka Connect with Debezium connector
	// 3. Monitor connector status
	// 4. Process change events

	connectorName, _ := config["connector_name"].(string)
	if connectorName == "" {
		connectorName = fmt.Sprintf("neuronip-cdc-%d", time.Now().Unix())
	}
	sourceType, _ := config["source_type"].(string)
	if sourceType == "" {
		sourceType = "postgresql"
	}

	// Store connector configuration
	configJSON, _ := json.Marshal(config)
	query := `
		INSERT INTO neuronip.cdc_connectors (connector_name, connector_type, source_type, config, status, created_at, updated_at)
		VALUES ($1, 'debezium', $2, $3, 'active', NOW(), NOW())
		ON CONFLICT (connector_name) DO UPDATE SET
			source_type = EXCLUDED.source_type,
			config = EXCLUDED.config,
			status = EXCLUDED.status,
			updated_at = NOW()`

	_, err := d.pool.Exec(ctx, query, connectorName, sourceType, configJSON)
	return err
}

/* StopCDC stops CDC */
func (d *DebeziumCDC) StopCDC(ctx context.Context, connectorName string) error {
	query := `
		UPDATE neuronip.cdc_connectors
		SET status = 'stopped', updated_at = NOW()
		WHERE connector_name = $1`

	_, err := d.pool.Exec(ctx, query, connectorName)
	return err
}

/* GetChangeEvents retrieves change events from neuronip.cdc_events for the connector. Events are populated by a Kafka consumer or DB triggers when wired. */
func (d *DebeziumCDC) GetChangeEvents(ctx context.Context, connectorName string, limit int) ([]ChangeEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT table_name, operation, lsn, created_at, old_data, new_data
		FROM neuronip.cdc_events
		WHERE connector_name = $1
		ORDER BY created_at DESC
		LIMIT $2`
	rows, err := d.pool.Query(ctx, query, connectorName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ChangeEvent
	for rows.Next() {
		var ev ChangeEvent
		var lsn *string
		var oldData, newData []byte
		if err := rows.Scan(&ev.Table, &ev.Operation, &lsn, &ev.Timestamp, &oldData, &newData); err != nil {
			continue
		}
		if lsn != nil {
			ev.LSN = *lsn
		}
		if len(oldData) > 0 {
			_ = json.Unmarshal(oldData, &ev.OldData)
		}
		if len(newData) > 0 {
			_ = json.Unmarshal(newData, &ev.NewData)
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}
