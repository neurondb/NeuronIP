package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
)

/* GlossaryLinker provides functionality to link glossary terms to columns, metrics, and documents */
type GlossaryLinker struct {
	pool *pgxpool.Pool
}

/* NewGlossaryLinker creates a new glossary linker */
func NewGlossaryLinker(pool *pgxpool.Pool) *GlossaryLinker {
	return &GlossaryLinker{pool: pool}
}

/* GlossaryLink represents a link between a glossary term and an entity */
type GlossaryLink struct {
	ID           uuid.UUID              `json:"id"`
	GlossaryID   uuid.UUID              `json:"glossary_id"`
	EntityType   string                 `json:"entity_type"` // "column", "metric", "document", "table"
	EntityID     uuid.UUID              `json:"entity_id"`
	EntityName   string                 `json:"entity_name"`
	Relationship string                 `json:"relationship"` // "defines", "related_to", "synonym_of"
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

/* LinkGlossaryToColumn links a glossary term to a column */
func (gl *GlossaryLinker) LinkGlossaryToColumn(ctx context.Context, glossaryID uuid.UUID, schemaName, tableName, columnName string, relationship string, confidence float64) error {
	// Create entity reference
	entityID := uuid.New()
	entityName := fmt.Sprintf("%s.%s.%s", schemaName, tableName, columnName)

	metadata := map[string]interface{}{
		"schema_name": schemaName,
		"table_name":  tableName,
		"column_name": columnName,
	}

	return gl.createLink(ctx, glossaryID, "column", entityID, entityName, relationship, confidence, metadata)
}

/* LinkGlossaryToMetric links a glossary term to a metric */
func (gl *GlossaryLinker) LinkGlossaryToMetric(ctx context.Context, glossaryID, metricID uuid.UUID, relationship string, confidence float64) error {
	// Get metric name
	var metricName string
	err := gl.pool.QueryRow(ctx, `SELECT name FROM neuronip.metric_catalog WHERE id = $1`, metricID).Scan(&metricName)
	if err != nil {
		return fmt.Errorf("failed to get metric: %w", err)
	}

	return gl.createLink(ctx, glossaryID, "metric", metricID, metricName, relationship, confidence, nil)
}

/* LinkGlossaryToDocument links a glossary term to a document */
func (gl *GlossaryLinker) LinkGlossaryToDocument(ctx context.Context, glossaryID, documentID uuid.UUID, relationship string, confidence float64) error {
	// Get document title
	var documentTitle string
	err := gl.pool.QueryRow(ctx, `SELECT title FROM neuronip.knowledge_documents WHERE id = $1`, documentID).Scan(&documentTitle)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	return gl.createLink(ctx, glossaryID, "document", documentID, documentTitle, relationship, confidence, nil)
}

/* LinkGlossaryToTable links a glossary term to a table */
func (gl *GlossaryLinker) LinkGlossaryToTable(ctx context.Context, glossaryID uuid.UUID, schemaName, tableName string, relationship string, confidence float64) error {
	entityID := uuid.New()
	entityName := fmt.Sprintf("%s.%s", schemaName, tableName)

	metadata := map[string]interface{}{
		"schema_name": schemaName,
		"table_name":  tableName,
	}

	return gl.createLink(ctx, glossaryID, "table", entityID, entityName, relationship, confidence, metadata)
}

/* createLink creates a glossary link */
func (gl *GlossaryLinker) createLink(ctx context.Context, glossaryID uuid.UUID, entityType string, entityID uuid.UUID, entityName, relationship string, confidence float64, metadata map[string]interface{}) error {
	linkID := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)
	now := time.Now()

	query := `
		INSERT INTO neuronip.glossary_links 
		(id, glossary_id, entity_type, entity_id, entity_name, relationship, confidence, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (glossary_id, entity_type, entity_id) 
		DO UPDATE SET 
			relationship = EXCLUDED.relationship,
			confidence = EXCLUDED.confidence,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err := gl.pool.Exec(ctx, query, linkID, glossaryID, entityType, entityID, entityName, relationship, confidence, metadataJSON, now, now)
	return err
}

/* GetGlossaryLinks retrieves links for a glossary term */
func (gl *GlossaryLinker) GetGlossaryLinks(ctx context.Context, glossaryID uuid.UUID, entityType *string) ([]GlossaryLink, error) {
	query := `
		SELECT id, glossary_id, entity_type, entity_id, entity_name, relationship, confidence, metadata, created_at, updated_at
		FROM neuronip.glossary_links
		WHERE glossary_id = $1
	`

	args := []interface{}{glossaryID}
	if entityType != nil {
		query += " AND entity_type = $2"
		args = append(args, *entityType)
	}

	query += " ORDER BY confidence DESC, created_at DESC"

	rows, err := gl.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get glossary links: %w", err)
	}
	defer rows.Close()

	var links []GlossaryLink
	for rows.Next() {
		var link GlossaryLink
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&link.ID, &link.GlossaryID, &link.EntityType, &link.EntityID, &link.EntityName,
			&link.Relationship, &link.Confidence, &metadataJSON, &link.CreatedAt, &link.UpdatedAt,
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

/* GetEntityLinks retrieves glossary links for an entity */
func (gl *GlossaryLinker) GetEntityLinks(ctx context.Context, entityType string, entityID uuid.UUID) ([]GlossaryLink, error) {
	query := `
		SELECT id, glossary_id, entity_type, entity_id, entity_name, relationship, confidence, metadata, created_at, updated_at
		FROM neuronip.glossary_links
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY confidence DESC, created_at DESC
	`

	rows, err := gl.pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity links: %w", err)
	}
	defer rows.Close()

	var links []GlossaryLink
	for rows.Next() {
		var link GlossaryLink
		var metadataJSON json.RawMessage

		err := rows.Scan(
			&link.ID, &link.GlossaryID, &link.EntityType, &link.EntityID, &link.EntityName,
			&link.Relationship, &link.Confidence, &metadataJSON, &link.CreatedAt, &link.UpdatedAt,
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

/* AutoLinkGlossaryTerms automatically links glossary terms to entities based on name matching */
func (gl *GlossaryLinker) AutoLinkGlossaryTerms(ctx context.Context, entityType string, minConfidence float64) error {
	// Get all glossary terms
	glossaryQuery := `SELECT id, term FROM neuronip.glossary`
	rows, err := gl.pool.Query(ctx, glossaryQuery)
	if err != nil {
		return fmt.Errorf("failed to get glossary terms: %w", err)
	}
	defer rows.Close()

	type GlossaryTerm struct {
		ID   uuid.UUID
		Term string
	}

	var terms []GlossaryTerm
	for rows.Next() {
		var term GlossaryTerm
		if err := rows.Scan(&term.ID, &term.Term); err == nil {
			terms = append(terms, term)
		}
	}

	// Convert to knowledgegraph.GlossaryTerm
	kgTerms := make([]knowledgegraph.GlossaryTerm, len(terms))
	for i, term := range terms {
		kgTerms[i] = knowledgegraph.GlossaryTerm{
			ID:   term.ID,
			Term: term.Term,
		}
	}

	// For each entity type, find matches
	switch entityType {
	case "column":
		return gl.autoLinkColumns(ctx, kgTerms, minConfidence)
	case "metric":
		return gl.autoLinkMetrics(ctx, kgTerms, minConfidence)
	case "document":
		return gl.autoLinkDocuments(ctx, kgTerms, minConfidence)
	}

	return nil
}

/* autoLinkColumns automatically links glossary terms to columns by querying information_schema */
func (gl *GlossaryLinker) autoLinkColumns(ctx context.Context, terms []knowledgegraph.GlossaryTerm, minConfidence float64) error {
	columnsQuery := `
		SELECT table_schema, table_name, column_name
		FROM information_schema.columns
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		AND table_schema LIKE 'neuronip%'
		ORDER BY table_schema, table_name, column_name`

	rows, err := gl.pool.Query(ctx, columnsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName, tableName, columnName string
		if err := rows.Scan(&schemaName, &tableName, &columnName); err != nil {
			continue
		}

		fullColumnName := fmt.Sprintf("%s.%s.%s", schemaName, tableName, columnName)

		// Match against glossary terms
		for _, term := range terms {
			confidence := gl.calculateMatchConfidence(term.Term, columnName, fullColumnName)
			if confidence >= minConfidence {
				// Store column link in glossary_links table
				_, err := gl.pool.Exec(ctx, `
					INSERT INTO neuronip.glossary_links (glossary_id, resource_type, resource_id, link_type, confidence)
					VALUES ($1, 'column', $2, 'defines', $3)
					ON CONFLICT DO NOTHING`,
					term.ID, fullColumnName, confidence)
				if err != nil {
					continue
				}
			}
		}
	}

	return nil
}

/* autoLinkMetrics automatically links glossary terms to metrics */
func (gl *GlossaryLinker) autoLinkMetrics(ctx context.Context, terms []knowledgegraph.GlossaryTerm, minConfidence float64) error {
	metricsQuery := `SELECT id, name, display_name FROM neuronip.metric_catalog`
	rows, err := gl.pool.Query(ctx, metricsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var metricID uuid.UUID
		var name, displayName string
		if err := rows.Scan(&metricID, &name, &displayName); err != nil {
			continue
		}

		// Match against glossary terms
		for _, term := range terms {
			confidence := gl.calculateMatchConfidence(term.Term, name, displayName)
			if confidence >= minConfidence {
				gl.LinkGlossaryToMetric(ctx, term.ID, metricID, "defines", confidence)
			}
		}
	}

	return nil
}

/* autoLinkDocuments automatically links glossary terms to documents */
func (gl *GlossaryLinker) autoLinkDocuments(ctx context.Context, terms []knowledgegraph.GlossaryTerm, minConfidence float64) error {
	docsQuery := `SELECT id, title FROM neuronip.knowledge_documents`
	rows, err := gl.pool.Query(ctx, docsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var docID uuid.UUID
		var title string
		if err := rows.Scan(&docID, &title); err != nil {
			continue
		}

		// Match against glossary terms
		for _, term := range terms {
			confidence := gl.calculateMatchConfidence(term.Term, title, "")
			if confidence >= minConfidence {
				gl.LinkGlossaryToDocument(ctx, term.ID, docID, "related_to", confidence)
			}
		}
	}

	return nil
}

/* calculateMatchConfidence calculates confidence score for term matching */
func (gl *GlossaryLinker) calculateMatchConfidence(term, name1, name2 string) float64 {
	termLower := strings.ToLower(term)
	name1Lower := strings.ToLower(name1)

	// Exact match
	if termLower == name1Lower {
		return 1.0
	}

	// Contains match
	if strings.Contains(name1Lower, termLower) || strings.Contains(termLower, name1Lower) {
		return 0.8
	}

	// Check second name if provided
	if name2 != "" {
		name2Lower := strings.ToLower(name2)
		if termLower == name2Lower {
			return 1.0
		}
		if strings.Contains(name2Lower, termLower) || strings.Contains(termLower, name2Lower) {
			return 0.8
		}
	}

	// Word overlap
	termWords := strings.Fields(termLower)
	nameWords := strings.Fields(name1Lower)
	overlap := 0
	for _, tw := range termWords {
		for _, nw := range nameWords {
			if tw == nw {
				overlap++
				break
			}
		}
	}

	if len(termWords) > 0 {
		return float64(overlap) / float64(len(termWords))
	}

	return 0.0
}
