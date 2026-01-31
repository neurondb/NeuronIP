package semantic

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
	"github.com/neurondb/NeuronIP/api/internal/policy"
)

const defaultEmbeddingModel = "sentence-transformers/all-MiniLM-L6-v2"

/* PolicyAwareSearchService provides policy-aware semantic search */
type PolicyAwareSearchService struct {
	pool           *pgxpool.Pool
	policyEngine   *policy.PolicyEngine
	neurondbClient *neurondb.Client
}

/* NewPolicyAwareSearchService creates a new policy-aware search service (neurondbClient may be nil) */
func NewPolicyAwareSearchService(pool *pgxpool.Pool, policyEngine *policy.PolicyEngine, neurondbClient *neurondb.Client) *PolicyAwareSearchService {
	return &PolicyAwareSearchService{
		pool:           pool,
		policyEngine:   policyEngine,
		neurondbClient: neurondbClient,
	}
}

/* PolicyAwareSearchRequest represents a policy-aware search request */
type PolicyAwareSearchRequest struct {
	Query         string
	CollectionID  *string
	UserID        string
	TenantID      *string
	Limit         int
	Threshold     float64
	ApplyPolicies bool
}

/* Search performs policy-aware semantic search */
func (pass *PolicyAwareSearchService) Search(ctx context.Context, req PolicyAwareSearchRequest) ([]map[string]interface{}, error) {
	// Generate embedding for query
	queryEmbedding, err := pass.generateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Build base search query
	baseQuery := pass.buildBaseSearchQuery(req, queryEmbedding)

	// Apply policy filters if enabled
	if req.ApplyPolicies {
		policyFilter, err := pass.buildPolicyFilter(ctx, req)
		if err == nil && policyFilter != "" {
			baseQuery = pass.addPolicyFilter(baseQuery, policyFilter)
		}
	}

	// Execute search
	rows, err := pass.pool.Query(ctx, baseQuery, queryEmbedding, req.Threshold, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	fieldDescriptions := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, desc := range fieldDescriptions {
			row[desc.Name] = values[i]
		}

		// Additional policy check on results
		if req.ApplyPolicies {
			if allowed, _ := pass.checkResultAccess(ctx, row, req); !allowed {
				continue
			}
		}

		results = append(results, row)
	}

	return results, nil
}

/* buildBaseSearchQuery builds the base semantic search query */
func (pass *PolicyAwareSearchService) buildBaseSearchQuery(req PolicyAwareSearchRequest, queryEmbedding string) string {
	var query strings.Builder

	query.WriteString(`
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
	`)

	if req.CollectionID != nil {
		query.WriteString(" AND kd.collection_id = $4")
	}

	query.WriteString(" ORDER BY ke.embedding <=> $1::vector LIMIT $3")

	return query.String()
}

/* buildPolicyFilter builds a policy filter for the search */
func (pass *PolicyAwareSearchService) buildPolicyFilter(ctx context.Context, req PolicyAwareSearchRequest) (string, error) {
	if pass.policyEngine == nil {
		return "", nil
	}

	// Create policy request
	policyReq := policy.PolicyRequest{
		ResourceType: "knowledge_document",
		UserID:       req.UserID,
		Action:       "read",
	}

	// Evaluate policies
	result, err := pass.policyEngine.EvaluatePolicies(ctx, policyReq)
	if err != nil {
		return "", err
	}

	// Build filter from policy result
	if !result.Allowed {
		return "1=0", nil // No results allowed
	}

	if result.Filtered && result.ModifiedRequest != nil && result.ModifiedRequest.QueryFilters != nil {
		// Build filter from query filters
		// In production, convert QueryFilters to SQL WHERE clause
		return "", nil
	}

	return "", nil
}

/* addPolicyFilter adds policy filter to query */
func (pass *PolicyAwareSearchService) addPolicyFilter(query, filter string) string {
	// Insert filter before ORDER BY
	orderByIndex := strings.Index(query, "ORDER BY")
	if orderByIndex > 0 {
		return query[:orderByIndex] + " AND " + filter + " " + query[orderByIndex:]
	}
	return query + " AND " + filter
}

/* checkResultAccess checks if user has access to a result */
func (pass *PolicyAwareSearchService) checkResultAccess(ctx context.Context, result map[string]interface{}, req PolicyAwareSearchRequest) (bool, error) {
	if pass.policyEngine == nil {
		return true, nil
	}

	// Extract document ID
	docID, ok := result["id"].(string)
	if !ok {
		return true, nil
	}

	// Create policy request
	policyReq := policy.PolicyRequest{
		ResourceType: "knowledge_document",
		ResourceID:   docID,
		UserID:       req.UserID,
		Action:       "read",
	}

	policyResult, err := pass.policyEngine.EvaluatePolicies(ctx, policyReq)
	if err != nil {
		return false, err
	}

	return policyResult.Allowed, nil
}

/* generateEmbedding generates embedding for query using NeuronDB when available */
func (pass *PolicyAwareSearchService) generateEmbedding(ctx context.Context, text string) (string, error) {
	if pass.neurondbClient == nil {
		return "", fmt.Errorf("embedding not available: NeuronDB client not configured or disabled")
	}
	return pass.neurondbClient.GenerateEmbedding(ctx, text, defaultEmbeddingModel)
}
