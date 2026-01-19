package neurondb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Client provides NeuronDB integration */
type Client struct {
	pool *pgxpool.Pool
}

/* NewClient creates a new NeuronDB client */
func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

/* Helper function to convert rows to map slice */
func rowsToMapSlice(ctx context.Context, rows pgx.Rows) ([]map[string]interface{}, error) {
	defer rows.Close()
	var results []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		result := make(map[string]interface{})
		for i, desc := range fieldDescriptions {
			result[desc.Name] = values[i]
		}
		results = append(results, result)
	}
	return results, nil
}

/* GenerateEmbedding generates an embedding using NeuronDB */
func (c *Client) GenerateEmbedding(ctx context.Context, text string, model string) (string, error) {
	var embedding string
	query := `SELECT neurondb_embed($1, $2)::text`
	err := c.pool.QueryRow(ctx, query, text, model).Scan(&embedding)
	if err != nil {
		return "", fmt.Errorf("failed to generate embedding: %w", err)
	}
	return embedding, nil
}

/* VectorSearch performs vector similarity search */
func (c *Client) VectorSearch(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if embeddingColumn == "" {
		embeddingColumn = "embedding"
	}

	// Generic vector similarity search query
	query := fmt.Sprintf(`
		SELECT *, 1 - (%s <=> $1::vector) as similarity
		FROM %s
		WHERE %s IS NOT NULL
		ORDER BY %s <=> $1::vector
		LIMIT $2
	`, embeddingColumn, tableName, embeddingColumn, embeddingColumn)

	rows, err := c.pool.Query(ctx, query, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform vector search: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		result := make(map[string]interface{})
		for i, desc := range fieldDescriptions {
			result[desc.Name] = values[i]
		}
		results = append(results, result)
	}

	return results, nil
}

/* Classify performs classification using NeuronDB ML functions */
func (c *Client) Classify(ctx context.Context, text string, model string) (map[string]interface{}, error) {
	var result map[string]interface{}
	query := `SELECT neurondb_classify($1, $2)::jsonb as result`
	err := c.pool.QueryRow(ctx, query, text, model).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to classify: %w", err)
	}
	return result, nil
}

/* Regress performs regression using NeuronDB ML functions */
func (c *Client) Regress(ctx context.Context, features map[string]interface{}, model string) (float64, error) {
	var result float64
	query := `SELECT neurondb_regress($1::jsonb, $2)::float8 as result`
	err := c.pool.QueryRow(ctx, query, features, model).Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to regress: %w", err)
	}
	return result, nil
}

// ============================================================================
// Vector Operations - Advanced
// ============================================================================

/* BatchGenerateEmbedding generates embeddings for multiple texts */
func (c *Client) BatchGenerateEmbedding(ctx context.Context, texts []string, model string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}

	textsJSON, err := json.Marshal(texts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal texts: %w", err)
	}

	query := `SELECT neurondb_embed_batch($1::jsonb, $2)::text[] as embeddings`
	var embeddings []string
	err = c.pool.QueryRow(ctx, query, string(textsJSON), model).Scan(&embeddings)
	if err != nil {
		// Fallback to individual embeddings if batch function not available
		embeddings = make([]string, len(texts))
		for i, text := range texts {
			emb, err := c.GenerateEmbedding(ctx, text, model)
			if err != nil {
				return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
			}
			embeddings[i] = emb
		}
	}
	return embeddings, nil
}

/* GenerateMultimodalEmbedding generates embedding for text + image */
func (c *Client) GenerateMultimodalEmbedding(ctx context.Context, text string, imageData []byte, model string) (string, error) {
	var embedding string
	query := `SELECT neurondb_embed_multimodal($1, $2, $3)::text`
	err := c.pool.QueryRow(ctx, query, text, imageData, model).Scan(&embedding)
	if err != nil {
		return "", fmt.Errorf("failed to generate multimodal embedding: %w", err)
	}
	return embedding, nil
}

/* GenerateImageEmbedding generates embedding for image only */
func (c *Client) GenerateImageEmbedding(ctx context.Context, imageData []byte, model string) (string, error) {
	var embedding string
	query := `SELECT neurondb_embed_image($1, $2)::text`
	err := c.pool.QueryRow(ctx, query, imageData, model).Scan(&embedding)
	if err != nil {
		return "", fmt.Errorf("failed to generate image embedding: %w", err)
	}
	return embedding, nil
}

/* VectorSearchL2 performs vector search using L2 distance */
func (c *Client) VectorSearchL2(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if embeddingColumn == "" {
		embeddingColumn = "embedding"
	}

	query := fmt.Sprintf(`
		SELECT *, (%s <-> $1::vector) as distance
		FROM %s
		WHERE %s IS NOT NULL
		ORDER BY %s <-> $1::vector
		LIMIT $2
	`, embeddingColumn, tableName, embeddingColumn, embeddingColumn)

	rows, err := c.pool.Query(ctx, query, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform L2 vector search: %w", err)
	}
	return rowsToMapSlice(ctx, rows)
}

/* VectorSearchCosine performs vector search using cosine similarity */
func (c *Client) VectorSearchCosine(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if embeddingColumn == "" {
		embeddingColumn = "embedding"
	}

	query := fmt.Sprintf(`
		SELECT *, 1 - (%s <=> $1::vector) as similarity
		FROM %s
		WHERE %s IS NOT NULL
		ORDER BY %s <=> $1::vector
		LIMIT $2
	`, embeddingColumn, tableName, embeddingColumn, embeddingColumn)

	rows, err := c.pool.Query(ctx, query, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform cosine vector search: %w", err)
	}
	return rowsToMapSlice(ctx, rows)
}

/* VectorSearchInnerProduct performs vector search using inner product */
func (c *Client) VectorSearchInnerProduct(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if embeddingColumn == "" {
		embeddingColumn = "embedding"
	}

	query := fmt.Sprintf(`
		SELECT *, (%s <#> $1::vector) as distance
		FROM %s
		WHERE %s IS NOT NULL
		ORDER BY %s <#> $1::vector
		LIMIT $2
	`, embeddingColumn, tableName, embeddingColumn, embeddingColumn)

	rows, err := c.pool.Query(ctx, query, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to perform inner product vector search: %w", err)
	}
	return rowsToMapSlice(ctx, rows)
}

/* HybridSearch performs hybrid semantic + keyword search */
func (c *Client) HybridSearch(ctx context.Context, queryEmbedding string, keywordQuery string, tableName string, embeddingColumn string, textColumn string, limit int, weights map[string]float64) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if embeddingColumn == "" {
		embeddingColumn = "embedding"
	}
	if textColumn == "" {
		textColumn = "content"
	}
	semanticWeight := 0.7
	keywordWeight := 0.3
	if weights != nil {
		if w, ok := weights["semantic"]; ok {
			semanticWeight = w
		}
		if w, ok := weights["keyword"]; ok {
			keywordWeight = w
		}
	}

	query := fmt.Sprintf(`
		WITH semantic_results AS (
			SELECT *, 1 - (%s <=> $1::vector) as semantic_score
			FROM %s
			WHERE %s IS NOT NULL
			ORDER BY %s <=> $1::vector
			LIMIT $3
		),
		keyword_results AS (
			SELECT *, ts_rank(to_tsvector('english', %s), plainto_tsquery('english', $2)) as keyword_score
			FROM %s
			WHERE to_tsvector('english', %s) @@ plainto_tsquery('english', $2)
			ORDER BY keyword_score DESC
			LIMIT $3
		)
		SELECT DISTINCT ON (COALESCE(s.id, k.id)) 
			COALESCE(s.*, k.*),
			(COALESCE(s.semantic_score, 0) * $4 + COALESCE(k.keyword_score, 0) * $5) as combined_score
		FROM semantic_results s
		FULL OUTER JOIN keyword_results k ON s.id = k.id
		ORDER BY combined_score DESC
		LIMIT $3
	`, embeddingColumn, tableName, embeddingColumn, embeddingColumn, textColumn, tableName, textColumn)

	rows, err := c.pool.Query(ctx, query, queryEmbedding, keywordQuery, limit, semanticWeight, keywordWeight)
	if err != nil {
		return nil, fmt.Errorf("failed to perform hybrid search: %w", err)
	}
	return rowsToMapSlice(ctx, rows)
}

/* CreateVectorIndex creates HNSW or IVF index on a vector column */
func (c *Client) CreateVectorIndex(ctx context.Context, indexName string, tableName string, columnName string, indexType string, options map[string]interface{}) error {
	var query string
	if indexType == "hnsw" {
		m := 16
		efConstruction := 64
		if options != nil {
			if mVal, ok := options["m"].(float64); ok {
				m = int(mVal)
			}
			if efVal, ok := options["ef_construction"].(float64); ok {
				efConstruction = int(efVal)
			}
		}
		query = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING hnsw (%s vector_cosine_ops) WITH (m = %d, ef_construction = %d)`,
			indexName, tableName, columnName, m, efConstruction)
	} else if indexType == "ivf" {
		lists := 100
		if options != nil {
			if lVal, ok := options["lists"].(float64); ok {
				lists = int(lVal)
			}
		}
		query = fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s USING ivfflat (%s vector_cosine_ops) WITH (lists = %d)`,
			indexName, tableName, columnName, lists)
	} else {
		return fmt.Errorf("unsupported index type: %s (supported: hnsw, ivf)", indexType)
	}

	_, err := c.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %w", err)
	}
	return nil
}

// ============================================================================
// ML Operations
// ============================================================================

/* TrainModel trains an ML model using NeuronDB */
func (c *Client) TrainModel(ctx context.Context, algorithm string, tableName string, targetColumn string, featureColumns []string, options map[string]interface{}) (string, error) {
	featuresJSON, err := json.Marshal(featureColumns)
	if err != nil {
		return "", fmt.Errorf("failed to marshal feature columns: %w", err)
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to marshal options: %w", err)
	}

	query := `SELECT neurondb_train_model($1, $2, $3, $4::text[], $5::jsonb)::text as model_path`
	var modelPath string
	err = c.pool.QueryRow(ctx, query, algorithm, tableName, targetColumn, string(featuresJSON), string(optionsJSON)).Scan(&modelPath)
	if err != nil {
		return "", fmt.Errorf("failed to train model: %w", err)
	}
	return modelPath, nil
}

/* Predict makes predictions using a trained model */
func (c *Client) Predict(ctx context.Context, modelPath string, features map[string]interface{}) (interface{}, error) {
	featuresJSON, err := json.Marshal(features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	query := `SELECT neurondb_predict($1, $2::jsonb) as prediction`
	var result interface{}
	err = c.pool.QueryRow(ctx, query, modelPath, string(featuresJSON)).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to predict: %w", err)
	}
	return result, nil
}

/* BatchPredict makes batch predictions using a trained model */
func (c *Client) BatchPredict(ctx context.Context, modelPath string, featuresList []map[string]interface{}) ([]interface{}, error) {
	featuresJSON, err := json.Marshal(featuresList)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features list: %w", err)
	}

	query := `SELECT neurondb_predict_batch($1, $2::jsonb)::jsonb as predictions`
	var predictions []interface{}
	var predictionsJSON []byte
	err = c.pool.QueryRow(ctx, query, modelPath, string(featuresJSON)).Scan(&predictionsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to batch predict: %w", err)
	}
	err = json.Unmarshal(predictionsJSON, &predictions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal predictions: %w", err)
	}
	return predictions, nil
}

/* EvaluateModel evaluates model performance */
func (c *Client) EvaluateModel(ctx context.Context, modelPath string, testTable string, targetColumn string) (map[string]interface{}, error) {
	query := `SELECT neurondb_evaluate_model($1, $2, $3)::jsonb as metrics`
	var result map[string]interface{}
	var metricsJSON []byte
	err := c.pool.QueryRow(ctx, query, modelPath, testTable, targetColumn).Scan(&metricsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate model: %w", err)
	}
	err = json.Unmarshal(metricsJSON, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}
	return result, nil
}

/* ListModels lists all trained models */
func (c *Client) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	query := `SELECT * FROM neurondb.models ORDER BY created_at DESC`
	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		// If table doesn't exist, return empty list
		return []map[string]interface{}{}, nil
	}
	return rowsToMapSlice(ctx, rows)
}

/* GetModelInfo retrieves information about a model */
func (c *Client) GetModelInfo(ctx context.Context, modelPath string) (map[string]interface{}, error) {
	query := `SELECT * FROM neurondb.models WHERE model_path = $1`
	rows, err := c.pool.Query(ctx, query, modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}
	results, err := rowsToMapSlice(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("model not found: %s", modelPath)
	}
	return results[0], nil
}

/* DeleteModel deletes a trained model */
func (c *Client) DeleteModel(ctx context.Context, modelPath string) error {
	query := `DELETE FROM neurondb.models WHERE model_path = $1`
	_, err := c.pool.Exec(ctx, query, modelPath)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	return nil
}

// ============================================================================
// RAG Operations
// ============================================================================

/* ProcessDocument processes and chunks a document for RAG */
func (c *Client) ProcessDocument(ctx context.Context, content string, chunkSize int, chunkOverlap int) ([]map[string]interface{}, error) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 {
		chunkOverlap = 200
	}

	query := `SELECT neurondb_process_document($1, $2, $3)::jsonb as chunks`
	var chunksJSON []byte
	err := c.pool.QueryRow(ctx, query, content, chunkSize, chunkOverlap).Scan(&chunksJSON)
	if err != nil {
		// Fallback to simple chunking if function not available
		return c.simpleChunkDocument(content, chunkSize, chunkOverlap), nil
	}

	var chunks []map[string]interface{}
	err = json.Unmarshal(chunksJSON, &chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chunks: %w", err)
	}
	return chunks, nil
}

/* simpleChunkDocument provides fallback chunking */
func (c *Client) simpleChunkDocument(content string, chunkSize int, chunkOverlap int) []map[string]interface{} {
	chunks := []map[string]interface{}{}
	start := 0
	contentLen := len(content)

	for start < contentLen {
		end := start + chunkSize
		if end > contentLen {
			end = contentLen
		}

		chunk := content[start:end]
		chunks = append(chunks, map[string]interface{}{
			"text":     chunk,
			"index":    len(chunks),
			"start":    start,
			"end":      end,
			"overlap":  chunkOverlap,
		})

		if end >= contentLen {
			break
		}
		start = end - chunkOverlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

/* RetrieveContext retrieves relevant context for a query */
func (c *Client) RetrieveContext(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int) ([]map[string]interface{}, error) {
	return c.VectorSearch(ctx, queryEmbedding, tableName, embeddingColumn, limit)
}

/* GenerateResponse generates RAG response using retrieved context */
func (c *Client) GenerateResponse(ctx context.Context, query string, context []string, model string) (string, error) {
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	querySQL := `SELECT neurondb_generate_response($1, $2::jsonb, $3)::text as response`
	var response string
	err = c.pool.QueryRow(ctx, querySQL, query, string(contextJSON), model).Scan(&response)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}
	return response, nil
}

/* AnswerWithCitations generates answer with citations */
func (c *Client) AnswerWithCitations(ctx context.Context, query string, context []map[string]interface{}, model string) (map[string]interface{}, error) {
	contextJSON, err := json.Marshal(context)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context: %w", err)
	}

	querySQL := `SELECT neurondb_answer_with_citations($1, $2::jsonb, $3)::jsonb as result`
	var result map[string]interface{}
	var resultJSON []byte
	err = c.pool.QueryRow(ctx, querySQL, query, string(contextJSON), model).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer with citations: %w", err)
	}
	err = json.Unmarshal(resultJSON, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}
	return result, nil
}

// ============================================================================
// Analytics Operations
// ============================================================================

/* ClusterData performs clustering on data */
func (c *Client) ClusterData(ctx context.Context, algorithm string, tableName string, featureColumns []string, numClusters int, options map[string]interface{}) ([]map[string]interface{}, error) {
	featuresJSON, err := json.Marshal(featureColumns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal feature columns: %w", err)
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	query := `SELECT neurondb_cluster_data($1, $2, $3::text[], $4, $5::jsonb)::jsonb as clusters`
	var clustersJSON []byte
	err = c.pool.QueryRow(ctx, query, algorithm, tableName, string(featuresJSON), numClusters, string(optionsJSON)).Scan(&clustersJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to cluster data: %w", err)
	}

	var clusters []map[string]interface{}
	err = json.Unmarshal(clustersJSON, &clusters)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal clusters: %w", err)
	}
	return clusters, nil
}

/* DetectOutliers detects outliers in data */
func (c *Client) DetectOutliers(ctx context.Context, tableName string, featureColumns []string, method string) ([]map[string]interface{}, error) {
	featuresJSON, err := json.Marshal(featureColumns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal feature columns: %w", err)
	}

	query := `SELECT neurondb_detect_outliers($1, $2::text[], $3)::jsonb as outliers`
	var outliersJSON []byte
	err = c.pool.QueryRow(ctx, query, tableName, string(featuresJSON), method).Scan(&outliersJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to detect outliers: %w", err)
	}

	var outliers []map[string]interface{}
	err = json.Unmarshal(outliersJSON, &outliers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal outliers: %w", err)
	}
	return outliers, nil
}

/* ReduceDimensionality reduces data dimensionality using PCA */
func (c *Client) ReduceDimensionality(ctx context.Context, tableName string, featureColumns []string, targetDimensions int) ([]map[string]interface{}, error) {
	featuresJSON, err := json.Marshal(featureColumns)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal feature columns: %w", err)
	}

	query := `SELECT neurondb_reduce_dimensionality($1, $2::text[], $3)::jsonb as reduced`
	var reducedJSON []byte
	err = c.pool.QueryRow(ctx, query, tableName, string(featuresJSON), targetDimensions).Scan(&reducedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to reduce dimensionality: %w", err)
	}

	var reduced []map[string]interface{}
	err = json.Unmarshal(reducedJSON, &reduced)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reduced data: %w", err)
	}
	return reduced, nil
}

/* TimeSeriesAnalysis performs time series analysis */
func (c *Client) TimeSeriesAnalysis(ctx context.Context, tableName string, timeColumn string, valueColumn string, method string, options map[string]interface{}) (map[string]interface{}, error) {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	query := `SELECT neurondb_timeseries_analysis($1, $2, $3, $4, $5::jsonb)::jsonb as result`
	var result map[string]interface{}
	var resultJSON []byte
	err = c.pool.QueryRow(ctx, query, tableName, timeColumn, valueColumn, method, string(optionsJSON)).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to perform time series analysis: %w", err)
	}
	err = json.Unmarshal(resultJSON, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}
	return result, nil
}
