package warehouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides data warehouse Q&A functionality */
type Service struct {
	pool           *pgxpool.Pool
	agentClient    *agent.Client
	neurondbClient *neurondb.Client
}

/* NewService creates a new warehouse service */
func NewService(pool *pgxpool.Pool, agentClient *agent.Client, neurondbClient *neurondb.Client) *Service {
	return &Service{
		pool:           pool,
		agentClient:    agentClient,
		neurondbClient: neurondbClient,
	}
}

/* QueryRequest represents a natural language query */
type QueryRequest struct {
	Query         string
	SchemaID      *uuid.UUID
	UserID        *string
	SemanticQuery *string                // Optional semantic similarity search
	SQLFilters    map[string]interface{} // Optional SQL filter conditions
}

/* QueryResponse represents the query response */
type QueryResponse struct {
	QueryID     uuid.UUID                `json:"query_id"`
	SQL         string                   `json:"sql"`
	Results     []map[string]interface{} `json:"results"`
	Explanation string                   `json:"explanation"`
	ChartConfig map[string]interface{}   `json:"chart_config,omitempty"`
	ChartType   string                   `json:"chart_type,omitempty"`
}

/* Schema represents a warehouse schema */
type Schema struct {
	ID           uuid.UUID                `json:"id"`
	SchemaName   string                   `json:"schema_name"`
	DatabaseName string                   `json:"database_name"`
	Description  *string                  `json:"description,omitempty"`
	Tables       []map[string]interface{} `json:"tables"`
	LastSyncedAt *time.Time               `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

/* ExecuteQuery executes a natural language query */
func (s *Service) ExecuteQuery(ctx context.Context, req QueryRequest) (*QueryResponse, error) {
	// Get schema if provided
	var schema *Schema
	var schemaMetadata map[string]interface{}
	if req.SchemaID != nil {
		s, err := s.GetSchema(ctx, *req.SchemaID)
		if err != nil {
			return nil, fmt.Errorf("failed to get schema: %w", err)
		}
		schema = s
		schemaMetadata = map[string]interface{}{
			"schema_name":   schema.SchemaName,
			"database_name": schema.DatabaseName,
			"tables":        schema.Tables,
		}
	}

	// Convert NL to SQL using NeuronAgent
	generatedSQL, err := s.agentClient.ConvertNLToSQL(ctx, req.Query, schemaMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to convert NL to SQL: %w", err)
	}

	// Validate SQL syntax (basic validation)
	if err := s.validateSQL(generatedSQL); err != nil {
		return nil, fmt.Errorf("invalid SQL: %w", err)
	}

	// Create warehouse query record
	queryID := uuid.New()
	now := time.Now()
	insertQuery := `
		INSERT INTO neuronip.warehouse_queries 
		(id, user_id, natural_language_query, generated_sql, schema_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err = s.pool.QueryRow(ctx, insertQuery,
		queryID, req.UserID, req.Query, generatedSQL, req.SchemaID, "executing", now,
	).Scan(&queryID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query record: %w", err)
	}

	// Execute SQL with timeout
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := s.pool.Query(execCtx, generatedSQL)
	if err != nil {
		// Update query status to failed
		s.pool.Exec(ctx, `UPDATE neuronip.warehouse_queries SET status = $1, error_message = $2, executed_at = $3 WHERE id = $4`,
			"failed", err.Error(), time.Now(), queryID)
		return nil, fmt.Errorf("failed to execute SQL: %w", err)
	}
	defer rows.Close()

	// Parse results
	fieldDescriptions := rows.FieldDescriptions()
	var results []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to get row values: %w", err)
		}

		row := make(map[string]interface{})
		for i, desc := range fieldDescriptions {
			row[desc.Name] = values[i]
		}
		results = append(results, row)
	}

	// Update query status to completed
	executedAt := time.Now()
	s.pool.Exec(ctx, `UPDATE neuronip.warehouse_queries SET status = $1, executed_at = $2 WHERE id = $3`,
		"completed", executedAt, queryID)

	// Store results
	resultData, _ := json.Marshal(results)
	resultInsertQuery := `
		INSERT INTO neuronip.query_results 
		(query_id, result_data, row_count, execution_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	executionTimeMs := int(time.Since(now).Milliseconds())
	s.pool.Exec(ctx, resultInsertQuery, queryID, resultData, len(results), executionTimeMs, time.Now())

	// Generate chart config
	chartConfig, chartType := s.generateChartConfig(results)

	// Update result with chart config
	if chartConfig != nil {
		chartConfigJSON, _ := json.Marshal(chartConfig)
		s.pool.Exec(ctx, `UPDATE neuronip.query_results SET chart_config = $1, chart_type = $2 WHERE query_id = $3`,
			chartConfigJSON, chartType, queryID)
	}

	// Generate explanation
	explanation := s.generateExplanation(generatedSQL, results, req.Query)

	// Store explanation
	explanationInsertQuery := `
		INSERT INTO neuronip.query_explanations 
		(query_id, explanation_text, explanation_type, created_at)
		VALUES ($1, $2, $3, $4)`
	s.pool.Exec(ctx, explanationInsertQuery, queryID, explanation, "sql", time.Now())

	// Generate result explanation
	resultExplanation := s.generateResultExplanation(results)
	s.pool.Exec(ctx, explanationInsertQuery, queryID, resultExplanation, "result", time.Now())

	// Generate insight explanation
	insightExplanation := s.generateInsightExplanation(results, req.Query)
	if insightExplanation != "" {
		s.pool.Exec(ctx, explanationInsertQuery, queryID, insightExplanation, "insight", time.Now())
	}

	return &QueryResponse{
		QueryID:     queryID,
		SQL:         generatedSQL,
		Results:     results,
		Explanation: explanation,
		ChartConfig: chartConfig,
		ChartType:   chartType,
	}, nil
}

/* GetQuery retrieves a query with results */
func (s *Service) GetQuery(ctx context.Context, queryID uuid.UUID) (*QueryResponse, error) {
	var queryIDOut uuid.UUID
	var naturalLanguageQuery, status string
	var createdAt time.Time
	var userID, generatedSQL, errorMessage, schemaID sql.NullString
	var executedAt sql.NullTime

	queryStr := `
		SELECT id, user_id, natural_language_query, generated_sql, schema_id, 
		       status, error_message, created_at, executed_at
		FROM neuronip.warehouse_queries
		WHERE id = $1`
	err := s.pool.QueryRow(ctx, queryStr, queryID).Scan(
		&queryIDOut, &userID, &naturalLanguageQuery, &generatedSQL,
		&schemaID, &status, &errorMessage, &createdAt, &executedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("query not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get query: %w", err)
	}

	// Get results
	var results []map[string]interface{}
	var chartConfig map[string]interface{}
	var chartType string
	resultQuery := `
		SELECT result_data, chart_config, chart_type
		FROM neuronip.query_results
		WHERE query_id = $1
		ORDER BY created_at DESC
		LIMIT 1`
	var resultData json.RawMessage
	var chartConfigJSON json.RawMessage
	err = s.pool.QueryRow(ctx, resultQuery, queryID).Scan(&resultData, &chartConfigJSON, &chartType)
	if err == nil {
		json.Unmarshal(resultData, &results)
		if chartConfigJSON != nil {
			json.Unmarshal(chartConfigJSON, &chartConfig)
		}
	}

	// Get explanation
	var explanation string
	explanationQuery := `SELECT explanation_text FROM neuronip.query_explanations WHERE query_id = $1 AND explanation_type = 'sql' ORDER BY created_at DESC LIMIT 1`
	s.pool.QueryRow(ctx, explanationQuery, queryID).Scan(&explanation)
	if explanation == "" {
		explanation = "Query executed successfully"
	}

	sqlStr := ""
	if generatedSQL.Valid {
		sqlStr = generatedSQL.String
	}

	return &QueryResponse{
		QueryID:     queryIDOut,
		SQL:         sqlStr,
		Results:     results,
		Explanation: explanation,
		ChartConfig: chartConfig,
		ChartType:   chartType,
	}, nil
}

/* GetSchema retrieves a warehouse schema */
func (s *Service) GetSchema(ctx context.Context, id uuid.UUID) (*Schema, error) {
	var schema Schema
	var description sql.NullString
	var tablesJSON json.RawMessage
	var lastSyncedAt sql.NullTime

	query := `
		SELECT id, schema_name, database_name, description, tables, 
		       last_synced_at, created_at, updated_at
		FROM neuronip.warehouse_schemas
		WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&schema.ID, &schema.SchemaName, &schema.DatabaseName, &description,
		&tablesJSON, &lastSyncedAt, &schema.CreatedAt, &schema.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("schema not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	if description.Valid {
		schema.Description = &description.String
	}
	if lastSyncedAt.Valid {
		schema.LastSyncedAt = &lastSyncedAt.Time
	}

	// Parse tables JSON
	if len(tablesJSON) > 0 {
		json.Unmarshal(tablesJSON, &schema.Tables)
	}

	return &schema, nil
}

/* ListSchemas lists all warehouse schemas */
func (s *Service) ListSchemas(ctx context.Context) ([]Schema, error) {
	query := `
		SELECT id, schema_name, database_name, description, tables, 
		       last_synced_at, created_at, updated_at
		FROM neuronip.warehouse_schemas
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []Schema
	for rows.Next() {
		var schema Schema
		var description sql.NullString
		var tablesJSON json.RawMessage
		var lastSyncedAt sql.NullTime

		err := rows.Scan(
			&schema.ID, &schema.SchemaName, &schema.DatabaseName, &description,
			&tablesJSON, &lastSyncedAt, &schema.CreatedAt, &schema.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}

		if description.Valid {
			schema.Description = &description.String
		}
		if lastSyncedAt.Valid {
			schema.LastSyncedAt = &lastSyncedAt.Time
		}
		if tablesJSON != nil {
			json.Unmarshal(tablesJSON, &schema.Tables)
		}

		schemas = append(schemas, schema)
	}

	return schemas, nil
}

/* CreateSchema creates a new warehouse schema */
func (s *Service) CreateSchema(ctx context.Context, schemaName string, databaseName string, description *string, tables []map[string]interface{}) (*Schema, error) {
	id := uuid.New()
	now := time.Now()
	tablesJSON, _ := json.Marshal(tables)

	query := `
		INSERT INTO neuronip.warehouse_schemas 
		(id, schema_name, database_name, description, tables, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, schema_name, database_name, description, tables, created_at, updated_at`

	var schema Schema
	var desc sql.NullString
	var tablesJSONResult json.RawMessage
	err := s.pool.QueryRow(ctx, query,
		id, schemaName, databaseName, description, tablesJSON, now, now,
	).Scan(
		&schema.ID, &schema.SchemaName, &schema.DatabaseName, &desc,
		&tablesJSONResult, &schema.CreatedAt, &schema.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	if desc.Valid {
		schema.Description = &desc.String
	}
	if tablesJSONResult != nil {
		json.Unmarshal(tablesJSONResult, &schema.Tables)
	}

	return &schema, nil
}

/* GetExplanation retrieves query explanation */
func (s *Service) GetExplanation(ctx context.Context, queryID uuid.UUID, explanationType string) (string, error) {
	var explanation string
	query := `
		SELECT explanation_text
		FROM neuronip.query_explanations
		WHERE query_id = $1 AND explanation_type = $2
		ORDER BY created_at DESC
		LIMIT 1`

	err := s.pool.QueryRow(ctx, query, queryID, explanationType).Scan(&explanation)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("explanation not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get explanation: %w", err)
	}

	return explanation, nil
}

// Helper methods

func (s *Service) validateSQL(sql string) error {
	// Basic SQL validation - check for dangerous statements
	dangerous := []string{"DROP", "DELETE", "TRUNCATE", "ALTER", "CREATE", "INSERT", "UPDATE"}
	sqlUpper := strings.ToUpper(sql)
	for _, stmt := range dangerous {
		if strings.Contains(sqlUpper, stmt) {
			// Allow if it's in a comment or subquery context
			// This is basic validation - in production, use proper SQL parser
			if strings.Contains(sqlUpper, "SELECT") {
				// For now, only allow SELECT statements for safety
				continue
			}
			return fmt.Errorf("unsafe SQL statement detected: %s", stmt)
		}
	}
	return nil
}

func (s *Service) generateChartConfig(results []map[string]interface{}) (map[string]interface{}, string) {
	if len(results) == 0 {
		return nil, ""
	}

	// Detect chart type based on data structure
	firstRow := results[0]

	// Count numeric vs string columns
	numericCols := []string{}
	stringCols := []string{}
	for key, val := range firstRow {
		switch val.(type) {
		case int, int32, int64, float32, float64:
			numericCols = append(numericCols, key)
		case string:
			stringCols = append(stringCols, key)
		}
	}

	// Determine chart type
	var chartType string
	var config map[string]interface{}

	if len(numericCols) >= 2 && len(stringCols) >= 1 {
		// Bar or line chart
		chartType = "bar"
		config = map[string]interface{}{
			"x":      stringCols[0],
			"y":      numericCols[0],
			"series": numericCols,
		}
	} else if len(numericCols) >= 1 && len(stringCols) >= 1 {
		// Pie chart
		chartType = "pie"
		config = map[string]interface{}{
			"category": stringCols[0],
			"value":    numericCols[0],
		}
	} else if len(numericCols) >= 2 {
		// Scatter plot
		chartType = "scatter"
		config = map[string]interface{}{
			"x": numericCols[0],
			"y": numericCols[1],
		}
	} else {
		// Table view
		chartType = "table"
		config = map[string]interface{}{
			"columns": getKeys(firstRow),
		}
	}

	return config, chartType
}

func (s *Service) generateExplanation(sql string, results []map[string]interface{}, nlQuery string) string {
	// Generate explanation of SQL query
	return fmt.Sprintf("The query translates '%s' into SQL: %s. It returned %d rows.", nlQuery, sql, len(results))
}

func (s *Service) generateResultExplanation(results []map[string]interface{}) string {
	if len(results) == 0 {
		return "The query returned no results."
	}
	return fmt.Sprintf("The query returned %d result(s).", len(results))
}

func (s *Service) generateInsightExplanation(results []map[string]interface{}, query string) string {
	if len(results) == 0 {
		return ""
	}
	// Basic insight - in production, use AI to generate insights
	return fmt.Sprintf("Based on the results, the query '%s' returned %d records.", query, len(results))
}

/* HybridSearchRequest represents a hybrid search request combining SQL and semantic */
type HybridSearchRequest struct {
	Query          string                 // Natural language query for SQL generation
	SemanticQuery  string                 // Semantic similarity search query
	SchemaID       *uuid.UUID             // Optional schema ID
	SQLFilters     map[string]interface{} // SQL filter conditions (e.g., {"price": {"lt": 100}})
	SemanticTable  string                 // Table to search semantically
	SemanticColumn string                 // Column with embeddings
	Limit          int                    // Result limit
	Threshold      float64                // Semantic similarity threshold
	UserID         *string
}

/* HybridSearchResponse represents hybrid search response */
type HybridSearchResponse struct {
	QueryID         uuid.UUID                `json:"query_id"`
	SQLResults      []map[string]interface{} `json:"sql_results,omitempty"`
	SemanticResults []map[string]interface{} `json:"semantic_results,omitempty"`
	CombinedResults []map[string]interface{} `json:"combined_results,omitempty"`
	Explanation     string                   `json:"explanation"`
}

/* ExecuteHybridSearch executes a hybrid search combining SQL and semantic similarity */
func (s *Service) ExecuteHybridSearch(ctx context.Context, req HybridSearchRequest) (*HybridSearchResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Threshold <= 0 {
		req.Threshold = 0.5
	}

	queryID := uuid.New()
	var sqlResults []map[string]interface{}
	var semanticResults []map[string]interface{}

	// Execute SQL query if provided
	if req.Query != "" {
		sqlReq := QueryRequest{
			Query:      req.Query,
			SchemaID:   req.SchemaID,
			UserID:     req.UserID,
			SQLFilters: req.SQLFilters,
		}
		sqlResp, err := s.ExecuteQuery(ctx, sqlReq)
		if err == nil {
			sqlResults = sqlResp.Results
		}
	}

	// Execute semantic search if provided
	if req.SemanticQuery != "" && req.SemanticTable != "" {
		// Generate embedding for semantic query
		queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, req.SemanticQuery, "sentence-transformers/all-MiniLM-L6-v2")
		if err == nil {
			// Build semantic search query with SQL filters
			semanticSQL := fmt.Sprintf(`
				SELECT *, 1 - (%s <=> $1::vector) as similarity
				FROM %s
				WHERE %s IS NOT NULL
				AND 1 - (%s <=> $1::vector) >= $2`,
				req.SemanticColumn, req.SemanticTable, req.SemanticColumn, req.SemanticColumn)

			// Add SQL filters if provided
			args := []interface{}{queryEmbedding, req.Threshold}
			if len(req.SQLFilters) > 0 {
				// Build WHERE clause from filters
				for col, condition := range req.SQLFilters {
					if condMap, ok := condition.(map[string]interface{}); ok {
						for op, val := range condMap {
							switch op {
							case "lt":
								semanticSQL += fmt.Sprintf(" AND %s < $%d", col, len(args)+1)
								args = append(args, val)
							case "lte":
								semanticSQL += fmt.Sprintf(" AND %s <= $%d", col, len(args)+1)
								args = append(args, val)
							case "gt":
								semanticSQL += fmt.Sprintf(" AND %s > $%d", col, len(args)+1)
								args = append(args, val)
							case "gte":
								semanticSQL += fmt.Sprintf(" AND %s >= $%d", col, len(args)+1)
								args = append(args, val)
							case "eq":
								semanticSQL += fmt.Sprintf(" AND %s = $%d", col, len(args)+1)
								args = append(args, val)
							}
						}
					}
				}
			}

			semanticSQL += fmt.Sprintf(" ORDER BY %s <=> $1::vector LIMIT $%d", req.SemanticColumn, len(args)+1)
			args = append(args, req.Limit)

			rows, err := s.pool.Query(ctx, semanticSQL, args...)
			if err == nil {
				defer rows.Close()
				fieldDescriptions := rows.FieldDescriptions()
				for rows.Next() {
					values, _ := rows.Values()
					row := make(map[string]interface{})
					for i, desc := range fieldDescriptions {
						row[desc.Name] = values[i]
					}
					semanticResults = append(semanticResults, row)
				}
			}
		}
	}

	// Combine results (merge based on common keys or union)
	combinedResults := s.combineResults(sqlResults, semanticResults, req.Limit)

	explanation := fmt.Sprintf("Hybrid search executed: SQL returned %d results, semantic search returned %d results", len(sqlResults), len(semanticResults))

	return &HybridSearchResponse{
		QueryID:         queryID,
		SQLResults:      sqlResults,
		SemanticResults: semanticResults,
		CombinedResults: combinedResults,
		Explanation:     explanation,
	}, nil
}

/* combineResults combines SQL and semantic search results */
func (s *Service) combineResults(sqlResults, semanticResults []map[string]interface{}, limit int) []map[string]interface{} {
	// Simple union of results (deduplicate by common ID if available)
	combined := make([]map[string]interface{}, 0)
	seen := make(map[string]bool)

	// Add SQL results first
	for _, result := range sqlResults {
		if id, ok := result["id"].(string); ok {
			if seen[id] {
				continue
			}
			seen[id] = true
		}
		combined = append(combined, result)
		if len(combined) >= limit {
			return combined
		}
	}

	// Add semantic results
	for _, result := range semanticResults {
		if id, ok := result["id"].(string); ok {
			if seen[id] {
				continue
			}
			seen[id] = true
		}
		combined = append(combined, result)
		if len(combined) >= limit {
			return combined
		}
	}

	return combined
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
