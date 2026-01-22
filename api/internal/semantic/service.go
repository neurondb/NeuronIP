package semantic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides semantic search functionality */
type Service struct {
	queries        *db.Queries
	pool           *pgxpool.Pool
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
	auditService   *compliance.AuditService
}

/* NewService creates a new semantic search service */
func NewService(queries *db.Queries, pool *pgxpool.Pool, neurondbClient *neurondb.Client, mcpClient *mcp.Client) *Service {
	return &Service{
		queries:        queries,
		pool:           pool,
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
		auditService:   compliance.NewAuditService(pool),
	}
}

/* SearchRequest represents a semantic search request */
type SearchRequest struct {
	Query        string
	CollectionID *uuid.UUID
	Limit        int
	Threshold    float64
	DistanceMetric string // "cosine" (default), "l2", "inner_product"
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
	if req.DistanceMetric == "" {
		req.DistanceMetric = "cosine"
	}

	// Generate embedding for the query using NeuronDB
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, req.Query, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Use appropriate search method based on distance metric
	switch req.DistanceMetric {
	case "l2":
		return s.searchWithL2(ctx, req, queryEmbedding)
	case "inner_product":
		return s.searchWithInnerProduct(ctx, req, queryEmbedding)
	default: // cosine (default)
		return s.searchWithCosine(ctx, req, queryEmbedding)
	}
}

/* searchWithCosine performs cosine similarity search */
func (s *Service) searchWithCosine(ctx context.Context, req SearchRequest, queryEmbedding string) ([]SearchResult, error) {
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
		return nil, fmt.Errorf("failed to perform cosine search: %w", err)
	}
	defer rows.Close()

	return s.scanSearchResults(rows)
}

/* searchWithL2 performs L2 distance search */
func (s *Service) searchWithL2(ctx context.Context, req SearchRequest, queryEmbedding string) ([]SearchResult, error) {
	// Try MCP first if available
	if s.mcpClient != nil {
		// Convert embedding string to float64 slice (simplified - in production, parse properly)
		// For now, use NeuronDB directly
	}

	// Use NeuronDB VectorSearchL2 method
	results, err := s.neurondbClient.VectorSearchL2(ctx, queryEmbedding, 
		"neuronip.knowledge_embeddings", "embedding", req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform L2 search: %w", err)
	}

	// Filter by collection if specified and apply threshold
	return s.filterAndConvertResults(ctx, results, req.CollectionID, req.Threshold, "distance")
}

/* SearchWithMCPVectorTools performs vector search using MCP tools */
func (s *Service) SearchWithMCPVectorTools(ctx context.Context, req SearchRequest, metric string) ([]SearchResult, error) {
	if s.mcpClient == nil {
		// Fallback to NeuronDB methods
		return s.Search(ctx, req)
	}

	if metric == "" {
		metric = "cosine"
	}

	// Generate embedding for the query
	queryEmbeddingStr, err := s.neurondbClient.GenerateEmbedding(ctx, req.Query, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Parse embedding string to float64 slice for MCP tools
	queryVector, err := parseEmbeddingString(queryEmbeddingStr)
	if err != nil {
		// Fallback to NeuronDB if parsing fails
		return s.Search(ctx, req)
	}

	// Build filters for collection if specified
	filters := make(map[string]interface{})
	if req.CollectionID != nil {
		filters["collection_id"] = req.CollectionID.String()
	}

	// Use appropriate MCP tool based on metric
	var mcpResult map[string]interface{}
	var mcpErr error

	switch metric {
	case "l2":
		mcpResult, mcpErr = s.mcpClient.VectorSearchL2(ctx, queryVector, "neuronip.knowledge_embeddings", "embedding", req.Limit, filters)
	case "inner_product":
		mcpResult, mcpErr = s.mcpClient.VectorSearchInnerProduct(ctx, queryVector, "neuronip.knowledge_embeddings", "embedding", req.Limit, filters)
	default: // cosine
		mcpResult, mcpErr = s.mcpClient.VectorSearchCosine(ctx, queryVector, "neuronip.knowledge_embeddings", "embedding", req.Limit, filters)
	}

	if mcpErr != nil {
		// Fallback to NeuronDB methods
		return s.Search(ctx, req)
	}

	// Extract results from MCP response
	var results []map[string]interface{}
	if docs, ok := mcpResult["documents"].([]interface{}); ok {
		for _, doc := range docs {
			if docMap, ok := doc.(map[string]interface{}); ok {
				results = append(results, docMap)
			}
		}
	} else if mcpResult != nil {
		// Single result or different format
		results = []map[string]interface{}{mcpResult}
	}

	// Convert and filter results
	return s.filterAndConvertResults(ctx, results, req.CollectionID, req.Threshold, "similarity")
}

/* parseEmbeddingString parses embedding string to float64 slice */
func parseEmbeddingString(embeddingStr string) ([]float64, error) {
	if embeddingStr == "" {
		return nil, fmt.Errorf("empty embedding string")
	}

	// Remove brackets if present
	embeddingStr = strings.TrimSpace(embeddingStr)
	embeddingStr = strings.TrimPrefix(embeddingStr, "[")
	embeddingStr = strings.TrimSuffix(embeddingStr, "]")

	// Split by comma
	parts := strings.Split(embeddingStr, ",")
	embedding := make([]float64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedding value '%s': %w", part, err)
		}
		embedding = append(embedding, val)
	}

	return embedding, nil
}

/* CalculateVectorSimilarity calculates similarity between two embeddings using MCP */
func (s *Service) CalculateVectorSimilarity(ctx context.Context, embedding1Str, embedding2Str string, metric string) (float64, error) {
	if s.mcpClient == nil {
		// Fallback to direct SQL comparison
		return s.CompareDocuments(ctx, uuid.Nil, uuid.Nil, metric)
	}

	if metric == "" {
		metric = "cosine"
	}

	// Parse embeddings
	vec1, err := parseEmbeddingString(embedding1Str)
	if err != nil {
		return 0, fmt.Errorf("failed to parse embedding 1: %w", err)
	}

	vec2, err := parseEmbeddingString(embedding2Str)
	if err != nil {
		return 0, fmt.Errorf("failed to parse embedding 2: %w", err)
	}

	// Use MCP VectorSimilarity
	result, err := s.mcpClient.VectorSimilarity(ctx, vec1, vec2, metric)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate similarity: %w", err)
	}

	// Extract similarity score
	if similarity, ok := result["similarity"].(float64); ok {
		return similarity, nil
	}
	if distance, ok := result["distance"].(float64); ok {
		// Convert distance to similarity
		return 1.0 / (1.0 + distance), nil
	}

	return 0, fmt.Errorf("similarity score not found in MCP result")
}

/* parseEmbeddingString parses an embedding string to float64 slice */
func parseEmbeddingString(embeddingStr string) ([]float64, error) {
	// Remove brackets and whitespace
	embeddingStr = strings.TrimSpace(embeddingStr)
	embeddingStr = strings.TrimPrefix(embeddingStr, "[")
	embeddingStr = strings.TrimSuffix(embeddingStr, "]")
	
	// Split by comma
	parts := strings.Split(embeddingStr, ",")
	result := make([]float64, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedding value '%s': %w", part, err)
		}
		result = append(result, val)
	}
	
	return result, nil
}

/* PerformVectorArithmetic performs arithmetic operations on vectors using MCP */
func (s *Service) PerformVectorArithmetic(ctx context.Context, embedding1Str, embedding2Str string, operation string) ([]float64, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}

	if operation == "" {
		operation = "add"
	}

	// Parse embeddings
	vec1, err := parseEmbeddingString(embedding1Str)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding 1: %w", err)
	}

	vec2, err := parseEmbeddingString(embedding2Str)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding 2: %w", err)
	}

	// Use MCP VectorArithmetic
	result, err := s.mcpClient.VectorArithmetic(ctx, vec1, vec2, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector arithmetic: %w", err)
	}

	// Extract result vector
	if resultVec, ok := result["result"].([]interface{}); ok {
		output := make([]float64, len(resultVec))
		for i, v := range resultVec {
			if f, ok := v.(float64); ok {
				output[i] = f
			}
		}
		return output, nil
	}

	return nil, fmt.Errorf("result vector not found in MCP result")
}

/* searchWithInnerProduct performs inner product search */
func (s *Service) searchWithInnerProduct(ctx context.Context, req SearchRequest, queryEmbedding string) ([]SearchResult, error) {
	// Use NeuronDB VectorSearchInnerProduct method
	results, err := s.neurondbClient.VectorSearchInnerProduct(ctx, queryEmbedding,
		"neuronip.knowledge_embeddings", "embedding", req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform inner product search: %w", err)
	}

	// Filter by collection if specified and apply threshold
	return s.filterAndConvertResults(ctx, results, req.CollectionID, req.Threshold, "distance")
}

/* filterAndConvertResults filters results by collection and threshold, then converts to SearchResult */
func (s *Service) filterAndConvertResults(ctx context.Context, results []map[string]interface{}, 
	collectionID *uuid.UUID, threshold float64, scoreKey string) ([]SearchResult, error) {
	var searchResults []SearchResult

	for _, result := range results {
		// Get document ID from result
		var docID uuid.UUID
		if id, ok := result["document_id"].(uuid.UUID); ok {
			docID = id
		} else if id, ok := result["id"].(uuid.UUID); ok {
			docID = id
		} else {
			continue
		}

		// Get similarity/distance score
		similarity := 0.0
		if sim, ok := result["similarity"].(float64); ok {
			similarity = sim
		} else if dist, ok := result[scoreKey].(float64); ok {
			// Convert distance to similarity (for L2, lower distance = higher similarity)
			if scoreKey == "distance" {
				similarity = 1.0 / (1.0 + dist) // Normalize distance to similarity
			} else {
				similarity = dist
			}
		}

		// Apply threshold
		if similarity < threshold {
			continue
		}

		// Get document details
		var title, content, contentType string
		var metadata map[string]interface{}

		// Try to get from result first
		if t, ok := result["title"].(string); ok {
			title = t
		}
		if c, ok := result["content"].(string); ok {
			content = c
		}
		if ct, ok := result["content_type"].(string); ok {
			contentType = ct
		}
		if m, ok := result["metadata"].(map[string]interface{}); ok {
			metadata = m
		}

		// If not in result, query database
		if title == "" || content == "" {
			docQuery := `
				SELECT title, content, content_type, metadata
				FROM neuronip.knowledge_documents
				WHERE id = $1`
			if collectionID != nil {
				docQuery += ` AND collection_id = $2`
				err := s.pool.QueryRow(ctx, docQuery, docID, collectionID).Scan(
					&title, &content, &contentType, &metadata)
				if err != nil {
					continue
				}
			} else {
				err := s.pool.QueryRow(ctx, docQuery, docID).Scan(
					&title, &content, &contentType, &metadata)
				if err != nil {
					continue
				}
			}
		}

		searchResults = append(searchResults, SearchResult{
			DocumentID:  docID,
			Title:       title,
			Content:     content,
			ContentType: contentType,
			Similarity:  similarity,
			Metadata:    metadata,
		})
	}

	return searchResults, nil
}

/* scanSearchResults scans rows into SearchResult slice */
func (s *Service) scanSearchResults(rows pgx.Rows) ([]SearchResult, error) {
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

/* HybridSearch performs hybrid semantic + keyword search using NeuronDB */
func (s *Service) HybridSearch(ctx context.Context, req SearchRequest, keywordQuery string) ([]SearchResult, error) {
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

	// Determine table and column names based on collection
	tableName := "neuronip.knowledge_documents kd JOIN neuronip.knowledge_embeddings ke ON ke.document_id = kd.id"
	embeddingColumn := "ke.embedding"
	textColumn := "kd.content"

	var weights map[string]float64
	if keywordQuery != "" {
		weights = map[string]float64{
			"semantic": 0.7,
			"keyword":  0.3,
		}
	}

	// Use NeuronDB HybridSearch
	results, err := s.neurondbClient.HybridSearch(ctx, queryEmbedding, keywordQuery, tableName, embeddingColumn, textColumn, req.Limit, weights)
	if err != nil {
		// Fallback to regular semantic search if hybrid fails
		return s.Search(ctx, req)
	}

	// Convert results to SearchResult format
	var searchResults []SearchResult
	for _, result := range results {
		if docID, ok := result["id"].(uuid.UUID); ok {
			similarity := 0.0
			if sim, ok := result["combined_score"].(float64); ok {
				similarity = sim
			} else if sim, ok := result["similarity"].(float64); ok {
				similarity = sim
			}

			searchResults = append(searchResults, SearchResult{
				DocumentID: docID,
				Title:      getStringValue(result, "title"),
				Content:    getStringValue(result, "content"),
				Similarity: similarity,
			})
		}
	}

	return searchResults, nil
}

/* CompareDocuments compares two documents using vector similarity */
func (s *Service) CompareDocuments(ctx context.Context, docID1, docID2 uuid.UUID, metric string) (float64, error) {
	if metric == "" {
		metric = "cosine"
	}

	// Get embeddings for both documents
	var embedding1, embedding2 string
	query := `
		SELECT embedding::text
		FROM neuronip.knowledge_embeddings
		WHERE document_id = $1
		ORDER BY chunk_index
		LIMIT 1`
	
	err := s.pool.QueryRow(ctx, query, docID1).Scan(&embedding1)
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding for document 1: %w", err)
	}

	err = s.pool.QueryRow(ctx, query, docID2).Scan(&embedding2)
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding for document 2: %w", err)
	}

	// Try MCP VectorSimilarity first if available
	if s.mcpClient != nil {
		similarity, err := s.CalculateVectorSimilarity(ctx, embedding1, embedding2, metric)
		if err == nil {
			return similarity, nil
		}
		// Fall through to SQL if MCP fails
	}

	// Fallback to direct SQL comparison
	var similarity float64
	var similarityQuery string
	switch metric {
	case "l2":
		similarityQuery = `SELECT 1.0 / (1.0 + ($1::vector <-> $2::vector)) as similarity`
	case "inner_product":
		similarityQuery = `SELECT ($1::vector <#> $2::vector) * -1 as similarity`
	default: // cosine
		similarityQuery = `SELECT 1 - ($1::vector <=> $2::vector) as similarity`
	}
	
	err = s.pool.QueryRow(ctx, similarityQuery, embedding1, embedding2).Scan(&similarity)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate similarity: %w", err)
	}
	return similarity, nil
}

/* ensureVectorIndex ensures a vector index exists on the specified table and column */
func (s *Service) ensureVectorIndex(ctx context.Context, tableName string, columnName string) error {
	// Check if index already exists
	indexName := fmt.Sprintf("idx_%s_%s_vector", tableName, columnName)
	checkQuery := `
		SELECT COUNT(*) FROM pg_indexes 
		WHERE indexname = $1 AND tablename = $2`
	
	var count int
	err := s.pool.QueryRow(ctx, checkQuery, indexName, tableName).Scan(&count)
	if err == nil && count > 0 {
		return nil // Index already exists
	}

	// Get row count to determine index parameters
	var rowCount int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM neuronip.%s WHERE %s IS NOT NULL`, tableName, columnName)
	err = s.pool.QueryRow(ctx, countQuery).Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("failed to get row count: %w", err)
	}

	// Choose index type and parameters based on data size
	var indexType string
	var options map[string]interface{}
	
	if rowCount < 10000 {
		// Small dataset - use IVF
		indexType = "ivf"
		options = map[string]interface{}{
			"lists": 100,
		}
	} else {
		// Large dataset - use HNSW
		indexType = "hnsw"
		options = map[string]interface{}{
			"m":              16,
			"ef_construction": 64,
		}
	}

	// Try MCP first if available
	if s.mcpClient != nil {
		fullTableName := fmt.Sprintf("neuronip.%s", tableName)
		var result map[string]interface{}
		var mcpErr error

		if indexType == "hnsw" {
			result, mcpErr = s.mcpClient.CreateHNSWIndex(ctx, indexName, fullTableName, columnName, options)
		} else {
			result, mcpErr = s.mcpClient.CreateIVFIndex(ctx, indexName, fullTableName, columnName, options)
		}

		if mcpErr == nil && result != nil {
			return nil // Successfully created via MCP
		}
	}

	// Fallback to NeuronDB client
	fullTableName := fmt.Sprintf("neuronip.%s", tableName)
	err = s.neurondbClient.CreateVectorIndex(ctx, indexName, fullTableName, columnName, indexType, options)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}

	return nil
}

/* GetIndexStatus gets the status of a vector index using MCP */
func (s *Service) GetIndexStatus(ctx context.Context, indexName string) (map[string]interface{}, error) {
	if s.mcpClient != nil {
		return s.mcpClient.IndexStatus(ctx, indexName)
	}
	return nil, fmt.Errorf("MCP client not configured")
}

/* TuneVectorIndex tunes a vector index using MCP */
func (s *Service) TuneVectorIndex(ctx context.Context, indexName string, tableName string, columnName string, indexType string) error {
	if s.mcpClient == nil {
		return fmt.Errorf("MCP client not configured")
	}

	var options map[string]interface{}
	if indexType == "hnsw" {
		options = map[string]interface{}{
			"m":              32, // Increase for better recall
			"ef_construction": 128, // Increase for better quality
		}
		_, err := s.mcpClient.CreateHNSWIndex(ctx, indexName, tableName, columnName, options)
		return err
	} else if indexType == "ivf" {
		options = map[string]interface{}{
			"lists": 200, // Increase for larger datasets
		}
		_, err := s.mcpClient.CreateIVFIndex(ctx, indexName, tableName, columnName, options)
		return err
	}

	return fmt.Errorf("unsupported index type: %s", indexType)
}

/* GenerateEmbeddingWithMCP generates embedding using MCP tools with fallback */
func (s *Service) GenerateEmbeddingWithMCP(ctx context.Context, text string, model string, useCache bool) (string, error) {
	if s.mcpClient == nil {
		// Fallback to NeuronDB
		return s.neurondbClient.GenerateEmbedding(ctx, text, model)
	}

	if model == "" {
		model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	var result map[string]interface{}
	var err error

	if useCache {
		// Use cached embedding
		result, err = s.mcpClient.EmbedCached(ctx, text, model)
	} else {
		// Use regular embedding (via batch_embedding with single text)
		texts := []string{text}
		result, err = s.mcpClient.BatchEmbedding(ctx, texts, model, nil)
	}

	if err != nil {
		// Fallback to NeuronDB
		return s.neurondbClient.GenerateEmbedding(ctx, text, model)
	}

	// Extract embedding from result
	if embeddings, ok := result["embeddings"].([]interface{}); ok && len(embeddings) > 0 {
		if emb, ok := embeddings[0].(string); ok {
			return emb, nil
		}
		if emb, ok := embeddings[0].([]float64); ok {
			// Convert float64 slice to string format
			return formatEmbeddingToString(emb), nil
		}
	}

	if embedding, ok := result["embedding"].(string); ok {
		return embedding, nil
	}

	// Fallback to NeuronDB if extraction fails
	return s.neurondbClient.GenerateEmbedding(ctx, text, model)
}

/* ConfigureEmbeddingModel configures embedding model using MCP */
func (s *Service) ConfigureEmbeddingModel(ctx context.Context, modelName string, config map[string]interface{}) error {
	if s.mcpClient == nil {
		return fmt.Errorf("MCP client not configured")
	}

	_, err := s.mcpClient.ConfigureEmbeddingModel(ctx, modelName, config)
	return err
}

/* formatEmbeddingToString formats float64 slice to PostgreSQL vector string */
func formatEmbeddingToString(embedding []float64) string {
	if len(embedding) == 0 {
		return "[]"
	}

	parts := make([]string, len(embedding))
	for i, val := range embedding {
		parts[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}

	return "[" + strings.Join(parts, ",") + "]"
}

/* getStringValue safely extracts string value from map */
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

/* AnswerWithRAG generates an answer with citations using RAG pipeline */
func (s *Service) AnswerWithRAG(ctx context.Context, query string, collectionID *uuid.UUID, model string) (map[string]interface{}, error) {
	if model == "" {
		model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	// First, retrieve relevant context using semantic search
	searchReq := SearchRequest{
		Query:        query,
		CollectionID: collectionID,
		Limit:        5,
		Threshold:    0.5,
	}
	
	searchResults, err := s.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve context: %w", err)
	}

	// Build context array for RAG
	context := make([]map[string]interface{}, 0, len(searchResults))
	for _, result := range searchResults {
		context = append(context, map[string]interface{}{
			"content":  result.Content,
			"title":    result.Title,
			"document_id": result.DocumentID.String(),
		})
	}

	// Use NeuronDB AnswerWithCitations for RAG response
	answer, err := s.neurondbClient.AnswerWithCitations(ctx, query, context, model)
	if err != nil {
		// Fallback to GenerateResponse if AnswerWithCitations not available
		contextTexts := make([]string, 0, len(searchResults))
		for _, result := range searchResults {
			contextTexts = append(contextTexts, result.Content)
		}
		response, err := s.neurondbClient.GenerateResponse(ctx, query, contextTexts, model)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RAG response: %w", err)
		}
		return map[string]interface{}{
			"answer":    response,
			"context":   context,
			"citations": []string{},
		}, nil
	}

	return answer, nil
}

/* RerankSearchResults reranks search results using MCP reranking tools */
func (s *Service) RerankSearchResults(ctx context.Context, query string, documents []string, topK int) ([]map[string]interface{}, error) {
	return s.RerankSearchResultsWithMethod(ctx, query, documents, topK, "cross_encoder")
}

/* RerankSearchResultsWithMethod reranks search results using specified MCP reranking method */
func (s *Service) RerankSearchResultsWithMethod(ctx context.Context, query string, documents []string, topK int, method string) ([]map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}
	if topK <= 0 {
		topK = 10
	}
	if method == "" {
		method = "cross_encoder"
	}

	var result map[string]interface{}
	var err error

	// Use appropriate reranking method
	switch method {
	case "llm":
		result, err = s.mcpClient.RerankLLM(ctx, query, documents, topK, "")
	case "cohere":
		result, err = s.mcpClient.RerankCohere(ctx, query, documents, topK)
	case "ensemble":
		// Use ensemble reranking with multiple methods
		methods := []string{"cross_encoder", "llm"}
		weights := map[string]float64{
			"cross_encoder": 0.7,
			"llm":           0.3,
		}
		result, err = s.mcpClient.RerankEnsemble(ctx, query, documents, topK, methods, weights)
	default: // cross_encoder
		result, err = s.mcpClient.RerankCrossEncoder(ctx, query, documents, topK)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to rerank results with method %s: %w", method, err)
	}

	// Extract reranked documents from result
	if rerankedDocs, ok := result["documents"].([]interface{}); ok {
		results := make([]map[string]interface{}, 0, len(rerankedDocs))
		for _, doc := range rerankedDocs {
			if docMap, ok := doc.(map[string]interface{}); ok {
				results = append(results, docMap)
			}
		}
		return results, nil
	}

	return []map[string]interface{}{result}, nil
}

/* RerankWithReciprocalRankFusion performs reciprocal rank fusion on multiple result sets */
func (s *Service) RerankWithReciprocalRankFusion(ctx context.Context, resultSets [][]map[string]interface{}, k int) ([]map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}
	if k <= 0 {
		k = 60 // Default RRF constant
	}

	result, err := s.mcpClient.ReciprocalRankFusion(ctx, resultSets, k)
	if err != nil {
		return nil, fmt.Errorf("failed to perform reciprocal rank fusion: %w", err)
	}

	// Extract fused results
	if fusedDocs, ok := result["documents"].([]interface{}); ok {
		results := make([]map[string]interface{}, 0, len(fusedDocs))
		for _, doc := range fusedDocs {
			if docMap, ok := doc.(map[string]interface{}); ok {
				results = append(results, docMap)
			}
		}
		return results, nil
	}

	return []map[string]interface{}{result}, nil
}

/* GenerateRAGResponse generates a response using RAG pipeline */
func (s *Service) GenerateRAGResponse(ctx context.Context, query string, context []string, model string) (string, error) {
	if model == "" {
		model = "sentence-transformers/all-MiniLM-L6-v2"
	}
	return s.neurondbClient.GenerateResponse(ctx, query, context, model)
}

/* CreateDocumentWithImage creates a document with image support using multimodal embeddings */
func (s *Service) CreateDocumentWithImage(ctx context.Context, doc *db.KnowledgeDocument, imageData []byte, config *ChunkingConfig, userID *string) error {
	if config == nil {
		defaultConfig := DefaultChunkingConfig()
		config = &defaultConfig
	}

	// Initialize metadata if nil
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}
	doc.Metadata["version"] = 1
	doc.Metadata["has_image"] = true

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

	// Generate multimodal embedding (text + image)
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	var multimodalEmbedding string
	var err error

	// Try MCP EmbedMultimodal first if available
	if s.mcpClient != nil && imageData != nil {
		// MCP expects file path, so we'll save to temp file
		// For now, fallback to NeuronDB which accepts byte data
		// In production, you might want to save to temp file and pass path
		multimodalEmbedding, err = s.neurondbClient.GenerateMultimodalEmbedding(ctx, doc.Content, imageData, modelName)
	} else {
		multimodalEmbedding, err = s.neurondbClient.GenerateMultimodalEmbedding(ctx, doc.Content, imageData, modelName)
	}

	if err != nil {
		// Fallback to text-only embedding if multimodal fails
		multimodalEmbedding, err = s.neurondbClient.GenerateEmbedding(ctx, doc.Content, modelName)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
	}

	// Store multimodal embedding
	embedInsertQuery := `
		INSERT INTO neuronip.knowledge_embeddings 
		(document_id, embedding, model_name, chunk_index, chunk_text)
		VALUES ($1, $2::vector, $3, $4, $5)`
	
	_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, multimodalEmbedding, modelName, 0, doc.Content)
	if err != nil {
		return fmt.Errorf("failed to insert multimodal embedding: %w", err)
	}

	// Log audit event
	s.auditService.LogDocumentEvent(ctx, "create", doc.ID, userID, map[string]interface{}{
		"title":       doc.Title,
		"content_type": doc.ContentType,
		"version":     1,
		"has_image":   true,
	})

	return nil
}

/* CreateDocumentWithCachedEmbedding creates a document using cached embedding via MCP */
func (s *Service) CreateDocumentWithCachedEmbedding(ctx context.Context, doc *db.KnowledgeDocument, config *ChunkingConfig, userID *string) error {
	if s.mcpClient == nil {
		// Fallback to regular document creation
		return s.CreateDocument(ctx, doc, config, userID)
	}

	if config == nil {
		defaultConfig := DefaultChunkingConfig()
		config = &defaultConfig
	}

	// Initialize metadata if nil
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}
	doc.Metadata["version"] = 1
	doc.Metadata["used_cached_embedding"] = true

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

	// Chunk the document
	chunks := ChunkText(doc.Content, *config)
	modelName := "sentence-transformers/all-MiniLM-L6-v2"

	// Use MCP cached embedding for each chunk
	embedInsertQuery := `
		INSERT INTO neuronip.knowledge_embeddings 
		(document_id, embedding, model_name, chunk_index, chunk_text)
		VALUES ($1, $2::vector, $3, $4, $5)`

	for i, chunk := range chunks {
		// Use MCP EmbedCached
		result, err := s.mcpClient.EmbedCached(ctx, chunk, modelName)
		if err != nil {
			// Fallback to regular embedding
			embedding, err := s.neurondbClient.GenerateEmbedding(ctx, chunk, modelName)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
			}
			_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, embedding, modelName, i, chunk)
			if err != nil {
				return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
			}
		} else {
			// Extract embedding from MCP result
			if emb, ok := result["embedding"].(string); ok {
				_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, emb, modelName, i, chunk)
				if err != nil {
					return fmt.Errorf("failed to insert embedding for chunk %d: %w", i, err)
				}
			}
		}
	}

	// Log audit event
	s.auditService.LogDocumentEvent(ctx, "create", doc.ID, userID, map[string]interface{}{
		"title":       doc.Title,
		"content_type": doc.ContentType,
		"version":     1,
		"chunks_count": len(chunks),
		"used_cached": true,
	})

	return nil
}

/* CreateDocumentWithImageOnly creates a document with image-only embedding */
func (s *Service) CreateDocumentWithImageOnly(ctx context.Context, doc *db.KnowledgeDocument, imageData []byte, userID *string) error {
	// Initialize metadata if nil
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}
	doc.Metadata["version"] = 1
	doc.Metadata["has_image"] = true
	doc.Metadata["image_only"] = true

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

	// Generate image-only embedding
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	var imageEmbedding string
	var err error

	// Try MCP EmbedImage first if available
	if s.mcpClient != nil && imageData != nil {
		// MCP expects file path, so we'll save to temp file
		// For now, fallback to NeuronDB which accepts byte data
		imageEmbedding, err = s.neurondbClient.GenerateImageEmbedding(ctx, imageData, modelName)
	} else {
		imageEmbedding, err = s.neurondbClient.GenerateImageEmbedding(ctx, imageData, modelName)
	}

	if err != nil {
		return fmt.Errorf("failed to generate image embedding: %w", err)
	}

	// Store image embedding
	embedInsertQuery := `
		INSERT INTO neuronip.knowledge_embeddings 
		(document_id, embedding, model_name, chunk_index, chunk_text)
		VALUES ($1, $2::vector, $3, $4, $5)`
	
	_, err = s.pool.Exec(ctx, embedInsertQuery, doc.ID, imageEmbedding, modelName, 0, "[IMAGE]")
	if err != nil {
		return fmt.Errorf("failed to insert image embedding: %w", err)
	}

	// Log audit event
	s.auditService.LogDocumentEvent(ctx, "create", doc.ID, userID, map[string]interface{}{
		"title":       doc.Title,
		"content_type": doc.ContentType,
		"version":     1,
		"has_image":   true,
		"image_only":  true,
	})

	return nil
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

		// Auto-create vector index if it doesn't exist
		if err := s.ensureVectorIndex(ctx, "knowledge_embeddings", "embedding"); err != nil {
			// Log error but don't fail document creation
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

	// Chunk the updated document using NeuronDB ProcessDocument if available
	var chunks []string
	var chunkData []map[string]interface{}
	
	if config.EnableChunking {
		// Try using NeuronDB ProcessDocument first
		processedChunks, err := s.neurondbClient.ProcessDocument(ctx, currentDoc.Content, config.ChunkSize, config.ChunkOverlap)
		if err == nil && len(processedChunks) > 0 {
			// Extract text from processed chunks
			for _, chunk := range processedChunks {
				if text, ok := chunk["text"].(string); ok {
					chunks = append(chunks, text)
				}
			}
			chunkData = processedChunks
		} else {
			// Fallback to local chunking
			chunks = ChunkText(currentDoc.Content, *config)
		}
	} else {
		chunks = []string{currentDoc.Content}
	}

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

/* TimeGrain represents a time grain */
type TimeGrain struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Duration    string    `json:"duration"`
	IsDefault   bool      `json:"is_default"`
}

/* GetTimeGrains retrieves all available time grains */
func (s *Service) GetTimeGrains(ctx context.Context) ([]TimeGrain, error) {
	query := `
		SELECT id, name, display_name, duration, is_default
		FROM neuronip.time_grains
		ORDER BY is_default DESC, name`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get time grains: %w", err)
	}
	defer rows.Close()

	var grains []TimeGrain
	for rows.Next() {
		var grain TimeGrain
		err := rows.Scan(&grain.ID, &grain.Name, &grain.DisplayName, &grain.Duration, &grain.IsDefault)
		if err != nil {
			continue
		}
		grains = append(grains, grain)
	}

	return grains, nil
}

/* AddMetricTimeGrain adds a time grain to a metric */
func (s *Service) AddMetricTimeGrain(ctx context.Context, metricID uuid.UUID, timeGrainID uuid.UUID, isDefault bool) error {
	query := `
		INSERT INTO neuronip.metric_time_grains (metric_id, time_grain_id, is_default, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (metric_id, time_grain_id) 
		DO UPDATE SET is_default = EXCLUDED.is_default`

	_, err := s.pool.Exec(ctx, query, metricID, timeGrainID, isDefault)
	if err != nil {
		return fmt.Errorf("failed to add metric time grain: %w", err)
	}

	return nil
}

/* DeleteMetricTimeGrain removes a time grain from a metric */
func (s *Service) DeleteMetricTimeGrain(ctx context.Context, metricID uuid.UUID, timeGrainID uuid.UUID) error {
	query := `
		DELETE FROM neuronip.metric_time_grains
		WHERE metric_id = $1 AND time_grain_id = $2
	`
	
	result, err := s.pool.Exec(ctx, query, metricID, timeGrainID)
	if err != nil {
		return fmt.Errorf("failed to delete metric time grain: %w", err)
	}
	
	// Check if any rows were deleted
	if result.RowsAffected() == 0 {
		return fmt.Errorf("metric time grain not found")
	}
	
	return nil
}

/* GetMetricTimeGrains retrieves time grains for a metric */
func (s *Service) GetMetricTimeGrains(ctx context.Context, metricID uuid.UUID) ([]TimeGrain, error) {
	query := `
		SELECT tg.id, tg.name, tg.display_name, tg.duration, mtg.is_default
		FROM neuronip.time_grains tg
		JOIN neuronip.metric_time_grains mtg ON tg.id = mtg.time_grain_id
		WHERE mtg.metric_id = $1
		ORDER BY mtg.is_default DESC, tg.name`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric time grains: %w", err)
	}
	defer rows.Close()

	var grains []TimeGrain
	for rows.Next() {
		var grain TimeGrain
		err := rows.Scan(&grain.ID, &grain.Name, &grain.DisplayName, &grain.Duration, &grain.IsDefault)
		if err != nil {
			continue
		}
		grains = append(grains, grain)
	}

	return grains, nil
}

/* MetricFilter represents a metric filter */
type MetricFilter struct {
	ID              uuid.UUID `json:"id"`
	MetricID        uuid.UUID `json:"metric_id"`
	FilterExpression string    `json:"filter_expression"`
	FilterName      *string   `json:"filter_name,omitempty"`
	IsDefault       bool      `json:"is_default"`
}

/* AddMetricFilter adds a default filter to a metric */
func (s *Service) AddMetricFilter(ctx context.Context, metricID uuid.UUID, filterExpression string, filterName *string, isDefault bool) error {
	id := uuid.New()

	query := `
		INSERT INTO neuronip.metric_filters (id, metric_id, filter_expression, filter_name, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`

	_, err := s.pool.Exec(ctx, query, id, metricID, filterExpression, filterName, isDefault)
	if err != nil {
		return fmt.Errorf("failed to add metric filter: %w", err)
	}

	return nil
}

/* GetMetricFilters retrieves filters for a metric */
func (s *Service) GetMetricFilters(ctx context.Context, metricID uuid.UUID) ([]MetricFilter, error) {
	query := `
		SELECT id, metric_id, filter_expression, filter_name, is_default
		FROM neuronip.metric_filters
		WHERE metric_id = $1
		ORDER BY is_default DESC, created_at`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric filters: %w", err)
	}
	defer rows.Close()

	var filters []MetricFilter
	for rows.Next() {
		var filter MetricFilter
		var filterName *string
		err := rows.Scan(&filter.ID, &filter.MetricID, &filter.FilterExpression, &filterName, &filter.IsDefault)
		if err != nil {
			continue
		}
		filter.FilterName = filterName
		filters = append(filters, filter)
	}

	return filters, nil
}

/* GetDefaultMetricFilters retrieves default filters for a metric */
func (s *Service) GetDefaultMetricFilters(ctx context.Context, metricID uuid.UUID) ([]MetricFilter, error) {
	query := `
		SELECT id, metric_id, filter_expression, filter_name, is_default
		FROM neuronip.metric_filters
		WHERE metric_id = $1 AND is_default = true
		ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get default metric filters: %w", err)
	}
	defer rows.Close()

	var filters []MetricFilter
	for rows.Next() {
		var filter MetricFilter
		var filterName *string
		err := rows.Scan(&filter.ID, &filter.MetricID, &filter.FilterExpression, &filterName, &filter.IsDefault)
		if err != nil {
			continue
		}
		filter.FilterName = filterName
		filters = append(filters, filter)
	}

	return filters, nil
}

/* LinkMetricToGlossary links a metric to a glossary term */
func (s *Service) LinkMetricToGlossary(ctx context.Context, metricID uuid.UUID, glossaryTermID uuid.UUID) error {
	query := `
		INSERT INTO neuronip.metric_glossary_links (metric_id, glossary_term_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (metric_id, glossary_term_id) DO NOTHING`

	_, err := s.pool.Exec(ctx, query, metricID, glossaryTermID)
	if err != nil {
		return fmt.Errorf("failed to link metric to glossary: %w", err)
	}

	return nil
}

/* GetMetricGlossaryTerms retrieves glossary terms linked to a metric */
func (s *Service) GetMetricGlossaryTerms(ctx context.Context, metricID uuid.UUID) ([]map[string]interface{}, error) {
	query := `
		SELECT gt.id, gt.name, gt.definition, gt.category
		FROM neuronip.glossary_terms gt
		JOIN neuronip.metric_glossary_links mgl ON gt.id = mgl.glossary_term_id
		WHERE mgl.metric_id = $1
		ORDER BY gt.name`

	rows, err := s.pool.Query(ctx, query, metricID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric glossary terms: %w", err)
	}
	defer rows.Close()

	var terms []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, definition string
		var category *string

		err := rows.Scan(&id, &name, &definition, &category)
		if err != nil {
			continue
		}

		term := map[string]interface{}{
			"id":         id,
			"name":       name,
			"definition": definition,
		}
		if category != nil {
			term["category"] = *category
		}
		terms = append(terms, term)
	}

	return terms, nil
}

/* UnlinkMetricFromGlossary removes a glossary link from a metric */
func (s *Service) UnlinkMetricFromGlossary(ctx context.Context, metricID uuid.UUID, glossaryTermID uuid.UUID) error {
	query := `
		DELETE FROM neuronip.metric_glossary_links
		WHERE metric_id = $1 AND glossary_term_id = $2`

	_, err := s.pool.Exec(ctx, query, metricID, glossaryTermID)
	if err != nil {
		return fmt.Errorf("failed to unlink metric from glossary: %w", err)
	}

	return nil
}
