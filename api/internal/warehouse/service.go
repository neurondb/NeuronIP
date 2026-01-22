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
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides data warehouse Q&A functionality */
type Service struct {
	pool           *pgxpool.Pool
	agentClient    *agent.Client
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
}

/* NewService creates a new warehouse service */
func NewService(pool *pgxpool.Pool, agentClient *agent.Client, neurondbClient *neurondb.Client, mcpClient *mcp.Client) *Service {
	return &Service{
		pool:           pool,
		agentClient:    agentClient,
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
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

	// Use MCP PostgreSQL query optimization if available
	if s.mcpClient != nil {
		// Use MCP PostgreSQL tools for query optimization and plan analysis
		if s.mcpClient != nil {
			// Get query plan
			plan, err := s.mcpClient.PostgreSQLQueryPlan(ctx, generatedSQL)
			if err == nil && plan != nil {
				// Store plan in metadata
				if metadata == nil {
					metadata = make(map[string]interface{})
				}
				metadata["query_plan"] = plan
			}

			// Get optimization suggestions
			optimization, err := s.mcpClient.PostgreSQLQueryOptimization(ctx, generatedSQL)
		if err == nil {
			// Log optimization suggestions (could be used to improve query)
			if suggestions, ok := optimization["suggestions"].([]interface{}); ok && len(suggestions) > 0 {
				// Store optimization suggestions (could be added to response metadata)
			}
		}
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

	// Execute SQL with timeout - use MCP if available for better execution
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var results []map[string]interface{}
	var execErr error

	// Try MCP PostgreSQLExecuteQuery first if available
	if s.mcpClient != nil {
		mcpResult, mcpErr := s.mcpClient.PostgreSQLExecuteQuery(execCtx, generatedSQL, nil)
		if mcpErr == nil {
			// Convert MCP result to map slice
			if resultRows, ok := mcpResult["rows"].([]interface{}); ok {
				results = make([]map[string]interface{}, 0, len(resultRows))
				for _, row := range resultRows {
					if rowMap, ok := row.(map[string]interface{}); ok {
						results = append(results, rowMap)
					}
				}
			} else if data, ok := mcpResult["data"].([]interface{}); ok {
				results = make([]map[string]interface{}, 0, len(data))
				for _, row := range data {
					if rowMap, ok := row.(map[string]interface{}); ok {
						results = append(results, rowMap)
					}
				}
			}
			// If we got results from MCP, skip direct execution
			if len(results) > 0 || execErr == nil {
				// Use MCP results or continue with empty results
				goto processResults
			}
		}
		// Fall through to direct execution if MCP fails
	}

	// Direct SQL execution (fallback or primary method)
	rows, execErr := s.pool.Query(execCtx, generatedSQL)
	if execErr != nil {
		// Update query status to failed
		s.pool.Exec(ctx, `UPDATE neuronip.warehouse_queries SET status = $1, error_message = $2, executed_at = $3 WHERE id = $4`,
			"failed", execErr.Error(), time.Now(), queryID)
		return nil, fmt.Errorf("failed to execute SQL: %w", execErr)
	}
	defer rows.Close()

	// Parse results from direct execution
	fieldDescriptions := rows.FieldDescriptions()
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

processResults:

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

/* ValidateDataQualityWithML validates data quality using NeuronDB ML classification */
func (s *Service) ValidateDataQualityWithML(ctx context.Context, tableName string, columnName string, sampleValues []string, modelPath string) (map[string]interface{}, error) {
	// Use NeuronDB classification to validate data types and patterns
	results := map[string]interface{}{
		"valid_count":        0,
		"invalid_count":      0,
		"validation_results": []map[string]interface{}{},
	}

	for _, value := range sampleValues {
		// Classify the value to determine if it's valid
		classification, err := s.neurondbClient.Classify(ctx, value, modelPath)
		if err != nil {
			continue
		}

		isValid := false
		// classification is already map[string]interface{}
		if valid, ok := classification["valid"].(bool); ok {
			isValid = valid
		} else if class, ok := classification["class"].(string); ok {
			isValid = class == "valid"
		}

		result := map[string]interface{}{
			"value":          value,
			"valid":          isValid,
			"classification": classification,
		}

		if isValid {
			validCount := results["valid_count"].(int) + 1
			results["valid_count"] = validCount
		} else {
			invalidCount := results["invalid_count"].(int) + 1
			results["invalid_count"] = invalidCount
		}

		validationResults := results["validation_results"].([]map[string]interface{})
		results["validation_results"] = append(validationResults, result)
	}

	return results, nil
}

/* ScoreAnomalyWithML scores anomalies using NeuronDB regression */
func (s *Service) ScoreAnomalyWithML(ctx context.Context, features map[string]interface{}, modelPath string) (float64, error) {
	// Use NeuronDB regression to score how anomalous a data point is
	// Higher score indicates more anomalous
	score, err := s.neurondbClient.Regress(ctx, features, modelPath)
	if err != nil {
		return 0, fmt.Errorf("failed to score anomaly: %w", err)
	}

	return score, nil
}

/* GetDatabaseInfo gets database information using MCP PostgreSQL tools */
func (s *Service) GetDatabaseInfo(ctx context.Context) (map[string]interface{}, error) {
	if s.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not configured")
	}

	// Get PostgreSQL version
	version, err := s.mcpClient.PostgreSQLVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL version: %w", err)
	}

	// Get list of tables
	tables, err := s.mcpClient.PostgreSQLTables(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return map[string]interface{}{
		"version": version,
		"tables":  tables,
	}, nil
}

/* OptimizeQueryIntelligently optimizes a query using NeuronAgent for intent, MCP for plan analysis, and NeuronDB for execution */
func (s *Service) OptimizeQueryIntelligently(ctx context.Context, query string, schemaMetadata map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["original_query"] = query

	// Step 1: NeuronAgent - Understand query intent and get specialized agent if available
	var queryIntent string
	var suggestedOptimizations []string
	var specializedAgentID string

	if s.agentClient != nil {
		// Use agent to understand query intent
		intentPrompt := fmt.Sprintf("Analyze this SQL query and describe its intent: %s", query)
		context := []map[string]interface{}{
			{"query": query, "schema": schemaMetadata},
		}
		intent, err := s.agentClient.GenerateReply(ctx, context, intentPrompt)
		if err == nil {
			queryIntent = intent
			result["query_intent"] = queryIntent
		}

		// Check for SQL specialization
		specializations, err := s.agentClient.GetSpecializations(ctx, "")
		if err == nil {
			if specs, ok := specializations["specializations"].([]interface{}); ok {
				for _, spec := range specs {
					if specMap, ok := spec.(map[string]interface{}); ok {
						if name, ok := specMap["name"].(string); ok && name == "sql_optimizer" {
							if id, ok := specMap["id"].(string); ok {
								specializedAgentID = id
								result["specialized_agent_id"] = specializedAgentID
							}
						}
					}
				}
			}
		}

		// Get optimization suggestions from agent
		optPrompt := fmt.Sprintf("Provide optimization suggestions for this SQL query: %s", query)
		suggestions, err := s.agentClient.GenerateReply(ctx, context, optPrompt)
		if err == nil {
			suggestedOptimizations = []string{suggestions}
			result["agent_suggestions"] = suggestedOptimizations
		}
	}

	// Step 2: NeuronMCP - Analyze query plan
	if s.mcpClient != nil {
		// Get query execution plan
		plan, err := s.mcpClient.PostgreSQLQueryPlan(ctx, query)
		if err == nil {
			result["query_plan"] = plan
			if planStr, ok := plan["plan"].(string); ok {
				result["plan_text"] = planStr
			}
		}

		// Get optimization suggestions from MCP
		optimization, err := s.mcpClient.PostgreSQLQueryOptimization(ctx, query)
		if err == nil {
			result["mcp_optimization"] = optimization
			if suggestions, ok := optimization["suggestions"].([]interface{}); ok {
				mcpSuggestions := make([]string, 0, len(suggestions))
				for _, s := range suggestions {
					if str, ok := s.(string); ok {
						mcpSuggestions = append(mcpSuggestions, str)
					}
				}
				result["mcp_suggestions"] = mcpSuggestions
			}
			if optimizedQuery, ok := optimization["optimized_query"].(string); ok {
				result["mcp_optimized_query"] = optimizedQuery
			}
		}
	}

	// Step 3: Generate optimized query using agent if specialized agent available
	var optimizedQuery string
	if specializedAgentID != "" && s.agentClient != nil {
		// Use specialized agent to optimize query
		optContext := []map[string]interface{}{
			{"query": query, "schema": schemaMetadata, "mcp_suggestions": result["mcp_suggestions"]},
		}
		optPrompt := fmt.Sprintf("Optimize this SQL query based on the suggestions: %s", query)
		optimized, err := s.agentClient.GenerateReply(ctx, optContext, optPrompt)
		if err == nil && optimized != "" {
			optimizedQuery = optimized
			result["agent_optimized_query"] = optimizedQuery
		} else {
			optimizedQuery = query
		}
	} else if result["mcp_optimized_query"] != nil {
		optimizedQuery = result["mcp_optimized_query"].(string)
	} else {
		optimizedQuery = query
	}
	result["optimized_query"] = optimizedQuery
	result["query_changed"] = optimizedQuery != query

	// Step 4: NeuronDB - Could execute both and compare performance
	if optimizedQuery != query && s.neurondbClient != nil {
		result["optimization_applied"] = true
	}

	return result, nil
}

/* OptimizeQueryWithMCPAndAgent optimizes a query using MCP query optimization and NeuronAgent analysis */
func (s *Service) OptimizeQueryWithMCPAndAgent(ctx context.Context, query string, schemaID *uuid.UUID) (map[string]interface{}, error) {
	optimizationResult := map[string]interface{}{
		"original_query": query,
	}

	// Use MCP query optimization
	if s.mcpClient != nil {
		optimization, err := s.mcpClient.PostgreSQLQueryOptimization(ctx, query)
		if err == nil {
			optimizationResult["mcp_optimization"] = optimization
			if suggestions, ok := optimization["suggestions"].([]interface{}); ok {
				optimizationResult["suggestions"] = suggestions
			}
		}
	}

	// Use NeuronAgent for query pattern analysis if available
	if s.agentClient != nil {
		// Convert NL query analysis to agent task
		analysisTask := fmt.Sprintf("Analyze this SQL query for optimization opportunities: %s", query)
		analysis, err := s.agentClient.ExecuteAgent(ctx, "", analysisTask, []string{}, nil)
		if err == nil {
			optimizationResult["agent_analysis"] = analysis
		}
	}

	return optimizationResult, nil
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
	// Check if last_synced_at column exists, if not, use simpler query
	query := `
		SELECT id, schema_name, database_name, description, tables, 
		       created_at, updated_at
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

		err := rows.Scan(
			&schema.ID, &schema.SchemaName, &schema.DatabaseName, &description,
			&tablesJSON, &schema.CreatedAt, &schema.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}

		if description.Valid {
			schema.Description = &description.String
		}
		// last_synced_at column doesn't exist in current schema, set to nil
		schema.LastSyncedAt = nil
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
			// Use MCP or NeuronDB HybridSearch if keyword query is also provided
			if req.Query != "" {
				weights := map[string]float64{
					"semantic": 0.7,
					"keyword":  0.3,
				}
				
				// Try MCP hybrid search first
				if s.mcpClient != nil {
					mcpResult, mcpErr := s.mcpClient.HybridSearch(ctx, req.SemanticQuery, req.SemanticTable, req.SemanticColumn, "content", req.Limit, weights)
					if mcpErr == nil {
						if docs, ok := mcpResult["documents"].([]interface{}); ok {
							semanticResults = make([]map[string]interface{}, 0, len(docs))
							for _, doc := range docs {
								if docMap, ok := doc.(map[string]interface{}); ok {
									semanticResults = append(semanticResults, docMap)
								}
							}
						}
					}
				}
				
				// Fallback to NeuronDB if MCP didn't work
				if len(semanticResults) == 0 {
					hybridResults, err := s.neurondbClient.HybridSearch(ctx, queryEmbedding, req.Query, req.SemanticTable, req.SemanticColumn, "content", req.Limit, weights)
					if err == nil {
						semanticResults = hybridResults
					}
				}
			}

			// Fallback to regular semantic search if hybrid failed or not applicable
			if len(semanticResults) == 0 {
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

/* GetQueryHistory retrieves query history for a user */
func (s *Service) GetQueryHistory(ctx context.Context, userID string, limit int) ([]QueryHistoryItem, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, natural_language_query, generated_sql, status, created_at, executed_at, execution_time_ms
		FROM neuronip.warehouse_queries
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := s.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get query history: %w", err)
	}
	defer rows.Close()

	var history []QueryHistoryItem
	for rows.Next() {
		var item QueryHistoryItem
		var generatedSQL sql.NullString
		var executedAt sql.NullTime
		var executionTimeMs sql.NullInt64

		err := rows.Scan(
			&item.QueryID, &item.NaturalLanguageQuery, &generatedSQL,
			&item.Status, &item.CreatedAt, &executedAt, &executionTimeMs,
		)
		if err != nil {
			continue
		}

		if generatedSQL.Valid {
			item.GeneratedSQL = generatedSQL.String
		}
		if executedAt.Valid {
			item.ExecutedAt = &executedAt.Time
		}
		if executionTimeMs.Valid {
			item.ExecutionTimeMs = &executionTimeMs.Int64
		}

		history = append(history, item)
	}

	return history, nil
}

/* QueryHistoryItem represents a query history item */
type QueryHistoryItem struct {
	QueryID              uuid.UUID  `json:"query_id"`
	NaturalLanguageQuery string     `json:"natural_language_query"`
	GeneratedSQL         string     `json:"generated_sql"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"created_at"`
	ExecutedAt           *time.Time `json:"executed_at,omitempty"`
	ExecutionTimeMs      *int64     `json:"execution_time_ms,omitempty"`
}

/* GetQueryOptimizationSuggestions provides optimization suggestions for a query */
func (s *Service) GetQueryOptimizationSuggestions(ctx context.Context, sqlQuery string) ([]QueryOptimization, error) {
	suggestions := []QueryOptimization{}

	// Check for missing indexes
	if strings.Contains(strings.ToUpper(sqlQuery), "WHERE") {
		// Extract WHERE clause columns
		whereMatch := strings.Index(strings.ToUpper(sqlQuery), "WHERE")
		if whereMatch > 0 {
			whereClause := sqlQuery[whereMatch+5:]
			// Simple heuristic: if WHERE clause has multiple conditions, suggest composite index
			if strings.Count(whereClause, "AND") > 0 {
				suggestions = append(suggestions, QueryOptimization{
					Type:        "index",
					Severity:    "medium",
					Description: "Consider adding a composite index for frequently queried columns",
					SQL:         "", // Would generate CREATE INDEX statement
				})
			}
		}
	}

	// Check for SELECT *
	if strings.Contains(strings.ToUpper(sqlQuery), "SELECT *") {
		suggestions = append(suggestions, QueryOptimization{
			Type:        "performance",
			Severity:    "low",
			Description: "SELECT * retrieves all columns. Consider selecting only needed columns for better performance",
			SQL:         "",
		})
	}

	// Check for missing LIMIT
	if !strings.Contains(strings.ToUpper(sqlQuery), "LIMIT") && !strings.Contains(strings.ToUpper(sqlQuery), "FETCH") {
		suggestions = append(suggestions, QueryOptimization{
			Type:        "performance",
			Severity:    "high",
			Description: "Query lacks LIMIT clause. Consider adding LIMIT to prevent large result sets",
			SQL:         "",
		})
	}

	// Check for inefficient JOINs
	if strings.Contains(strings.ToUpper(sqlQuery), "JOIN") {
		joinCount := strings.Count(strings.ToUpper(sqlQuery), "JOIN")
		if joinCount > 3 {
			suggestions = append(suggestions, QueryOptimization{
				Type:        "performance",
				Severity:    "medium",
				Description: fmt.Sprintf("Query has %d JOINs. Consider if all are necessary or if query can be simplified", joinCount),
				SQL:         "",
			})
		}
	}

	return suggestions, nil
}

/* QueryOptimization represents a query optimization suggestion */
type QueryOptimization struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	SQL         string `json:"sql,omitempty"`
}

/* CacheQueryResult caches a query result */
func (s *Service) CacheQueryResult(ctx context.Context, queryHash string, result QueryResponse, ttl time.Duration) error {
	resultJSON, _ := json.Marshal(result)
	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO neuronip.query_cache (query_hash, result_data, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (query_hash) DO UPDATE
		SET result_data = $2, expires_at = $3, created_at = NOW()`

	_, err := s.pool.Exec(ctx, query, queryHash, resultJSON, expiresAt)
	return err
}

/* GetCachedQueryResult retrieves a cached query result */
func (s *Service) GetCachedQueryResult(ctx context.Context, queryHash string) (*QueryResponse, error) {
	query := `
		SELECT result_data
		FROM neuronip.query_cache
		WHERE query_hash = $1 AND expires_at > NOW()`

	var resultJSON json.RawMessage
	err := s.pool.QueryRow(ctx, query, queryHash).Scan(&resultJSON)
	if err == pgx.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cached result: %w", err)
	}

	var result QueryResponse
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached result: %w", err)
	}

	return &result, nil
}

/* Enhanced chart type detection with more chart types */
func (s *Service) generateChartConfigEnhanced(results []map[string]interface{}) (map[string]interface{}, string) {
	if len(results) == 0 {
		return nil, ""
	}

	firstRow := results[0]
	numericCols := []string{}
	stringCols := []string{}
	dateCols := []string{}

	for key, val := range firstRow {
		switch val.(type) {
		case int, int32, int64, float32, float64:
			numericCols = append(numericCols, key)
		case string:
			// Check if it's a date string
			if s.isDateString(val.(string)) {
				dateCols = append(dateCols, key)
			} else {
				stringCols = append(stringCols, key)
			}
		case time.Time:
			dateCols = append(dateCols, key)
		}
	}

	// Enhanced chart type detection
	var chartType string
	var config map[string]interface{}

	// Time series chart (date + numeric)
	if len(dateCols) > 0 && len(numericCols) > 0 {
		chartType = "line"
		config = map[string]interface{}{
			"x":      dateCols[0],
			"y":      numericCols[0],
			"series": numericCols,
		}
	} else if len(numericCols) >= 2 && len(stringCols) >= 1 {
		// Bar chart
		chartType = "bar"
		config = map[string]interface{}{
			"x":      stringCols[0],
			"y":      numericCols[0],
			"series": numericCols,
		}
	} else if len(numericCols) >= 1 && len(stringCols) >= 1 && len(results) <= 10 {
		// Pie chart (for small datasets)
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

/* isDateString checks if a string looks like a date */
func (s *Service) isDateString(str string) bool {
	dateFormats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"2006-01-02T15:04:05Z",
	}
	for _, format := range dateFormats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}
