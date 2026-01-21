package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* SchemaEvolutionService provides schema evolution tracking functionality */
type SchemaEvolutionService struct {
	pool *pgxpool.Pool
}

/* NewSchemaEvolutionService creates a new schema evolution service */
func NewSchemaEvolutionService(pool *pgxpool.Pool) *SchemaEvolutionService {
	return &SchemaEvolutionService{pool: pool}
}

/* SchemaVersion represents a version of a schema */
type SchemaVersion struct {
	ID            uuid.UUID              `json:"id"`
	ConnectorID   uuid.UUID              `json:"connector_id"`
	SchemaName    string                 `json:"schema_name"`
	TableName     string                 `json:"table_name"`
	Version       int                    `json:"version"`
	SchemaDefinition map[string]interface{} `json:"schema_definition"`
	Changes       []SchemaChange         `json:"changes,omitempty"`
	DetectedAt    time.Time              `json:"detected_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* SchemaChange represents a change between schema versions */
type SchemaChange struct {
	ID            uuid.UUID              `json:"id"`
	ChangeType    string                 `json:"change_type"` // "column_added", "column_removed", "column_modified", "type_changed"
	ColumnName    string                 `json:"column_name,omitempty"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value,omitempty"`
	Severity      string                 `json:"severity"` // "breaking", "non_breaking", "additive"
	Impact        string                 `json:"impact,omitempty"`
	DetectedAt    time.Time              `json:"detected_at"`
}

/* TrackSchemaEvolution tracks schema changes */
func (s *SchemaEvolutionService) TrackSchemaEvolution(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName string,
	currentSchema map[string]interface{}) (*SchemaVersion, error) {

	// Get latest version
	var latestVersion int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0)
		FROM neuronip.schema_versions
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3`,
		connectorID, schemaName, tableName,
	).Scan(&latestVersion)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	// Get previous schema
	var prevSchemaJSON []byte
	var prevVersionID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id, schema_definition
		FROM neuronip.schema_versions
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3 AND version = $4`,
		connectorID, schemaName, tableName, latestVersion,
	).Scan(&prevVersionID, &prevSchemaJSON)

	var changes []SchemaChange
	if err == nil && prevSchemaJSON != nil {
		var prevSchema map[string]interface{}
		json.Unmarshal(prevSchemaJSON, &prevSchema)
		changes = s.detectChanges(prevSchema, currentSchema)
	}

	// Create new version
	version := &SchemaVersion{
		ID:              uuid.New(),
		ConnectorID:     connectorID,
		SchemaName:      schemaName,
		TableName:       tableName,
		Version:         latestVersion + 1,
		SchemaDefinition: currentSchema,
		Changes:         changes,
		DetectedAt:      time.Now(),
	}

	schemaJSON, _ := json.Marshal(currentSchema)
	metadataJSON, _ := json.Marshal(version.Metadata)

	_, err = s.pool.Exec(ctx, `
		INSERT INTO neuronip.schema_versions
		(id, connector_id, schema_name, table_name, version, schema_definition, detected_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		version.ID, version.ConnectorID, version.SchemaName, version.TableName,
		version.Version, schemaJSON, version.DetectedAt, metadataJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to track schema evolution: %w", err)
	}

	// Save changes
	for i := range changes {
		changes[i].ID = uuid.New()
		changes[i].DetectedAt = time.Now()

		oldValueJSON, _ := json.Marshal(changes[i].OldValue)
		newValueJSON, _ := json.Marshal(changes[i].NewValue)

		s.pool.Exec(ctx, `
			INSERT INTO neuronip.schema_changes
			(id, schema_version_id, change_type, column_name, old_value, new_value, severity, impact, detected_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			changes[i].ID, version.ID, changes[i].ChangeType, changes[i].ColumnName,
			oldValueJSON, newValueJSON, changes[i].Severity, changes[i].Impact, changes[i].DetectedAt,
		)
	}

	return version, nil
}

/* detectChanges detects changes between schema versions */
func (s *SchemaEvolutionService) detectChanges(prevSchema, currentSchema map[string]interface{}) []SchemaChange {
	var changes []SchemaChange

	prevColumns, _ := prevSchema["columns"].(map[string]interface{})
	currentColumns, _ := currentSchema["columns"].(map[string]interface{})

	// Detect removed columns
	if prevColumns != nil {
		for colName := range prevColumns {
			if currentColumns == nil {
				changes = append(changes, SchemaChange{
					ChangeType: "column_removed",
					ColumnName: colName,
					Severity:   "breaking",
					Impact:     "Data loss risk",
				})
			} else if _, exists := currentColumns[colName]; !exists {
				changes = append(changes, SchemaChange{
					ChangeType: "column_removed",
					ColumnName: colName,
					Severity:   "breaking",
					Impact:     "Data loss risk",
				})
			}
		}
	}

	// Detect added/modified columns
	if currentColumns != nil {
		for colName, colDef := range currentColumns {
			colMap, _ := colDef.(map[string]interface{})

			if prevColumns == nil {
				changes = append(changes, SchemaChange{
					ChangeType: "column_added",
					ColumnName: colName,
					NewValue:   colDef,
					Severity:   "additive",
					Impact:     "No breaking changes",
				})
			} else if prevColDef, exists := prevColumns[colName]; !exists {
				changes = append(changes, SchemaChange{
					ChangeType: "column_added",
					ColumnName: colName,
					NewValue:   colDef,
					Severity:   "additive",
					Impact:     "No breaking changes",
				})
			} else {
				// Check for type changes
				prevColMap, _ := prevColDef.(map[string]interface{})
				if prevColMap != nil && colMap != nil {
					prevType, _ := prevColMap["type"].(string)
					currType, _ := colMap["type"].(string)

					if prevType != currType {
						changes = append(changes, SchemaChange{
							ChangeType: "type_changed",
							ColumnName: colName,
							OldValue:   prevType,
							NewValue:   currType,
							Severity:   s.determineSeverity(prevType, currType),
							Impact:     "Potential data conversion issues",
						})
					}
				}
			}
		}
	}

	return changes
}

/* determineSeverity determines change severity */
func (s *SchemaEvolutionService) determineSeverity(oldType, newType string) string {
	// Type compatibility matrix
	compatibleTypes := map[string][]string{
		"integer": {"bigint", "numeric", "text"},
		"text":    {"varchar", "char"},
		"varchar": {"text"},
	}

	if compatible, ok := compatibleTypes[oldType]; ok {
		for _, ct := range compatible {
			if ct == newType {
				return "non_breaking"
			}
		}
	}

	return "breaking"
}

/* GetSchemaHistory retrieves schema evolution history */
func (s *SchemaEvolutionService) GetSchemaHistory(ctx context.Context,
	connectorID uuid.UUID, schemaName, tableName string) ([]SchemaVersion, error) {

	query := `
		SELECT id, connector_id, schema_name, table_name, version, 
		       schema_definition, detected_at, metadata
		FROM neuronip.schema_versions
		WHERE connector_id = $1 AND schema_name = $2 AND table_name = $3
		ORDER BY version ASC`

	rows, err := s.pool.Query(ctx, query, connectorID, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema history: %w", err)
	}
	defer rows.Close()

	var versions []SchemaVersion
	for rows.Next() {
		var version SchemaVersion
		var schemaJSON, metadataJSON []byte

		err := rows.Scan(&version.ID, &version.ConnectorID, &version.SchemaName,
			&version.TableName, &version.Version, &schemaJSON, &version.DetectedAt, &metadataJSON)
		if err != nil {
			continue
		}

		json.Unmarshal(schemaJSON, &version.SchemaDefinition)
		json.Unmarshal(metadataJSON, &version.Metadata)

		// Get changes for this version
		changeRows, _ := s.pool.Query(ctx, `
			SELECT id, change_type, column_name, old_value, new_value, severity, impact, detected_at
			FROM neuronip.schema_changes
			WHERE schema_version_id = $1
			ORDER BY detected_at ASC`, version.ID)

		var changes []SchemaChange
		for changeRows.Next() {
			var change SchemaChange
			var oldValueJSON, newValueJSON []byte

			err := changeRows.Scan(&change.ID, &change.ChangeType, &change.ColumnName,
				&oldValueJSON, &newValueJSON, &change.Severity, &change.Impact, &change.DetectedAt)
			if err != nil {
				continue
			}

			json.Unmarshal(oldValueJSON, &change.OldValue)
			json.Unmarshal(newValueJSON, &change.NewValue)
			changes = append(changes, change)
		}
		changeRows.Close()

		version.Changes = changes
		versions = append(versions, version)
	}

	return versions, nil
}
