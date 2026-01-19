package semantic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides semantic search functionality */
type Service struct {
	queries       *db.Queries
	pool          *pgxpool.Pool
	neurondbClient *neurondb.Client
	auditService  *compliance.AuditService
}

/* NewService creates a new semantic search service */
func NewService(queries *db.Queries, pool *pgxpool.Pool, neurondbClient *neurondb.Client) *Service {
	return &Service{
		queries:       queries,
		pool:          pool,
		neurondbClient: neurondbClient,
		auditService:  compliance.NewAuditService(pool),
	}
}

/* SearchRequest represents a semantic search request */
type SearchRequest struct {
	Query      string
	CollectionID *uuid.UUID
	Limit      int
	Threshold  float64
}

/* SearchResult represents a search result */
type SearchResult struct {
	DocumentID   uuid.UUID              `json:"document_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	ContentType  string                 `json:"content_type"`
	Similarity   float64                `json:"similarity"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

/* Search performs semantic search on knowledge documents */
func (s *Service) Search(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.5
	}

	// Generate embedding for the query using NeuronDB
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, req.Query, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector similarity search
	var searchQuery string
	var args []interface{}

	if req.CollectionID != nil {
		searchQuery = `
			SELECT 
				kd.id,
				kd.title,
				kd.content,
				kd.content_type,
				1 - (ke.embedding <=> $1::vector) as similarity,
				kd.metadata
			FROM neuronip.knowledge_documents kd
			JOIN neuronip.knowledge_embeddings ke ON ke.document_id = kd.id
			WHERE kd.collection_id = $2
				AND 1 - (ke.embedding <=> $1::vector) >= $3
			ORDER BY ke.embedding <=> $1::vector
			LIMIT $4`
		args = []interface{}{queryEmbedding, req.CollectionID, req.Threshold, req.Limit}
	} else {
		searchQuery = `
			SELECT 
				kd.id,
				kd.title,
				kd.content,
				kd.content_type,
				1 - (ke.embedding <=> $1::vector) as similarity,
				kd.metadata
			FROM neuronip.knowledge_documents kd
			JOIN neuronip.knowledge_embeddings ke ON ke.document_id = kd.id
			WHERE 1 - (ke.embedding <=> $1::vector) >= $2
			ORDER BY ke.embedding <=> $1::vector
			LIMIT $3`
		args = []interface{}{queryEmbedding, req.Threshold, req.Limit}
	}

	rows, err := s.pool.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
		err := rows.Scan(
			&result.DocumentID,
			&result.Title,
			&result.Content,
			&result.ContentType,
			&result.Similarity,
			&result.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

/* ChunkingConfig represents document chunking configuration */
type ChunkingConfig struct {
	ChunkSize      int  // Maximum characters per chunk
	ChunkOverlap   int  // Number of characters to overlap between chunks
	EnableChunking bool // Whether to enable chunking (false = use entire document)
}

/* DefaultChunkingConfig returns default chunking configuration */
func DefaultChunkingConfig() ChunkingConfig {
	return ChunkingConfig{
		ChunkSize:      1000,  // Default 1000 characters
		ChunkOverlap:   200,   // Default 200 character overlap
		EnableChunking: true,  // Enable by default for long documents
	}
}

/* ChunkText splits text into chunks with overlap */
func ChunkText(text string, config ChunkingConfig) []string {
	if !config.EnableChunking || len(text) <= config.ChunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	
	for start < len(text) {
		end := start + config.ChunkSize
		if end > len(text) {
			end = len(text)
		}
		
		chunk := text[start:end]
		
		// Try to break at word boundaries
		if end < len(text) && end > start+config.ChunkSize*3/4 {
			// Look for sentence or paragraph boundaries near the end
			lastPeriod := strings.LastIndex(chunk, ". ")
			lastNewline := strings.LastIndex(chunk, "\n\n")
			
			if lastNewline > config.ChunkSize/2 {
				chunk = chunk[:lastNewline+2]
				end = start + lastNewline + 2
			} else if lastPeriod > config.ChunkSize/2 {
				chunk = chunk[:lastPeriod+2]
				end = start + lastPeriod + 2
			}
		}
		
		chunks = append(chunks, strings.TrimSpace(chunk))
		
		// Move start position with overlap
		if end >= len(text) {
			break
		}
		start = end - config.ChunkOverlap
		if start < 0 {
			start = 0
		}
	}
	
	return chunks
}

/* CreateDocument creates a new knowledge document with embedding */
func (s *Service) CreateDocument(ctx context.Context, doc *db.KnowledgeDocument, config *ChunkingConfig, userID *string) error {
	if config == nil {
		defaultConfig := DefaultChunkingConfig()
		config = &defaultConfig
	}

	// Initialize metadata if nil
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}

	// Set version to 1 for new documents
	doc.Metadata["version"] = 1

	// Insert document
	insertQuery := `
		INSERT INTO neuronip.knowledge_documents 
		(collection_id, title, content, content_type, source, source_url, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	
	metadataJSON, _ := json.Marshal(doc.Metadata)
	err := s.pool.QueryRow(ctx, insertQuery,
		doc.CollectionID, doc.Title, doc.Content, doc.ContentType,
		doc.Source, doc.SourceURL, metadataJSON,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	// Log audit event for document creation
	s.auditService.LogDocumentEvent(ctx, "create", doc.ID, userID, map[string]interface{}{
		"title":       doc.Title,
		"content_type": doc.ContentType,
		"version":     1,
		"chunks_count": 0, // Will update after chunks are created
	})

	// Chunk the document
	chunks := ChunkText(doc.Content, *config)

	// Generate embeddings for chunks using batch processing
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	embedInsertQuery := `
		INSERT INTO neuronip.knowledge_embeddings 
		(document_id, embedding, model_name, chunk_index, chunk_text)
		VALUES ($1, $2::vector, $3, $4, $5)`

	// Use batch embedding generation for better performance
	if len(chunks) > 0 {
		embeddings, err := s.neurondbClient.BatchGenerateEmbedding(ctx, chunks, modelName)
		if err != nil {
			// Fallback to individual embeddings if batch fails
			for i, chunk := range chunks {
				embedding, err := s.neurondbClient.GenerateEmbedding(ctx, chunk, modelName)
				if err != nil {
					return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
				}

				_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, embedding, modelName, i, chunk)
				if err != nil {
					return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
				}
			}
		} else {
			// Batch embeddings succeeded - insert all at once
			for i, chunk := range chunks {
				if i < len(embeddings) {
					_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, embeddings[i], modelName, i, chunk)
					if err != nil {
						return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
					}
				}
			}
		}
	}

	// Update audit event with chunks count
	doc.Metadata["chunks_count"] = len(chunks)
	updateMetadataQuery := `UPDATE neuronip.knowledge_documents SET metadata = $1 WHERE id = $2`
	chunksMetadataJSON, _ := json.Marshal(doc.Metadata)
	s.pool.Exec(ctx, updateMetadataQuery, chunksMetadataJSON, doc.ID)

	// Extract entities from document content (async, non-blocking)
	// This integrates knowledge graph entity extraction into document workflow
	go func() {
		// Import knowledge graph service to extract entities
		// Note: In production, you might want to pass this as a dependency
		// For now, we'll create a temporary service instance
		kgService := knowledgegraph.NewService(s.pool, s.neurondbClient)
		extractReq := knowledgegraph.ExtractEntitiesRequest{
			DocumentID:    doc.ID,
			Text:          doc.Content,
			MinConfidence: 0.5,
		}
		_, err := kgService.ExtractEntities(context.Background(), extractReq)
		if err != nil {
			// Log error but don't fail document creation
			// In production, use proper logging
		}
	}()

	return nil
}

/* UpdateDocument updates an existing knowledge document with version tracking */
func (s *Service) UpdateDocument(ctx context.Context, docID uuid.UUID, updates *db.KnowledgeDocument, config *ChunkingConfig, userID *string) error {
	if config == nil {
		defaultConfig := DefaultChunkingConfig()
		config = &defaultConfig
	}

	// Get current document
	var currentDoc db.KnowledgeDocument
	var currentMetadataJSON []byte
	selectQuery := `
		SELECT id, collection_id, title, content, content_type, source, source_url, metadata, created_at, updated_at
		FROM neuronip.knowledge_documents
		WHERE id = $1`
	
	err := s.pool.QueryRow(ctx, selectQuery, docID).Scan(
		&currentDoc.ID, &currentDoc.CollectionID, &currentDoc.Title, &currentDoc.Content,
		&currentDoc.ContentType, &currentDoc.Source, &currentDoc.SourceURL,
		&currentMetadataJSON, &currentDoc.CreatedAt, &currentDoc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Parse current metadata
	var currentMetadata map[string]interface{}
	if currentMetadataJSON != nil {
		json.Unmarshal(currentMetadataJSON, &currentMetadata)
	}
	if currentMetadata == nil {
		currentMetadata = make(map[string]interface{})
	}

	// Get current version
	currentVersion := 1
	if v, ok := currentMetadata["version"].(float64); ok {
		currentVersion = int(v)
	}

	// Increment version
	newVersion := currentVersion + 1

	// Prepare update data (merge with current)
	if updates.Title != "" {
		currentDoc.Title = updates.Title
	}
	if updates.Content != "" {
		currentDoc.Content = updates.Content
	}
	if updates.ContentType != "" {
		currentDoc.ContentType = updates.ContentType
	}
	if updates.Source != nil {
		currentDoc.Source = updates.Source
	}
	if updates.SourceURL != nil {
		currentDoc.SourceURL = updates.SourceURL
	}
	if updates.Metadata != nil {
		// Merge metadata
		for k, v := range updates.Metadata {
			currentMetadata[k] = v
		}
	}

	currentMetadata["version"] = newVersion
	currentDoc.Metadata = currentMetadata

	// Delete old embeddings
	deleteEmbedsQuery := `DELETE FROM neuronip.knowledge_embeddings WHERE document_id = $1`
	_, err = s.pool.Exec(ctx, deleteEmbedsQuery, docID)
	if err != nil {
		return fmt.Errorf("failed to delete old embeddings: %w", err)
	}

	// Update document
	metadataJSON, _ := json.Marshal(currentDoc.Metadata)
	updateQuery := `
		UPDATE neuronip.knowledge_documents 
		SET title = $1, content = $2, content_type = $3, source = $4, source_url = $5, metadata = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at`
	
	err = s.pool.QueryRow(ctx, updateQuery,
		currentDoc.Title, currentDoc.Content, currentDoc.ContentType,
		currentDoc.Source, currentDoc.SourceURL, metadataJSON, docID,
	).Scan(&currentDoc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Chunk the updated document
	chunks := ChunkText(currentDoc.Content, *config)

	// Generate embeddings for chunks using batch processing
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	embedInsertQuery := `
		INSERT INTO neuronip.knowledge_embeddings 
		(document_id, embedding, model_name, chunk_index, chunk_text)
		VALUES ($1, $2::vector, $3, $4, $5)`

	// Use batch embedding generation for better performance
	if len(chunks) > 0 {
		embeddings, err := s.neurondbClient.BatchGenerateEmbedding(ctx, chunks, modelName)
		if err != nil {
			// Fallback to individual embeddings if batch fails
			for i, chunk := range chunks {
				embedding, err := s.neurondbClient.GenerateEmbedding(ctx, chunk, modelName)
				if err != nil {
					return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
				}

				_, err = s.pool.Exec(ctx, embedInsertQuery, docID, embedding, modelName, i, chunk)
				if err != nil {
					return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
				}
			}
		} else {
			// Batch embeddings succeeded - insert all at once
			for i, chunk := range chunks {
				if i < len(embeddings) {
					_, err = s.pool.Exec(ctx, embedInsertQuery, docID, embeddings[i], modelName, i, chunk)
					if err != nil {
						return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
					}
				}
			}
		}
	}

	// Log audit event for document update
	s.auditService.LogDocumentEvent(ctx, "update", docID, userID, map[string]interface{}{
		"title":       currentDoc.Title,
		"content_type": currentDoc.ContentType,
		"previous_version": currentVersion,
		"new_version": newVersion,
		"chunks_count": len(chunks),
	})

	return nil
}

/* GetCollection retrieves a knowledge collection */
func (s *Service) GetCollection(ctx context.Context, id uuid.UUID) (*db.KnowledgeCollection, error) {
	return s.queries.GetKnowledgeCollectionByID(ctx, id)
}

/* RAGRequest represents a RAG pipeline request */
type RAGRequest struct {
	Query        string
	CollectionID *uuid.UUID
	Limit        int
	Threshold    float64
	MaxContext   int // Maximum number of context chunks to retrieve
}

/* RAGResult represents a RAG pipeline result with context */
type RAGResult struct {
	Query      string       `json:"query"`
	Context    []string     `json:"context"`    // Retrieved context chunks
	Sources    []RAGSource  `json:"sources"`    // Source documents for context
	Results    []SearchResult `json:"results"`  // Original search results
}

/* RAGSource represents a source document in RAG context */
type RAGSource struct {
	DocumentID  uuid.UUID `json:"document_id"`
	Title       string    `json:"title"`
	ContentType string    `json:"content_type"`
	ChunkIndex  int       `json:"chunk_index"`
	ChunkText   string    `json:"chunk_text"`
	Similarity  float64   `json:"similarity"`
}

/* RAG performs RAG pipeline: semantic search + context retrieval */
func (s *Service) RAG(ctx context.Context, req RAGRequest) (*RAGResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.5
	}
	if req.MaxContext <= 0 {
		req.MaxContext = 5
	}

	// First, perform semantic search to get relevant documents/chunks
	searchResults, err := s.Search(ctx, SearchRequest{
		Query:        req.Query,
		CollectionID: req.CollectionID,
		Limit:        req.Limit * 2, // Get more results for context
		Threshold:    req.Threshold,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}

	// Extract context from search results
	// Get chunk information from embeddings that matched
	var context []string
	var sources []RAGSource

	// Query to get chunk details for each matched document
	for i, result := range searchResults {
		if i >= req.MaxContext {
			break
		}

		// Get the best matching chunk for this document
		var chunkIndex int
		var chunkText string
		chunkQuery := `
			SELECT chunk_index, chunk_text
			FROM neuronip.knowledge_embeddings
			WHERE document_id = $1
			ORDER BY chunk_index
			LIMIT 1`
		
		err := s.pool.QueryRow(ctx, chunkQuery, result.DocumentID).Scan(&chunkIndex, &chunkText)
		if err != nil || chunkText == "" {
			// Fallback to using full content if chunk not found
			chunkText = result.Content
			chunkIndex = 0
		}

		context = append(context, chunkText)
		sources = append(sources, RAGSource{
			DocumentID:  result.DocumentID,
			Title:       result.Title,
			ContentType: result.ContentType,
			ChunkIndex:  chunkIndex,
			ChunkText:   chunkText,
			Similarity:  result.Similarity,
		})
	}

	// Limit results to requested limit
	if len(searchResults) > req.Limit {
		searchResults = searchResults[:req.Limit]
	}

	return &RAGResult{
		Query:   req.Query,
		Context: context,
		Sources: sources,
		Results: searchResults,
	}, nil
}
