package semantic

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
)

/* QueryRewriter provides business semantic query rewriting functionality */
type QueryRewriter struct {
	pool              *pgxpool.Pool
	semanticService   *catalog.SemanticService
	metricsService    *catalog.MetricsService
}

/* NewQueryRewriter creates a new query rewriter */
func NewQueryRewriter(pool *pgxpool.Pool, semanticService *catalog.SemanticService, metricsService *catalog.MetricsService) *QueryRewriter {
	return &QueryRewriter{
		pool:            pool,
		semanticService: semanticService,
		metricsService:  metricsService,
	}
}

/* RewriteQuery rewrites a natural language or SQL query using business semantics */
func (qr *QueryRewriter) RewriteQuery(ctx context.Context, query string, queryType string) (string, error) {
	if queryType == "sql" {
		return qr.rewriteSQLQuery(ctx, query)
	}
	return qr.rewriteNLQuery(ctx, query)
}

/* rewriteSQLQuery rewrites SQL query by replacing business terms with SQL expressions */
func (qr *QueryRewriter) rewriteSQLQuery(ctx context.Context, sqlQuery string) (string, error) {
	// Extract potential business terms from SQL (table names, column names, function names)
	terms := qr.extractTermsFromSQL(sqlQuery)
	
	rewrittenQuery := sqlQuery
	
	// For each term, check if it has a semantic definition or metric mapping
	for _, term := range terms {
		// Check semantic definitions
		definitions, err := qr.semanticService.SearchSemanticDefinitions(ctx, term, nil, 1)
		if err == nil && len(definitions) > 0 {
			def := definitions[0]
			if def.SQLExpression != nil && *def.SQLExpression != "" {
				// Replace term with SQL expression
				rewrittenQuery = qr.replaceTermWithSQL(rewrittenQuery, term, *def.SQLExpression)
			}
		}
		
		// Check metric catalog
		metrics, err := qr.metricsService.ListMetrics(ctx, nil, nil, 100)
		if err == nil {
			for _, metric := range metrics {
				if strings.EqualFold(metric.Name, term) || strings.EqualFold(metric.DisplayName, term) {
					if metric.Status == "approved" {
						// Replace with metric SQL expression
						rewrittenQuery = qr.replaceTermWithSQL(rewrittenQuery, term, metric.SQLExpression)
					}
				}
			}
		}
	}
	
	return rewrittenQuery, nil
}

/* rewriteNLQuery rewrites natural language query using business semantics */
func (qr *QueryRewriter) rewriteNLQuery(ctx context.Context, nlQuery string) (string, error) {
	// Extract business terms from natural language query
	terms := qr.extractTermsFromNL(nlQuery)
	
	rewrittenQuery := nlQuery
	
	// Enhance query with business context
	for _, term := range terms {
		// Check semantic definitions for context
		definitions, err := qr.semanticService.SearchSemanticDefinitions(ctx, term, nil, 1)
		if err == nil && len(definitions) > 0 {
			def := definitions[0]
			// Add definition context to query
			if def.Definition != "" {
				rewrittenQuery = fmt.Sprintf("%s (context: %s)", rewrittenQuery, def.Definition)
			}
		}
	}
	
	return rewrittenQuery, nil
}

/* extractTermsFromSQL extracts potential business terms from SQL */
func (qr *QueryRewriter) extractTermsFromSQL(sql string) []string {
	terms := []string{}
	
	// Extract table names (FROM clause)
	fromRegex := regexp.MustCompile(`(?i)FROM\s+(\w+)`)
	matches := fromRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			terms = append(terms, match[1])
		}
	}
	
	// Extract column names (SELECT clause)
	selectRegex := regexp.MustCompile(`(?i)SELECT\s+([\w\s,]+?)\s+FROM`)
	matches = selectRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			columns := strings.Split(match[1], ",")
			for _, col := range columns {
				col = strings.TrimSpace(col)
				if col != "" && !strings.Contains(col, "(") {
					terms = append(terms, col)
				}
			}
		}
	}
	
	// Extract function/aggregate names
	funcRegex := regexp.MustCompile(`(?i)(\w+)\s*\(`)
	matches = funcRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			funcName := match[1]
			if !strings.EqualFold(funcName, "SELECT") && !strings.EqualFold(funcName, "FROM") {
				terms = append(terms, funcName)
			}
		}
	}
	
	return terms
}

/* extractTermsFromNL extracts business terms from natural language */
func (qr *QueryRewriter) extractTermsFromNL(nlQuery string) []string {
	terms := []string{}
	
	// Simple extraction: look for capitalized words and common business terms
	words := strings.Fields(nlQuery)
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && (strings.ToUpper(word) == word || qr.isBusinessTerm(word)) {
			terms = append(terms, word)
		}
	}
	
	return terms
}

/* isBusinessTerm checks if a word might be a business term */
func (qr *QueryRewriter) isBusinessTerm(word string) bool {
	businessKeywords := []string{
		"revenue", "sales", "profit", "cost", "customer", "order", "product",
		"metric", "kpi", "dashboard", "report", "analytics",
	}
	
	wordLower := strings.ToLower(word)
	for _, keyword := range businessKeywords {
		if strings.Contains(wordLower, keyword) {
			return true
		}
	}
	return false
}

/* replaceTermWithSQL replaces a business term with its SQL expression */
func (qr *QueryRewriter) replaceTermWithSQL(sql, term, sqlExpression string) string {
	// Simple replacement - in production, use AST parsing for better accuracy
	termRegex := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(term)))
	return termRegex.ReplaceAllString(sql, fmt.Sprintf("(%s)", sqlExpression))
}

/* RewriteQueryWithMetrics rewrites a query using metric definitions */
func (qr *QueryRewriter) RewriteQueryWithMetrics(ctx context.Context, query string, metricNames []string) (string, error) {
	rewrittenQuery := query
	
	for _, metricName := range metricNames {
		// Get metric by name
		metrics, err := qr.metricsService.ListMetrics(ctx, nil, nil, 100)
		if err != nil {
			continue
		}
		
		for _, metric := range metrics {
			if strings.EqualFold(metric.Name, metricName) || strings.EqualFold(metric.DisplayName, metricName) {
				if metric.Status == "approved" {
					// Replace metric name with its SQL expression
					rewrittenQuery = qr.replaceTermWithSQL(rewrittenQuery, metricName, metric.SQLExpression)
				}
			}
		}
	}
	
	return rewrittenQuery, nil
}

/* GetQuerySemantics extracts semantic information from a query */
func (qr *QueryRewriter) GetQuerySemantics(ctx context.Context, query string) (map[string]interface{}, error) {
	semantics := make(map[string]interface{})
	
	terms := qr.extractTermsFromNL(query)
	if len(terms) == 0 {
		terms = qr.extractTermsFromSQL(query)
	}
	
	matchedDefinitions := []map[string]interface{}{}
	matchedMetrics := []map[string]interface{}{}
	
	for _, term := range terms {
		// Check semantic definitions
		definitions, err := qr.semanticService.SearchSemanticDefinitions(ctx, term, nil, 5)
		if err == nil {
			for _, def := range definitions {
				matchedDefinitions = append(matchedDefinitions, map[string]interface{}{
					"term":       def.Term,
					"definition": def.Definition,
					"sql":        def.SQLExpression,
				})
			}
		}
		
		// Check metrics
		metrics, err := qr.metricsService.ListMetrics(ctx, nil, nil, 100)
		if err == nil {
			for _, metric := range metrics {
				if strings.Contains(strings.ToLower(metric.Name), strings.ToLower(term)) ||
					strings.Contains(strings.ToLower(metric.DisplayName), strings.ToLower(term)) {
					matchedMetrics = append(matchedMetrics, map[string]interface{}{
						"name":        metric.Name,
						"display_name": metric.DisplayName,
						"sql":         metric.SQLExpression,
						"status":      metric.Status,
					})
				}
			}
		}
	}
	
	semantics["matched_definitions"] = matchedDefinitions
	semantics["matched_metrics"] = matchedMetrics
	semantics["extracted_terms"] = terms
	
	return semantics, nil
}
