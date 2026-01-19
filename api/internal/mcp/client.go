package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

/* Client provides NeuronMCP integration */
type Client struct {
	binaryPath string
}

/* NewClient creates a new NeuronMCP client */
func NewClient(binaryPath string) *Client {
	return &Client{binaryPath: binaryPath}
}

/* ExecuteTool executes an MCP tool */
func (c *Client) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	// Create context with timeout for MCP execution
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Prepare MCP request message
	request := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name": toolName,
			"arguments": args,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	// Execute MCP binary with request as stdin
	cmd := exec.CommandContext(timeoutCtx, c.binaryPath, "call", toolName)
	cmd.Stdin = bytes.NewReader(requestJSON)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute MCP tool: %w, output: %s", err, string(output))
	}

	// Parse MCP response
	var response map[string]interface{}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}

	// Extract result from MCP response
	if result, ok := response["result"].(map[string]interface{}); ok {
		return result, nil
	}

	// Return response as-is if no result field
	return response, nil
}

/* ListTools lists available MCP tools */
func (c *Client) ListTools(ctx context.Context) ([]map[string]interface{}, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, c.binaryPath, "list-tools")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP tools: %w, output: %s", err, string(output))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}

	if tools, ok := result["tools"].([]interface{}); ok {
		toolsList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				toolsList = append(toolsList, toolMap)
			}
		}
		return toolsList, nil
	}

	return []map[string]interface{}{}, nil
}

// ============================================================================
// Vector Operations Tools (8 tools)
// ============================================================================

/* VectorSearchL2 performs vector search with L2 distance */
func (c *Client) VectorSearchL2(ctx context.Context, queryVector []float64, table string, vectorColumn string, limit int, filters map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query_vector": queryVector,
		"table":        table,
		"vector_column": vectorColumn,
		"limit":        limit,
	}
	if filters != nil {
		args["filters"] = filters
	}
	return c.ExecuteTool(ctx, "vector_search_l2", args)
}

/* VectorSearchCosine performs vector search with cosine similarity */
func (c *Client) VectorSearchCosine(ctx context.Context, queryVector []float64, table string, vectorColumn string, limit int, filters map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query_vector": queryVector,
		"table":        table,
		"vector_column": vectorColumn,
		"limit":        limit,
	}
	if filters != nil {
		args["filters"] = filters
	}
	return c.ExecuteTool(ctx, "vector_search_cosine", args)
}

/* VectorSearchInnerProduct performs vector search with inner product */
func (c *Client) VectorSearchInnerProduct(ctx context.Context, queryVector []float64, table string, vectorColumn string, limit int, filters map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query_vector": queryVector,
		"table":        table,
		"vector_column": vectorColumn,
		"limit":        limit,
	}
	if filters != nil {
		args["filters"] = filters
	}
	return c.ExecuteTool(ctx, "vector_search_inner_product", args)
}

/* VectorSimilarity calculates similarity between two vectors */
func (c *Client) VectorSimilarity(ctx context.Context, vector1 []float64, vector2 []float64, metric string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"vector1": vector1,
		"vector2": vector2,
		"metric":  metric,
	}
	return c.ExecuteTool(ctx, "vector_similarity", args)
}

/* VectorArithmetic performs arithmetic operations on vectors */
func (c *Client) VectorArithmetic(ctx context.Context, vector1 []float64, vector2 []float64, operation string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"vector1":   vector1,
		"vector2":   vector2,
		"operation": operation,
	}
	return c.ExecuteTool(ctx, "vector_arithmetic", args)
}

// ============================================================================
// Embedding Tools (9 tools)
// ============================================================================

/* BatchEmbedding generates embeddings for multiple texts */
func (c *Client) BatchEmbedding(ctx context.Context, texts []string, model string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"texts": texts,
		"model": model,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "batch_embedding", args)
}

/* EmbedImage generates embedding for an image */
func (c *Client) EmbedImage(ctx context.Context, imagePath string, model string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"image_path": imagePath,
		"model":      model,
	}
	return c.ExecuteTool(ctx, "embed_image", args)
}

/* EmbedMultimodal generates multimodal embedding */
func (c *Client) EmbedMultimodal(ctx context.Context, text string, imagePath string, model string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"text":      text,
		"image_path": imagePath,
		"model":     model,
	}
	return c.ExecuteTool(ctx, "embed_multimodal", args)
}

/* EmbedCached generates cached embedding */
func (c *Client) EmbedCached(ctx context.Context, text string, model string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"text":  text,
		"model": model,
	}
	return c.ExecuteTool(ctx, "embed_cached", args)
}

/* ConfigureEmbeddingModel configures embedding model */
func (c *Client) ConfigureEmbeddingModel(ctx context.Context, modelName string, config map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"model_name": modelName,
		"config":     config,
	}
	return c.ExecuteTool(ctx, "configure_embedding_model", args)
}

// ============================================================================
// Hybrid Search Tools (8 tools)
// ============================================================================

/* HybridSearch performs hybrid semantic + lexical search */
func (c *Client) HybridSearch(ctx context.Context, query string, table string, vectorColumn string, textColumn string, limit int, weights map[string]float64) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":        query,
		"table":        table,
		"vector_column": vectorColumn,
		"text_column":  textColumn,
		"limit":        limit,
	}
	if weights != nil {
		args["weights"] = weights
	}
	return c.ExecuteTool(ctx, "hybrid_search", args)
}

/* ReciprocalRankFusion performs reciprocal rank fusion */
func (c *Client) ReciprocalRankFusion(ctx context.Context, resultSets [][]map[string]interface{}, k int) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"result_sets": resultSets,
		"k":           k,
	}
	return c.ExecuteTool(ctx, "reciprocal_rank_fusion", args)
}

/* SemanticKeywordSearch performs semantic + keyword search */
func (c *Client) SemanticKeywordSearch(ctx context.Context, query string, table string, vectorColumn string, textColumn string, limit int) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":        query,
		"table":        table,
		"vector_column": vectorColumn,
		"text_column":  textColumn,
		"limit":        limit,
	}
	return c.ExecuteTool(ctx, "semantic_keyword_search", args)
}

// ============================================================================
// Reranking Tools (6 tools)
// ============================================================================

/* RerankCrossEncoder reranks using cross-encoder */
func (c *Client) RerankCrossEncoder(ctx context.Context, query string, documents []string, topK int) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":     query,
		"documents": documents,
		"top_k":     topK,
	}
	return c.ExecuteTool(ctx, "rerank_cross_encoder", args)
}

/* RerankLLM reranks using LLM */
func (c *Client) RerankLLM(ctx context.Context, query string, documents []string, topK int, model string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":     query,
		"documents": documents,
		"top_k":     topK,
		"model":     model,
	}
	return c.ExecuteTool(ctx, "rerank_llm", args)
}

/* RerankCohere reranks using Cohere */
func (c *Client) RerankCohere(ctx context.Context, query string, documents []string, topK int) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":     query,
		"documents": documents,
		"top_k":     topK,
	}
	return c.ExecuteTool(ctx, "rerank_cohere", args)
}

/* RerankEnsemble performs ensemble reranking */
func (c *Client) RerankEnsemble(ctx context.Context, query string, documents []string, topK int, methods []string, weights map[string]float64) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":   query,
		"documents": documents,
		"top_k":   topK,
		"methods": methods,
	}
	if weights != nil {
		args["weights"] = weights
	}
	return c.ExecuteTool(ctx, "rerank_ensemble", args)
}

// ============================================================================
// ML Tools (9 tools)
// ============================================================================

/* TrainModel trains an ML model */
func (c *Client) TrainModel(ctx context.Context, algorithm string, table string, targetColumn string, featureColumns []string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"algorithm":      algorithm,
		"table":          table,
		"target_column":  targetColumn,
		"feature_columns": featureColumns,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "train_model", args)
}

/* Predict makes predictions using a trained model */
func (c *Client) Predict(ctx context.Context, modelPath string, features map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"model_path": modelPath,
		"features":   features,
	}
	return c.ExecuteTool(ctx, "predict", args)
}

/* PredictBatch makes batch predictions */
func (c *Client) PredictBatch(ctx context.Context, modelPath string, featuresList []map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"model_path":   modelPath,
		"features_list": featuresList,
	}
	return c.ExecuteTool(ctx, "predict_batch", args)
}

/* EvaluateModel evaluates model performance */
func (c *Client) EvaluateModel(ctx context.Context, modelPath string, testTable string, targetColumn string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"model_path":  modelPath,
		"test_table":  testTable,
		"target_column": targetColumn,
	}
	return c.ExecuteTool(ctx, "evaluate_model", args)
}

/* ListModels lists all trained models */
func (c *Client) ListModels(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	args := make(map[string]interface{})
	if filters != nil {
		args["filters"] = filters
	}
	return c.ExecuteTool(ctx, "list_models", args)
}

/* AutoML performs automated machine learning */
func (c *Client) AutoML(ctx context.Context, table string, targetColumn string, featureColumns []string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"table":          table,
		"target_column":  targetColumn,
		"feature_columns": featureColumns,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "automl", args)
}

// ============================================================================
// Analytics Tools (7 tools)
// ============================================================================

/* ClusterData performs clustering on data */
func (c *Client) ClusterData(ctx context.Context, algorithm string, table string, featureColumns []string, numClusters int, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"algorithm":      algorithm,
		"table":          table,
		"feature_columns": featureColumns,
		"num_clusters":   numClusters,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "cluster_data", args)
}

/* DetectOutliers detects outliers in data */
func (c *Client) DetectOutliers(ctx context.Context, table string, featureColumns []string, method string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"table":          table,
		"feature_columns": featureColumns,
		"method":         method,
	}
	return c.ExecuteTool(ctx, "detect_outliers", args)
}

/* ReduceDimensionality reduces data dimensionality */
func (c *Client) ReduceDimensionality(ctx context.Context, table string, featureColumns []string, targetDimensions int) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"table":            table,
		"feature_columns":  featureColumns,
		"target_dimensions": targetDimensions,
	}
	return c.ExecuteTool(ctx, "reduce_dimensionality", args)
}

/* AnalyzeData analyzes data */
func (c *Client) AnalyzeData(ctx context.Context, table string, columns []string, analysisType string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"table":         table,
		"columns":       columns,
		"analysis_type": analysisType,
	}
	return c.ExecuteTool(ctx, "analyze_data", args)
}

/* TopicDiscovery discovers topics in text data */
func (c *Client) TopicDiscovery(ctx context.Context, table string, textColumn string, numTopics int, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"table":      table,
		"text_column": textColumn,
		"num_topics": numTopics,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "topic_discovery", args)
}

// ============================================================================
// Index Management Tools (8 tools)
// ============================================================================

/* CreateHNSWIndex creates HNSW index */
func (c *Client) CreateHNSWIndex(ctx context.Context, indexName string, table string, column string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"index_name": indexName,
		"table":      table,
		"column":     column,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "create_hnsw_index", args)
}

/* CreateIVFIndex creates IVF index */
func (c *Client) CreateIVFIndex(ctx context.Context, indexName string, table string, column string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"index_name": indexName,
		"table":      table,
		"column":     column,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "create_ivf_index", args)
}

/* IndexStatus gets index status */
func (c *Client) IndexStatus(ctx context.Context, indexName string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"index_name": indexName,
	}
	return c.ExecuteTool(ctx, "index_status", args)
}

// ============================================================================
// RAG Operations Tools (6 tools)
// ============================================================================

/* ProcessDocument processes and chunks a document */
func (c *Client) ProcessDocument(ctx context.Context, documentPath string, chunkSize int, chunkOverlap int, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"document_path": documentPath,
		"chunk_size":    chunkSize,
		"chunk_overlap": chunkOverlap,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "process_document", args)
}

/* RetrieveContext retrieves relevant context for a query */
func (c *Client) RetrieveContext(ctx context.Context, query string, table string, vectorColumn string, limit int, filters map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":         query,
		"table":         table,
		"vector_column": vectorColumn,
		"limit":         limit,
	}
	if filters != nil {
		args["filters"] = filters
	}
	return c.ExecuteTool(ctx, "retrieve_context", args)
}

/* GenerateResponse generates response using retrieved context */
func (c *Client) GenerateResponse(ctx context.Context, query string, context []string, model string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":   query,
		"context": context,
		"model":   model,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "generate_response", args)
}

/* AnswerWithCitations generates answer with citations */
func (c *Client) AnswerWithCitations(ctx context.Context, query string, context []map[string]interface{}, model string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query":   query,
		"context": context,
		"model":   model,
	}
	return c.ExecuteTool(ctx, "answer_with_citations", args)
}

// ============================================================================
// PostgreSQL Tools (29+ tools - Key ones)
// ============================================================================

/* PostgreSQLVersion gets PostgreSQL version information */
func (c *Client) PostgreSQLVersion(ctx context.Context) (map[string]interface{}, error) {
	return c.ExecuteTool(ctx, "postgresql_version", nil)
}

/* PostgreSQLTables lists all tables */
func (c *Client) PostgreSQLTables(ctx context.Context, schema string) (map[string]interface{}, error) {
	args := make(map[string]interface{})
	if schema != "" {
		args["schema"] = schema
	}
	return c.ExecuteTool(ctx, "postgresql_tables", args)
}

/* PostgreSQLExecuteQuery executes SQL query */
func (c *Client) PostgreSQLExecuteQuery(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query": query,
	}
	if params != nil {
		args["params"] = params
	}
	return c.ExecuteTool(ctx, "postgresql_execute_query", args)
}

/* PostgreSQLQueryPlan gets query execution plan */
func (c *Client) PostgreSQLQueryPlan(ctx context.Context, query string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query": query,
	}
	return c.ExecuteTool(ctx, "postgresql_query_plan", args)
}

/* PostgreSQLQueryOptimization gets query optimization suggestions */
func (c *Client) PostgreSQLQueryOptimization(ctx context.Context, query string) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"query": query,
	}
	return c.ExecuteTool(ctx, "postgresql_query_optimization", args)
}

/* LoadDataset loads dataset from HuggingFace or other sources */
func (c *Client) LoadDataset(ctx context.Context, sourceType string, sourcePath string, options map[string]interface{}) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"source_type": sourceType,
		"source_path": sourcePath,
	}
	if options != nil {
		args["options"] = options
	}
	return c.ExecuteTool(ctx, "load_dataset", args)
}
