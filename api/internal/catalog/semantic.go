package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* SemanticService provides semantic definitions functionality */
type SemanticService struct {
	pool *pgxpool.Pool
}

/* NewSemanticService creates a new semantic service */
func NewSemanticService(pool *pgxpool.Pool) *SemanticService {
	return &SemanticService{pool: pool}
}

/* SemanticDefinition represents a semantic definition */
type SemanticDefinition struct {
	ID              uuid.UUID              `json:"id"`
	Term            string                 `json:"term"`
	Definition      string                 `json:"definition"`
	Category        *string                `json:"category,omitempty"`
	SQLExpression   *string                `json:"sql_expression,omitempty"`
	AIModelMapping  map[string]interface{} `json:"ai_model_mapping,omitempty"`
	Synonyms        []string               `json:"synonyms,omitempty"`
	RelatedTerms    []string               `json:"related_terms,omitempty"`
	Examples        []interface{}          `json:"examples,omitempty"`
	OwnerID         *string                `json:"owner_id,omitempty"`
	Version         string                 `json:"version"`
	Status          string                 `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* CreateSemanticDefinition creates a new semantic definition */
func (s *SemanticService) CreateSemanticDefinition(ctx context.Context, def SemanticDefinition) (*SemanticDefinition, error) {
	def.ID = uuid.New()
	def.CreatedAt = time.Now()
	def.UpdatedAt = time.Now()
	
	if def.Version == "" {
		def.Version = "1.0.0"
	}
	if def.Status == "" {
		def.Status = "draft"
	}
	
	synonymsJSON, _ := json.Marshal(def.Synonyms)
	relatedTermsJSON, _ := json.Marshal(def.RelatedTerms)
	examplesJSON, _ := json.Marshal(def.Examples)
	aiMappingJSON, _ := json.Marshal(def.AIModelMapping)
	
	query := `
		INSERT INTO neuronip.semantic_definitions 
		(id, term, definition, category, sql_expression, ai_model_mapping, synonyms, 
		 related_terms, examples, owner_id, version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (term, version) DO UPDATE SET
			definition = EXCLUDED.definition,
			category = EXCLUDED.category,
			sql_expression = EXCLUDED.sql_expression,
			ai_model_mapping = EXCLUDED.ai_model_mapping,
			synonyms = EXCLUDED.synonyms,
			related_terms = EXCLUDED.related_terms,
			examples = EXCLUDED.examples,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`
	
	err := s.pool.QueryRow(ctx, query,
		def.ID, def.Term, def.Definition, def.Category, def.SQLExpression, aiMappingJSON,
		synonymsJSON, relatedTermsJSON, examplesJSON, def.OwnerID, def.Version, def.Status,
		def.CreatedAt, def.UpdatedAt,
	).Scan(&def.ID, &def.CreatedAt, &def.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create semantic definition: %w", err)
	}
	
	return &def, nil
}

/* GetSemanticDefinition retrieves a semantic definition */
func (s *SemanticService) GetSemanticDefinition(ctx context.Context, id uuid.UUID) (*SemanticDefinition, error) {
	query := `
		SELECT id, term, definition, category, sql_expression, ai_model_mapping, synonyms,
		       related_terms, examples, owner_id, version, status, created_at, updated_at
		FROM neuronip.semantic_definitions
		WHERE id = $1`
	
	var def SemanticDefinition
	var synonymsJSON, relatedTermsJSON, examplesJSON, aiMappingJSON json.RawMessage
	
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&def.ID, &def.Term, &def.Definition, &def.Category, &def.SQLExpression, &aiMappingJSON,
		&synonymsJSON, &relatedTermsJSON, &examplesJSON, &def.OwnerID, &def.Version, &def.Status,
		&def.CreatedAt, &def.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get semantic definition: %w", err)
	}
	
	if synonymsJSON != nil {
		json.Unmarshal(synonymsJSON, &def.Synonyms)
	}
	if relatedTermsJSON != nil {
		json.Unmarshal(relatedTermsJSON, &def.RelatedTerms)
	}
	if examplesJSON != nil {
		json.Unmarshal(examplesJSON, &def.Examples)
	}
	if aiMappingJSON != nil {
		json.Unmarshal(aiMappingJSON, &def.AIModelMapping)
	}
	
	return &def, nil
}

/* SearchSemanticDefinitions searches semantic definitions */
func (s *SemanticService) SearchSemanticDefinitions(ctx context.Context, query string, category *string, limit int) ([]SemanticDefinition, error) {
	sqlQuery := `
		SELECT id, term, definition, category, sql_expression, ai_model_mapping, synonyms,
		       related_terms, examples, owner_id, version, status, created_at, updated_at
		FROM neuronip.semantic_definitions
		WHERE (term ILIKE $1 OR definition ILIKE $1)`
	
	args := []interface{}{"%" + query + "%"}
	argIndex := 2
	
	if category != nil {
		sqlQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, *category)
		argIndex++
	}
	
	sqlQuery += " ORDER BY term"
	
	if limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}
	
	rows, err := s.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search semantic definitions: %w", err)
	}
	defer rows.Close()
	
	definitions := make([]SemanticDefinition, 0)
	for rows.Next() {
		var def SemanticDefinition
		var synonymsJSON, relatedTermsJSON, examplesJSON, aiMappingJSON json.RawMessage
		
		err := rows.Scan(
			&def.ID, &def.Term, &def.Definition, &def.Category, &def.SQLExpression, &aiMappingJSON,
			&synonymsJSON, &relatedTermsJSON, &examplesJSON, &def.OwnerID, &def.Version, &def.Status,
			&def.CreatedAt, &def.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if synonymsJSON != nil {
			json.Unmarshal(synonymsJSON, &def.Synonyms)
		}
		if relatedTermsJSON != nil {
			json.Unmarshal(relatedTermsJSON, &def.RelatedTerms)
		}
		if examplesJSON != nil {
			json.Unmarshal(examplesJSON, &def.Examples)
		}
		if aiMappingJSON != nil {
			json.Unmarshal(aiMappingJSON, &def.AIModelMapping)
		}
		
		definitions = append(definitions, def)
	}
	
	return definitions, nil
}
