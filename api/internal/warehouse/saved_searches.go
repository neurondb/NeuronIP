package warehouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* SavedSearchService provides saved search management */
type SavedSearchService struct {
	pool *pgxpool.Pool
}

/* NewSavedSearchService creates a new saved search service */
func NewSavedSearchService(pool *pgxpool.Pool) *SavedSearchService {
	return &SavedSearchService{pool: pool}
}

/* SavedSearch represents a saved hybrid search */
type SavedSearch struct {
	ID            uuid.UUID              `json:"id"`
	Name          string                 `json:"name"`
	Description   *string                `json:"description,omitempty"`
	Query         string                 `json:"query"` // Natural language query
	SemanticQuery *string                `json:"semantic_query,omitempty"` // Semantic similarity query
	SchemaID      *uuid.UUID             `json:"schema_id,omitempty"`
	SQLFilters    map[string]interface{} `json:"sql_filters,omitempty"`
	SemanticTable *string                `json:"semantic_table,omitempty"`
	SemanticColumn *string               `json:"semantic_column,omitempty"`
	Limit         int                    `json:"limit"`
	Threshold     float64                `json:"threshold"`
	IsPublic      bool                   `json:"is_public"`
	OwnerID       *string                `json:"owner_id,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

/* CreateSavedSearch creates a new saved search */
func (s *SavedSearchService) CreateSavedSearch(ctx context.Context, search SavedSearch) (*SavedSearch, error) {
	searchID := uuid.New()
	sqlFiltersJSON, _ := json.Marshal(search.SQLFilters)

	query := `
		INSERT INTO neuronip.saved_searches (
			id, name, description, query, semantic_query, schema_id,
			sql_filters, semantic_table, semantic_column, limit_count,
			threshold, is_public, owner_id, tags, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	var createdAt, updatedAt time.Time
	err := s.pool.QueryRow(ctx, query,
		searchID, search.Name, search.Description, search.Query, search.SemanticQuery,
		search.SchemaID, sqlFiltersJSON, search.SemanticTable, search.SemanticColumn,
		search.Limit, search.Threshold, search.IsPublic, search.OwnerID, search.Tags,
	).Scan(&createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create saved search: %w", err)
	}

	search.ID = searchID
	search.CreatedAt = createdAt
	search.UpdatedAt = updatedAt
	return &search, nil
}

/* GetSavedSearch retrieves a saved search by ID */
func (s *SavedSearchService) GetSavedSearch(ctx context.Context, searchID uuid.UUID) (*SavedSearch, error) {
	query := `
		SELECT id, name, description, query, semantic_query, schema_id,
		       sql_filters, semantic_table, semantic_column, limit_count,
		       threshold, is_public, owner_id, tags, created_at, updated_at
		FROM neuronip.saved_searches
		WHERE id = $1
	`
	var search SavedSearch
	var sqlFiltersJSON []byte
	var semanticQuery, description, ownerID *string
	var schemaID *uuid.UUID

	err := s.pool.QueryRow(ctx, query, searchID).Scan(
		&search.ID, &search.Name, &description, &search.Query, &semanticQuery,
		&schemaID, &sqlFiltersJSON, &search.SemanticTable, &search.SemanticColumn,
		&search.Limit, &search.Threshold, &search.IsPublic, &ownerID, &search.Tags,
		&search.CreatedAt, &search.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("saved search not found: %w", err)
	}

	json.Unmarshal(sqlFiltersJSON, &search.SQLFilters)
	search.SemanticQuery = semanticQuery
	search.Description = description
	search.SchemaID = schemaID
	search.OwnerID = ownerID

	return &search, nil
}

/* ListSavedSearches lists saved searches */
func (s *SavedSearchService) ListSavedSearches(ctx context.Context, userID *string, publicOnly bool) ([]SavedSearch, error) {
	var query string
	var args []interface{}

	if publicOnly {
		query = `
			SELECT id, name, description, query, semantic_query, schema_id,
			       sql_filters, semantic_table, semantic_column, limit_count,
			       threshold, is_public, owner_id, tags, created_at, updated_at
			FROM neuronip.saved_searches
			WHERE is_public = true
			ORDER BY created_at DESC
		`
		args = []interface{}{}
	} else if userID != nil {
		query = `
			SELECT id, name, description, query, semantic_query, schema_id,
			       sql_filters, semantic_table, semantic_column, limit_count,
			       threshold, is_public, owner_id, tags, created_at, updated_at
			FROM neuronip.saved_searches
			WHERE owner_id = $1 OR is_public = true
			ORDER BY created_at DESC
		`
		args = []interface{}{*userID}
	} else {
		query = `
			SELECT id, name, description, query, semantic_query, schema_id,
			       sql_filters, semantic_table, semantic_column, limit_count,
			       threshold, is_public, owner_id, tags, created_at, updated_at
			FROM neuronip.saved_searches
			ORDER BY created_at DESC
		`
		args = []interface{}{}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list saved searches: %w", err)
	}
	defer rows.Close()

	var searches []SavedSearch
	for rows.Next() {
		var search SavedSearch
		var sqlFiltersJSON []byte
		var semanticQuery, description, ownerID *string
		var schemaID *uuid.UUID

		err := rows.Scan(
			&search.ID, &search.Name, &description, &search.Query, &semanticQuery,
			&schemaID, &sqlFiltersJSON, &search.SemanticTable, &search.SemanticColumn,
			&search.Limit, &search.Threshold, &search.IsPublic, &ownerID, &search.Tags,
			&search.CreatedAt, &search.UpdatedAt,
		)
		if err != nil {
			continue
		}

		json.Unmarshal(sqlFiltersJSON, &search.SQLFilters)
		search.SemanticQuery = semanticQuery
		search.Description = description
		search.SchemaID = schemaID
		search.OwnerID = ownerID

		searches = append(searches, search)
	}

	return searches, nil
}

/* ExecuteSavedSearch executes a saved search */
func (s *SavedSearchService) ExecuteSavedSearch(ctx context.Context, searchID uuid.UUID, service *Service) (*HybridSearchResponse, error) {
	search, err := s.GetSavedSearch(ctx, searchID)
	if err != nil {
		return nil, err
	}

	req := HybridSearchRequest{
		Query:         search.Query,
		SemanticQuery: "",
		SchemaID:      search.SchemaID,
		SQLFilters:    search.SQLFilters,
		SemanticTable: "",
		SemanticColumn: "",
		Limit:         search.Limit,
		Threshold:     search.Threshold,
	}
	if search.SemanticQuery != nil {
		req.SemanticQuery = *search.SemanticQuery
	}
	if search.SemanticTable != nil {
		req.SemanticTable = *search.SemanticTable
	}
	if search.SemanticColumn != nil {
		req.SemanticColumn = *search.SemanticColumn
	}

	return service.ExecuteHybridSearch(ctx, req)
}

/* UpdateSavedSearch updates a saved search */
func (s *SavedSearchService) UpdateSavedSearch(ctx context.Context, searchID uuid.UUID, search SavedSearch) error {
	sqlFiltersJSON, _ := json.Marshal(search.SQLFilters)

	query := `
		UPDATE neuronip.saved_searches
		SET name = $1, description = $2, query = $3, semantic_query = $4,
		    schema_id = $5, sql_filters = $6, semantic_table = $7,
		    semantic_column = $8, limit_count = $9, threshold = $10,
		    is_public = $11, tags = $12, updated_at = NOW()
		WHERE id = $13
	`
	_, err := s.pool.Exec(ctx, query,
		search.Name, search.Description, search.Query, search.SemanticQuery,
		search.SchemaID, sqlFiltersJSON, search.SemanticTable, search.SemanticColumn,
		search.Limit, search.Threshold, search.IsPublic, search.Tags, searchID,
	)
	return err
}

/* DeleteSavedSearch deletes a saved search */
func (s *SavedSearchService) DeleteSavedSearch(ctx context.Context, searchID uuid.UUID) error {
	query := `DELETE FROM neuronip.saved_searches WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, searchID)
	return err
}
