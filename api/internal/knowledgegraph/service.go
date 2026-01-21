package knowledgegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides knowledge graph functionality */
type Service struct {
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
}

/* NewService creates a new knowledge graph service */
func NewService(pool *pgxpool.Pool, neurondbClient *neurondb.Client) *Service {
	return &Service{
		pool:           pool,
		neurondbClient: neurondbClient,
	}
}

/* Entity represents a knowledge graph entity */
type Entity struct {
	ID             uuid.UUID              `json:"id"`
	EntityName     string                 `json:"entity_name"`
	EntityTypeID   *uuid.UUID             `json:"entity_type_id,omitempty"`
	EntityValue    *string                `json:"entity_value,omitempty"`
	Description    *string                `json:"description,omitempty"`
	SourceDocumentID *uuid.UUID           `json:"source_document_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ConfidenceScore float64               `json:"confidence_score"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

/* EntityLink represents a relationship between entities */
type EntityLink struct {
	ID                 uuid.UUID              `json:"id"`
	SourceEntityID     uuid.UUID              `json:"source_entity_id"`
	TargetEntityID     uuid.UUID              `json:"target_entity_id"`
	RelationshipType   string                 `json:"relationship_type"`
	RelationshipStrength float64              `json:"relationship_strength"`
	Description        *string                `json:"description,omitempty"`
	SourceDocumentID   *uuid.UUID             `json:"source_document_id,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

/* EntityType represents an entity type classification */
type EntityType struct {
	ID           uuid.UUID  `json:"id"`
	TypeName     string     `json:"type_name"`
	Description  *string    `json:"description,omitempty"`
	ParentTypeID *uuid.UUID `json:"parent_type_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

/* ExtractEntitiesRequest represents entity extraction request */
type ExtractEntitiesRequest struct {
	DocumentID    uuid.UUID
	Text          string
	EntityTypes   []string // Optional: filter by entity types
	MinConfidence float64  // Minimum confidence score
}

/* ExtractEntities extracts entities from text */
func (s *Service) ExtractEntities(ctx context.Context, req ExtractEntitiesRequest) ([]Entity, error) {
	if req.MinConfidence <= 0 {
		req.MinConfidence = 0.5
	}

	// Use NeuronDB classify function to extract entities
	// In production, this would use a proper NER model
	classification, err := s.neurondbClient.Classify(ctx, req.Text, "entity-extraction")
	if err != nil {
		return nil, fmt.Errorf("failed to classify text for entity extraction: %w", err)
	}

	// Parse classification results to extract entities
	// This is a simplified implementation - in production, use proper NER model
	entities := s.parseClassificationResults(ctx, classification, req)

	return entities, nil
}

/* parseClassificationResults parses classification results into entities */
func (s *Service) parseClassificationResults(ctx context.Context, classification map[string]interface{}, req ExtractEntitiesRequest) []Entity {
	var entities []Entity

	// Get entity type IDs if entity types specified
	entityTypeMap := make(map[string]uuid.UUID)
	if len(req.EntityTypes) > 0 {
		for _, typeName := range req.EntityTypes {
			typeID, err := s.getEntityTypeID(ctx, typeName)
			if err == nil {
				entityTypeMap[typeName] = typeID
			}
		}
	}

	// Parse classification results
	// In production, this would parse structured NER output
	if entitiesData, ok := classification["entities"].([]interface{}); ok {
		for _, entityData := range entitiesData {
			if entityMap, ok := entityData.(map[string]interface{}); ok {
				entity := s.createEntityFromMap(ctx, entityMap, req, entityTypeMap)
				if entity != nil && entity.ConfidenceScore >= req.MinConfidence {
					entities = append(entities, *entity)
				}
			}
		}
	}

	return entities
}

/* createEntityFromMap creates an entity from parsed map data */
func (s *Service) createEntityFromMap(ctx context.Context, entityMap map[string]interface{}, req ExtractEntitiesRequest, entityTypeMap map[string]uuid.UUID) *Entity {
	name, ok := entityMap["name"].(string)
	if !ok || name == "" {
		return nil
	}

	entityType := ""
	if et, ok := entityMap["type"].(string); ok {
		entityType = et
	}

	var typeID *uuid.UUID
	if typeIDVal, ok := entityTypeMap[entityType]; ok {
		typeID = &typeIDVal
	}

	confidence := 1.0
	if conf, ok := entityMap["confidence"].(float64); ok {
		confidence = conf
	}

	// Generate embedding for entity
	embeddingText := name
	if desc, ok := entityMap["description"].(string); ok && desc != "" {
		embeddingText += " " + desc
	}

	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, embeddingText, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		embedding = "" // Continue without embedding
	}

	entity := &Entity{
		EntityName:     name,
		EntityTypeID:   typeID,
		ConfidenceScore: confidence,
		SourceDocumentID: &req.DocumentID,
	}

	if val, ok := entityMap["value"].(string); ok {
		entity.EntityValue = &val
	}

	if desc, ok := entityMap["description"].(string); ok {
		entity.Description = &desc
	}

	// Store entity in database
	if embedding != "" {
		s.storeEntity(ctx, entity, embedding)
	}

	return entity
}

/* storeEntity stores an entity in the database */
func (s *Service) storeEntity(ctx context.Context, entity *Entity, embedding string) error {
	metadataJSON, _ := json.Marshal(entity.Metadata)
	now := time.Now()

	query := `
		INSERT INTO neuronip.entities 
		(id, entity_name, entity_type_id, entity_value, description, source_document_id, metadata, embedding, confidence_score, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7::vector, $8, $9, $10)
		ON CONFLICT (entity_name, entity_type_id, source_document_id) 
		DO UPDATE SET 
			entity_value = EXCLUDED.entity_value,
			description = EXCLUDED.description,
			metadata = EXCLUDED.metadata,
			embedding = EXCLUDED.embedding,
			confidence_score = EXCLUDED.confidence_score,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		entity.EntityName, entity.EntityTypeID, entity.EntityValue, entity.Description,
		entity.SourceDocumentID, metadataJSON, embedding, entity.ConfidenceScore, now, now,
	).Scan(&entity.ID, &entity.CreatedAt, &entity.UpdatedAt)

	return err
}

/* LinkEntities links two entities with a relationship */
func (s *Service) LinkEntities(ctx context.Context, sourceEntityID uuid.UUID, targetEntityID uuid.UUID, relationshipType string, description *string, strength float64) (*EntityLink, error) {
	if strength <= 0 {
		strength = 1.0
	}

	link := &EntityLink{
		SourceEntityID:     sourceEntityID,
		TargetEntityID:     targetEntityID,
		RelationshipType:   relationshipType,
		RelationshipStrength: strength,
		Description:        description,
	}

	now := time.Now()
	metadataJSON, _ := json.Marshal(link.Metadata)

	query := `
		INSERT INTO neuronip.entity_links 
		(id, source_entity_id, target_entity_id, relationship_type, relationship_strength, description, source_document_id, metadata, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source_entity_id, target_entity_id, relationship_type) 
		DO UPDATE SET 
			relationship_strength = EXCLUDED.relationship_strength,
			description = EXCLUDED.description,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		sourceEntityID, targetEntityID, relationshipType, strength, description, link.SourceDocumentID,
		metadataJSON, now, now,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create entity link: %w", err)
	}

	return link, nil
}

/* GetEntity retrieves an entity by ID */
func (s *Service) GetEntity(ctx context.Context, entityID uuid.UUID) (*Entity, error) {
	query := `
		SELECT id, entity_name, entity_type_id, entity_value, description, source_document_id, metadata, confidence_score, created_at, updated_at
		FROM neuronip.entities
		WHERE id = $1`

	var entity Entity
	var metadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, entityID).Scan(
		&entity.ID, &entity.EntityName, &entity.EntityTypeID, &entity.EntityValue,
		&entity.Description, &entity.SourceDocumentID, &metadataJSON, &entity.ConfidenceScore,
		&entity.CreatedAt, &entity.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get entity: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &entity.Metadata)
	}

	return &entity, nil
}

/* GetEntityLinks retrieves links for an entity */
func (s *Service) GetEntityLinks(ctx context.Context, entityID uuid.UUID, direction string) ([]EntityLink, error) {
	var query string
	if direction == "outgoing" {
		query = `SELECT id, source_entity_id, target_entity_id, relationship_type, relationship_strength, description, source_document_id, metadata, created_at, updated_at
			FROM neuronip.entity_links
			WHERE source_entity_id = $1`
	} else if direction == "incoming" {
		query = `SELECT id, source_entity_id, target_entity_id, relationship_type, relationship_strength, description, source_document_id, metadata, created_at, updated_at
			FROM neuronip.entity_links
			WHERE target_entity_id = $1`
	} else {
		query = `SELECT id, source_entity_id, target_entity_id, relationship_type, relationship_strength, description, source_document_id, metadata, created_at, updated_at
			FROM neuronip.entity_links
			WHERE source_entity_id = $1 OR target_entity_id = $1`
	}

	rows, err := s.pool.Query(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity links: %w", err)
	}
	defer rows.Close()

	var links []EntityLink
	for rows.Next() {
		var link EntityLink
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&link.ID, &link.SourceEntityID, &link.TargetEntityID, &link.RelationshipType,
			&link.RelationshipStrength, &link.Description, &link.SourceDocumentID,
			&metadataJSON, &link.CreatedAt, &link.UpdatedAt,
		)

		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &link.Metadata)
		}

		links = append(links, link)
	}

	return links, nil
}

/* SearchEntities performs semantic search on entities */
func (s *Service) SearchEntities(ctx context.Context, query string, entityTypeID *uuid.UUID, limit int) ([]Entity, error) {
	if limit <= 0 {
		limit = 10
	}

	// Generate embedding for query
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, query, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Build query
	var searchQuery string
	var args []interface{}

	if entityTypeID != nil {
		searchQuery = `
			SELECT id, entity_name, entity_type_id, entity_value, description, source_document_id, metadata, confidence_score, created_at, updated_at
			FROM neuronip.entities
			WHERE entity_type_id = $2 AND embedding IS NOT NULL
			ORDER BY embedding <=> $1::vector
			LIMIT $3`
		args = []interface{}{queryEmbedding, entityTypeID, limit}
	} else {
		searchQuery = `
			SELECT id, entity_name, entity_type_id, entity_value, description, source_document_id, metadata, confidence_score, created_at, updated_at
			FROM neuronip.entities
			WHERE embedding IS NOT NULL
			ORDER BY embedding <=> $1::vector
			LIMIT $2`
		args = []interface{}{queryEmbedding, limit}
	}

	rows, err := s.pool.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&entity.ID, &entity.EntityName, &entity.EntityTypeID, &entity.EntityValue,
			&entity.Description, &entity.SourceDocumentID, &metadataJSON, &entity.ConfidenceScore,
			&entity.CreatedAt, &entity.UpdatedAt,
		)

		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &entity.Metadata)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

/* getEntityTypeID gets entity type ID by name */
func (s *Service) getEntityTypeID(ctx context.Context, typeName string) (uuid.UUID, error) {
	var id uuid.UUID
	query := `SELECT id FROM neuronip.entity_types WHERE type_name = $1`
	err := s.pool.QueryRow(ctx, query, typeName).Scan(&id)
	return id, err
}

/* CreateEntityType creates a new entity type */
func (s *Service) CreateEntityType(ctx context.Context, typeName string, description *string, parentTypeID *uuid.UUID) (*EntityType, error) {
	now := time.Now()

	query := `
		INSERT INTO neuronip.entity_types (id, type_name, description, parent_type_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		ON CONFLICT (type_name) DO UPDATE SET 
			description = EXCLUDED.description,
			parent_type_id = EXCLUDED.parent_type_id,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	entityType := &EntityType{
		TypeName:     typeName,
		Description:  description,
		ParentTypeID: parentTypeID,
	}

	err := s.pool.QueryRow(ctx, query, typeName, description, parentTypeID, now, now).Scan(
		&entityType.ID, &entityType.CreatedAt, &entityType.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create entity type: %w", err)
	}

	return entityType, nil
}

/* GlossaryTerm represents a glossary term */
type GlossaryTerm struct {
	ID              uuid.UUID              `json:"id"`
	Term            string                 `json:"term"`
	Definition      string                 `json:"definition"`
	Category        *string                `json:"category,omitempty"`
	RelatedEntityID *uuid.UUID             `json:"related_entity_id,omitempty"`
	Synonyms        []string               `json:"synonyms,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

/* GraphTraversalResult represents graph traversal result */
type GraphTraversalResult struct {
	StartEntityID uuid.UUID              `json:"start_entity_id"`
	Entities      []Entity               `json:"entities"`
	Links         []EntityLink           `json:"links"`
	Paths         [][]uuid.UUID          `json:"paths"`
	Depth         int                    `json:"depth"`
}

/* TraverseGraph performs graph traversal from a starting entity */
func (s *Service) TraverseGraph(ctx context.Context, startEntityID uuid.UUID, maxDepth int, relationshipTypes []string, direction string) (*GraphTraversalResult, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}

	visited := make(map[uuid.UUID]bool)
	entities := make(map[uuid.UUID]*Entity)
	links := []EntityLink{}
	paths := [][]uuid.UUID{}

	// BFS traversal
	queue := []struct {
		entityID uuid.UUID
		depth    int
		path     []uuid.UUID
	}{
		{startEntityID, 0, []uuid.UUID{startEntityID}},
	}

	visited[startEntityID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= maxDepth {
			continue
		}

		// Get entity if not already loaded
		if _, exists := entities[current.entityID]; !exists {
			entity, err := s.GetEntity(ctx, current.entityID)
			if err != nil {
				continue
			}
			entities[current.entityID] = entity
		}

		// Get links for current entity
		entityLinks, err := s.GetEntityLinks(ctx, current.entityID, direction)
		if err != nil {
			continue
		}

		for _, link := range entityLinks {
			// Filter by relationship types if specified
			if len(relationshipTypes) > 0 {
				found := false
				for _, rt := range relationshipTypes {
					if link.RelationshipType == rt {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			links = append(links, link)

			// Determine next entity ID
			var nextEntityID uuid.UUID
			if link.SourceEntityID == current.entityID {
				nextEntityID = link.TargetEntityID
			} else {
				nextEntityID = link.SourceEntityID
			}

			// Skip if already visited
			if visited[nextEntityID] {
				continue
			}

			visited[nextEntityID] = true

			// Create new path
			newPath := make([]uuid.UUID, len(current.path))
			copy(newPath, current.path)
			newPath = append(newPath, nextEntityID)
			paths = append(paths, newPath)

			// Add to queue
			queue = append(queue, struct {
				entityID uuid.UUID
				depth    int
				path     []uuid.UUID
			}{nextEntityID, current.depth + 1, newPath})
		}
	}

	// Convert entities map to slice
	entitySlice := make([]Entity, 0, len(entities))
	for _, entity := range entities {
		entitySlice = append(entitySlice, *entity)
	}

	return &GraphTraversalResult{
		StartEntityID: startEntityID,
		Entities:      entitySlice,
		Links:         links,
		Paths:         paths,
		Depth:         maxDepth,
	}, nil
}

/* CreateGlossaryTerm creates a new glossary term */
func (s *Service) CreateGlossaryTerm(ctx context.Context, term string, definition string, category *string, relatedEntityID *uuid.UUID, synonyms []string) (*GlossaryTerm, error) {
	now := time.Now()

	query := `
		INSERT INTO neuronip.glossary (id, term, definition, category, related_entity_id, synonyms, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (term) DO UPDATE SET 
			definition = EXCLUDED.definition,
			category = EXCLUDED.category,
			related_entity_id = EXCLUDED.related_entity_id,
			synonyms = EXCLUDED.synonyms,
			updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at`

	glossaryTerm := &GlossaryTerm{
		Term:            term,
		Definition:      definition,
		Category:        category,
		RelatedEntityID: relatedEntityID,
		Synonyms:        synonyms,
	}

	err := s.pool.QueryRow(ctx, query, term, definition, category, relatedEntityID, synonyms, now, now).Scan(
		&glossaryTerm.ID, &glossaryTerm.CreatedAt, &glossaryTerm.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create glossary term: %w", err)
	}

	return glossaryTerm, nil
}

/* GetGlossaryTerm retrieves a glossary term by ID */
func (s *Service) GetGlossaryTerm(ctx context.Context, termID uuid.UUID) (*GlossaryTerm, error) {
	query := `
		SELECT id, term, definition, category, related_entity_id, synonyms, metadata, created_at, updated_at
		FROM neuronip.glossary
		WHERE id = $1`

	var term GlossaryTerm
	var metadataJSON json.RawMessage

	err := s.pool.QueryRow(ctx, query, termID).Scan(
		&term.ID, &term.Term, &term.Definition, &term.Category, &term.RelatedEntityID,
		&term.Synonyms, &metadataJSON, &term.CreatedAt, &term.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get glossary term: %w", err)
	}

	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &term.Metadata)
	}

	return &term, nil
}

/* SearchGlossary performs search on glossary terms */
func (s *Service) SearchGlossary(ctx context.Context, query string, category *string, limit int) ([]GlossaryTerm, error) {
	if limit <= 0 {
		limit = 10
	}

	var searchQuery string
	var args []interface{}

	if category != nil {
		searchQuery = `
			SELECT id, term, definition, category, related_entity_id, synonyms, metadata, created_at, updated_at
			FROM neuronip.glossary
			WHERE category = $1 AND (term ILIKE $2 OR definition ILIKE $2)
			ORDER BY term
			LIMIT $3`
		args = []interface{}{*category, "%" + query + "%", limit}
	} else {
		searchQuery = `
			SELECT id, term, definition, category, related_entity_id, synonyms, metadata, created_at, updated_at
			FROM neuronip.glossary
			WHERE term ILIKE $1 OR definition ILIKE $1
			ORDER BY term
			LIMIT $2`
		args = []interface{}{"%" + query + "%", limit}
	}

	rows, err := s.pool.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search glossary: %w", err)
	}
	defer rows.Close()

	var terms []GlossaryTerm
	for rows.Next() {
		var term GlossaryTerm
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&term.ID, &term.Term, &term.Definition, &term.Category, &term.RelatedEntityID,
			&term.Synonyms, &metadataJSON, &term.CreatedAt, &term.UpdatedAt,
		)

		if err != nil {
			continue
		}

		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &term.Metadata)
		}

		terms = append(terms, term)
	}

	return terms, nil
}

/* ExecuteGraphQuery executes a graph query (Cypher-like) */
func (s *Service) ExecuteGraphQuery(ctx context.Context, queryStr string) (*GraphQueryResult, error) {
	// Simple graph query parser - supports basic MATCH patterns
	// Supported patterns:
	//   MATCH (a)-[r]->(b)
	//   MATCH (a)-[r]->(b) WHERE a.type = 'person'
	//   MATCH (a)-[r]->(b) RETURN a, b, r
	
	queryStr = strings.TrimSpace(queryStr)
	
	// Parse MATCH clause
	if !strings.HasPrefix(strings.ToUpper(queryStr), "MATCH") {
		return nil, fmt.Errorf("query must start with MATCH")
	}
	
	// Extract MATCH pattern (simplified parsing)
	pattern := ""
	matchIdx := strings.Index(strings.ToUpper(queryStr), "MATCH")
	if matchIdx >= 0 {
		whereIdx := strings.Index(strings.ToUpper(queryStr), "WHERE")
		returnIdx := strings.Index(strings.ToUpper(queryStr), "RETURN")
		
		endIdx := len(queryStr)
		if whereIdx >= 0 && whereIdx < endIdx {
			endIdx = whereIdx
		}
		if returnIdx >= 0 && returnIdx < endIdx {
			endIdx = returnIdx
		}
		
		if matchIdx+5 < endIdx {
			pattern = strings.TrimSpace(queryStr[matchIdx+5 : endIdx])
		}
	}
	
	// Parse WHERE clause (simplified)
	whereClause := ""
	whereIdx := strings.Index(strings.ToUpper(queryStr), "WHERE")
	if whereIdx >= 0 {
		returnIdx := strings.Index(strings.ToUpper(queryStr), "RETURN")
		endIdx := len(queryStr)
		if returnIdx >= 0 {
			endIdx = returnIdx
		}
		whereClause = strings.TrimSpace(queryStr[whereIdx+5 : endIdx])
	}
	
	// Execute query based on pattern
	if strings.Contains(pattern, "-") && strings.Contains(pattern, "->") {
		// Pattern: (a)-[r]->(b)
		return s.executeEdgeQuery(ctx, whereClause)
	} else if strings.Contains(pattern, "(") {
		// Pattern: (a) - single node
		return s.executeNodeQuery(ctx, whereClause)
	}
	
	return &GraphQueryResult{
		Nodes: []Entity{},
		Edges: []EntityLink{},
	}, fmt.Errorf("unsupported query pattern: %s", pattern)
}

/* executeEdgeQuery executes a query for edges */
func (s *Service) executeEdgeQuery(ctx context.Context, whereClause string) (*GraphQueryResult, error) {
	query := `
		SELECT 
			e.id, e.entity_name, e.entity_type_id, e.entity_value, e.metadata, e.created_at,
			el.id, el.source_entity_id, el.target_entity_id, el.relationship_type, el.metadata, el.created_at
		FROM neuronip.entity_links el
		JOIN neuronip.entities e ON e.id = el.source_entity_id OR e.id = el.target_entity_id`
	
	// Add WHERE conditions if present
	if whereClause != "" {
		// Simple WHERE parsing (basic type filtering)
		if strings.Contains(whereClause, "type") {
			// Extract type value - need to join entity_types table
			re := regexp.MustCompile(`type\s*=\s*['"]([^'"]+)['"]`)
			matches := re.FindStringSubmatch(whereClause)
			if len(matches) > 1 {
				entityType := matches[1]
				query += fmt.Sprintf(` 
					JOIN neuronip.entity_types et ON e.entity_type_id = et.id 
					WHERE et.type_name = '%s'`, entityType)
			}
		}
	}
	
	query += " LIMIT 1000"
	
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	nodesMap := make(map[uuid.UUID]*Entity)
	var edges []EntityLink
	
	for rows.Next() {
		var entity Entity
		var link EntityLink
		var entityTypeID *uuid.UUID
		var entityValue *string
		var entityMetadata, linkMetadata json.RawMessage
		
		if err := rows.Scan(
			&entity.ID, &entity.EntityName, &entityTypeID, &entityValue, &entityMetadata, &entity.CreatedAt,
			&link.ID, &link.SourceEntityID, &link.TargetEntityID, &link.RelationshipType, &linkMetadata, &link.CreatedAt,
		); err != nil {
			continue
		}
		
		entity.EntityTypeID = entityTypeID
		entity.EntityValue = entityValue
		
		if entityMetadata != nil {
			json.Unmarshal(entityMetadata, &entity.Metadata)
		}
		if linkMetadata != nil {
			json.Unmarshal(linkMetadata, &link.Metadata)
		}
		
		nodesMap[entity.ID] = &entity
		edges = append(edges, link)
	}
	
	nodes := make([]Entity, 0, len(nodesMap))
	for _, node := range nodesMap {
		nodes = append(nodes, *node)
	}
	
	return &GraphQueryResult{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

/* executeNodeQuery executes a query for nodes */
func (s *Service) executeNodeQuery(ctx context.Context, whereClause string) (*GraphQueryResult, error) {
	query := `SELECT id, entity_name, entity_type_id, entity_value, metadata, created_at 
			  FROM neuronip.entities WHERE 1=1`
	
	// Add WHERE conditions
	if whereClause != "" {
		if strings.Contains(whereClause, "type") {
			re := regexp.MustCompile(`type\s*=\s*['"]([^'"]+)['"]`)
			matches := re.FindStringSubmatch(whereClause)
			if len(matches) > 1 {
				query += fmt.Sprintf(` 
					AND entity_type_id IN (
						SELECT id FROM neuronip.entity_types WHERE type_name = '%s'
					)`, matches[1])
			}
		}
	}
	
	query += " LIMIT 1000"
	
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()
	
	var nodes []Entity
	for rows.Next() {
		var entity Entity
		var entityTypeID *uuid.UUID
		var entityValue *string
		var metadata json.RawMessage
		
		if err := rows.Scan(&entity.ID, &entity.EntityName, &entityTypeID, &entityValue, &metadata, &entity.CreatedAt); err != nil {
			continue
		}
		
		entity.EntityTypeID = entityTypeID
		entity.EntityValue = entityValue
		
		if metadata != nil {
			json.Unmarshal(metadata, &entity.Metadata)
		}
		
		nodes = append(nodes, entity)
	}
	
	return &GraphQueryResult{
		Nodes: nodes,
		Edges: []EntityLink{},
	}, nil
}

/* GraphQueryResult represents the result of a graph query */
type GraphQueryResult struct {
	Nodes []Entity     `json:"nodes"`
	Edges []EntityLink `json:"edges"`
}

/* CalculateCentrality calculates centrality metrics for entities */
func (s *Service) CalculateCentrality(ctx context.Context, entityID uuid.UUID, centralityType string) (float64, error) {
	// Calculate different types of centrality
	// - degree: number of connections
	// - betweenness: how often entity appears in shortest paths
	// - closeness: average distance to all other entities
	
	switch centralityType {
	case "degree":
		// Count connections
		query := `
			SELECT COUNT(*) 
			FROM neuronip.entity_links
			WHERE source_entity_id = $1 OR target_entity_id = $1`
		
		var count int
		err := s.pool.QueryRow(ctx, query, entityID).Scan(&count)
		return float64(count), err
		
	case "betweenness":
		// Betweenness centrality: fraction of shortest paths passing through this node
		// Simplified implementation using shortest path counts
		query := `
			WITH all_nodes AS (
				SELECT DISTINCT source_entity_id as entity_id FROM neuronip.entity_links
				UNION
				SELECT DISTINCT target_entity_id FROM neuronip.entity_links
			),
			shortest_paths AS (
				SELECT 
					n1.entity_id as from_node,
					n2.entity_id as to_node,
					CASE 
						WHEN EXISTS (
							SELECT 1 FROM neuronip.entity_links 
							WHERE source_entity_id = n1.entity_id AND target_entity_id = n2.entity_id
						) THEN 1
						ELSE NULL
					END as direct_path
				FROM all_nodes n1
				CROSS JOIN all_nodes n2
				WHERE n1.entity_id != n2.entity_id
			),
			paths_through_node AS (
				SELECT COUNT(*) as path_count
				FROM shortest_paths sp1
				JOIN neuronip.entity_links el ON el.source_entity_id = sp1.from_node AND el.target_entity_id = $1
				JOIN neuronip.entity_links el2 ON el2.source_entity_id = $1 AND el2.target_entity_id = sp1.to_node
				WHERE sp1.direct_path IS NULL
			)
			SELECT COALESCE((SELECT path_count FROM paths_through_node), 0)::float / 
				NULLIF((SELECT COUNT(*) FROM shortest_paths WHERE direct_path IS NULL), 0), 0.0) as betweenness`
		
		var betweenness float64
		err := s.pool.QueryRow(ctx, query, entityID).Scan(&betweenness)
		if err != nil {
			// Fallback to simpler calculation
			return s.calculateSimplifiedBetweenness(ctx, entityID)
		}
		return betweenness, nil
		
	case "closeness":
		// Closeness centrality: average distance to all other reachable nodes
		query := `
			WITH RECURSIVE node_distances AS (
				-- Start with direct connections
				SELECT 
					$1 as start_node,
					CASE WHEN source_entity_id = $1 THEN target_entity_id ELSE source_entity_id END as target_node,
					1 as distance
				FROM neuronip.entity_links
				WHERE source_entity_id = $1 OR target_entity_id = $1
				
				UNION
				
				-- Find nodes at distance 2+
				SELECT 
					nd.start_node,
					CASE 
						WHEN el.source_entity_id = nd.target_node THEN el.target_entity_id
						ELSE el.source_entity_id
					END as target_node,
					nd.distance + 1
				FROM node_distances nd
				JOIN neuronip.entity_links el ON (
					el.source_entity_id = nd.target_node OR el.target_entity_id = nd.target_node
				)
				WHERE nd.distance < 5 AND nd.target_node != nd.start_node
			),
			all_distances AS (
				SELECT DISTINCT ON (target_node) target_node, distance
				FROM node_distances
				WHERE target_node != $1
				ORDER BY target_node, distance
			),
			total_distance AS (
				SELECT SUM(distance) as sum_dist, COUNT(*) as node_count
				FROM all_distances
			)
			SELECT 
				CASE 
					WHEN node_count > 0 THEN node_count::float / sum_dist::float
					ELSE 0.0
				END as closeness
			FROM total_distance`
		
		var closeness float64
		err := s.pool.QueryRow(ctx, query, entityID).Scan(&closeness)
		if err != nil {
			// Fallback to direct connections count
			degree, _ := s.CalculateCentrality(ctx, entityID, "degree")
			return degree / 100.0, nil // Normalized approximation
		}
		return closeness, nil
		
	default:
		return 0.0, fmt.Errorf("unknown centrality type: %s", centralityType)
	}
}

/* DetectCommunities detects communities in the graph */
func (s *Service) DetectCommunities(ctx context.Context) ([]Community, error) {
	// Community detection using connected components as communities
	// This is a simplified approach - a full implementation would use Louvain or similar algorithm
	
	// Find all connected components (communities)
	query := `
		WITH RECURSIVE component_nodes AS (
			-- Start with each node as its own component
			SELECT DISTINCT source_entity_id as entity_id, 
				   ROW_NUMBER() OVER () as component_id
			FROM neuronip.entity_links
			
			UNION
			
			-- Expand components through edges
			SELECT DISTINCT 
				CASE WHEN el.source_entity_id = cn.entity_id THEN el.target_entity_id ELSE el.source_entity_id END as entity_id,
				cn.component_id
			FROM component_nodes cn
			JOIN neuronip.entity_links el ON (
				el.source_entity_id = cn.entity_id OR el.target_entity_id = cn.entity_id
			)
		),
		components AS (
			SELECT DISTINCT ON (entity_id) component_id, entity_id
			FROM component_nodes
			ORDER BY entity_id, component_id
		),
		community_sizes AS (
			SELECT component_id, COUNT(*) as size, array_agg(entity_id) as entities
			FROM components
			GROUP BY component_id
			ORDER BY size DESC
		)
		SELECT component_id::int, entities, size
		FROM community_sizes
		WHERE size > 1
		LIMIT 100`
	
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to detect communities: %w", err)
	}
	defer rows.Close()
	
	var communities []Community
	communityID := 0
	for rows.Next() {
		var compID int
		var size int
		var entityUUIDs []uuid.UUID
		
		if err := rows.Scan(&compID, &entityUUIDs, &size); err != nil {
			continue
		}
		
		communities = append(communities, Community{
			ID:       communityID,
			Entities: entityUUIDs,
			Size:     size,
		})
		communityID++
	}
	
	return communities, nil
}

/* calculateSimplifiedBetweenness calculates a simplified betweenness centrality */
func (s *Service) calculateSimplifiedBetweenness(ctx context.Context, entityID uuid.UUID) (float64, error) {
	// Count how many paths of length 2 pass through this node
	query := `
		SELECT COUNT(*)::float / NULLIF((
			SELECT COUNT(*) FROM neuronip.entity_links el1
			JOIN neuronip.entity_links el2 ON el1.target_entity_id = el2.source_entity_id
			WHERE el1.source_entity_id != $1 AND el2.target_entity_id != $1
		), 1)
		FROM neuronip.entity_links el1
		JOIN neuronip.entity_links el2 ON el1.target_entity_id = $1 AND el2.source_entity_id = $1
		WHERE el1.source_entity_id != $1 AND el2.target_entity_id != $1`
	
	var betweenness float64
	err := s.pool.QueryRow(ctx, query, entityID).Scan(&betweenness)
	if err != nil {
		return 0.0, nil // Return 0 if calculation fails
	}
	
	return betweenness, nil
}

/* Community represents a detected community */
type Community struct {
	ID       int       `json:"id"`
	Entities []uuid.UUID `json:"entities"`
	Size     int       `json:"size"`
}
