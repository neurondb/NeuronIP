package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* SchemaEvolutionService provides schema evolution functionality */
type SchemaEvolutionService struct {
	pool *pgxpool.Pool
}

/* NewSchemaEvolutionService creates a new schema evolution service */
func NewSchemaEvolutionService(pool *pgxpool.Pool) *SchemaEvolutionService {
	return &SchemaEvolutionService{pool: pool}
}

/* DetectSchemaChanges detects schema changes */
func (ses *SchemaEvolutionService) DetectSchemaChanges(ctx context.Context, connectorID string, currentSchema Schema) (*SchemaChanges, error) {
	// Get previous schema version
	previousSchema, err := ses.GetSchemaVersion(ctx, connectorID)
	if err != nil {
		// No previous schema, this is initial discovery
		return &SchemaChanges{
			ConnectorID: connectorID,
			ChangeType:  "initial",
			Changes:     []SchemaChange{},
		}, nil
	}

	changes := &SchemaChanges{
		ConnectorID: connectorID,
		ChangeType:  "evolution",
		Changes:     []SchemaChange{},
		DetectedAt:  time.Now(),
	}

	// Compare schemas and detect changes
	// 1. Detect new tables
	currentTables := make(map[string]TableSchema)
	for _, table := range currentSchema.Tables {
		currentTables[table.Name] = table
	}

	previousTables := make(map[string]TableSchema)
	for _, table := range previousSchema.Tables {
		previousTables[table.Name] = table
	}

	// Find new tables
	for name, table := range currentTables {
		if _, exists := previousTables[name]; !exists {
			changes.Changes = append(changes.Changes, SchemaChange{
				Type:      "table_added",
				TableName: name,
				Details:   map[string]interface{}{"table": table},
			})
		}
	}

	// Find removed tables
	for name := range previousTables {
		if _, exists := currentTables[name]; !exists {
			changes.Changes = append(changes.Changes, SchemaChange{
				Type:      "table_removed",
				TableName: name,
			})
		}
	}

	// Find column changes
	for name, currentTable := range currentTables {
		if previousTable, exists := previousTables[name]; exists {
			columnChanges := ses.detectColumnChanges(currentTable, previousTable)
			changes.Changes = append(changes.Changes, columnChanges...)
		}
	}

	return changes, nil
}

/* detectColumnChanges detects column-level changes */
func (ses *SchemaEvolutionService) detectColumnChanges(current, previous TableSchema) []SchemaChange {
	var changes []SchemaChange

	currentColumns := make(map[string]ColumnSchema)
	for _, col := range current.Columns {
		currentColumns[col.Name] = col
	}

	previousColumns := make(map[string]ColumnSchema)
	for _, col := range previous.Columns {
		previousColumns[col.Name] = col
	}

	// Find new columns
	for name, col := range currentColumns {
		if _, exists := previousColumns[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:      "column_added",
				TableName: current.Name,
				ColumnName: &name,
				Details:   map[string]interface{}{"column": col},
			})
		}
	}

	// Find removed columns
	for name := range previousColumns {
		if _, exists := currentColumns[name]; !exists {
			changes = append(changes, SchemaChange{
				Type:      "column_removed",
				TableName: current.Name,
				ColumnName: &name,
			})
		}
	}

	// Find type changes
	for name, currentCol := range currentColumns {
		if previousCol, exists := previousColumns[name]; exists {
			if currentCol.DataType != previousCol.DataType {
				changes = append(changes, SchemaChange{
					Type:      "column_type_changed",
					TableName: current.Name,
					ColumnName: &name,
					Details: map[string]interface{}{
						"old_type": previousCol.DataType,
						"new_type": currentCol.DataType,
					},
				})
			}
		}
	}

	return changes
}

/* GetSchemaVersion retrieves a schema version */
func (ses *SchemaEvolutionService) GetSchemaVersion(ctx context.Context, connectorID string) (*Schema, error) {
	query := `
		SELECT schema_data
		FROM neuronip.schema_versions
		WHERE connector_id = $1
		ORDER BY version DESC
		LIMIT 1`

	var schemaJSON []byte
	err := ses.pool.QueryRow(ctx, query, connectorID).Scan(&schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("schema version not found: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

/* SchemaChanges represents detected schema changes */
type SchemaChanges struct {
	ConnectorID string        `json:"connector_id"`
	ChangeType  string        `json:"change_type"` // "initial", "evolution"
	Changes     []SchemaChange `json:"changes"`
	DetectedAt  time.Time     `json:"detected_at"`
}

/* SchemaChange represents a single schema change */
type SchemaChange struct {
	Type       string                 `json:"type"` // "table_added", "table_removed", "column_added", "column_removed", "column_type_changed"
	TableName  string                 `json:"table_name"`
	ColumnName *string                `json:"column_name,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}
