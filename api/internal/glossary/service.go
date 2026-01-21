package glossary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Service provides business glossary functionality */
type Service struct {
	pool *pgxpool.Pool
}

/* NewService creates a new glossary service */
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

/* Term represents a business glossary term */
type Term struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Definition      string                 `json:"definition"`
	Category        *string                `json:"category,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	RelatedTerms    []uuid.UUID            `json:"related_terms,omitempty"`
	OwnedBy         *string                `json:"owned_by,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

/* DictionaryEntry represents a data dictionary entry */
type DictionaryEntry struct {
	ID              uuid.UUID              `json:"id"`
	ConnectorID     *uuid.UUID             `json:"connector_id,omitempty"`
	SchemaName      string                 `json:"schema_name"`
	TableName       string                 `json:"table_name"`
	ColumnName      *string                `json:"column_name,omitempty"`
	BusinessName    *string                `json:"business_name,omitempty"`
	Description     *string                `json:"description,omitempty"`
	DataType        *string                `json:"data_type,omitempty"`
	BusinessRules   []string               `json:"business_rules,omitempty"`
	Examples        []string               `json:"examples,omitempty"`
	TermID          *uuid.UUID             `json:"term_id,omitempty"` // Link to glossary term
	OwnedBy         *string                `json:"owned_by,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

/* CreateTerm creates a new glossary term */
func (s *Service) CreateTerm(ctx context.Context, term Term) (*Term, error) {
	term.ID = uuid.New()
	term.CreatedAt = time.Now()
	term.UpdatedAt = time.Now()

	tagsJSON, _ := json.Marshal(term.Tags)
	relatedTermsJSON, _ := json.Marshal(term.RelatedTerms)
	metadataJSON, _ := json.Marshal(term.Metadata)

	query := `
		INSERT INTO neuronip.glossary_terms
		(id, name, definition, category, tags, related_terms, owned_by, created_by,
		 created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		term.ID, term.Name, term.Definition, term.Category, tagsJSON,
		relatedTermsJSON, term.OwnedBy, term.CreatedBy,
		term.CreatedAt, term.UpdatedAt, metadataJSON,
	).Scan(&term.ID, &term.CreatedAt, &term.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create term: %w", err)
	}

	return &term, nil
}

/* GetTerm retrieves a glossary term by ID */
func (s *Service) GetTerm(ctx context.Context, id uuid.UUID) (*Term, error) {
	var term Term
	var category, ownedBy sql.NullString
	var tagsJSON, relatedTermsJSON, metadataJSON []byte

	query := `
		SELECT id, name, definition, category, tags, related_terms, owned_by,
		       created_by, created_at, updated_at, metadata
		FROM neuronip.glossary_terms
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&term.ID, &term.Name, &term.Definition, &category,
		&tagsJSON, &relatedTermsJSON, &ownedBy,
		&term.CreatedBy, &term.CreatedAt, &term.UpdatedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("term not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get term: %w", err)
	}

	if category.Valid {
		term.Category = &category.String
	}
	if ownedBy.Valid {
		term.OwnedBy = &ownedBy.String
	}
	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &term.Tags)
	}
	if relatedTermsJSON != nil {
		json.Unmarshal(relatedTermsJSON, &term.RelatedTerms)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &term.Metadata)
	}

	return &term, nil
}

/* ListTerms lists glossary terms */
func (s *Service) ListTerms(ctx context.Context, category *string, search *string, limit int) ([]Term, error) {
	query := `
		SELECT id, name, definition, category, tags, related_terms, owned_by,
		       created_by, created_at, updated_at, metadata
		FROM neuronip.glossary_terms
		WHERE 1=1`
	
	args := []interface{}{}
	argIdx := 1

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *category)
		argIdx++
	}
	if search != nil {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR definition ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*search+"%")
		argIdx++
	}

	query += " ORDER BY name ASC"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list terms: %w", err)
	}
	defer rows.Close()

	var terms []Term
	for rows.Next() {
		var term Term
		var category, ownedBy sql.NullString
		var tagsJSON, relatedTermsJSON, metadataJSON []byte

		err := rows.Scan(
			&term.ID, &term.Name, &term.Definition, &category,
			&tagsJSON, &relatedTermsJSON, &ownedBy,
			&term.CreatedBy, &term.CreatedAt, &term.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if category.Valid {
			term.Category = &category.String
		}
		if ownedBy.Valid {
			term.OwnedBy = &ownedBy.String
		}
		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &term.Tags)
		}
		if relatedTermsJSON != nil {
			json.Unmarshal(relatedTermsJSON, &term.RelatedTerms)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &term.Metadata)
		}

		terms = append(terms, term)
	}

	return terms, nil
}

/* CreateDictionaryEntry creates a new data dictionary entry */
func (s *Service) CreateDictionaryEntry(ctx context.Context, entry DictionaryEntry) (*DictionaryEntry, error) {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	businessRulesJSON, _ := json.Marshal(entry.BusinessRules)
	examplesJSON, _ := json.Marshal(entry.Examples)
	metadataJSON, _ := json.Marshal(entry.Metadata)

	query := `
		INSERT INTO neuronip.data_dictionary
		(id, connector_id, schema_name, table_name, column_name, business_name,
		 description, data_type, business_rules, examples, term_id, owned_by,
		 created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		entry.ID, entry.ConnectorID, entry.SchemaName, entry.TableName,
		entry.ColumnName, entry.BusinessName, entry.Description, entry.DataType,
		businessRulesJSON, examplesJSON, entry.TermID, entry.OwnedBy,
		entry.CreatedBy, entry.CreatedAt, entry.UpdatedAt, metadataJSON,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create dictionary entry: %w", err)
	}

	return &entry, nil
}

/* GetDictionaryEntry retrieves a dictionary entry by ID */
func (s *Service) GetDictionaryEntry(ctx context.Context, id uuid.UUID) (*DictionaryEntry, error) {
	var entry DictionaryEntry
	var connectorID sql.NullString
	var columnName, businessName, description, dataType, ownedBy sql.NullString
	var termID sql.NullString
	var businessRulesJSON, examplesJSON, metadataJSON []byte

	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, business_name,
		       description, data_type, business_rules, examples, term_id, owned_by,
		       created_by, created_at, updated_at, metadata
		FROM neuronip.data_dictionary
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&entry.ID, &connectorID, &entry.SchemaName, &entry.TableName,
		&columnName, &businessName, &description, &dataType,
		&businessRulesJSON, &examplesJSON, &termID, &ownedBy,
		&entry.CreatedBy, &entry.CreatedAt, &entry.UpdatedAt, &metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dictionary entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get dictionary entry: %w", err)
	}

	if connectorID.Valid {
		id, _ := uuid.Parse(connectorID.String)
		entry.ConnectorID = &id
	}
	if columnName.Valid {
		entry.ColumnName = &columnName.String
	}
	if businessName.Valid {
		entry.BusinessName = &businessName.String
	}
	if description.Valid {
		entry.Description = &description.String
	}
	if dataType.Valid {
		entry.DataType = &dataType.String
	}
	if ownedBy.Valid {
		entry.OwnedBy = &ownedBy.String
	}
	if termID.Valid {
		id, _ := uuid.Parse(termID.String)
		entry.TermID = &id
	}
	if businessRulesJSON != nil {
		json.Unmarshal(businessRulesJSON, &entry.BusinessRules)
	}
	if examplesJSON != nil {
		json.Unmarshal(examplesJSON, &entry.Examples)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &entry.Metadata)
	}

	return &entry, nil
}

/* ListDictionaryEntries lists dictionary entries */
func (s *Service) ListDictionaryEntries(ctx context.Context, connectorID *uuid.UUID, schemaName, tableName *string, limit int) ([]DictionaryEntry, error) {
	query := `
		SELECT id, connector_id, schema_name, table_name, column_name, business_name,
		       description, data_type, business_rules, examples, term_id, owned_by,
		       created_by, created_at, updated_at, metadata
		FROM neuronip.data_dictionary
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

	query += " ORDER BY schema_name, table_name, column_name"
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list dictionary entries: %w", err)
	}
	defer rows.Close()

	var entries []DictionaryEntry
	for rows.Next() {
		var entry DictionaryEntry
		var connectorID sql.NullString
		var columnName, businessName, description, dataType, ownedBy sql.NullString
		var termID sql.NullString
		var businessRulesJSON, examplesJSON, metadataJSON []byte

		err := rows.Scan(
			&entry.ID, &connectorID, &entry.SchemaName, &entry.TableName,
			&columnName, &businessName, &description, &dataType,
			&businessRulesJSON, &examplesJSON, &termID, &ownedBy,
			&entry.CreatedBy, &entry.CreatedAt, &entry.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			continue
		}

		if connectorID.Valid {
			id, _ := uuid.Parse(connectorID.String)
			entry.ConnectorID = &id
		}
		if columnName.Valid {
			entry.ColumnName = &columnName.String
		}
		if businessName.Valid {
			entry.BusinessName = &businessName.String
		}
		if description.Valid {
			entry.Description = &description.String
		}
		if dataType.Valid {
			entry.DataType = &dataType.String
		}
		if ownedBy.Valid {
			entry.OwnedBy = &ownedBy.String
		}
		if termID.Valid {
			id, _ := uuid.Parse(termID.String)
			entry.TermID = &id
		}
		if businessRulesJSON != nil {
			json.Unmarshal(businessRulesJSON, &entry.BusinessRules)
		}
		if examplesJSON != nil {
			json.Unmarshal(examplesJSON, &entry.Examples)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &entry.Metadata)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
