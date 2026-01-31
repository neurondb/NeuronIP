package databases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* Service handles database operations */
type Service struct {
	db *pgxpool.Pool
}

/* NewService creates a new databases service */
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

/* Database represents a database */
type Database struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	WorkspaceID *uuid.UUID             `json:"workspace_id,omitempty"`
	CreatedBy   *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* DatabaseColumn represents a database column */
type DatabaseColumn struct {
	ID         uuid.UUID   `json:"id"`
	DatabaseID uuid.UUID   `json:"database_id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"` // text, number, date, select, multiSelect, person, file, checkbox
	Options    []string    `json:"options,omitempty"`
	Order      int         `json:"order"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

/* DatabaseRow represents a database row */
type DatabaseRow struct {
	ID         uuid.UUID              `json:"id"`
	DatabaseID uuid.UUID              `json:"database_id"`
	Data       map[string]interface{} `json:"data"`
	CreatedBy  *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

/* ViewPreferences represents user view preferences */
type ViewPreferences struct {
	ViewType string                 `json:"view_type"` // table, kanban, calendar, gallery, list
	Filters  []map[string]interface{} `json:"filters,omitempty"`
	Sort     map[string]interface{}   `json:"sort,omitempty"`
}

/* CreateDatabaseRequest represents a request to create a database */
type CreateDatabaseRequest struct {
	Name        string                 `json:"name"`
	Description *string                 `json:"description,omitempty"`
	WorkspaceID *uuid.UUID             `json:"workspace_id,omitempty"`
	Columns     []CreateColumnRequest   `json:"columns"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

/* CreateColumnRequest represents a request to create a column */
type CreateColumnRequest struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Options []string `json:"options,omitempty"`
	Order   *int     `json:"order,omitempty"`
}

/* UpdateRowRequest represents a request to update a row */
type UpdateRowRequest struct {
	Data map[string]interface{} `json:"data"`
}

/* GetDatabase retrieves a database with columns and rows */
func (s *Service) GetDatabase(ctx context.Context, databaseID uuid.UUID) (*Database, []DatabaseColumn, []DatabaseRow, error) {
	// Get database
	dbQuery := `
		SELECT id, name, description, workspace_id, created_by, created_at, updated_at, metadata
		FROM neuronip.databases
		WHERE id = $1 AND deleted_at IS NULL
	`

	var db Database
	var description *string
	var metadataJSON []byte

	err := s.db.QueryRow(ctx, dbQuery, databaseID).Scan(
		&db.ID, &db.Name, &description, &db.WorkspaceID, &db.CreatedBy,
		&db.CreatedAt, &db.UpdatedAt, &metadataJSON,
	)
	if err == pgx.ErrNoRows {
		return nil, nil, nil, errors.NotFound("Database not found")
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get database: %w", err)
	}

	db.Description = description
	if err := json.Unmarshal(metadataJSON, &db.Metadata); err != nil {
		db.Metadata = make(map[string]interface{})
	}

	// Get columns
	columnsQuery := `
		SELECT id, database_id, name, type, options, order_index, created_at, updated_at
		FROM neuronip.database_columns
		WHERE database_id = $1
		ORDER BY order_index ASC
	`

	rows, err := s.db.Query(ctx, columnsQuery, databaseID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []DatabaseColumn
	for rows.Next() {
		var col DatabaseColumn
		var optionsJSON []byte

		err := rows.Scan(
			&col.ID, &col.DatabaseID, &col.Name, &col.Type,
			&optionsJSON, &col.Order, &col.CreatedAt, &col.UpdatedAt,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to scan column: %w", err)
		}

		if err := json.Unmarshal(optionsJSON, &col.Options); err != nil {
			col.Options = []string{}
		}

		columns = append(columns, col)
	}

	// Get rows
	rowsQuery := `
		SELECT id, database_id, data, created_by, created_at, updated_at
		FROM neuronip.database_rows
		WHERE database_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rowRows, err := s.db.Query(ctx, rowsQuery, databaseID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get rows: %w", err)
	}
	defer rowRows.Close()

	var dbRows []DatabaseRow
	for rowRows.Next() {
		var row DatabaseRow
		var dataJSON []byte
		var createdBy *uuid.UUID

		err := rowRows.Scan(
			&row.ID, &row.DatabaseID, &dataJSON, &createdBy,
			&row.CreatedAt, &row.UpdatedAt,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal(dataJSON, &row.Data); err != nil {
			row.Data = make(map[string]interface{})
		}

		row.CreatedBy = createdBy
		dbRows = append(dbRows, row)
	}

	return &db, columns, dbRows, nil
}

/* CreateDatabase creates a new database */
func (s *Service) CreateDatabase(ctx context.Context, req CreateDatabaseRequest, userID *uuid.UUID) (*Database, []DatabaseColumn, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create database
	databaseID := uuid.New()
	metadataJSON, _ := json.Marshal(req.Metadata)

	dbQuery := `
		INSERT INTO neuronip.databases (id, name, description, workspace_id, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, workspace_id, created_by, created_at, updated_at, metadata
	`

	var db Database
	var description *string
	var metadataJSONOut []byte

	err = tx.QueryRow(ctx, dbQuery, databaseID, req.Name, req.Description, req.WorkspaceID, userID, metadataJSON).Scan(
		&db.ID, &db.Name, &description, &db.WorkspaceID, &db.CreatedBy,
		&db.CreatedAt, &db.UpdatedAt, &metadataJSONOut,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database: %w", err)
	}

	db.Description = description
	if err := json.Unmarshal(metadataJSONOut, &db.Metadata); err != nil {
		db.Metadata = req.Metadata
	}

	// Create columns
	var columns []DatabaseColumn
	for i, colReq := range req.Columns {
		order := i
		if colReq.Order != nil {
			order = *colReq.Order
		}

		columnID := uuid.New()
		optionsJSON, _ := json.Marshal(colReq.Options)

		colQuery := `
			INSERT INTO neuronip.database_columns (id, database_id, name, type, options, order_index)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, database_id, name, type, options, order_index, created_at, updated_at
		`

		var col DatabaseColumn
		var optionsJSONOut []byte

		err := tx.QueryRow(ctx, colQuery, columnID, databaseID, colReq.Name, colReq.Type, optionsJSON, order).Scan(
			&col.ID, &col.DatabaseID, &col.Name, &col.Type,
			&optionsJSONOut, &col.Order, &col.CreatedAt, &col.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create column: %w", err)
		}

		if err := json.Unmarshal(optionsJSONOut, &col.Options); err != nil {
			col.Options = colReq.Options
		}

		columns = append(columns, col)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &db, columns, nil
}

/* UpdateRow updates a database row */
func (s *Service) UpdateRow(ctx context.Context, databaseID, rowID uuid.UUID, req UpdateRowRequest) (*DatabaseRow, error) {
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.BadRequest("Invalid data format")
	}

	query := `
		UPDATE neuronip.database_rows
		SET data = $1, updated_at = NOW()
		WHERE id = $2 AND database_id = $3 AND deleted_at IS NULL
		RETURNING id, database_id, data, created_by, created_at, updated_at
	`

	var row DatabaseRow
	var dataJSONOut []byte
	var createdBy *uuid.UUID

	err = s.db.QueryRow(ctx, query, dataJSON, rowID, databaseID).Scan(
		&row.ID, &row.DatabaseID, &dataJSONOut, &createdBy,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.NotFound("Row not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update row: %w", err)
	}

	if err := json.Unmarshal(dataJSONOut, &row.Data); err != nil {
		row.Data = req.Data
	}

	row.CreatedBy = createdBy

	return &row, nil
}

/* CreateRow creates a new database row */
func (s *Service) CreateRow(ctx context.Context, databaseID uuid.UUID, req UpdateRowRequest, userID *uuid.UUID) (*DatabaseRow, error) {
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.BadRequest("Invalid data format")
	}

	rowID := uuid.New()
	query := `
		INSERT INTO neuronip.database_rows (id, database_id, data, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, database_id, data, created_by, created_at, updated_at
	`

	var row DatabaseRow
	var dataJSONOut []byte
	var createdBy *uuid.UUID

	err = s.db.QueryRow(ctx, query, rowID, databaseID, dataJSON, userID).Scan(
		&row.ID, &row.DatabaseID, &dataJSONOut, &createdBy,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create row: %w", err)
	}

	if err := json.Unmarshal(dataJSONOut, &row.Data); err != nil {
		row.Data = req.Data
	}

	row.CreatedBy = createdBy

	return &row, nil
}

/* DeleteRow soft deletes a database row */
func (s *Service) DeleteRow(ctx context.Context, databaseID, rowID uuid.UUID) error {
	query := `UPDATE neuronip.database_rows SET deleted_at = NOW() WHERE id = $1 AND database_id = $2 AND deleted_at IS NULL`
	result, err := s.db.Exec(ctx, query, rowID, databaseID)
	if err != nil {
		return fmt.Errorf("failed to delete row: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFound("Row not found")
	}

	return nil
}

/* UpdateViewPreferences updates user view preferences */
func (s *Service) UpdateViewPreferences(ctx context.Context, databaseID uuid.UUID, userID uuid.UUID, prefs ViewPreferences) error {
	filtersJSON, _ := json.Marshal(prefs.Filters)
	sortJSON, _ := json.Marshal(prefs.Sort)

	query := `
		INSERT INTO neuronip.database_view_preferences (database_id, user_id, view_type, filters, sort)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (database_id, user_id)
		DO UPDATE SET view_type = $3, filters = $4, sort = $5, updated_at = NOW()
	`

	_, err := s.db.Exec(ctx, query, databaseID, userID, prefs.ViewType, filtersJSON, sortJSON)
	if err != nil {
		return fmt.Errorf("failed to update view preferences: %w", err)
	}

	return nil
}

/* GetViewPreferences retrieves user view preferences */
func (s *Service) GetViewPreferences(ctx context.Context, databaseID uuid.UUID, userID uuid.UUID) (*ViewPreferences, error) {
	query := `
		SELECT view_type, filters, sort
		FROM neuronip.database_view_preferences
		WHERE database_id = $1 AND user_id = $2
	`

	var prefs ViewPreferences
	var filtersJSON, sortJSON []byte

	err := s.db.QueryRow(ctx, query, databaseID, userID).Scan(
		&prefs.ViewType, &filtersJSON, &sortJSON,
	)
	if err == pgx.ErrNoRows {
		// Return default preferences
		return &ViewPreferences{
			ViewType: "table",
			Filters:  []map[string]interface{}{},
			Sort:     map[string]interface{}{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get view preferences: %w", err)
	}

	if err := json.Unmarshal(filtersJSON, &prefs.Filters); err != nil {
		prefs.Filters = []map[string]interface{}{}
	}
	if err := json.Unmarshal(sortJSON, &prefs.Sort); err != nil {
		prefs.Sort = map[string]interface{}{}
	}

	return &prefs, nil
}
