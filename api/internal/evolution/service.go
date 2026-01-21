package evolution

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides schema evolution tracking functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new schema evolution service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* SchemaChange represents a schema change event */
type SchemaChange struct {
	ID            uuid.UUID              `json:"id"`
	ConnectorID   *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName    string                 `json:"schema_name"`
	TableName     string                 `json:"table_name"`
	ChangeType    string                 `json:"change_type"` // "table_added", "table_dropped", "column_added", "column_dropped", "column_modified", "index_added", "index_dropped"
	ColumnName    *string                `json:"column_name,omitempty"`
	OldSchema     map[string]interface{} `json:"old_schema,omitempty"`
	NewSchema     map[string]interface{} `json:"new_schema,omitempty"`
	ChangeSummary string                 `json:"change_summary"`
	DetectedAt    time.Time              `json:"detected_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

/* TrackChange tracks a schema change */
func (s *Service) TrackChange(ctx context.Context, change SchemaChange) (*SchemaChange, error) {
	change.ID = uuid.New()
	change.DetectedAt = time.Now()

	oldSchemaJSON, _ := json.Marshal(change.OldSchema)
	newSchemaJSON, _ := json.Marshal(change.NewSchema)
	metadataJSON, _ := json.Marshal(change.Metadata)

	query := `
		INSERT INTO neuronip.schema_changes
		(id, connector_id, schema_name, table_name, change_type, column_name,
		 old_schema, new_schema, change_summary, detected_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, detected_at`

	err := s.pool.QueryRow(ctx, query,
		change.ID, change.ConnectorID, change.SchemaName, change.TableName,
		change.ChangeType, change.ColumnName, oldSchemaJSON, newSchemaJSON,
		change.ChangeSummary, change.DetectedAt, metadataJSON,
	).Scan(&change.ID, &change.DetectedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to track change: %w", err)
	}

	return &change, nil
}

/* GetChange retrieves a schema change by ID */
func (s *Service) GetChange(ctx context.Context, id uuid.UUID) (*SchemaChange, error) {
	var change SchemaChange
	var connectorID sql.NullString
	var columnName sql.NullString
	var oldSchemaJSON, newSchemaJSON, metadataJSON []byte

	query := `
		SELECT id, connector_id, schema_name, table_name, change_type, column_name,
		       old_schema, new_schema, change_summary, detected_at, metadata
		FROM neuronip.schema_changes
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&change.ID, &connectorID, &change.SchemaName, &change.TableName,
		&change.ChangeType, &columnName, &oldSchemaJSON, &newSchemaJSON,
		&change.ChangeSummary, &change.DetectedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("change not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get change: %w", err)
	}

	if connectorID.Valid {
		id, _ := uuid.Parse(connectorID.String)
		change.ConnectorID = &id
	}
	if columnName.Valid {
		change.ColumnName = &columnName.String
	}
	if oldSchemaJSON != nil {
		json.Unmarshal(oldSchemaJSON, &change.OldSchema)
	}
	if newSchemaJSON != nil {
		json.Unmarshal(newSchemaJSON, &change.NewSchema)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &change.Metadata)
	}

	return &change, nil
}

/* ListChanges lists schema changes */
func (s *Service) ListChanges(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName *string, changeType *string, limit int) ([]SchemaChange, error) {
	query := `
		SELECT id, connector_id, schema_name, table_name, change_type, column_name,
		       old_schema, new_schema, change_summary, detected_at, metadata
		FROM neuronip.schema_changes
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if connectorID != nil {
		query += fmt.Sprintf(" AND connector_id = $%d", argIdx)
		args = append(args, *connectorID)
		argIdx++
	}
	if schemaName != nil {
		query += fmt.Sprintf(" AND schema_name = $%d", argIdx)
		args = append(args, *schemaName)
		argIdx++
	}
	if tableName != nil {
		query += fmt.Sprintf(" AND table_name = $%d", argIdx)
		args = append(args, *tableName)
		argIdx++
	}
	if changeType != nil {
		query += fmt.Sprintf(" AND change_type = $%d", argIdx)
		args = append(args, *changeType)
		argIdx++
	}

	query += " ORDER BY detected_at DESC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list changes: %w", err)
	}
	defer rows.Close()

	var changes []SchemaChange
	for rows.Next() {
		var change SchemaChange
		var connectorID sql.NullString
		var columnName sql.NullString
		var oldSchemaJSON, newSchemaJSON, metadataJSON []byte

		err := rows.Scan(
			&change.ID, &connectorID, &change.SchemaName, &change.TableName,
			&change.ChangeType, &columnName, &oldSchemaJSON, &newSchemaJSON,
			&change.ChangeSummary, &change.DetectedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			change.ConnectorID = &id
		}
		if columnName.Valid {
			change.ColumnName = &columnName.String
		}
		if oldSchemaJSON != nil {
			json.Unmarshal(oldSchemaJSON, &change.OldSchema)
		}
		if newSchemaJSON != nil {
			json.Unmarshal(newSchemaJSON, &change.NewSchema)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &change.Metadata)
		}

		changes = append(changes, change)
	}

	return changes, nil
}

/* GetCurrentSchema gets the current schema for a table */
func (s *Service) GetCurrentSchema(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName string) (map[string]interface{}, error) {
	// This would typically query the information_schema or use connector to get schema
	// For now, return a placeholder
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	rows, err := s.pool.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}
	defer rows.Close()

	columns := []map[string]interface{}{}
	for rows.Next() {
		var columnName, dataType, isNullable sql.NullString
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			continue
		}

		column := map[string]interface{}{
			"name":     columnName.String,
			"type":     dataType.String,
			"nullable": isNullable.String == "YES",
		}
		
		if columnDefault.Valid {
			column["default"] = columnDefault.String
		}

		columns = append(columns, column)
	}

	return map[string]interface{}{
		"schema_name": schemaName,
		"table_name":  tableName,
		"columns":     columns,
	}, nil
}
