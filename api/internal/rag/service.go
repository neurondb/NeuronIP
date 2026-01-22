package rag

import (
	"context"
	"fmt"

	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* UnifiedRAGService provides unified RAG pipeline using all three components */
type UnifiedRAGService struct {
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
	agentClient    *agent.Client
}

/* NewUnifiedRAGService creates a new unified RAG service */
func NewUnifiedRAGService(neurondbClient *neurondb.Client, mcpClient *mcp.Client, agentClient *agent.Client) *UnifiedRAGService {
	return &UnifiedRAGService{
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
		agentClient:    agentClient,
	}
}

/* RAGRequest represents a RAG pipeline request */
type RAGRequest struct {
	Query        string
	CollectionID *string
	Limit        int
	UseReranking bool
	RerankMethod string // "cross_encoder", "llm", "cohere", "ensemble"
	DistanceMetric string // "cosine", "l2", "inner_product"
	Threshold   float64
}

/* RAGResult represents a RAG pipeline result */
type RAGResult struct {
	Answer     string
	Context    []string
	Sources    []map[string]interface{}
	Citations  []string
	Confidence float64
}

/* ExecuteRAGPipeline executes the unified RAG pipeline */
func (s *UnifiedRAGService) ExecuteRAGPipeline(ctx context.Context, req RAGRequest) (*RAGResult, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Step 1: NeuronAgent - Understand query intent
	var enhancedQuery string
	if s.agentClient != nil {
		// Use agent to enhance/understand the query
		enhancedQuery = req.Query // For now, use as-is
		// In production, could use agent to expand/refine query
	} else {
		enhancedQuery = req.Query
	}

	// Step 2: Generate query embedding using NeuronDB
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, enhancedQuery, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Step 3: NeuronMCP/NeuronDB - Vector search with reranking
	var documents []string
	var sources []map[string]interface{}

	// Determine distance metric
	distanceMetric := req.DistanceMetric
	if distanceMetric == "" {
		distanceMetric = "cosine"
	}

	// Use appropriate search method
	if s.mcpClient != nil {
		// Try MCP hybrid search first
		result, err := s.mcpClient.HybridSearch(ctx, enhancedQuery, "neuronip.knowledge_documents", "embedding", "content", req.Limit*2, nil)
		if err == nil {
			if docs, ok := result["documents"].([]interface{}); ok {
				for _, doc := range docs {
					if docMap, ok := doc.(map[string]interface{}); ok {
						if content, ok := docMap["content"].(string); ok {
							documents = append(documents, content)
							sources = append(sources, docMap)
						}
					}
				}
			}
		}
	}

	// Fallback to NeuronDB vector search if MCP didn't return results
	if len(documents) == 0 && s.neurondbClient != nil {
		var results []map[string]interface{}
		var err error

		switch distanceMetric {
		case "l2":
			results, err = s.neurondbClient.VectorSearchL2(ctx, queryEmbedding, "neuronip.knowledge_embeddings", "embedding", req.Limit*2)
		case "inner_product":
			results, err = s.neurondbClient.VectorSearchInnerProduct(ctx, queryEmbedding, "neuronip.knowledge_embeddings", "embedding", req.Limit*2)
		default: // cosine
			results, err = s.neurondbClient.VectorSearch(ctx, queryEmbedding, "neuronip.knowledge_embeddings", "embedding", req.Limit*2)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to perform vector search: %w", err)
		}

		for _, result := range results {
			if content, ok := result["chunk_text"].(string); ok {
				documents = append(documents, content)
				sources = append(sources, result)
			} else if content, ok := result["content"].(string); ok {
				documents = append(documents, content)
				sources = append(sources, result)
			}
		}
	}

	// Step 4: Rerank results using MCP
	if req.UseReranking && s.mcpClient != nil && len(documents) > 0 {
		rerankMethod := req.RerankMethod
		if rerankMethod == "" {
			rerankMethod = "cross_encoder"
		}

		var rerankedResult map[string]interface{}
		var err error

		switch rerankMethod {
		case "llm":
			rerankedResult, err = s.mcpClient.RerankLLM(ctx, enhancedQuery, documents, req.Limit, "")
		case "cohere":
			rerankedResult, err = s.mcpClient.RerankCohere(ctx, enhancedQuery, documents, req.Limit)
		case "ensemble":
			rerankedResult, err = s.mcpClient.RerankEnsemble(ctx, enhancedQuery, documents, req.Limit, []string{"cross_encoder", "llm"}, nil)
		default: // cross_encoder
			rerankedResult, err = s.mcpClient.RerankCrossEncoder(ctx, enhancedQuery, documents, req.Limit)
		}

		if err == nil {
			if rerankedDocs, ok := rerankedResult["documents"].([]interface{}); ok {
				documents = make([]string, 0, len(rerankedDocs))
				for _, doc := range rerankedDocs {
					if docMap, ok := doc.(map[string]interface{}); ok {
						if content, ok := docMap["content"].(string); ok {
							documents = append(documents, content)
						}
					}
				}
			}
		}
	}

	// Limit to requested limit
	if len(documents) > req.Limit {
		documents = documents[:req.Limit]
		sources = sources[:req.Limit]
	}

	// Step 5: NeuronDB - Retrieve context
	context := documents

	// Step 6: NeuronAgent - Generate response
	var answer string
	var citations []string

	if s.agentClient != nil {
		// Use agent to generate response with context
		reply, err := s.agentClient.GenerateReply(ctx, convertToStringMaps(sources), enhancedQuery)
		if err == nil {
			answer = reply
		}
	}

	// Fallback to NeuronDB if agent fails
	if answer == "" {
		answer, err = s.neurondbClient.GenerateResponse(ctx, enhancedQuery, context, "sentence-transformers/all-MiniLM-L6-v2")
		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}
	}

	// Step 7: NeuronMCP - Generate citations
	if s.mcpClient != nil && len(sources) > 0 {
		citationResult, err := s.mcpClient.AnswerWithCitations(ctx, enhancedQuery, sources, "sentence-transformers/all-MiniLM-L6-v2")
		if err == nil {
			if cites, ok := citationResult["citations"].([]interface{}); ok {
				for _, cite := range cites {
					if citeStr, ok := cite.(string); ok {
						citations = append(citations, citeStr)
					}
				}
			} else if answer, ok := citationResult["answer"].(string); ok {
				// If citations not in expected format, extract from answer
				answer = answer // Could parse citations from answer text
			}
		}
	}

	return &RAGResult{
		Answer:     answer,
		Context:    context,
		Sources:    sources,
		Citations:  citations,
		Confidence: 0.8, // Could be calculated from similarity scores
	}, nil
}

/* convertToStringMaps converts []map[string]interface{} to format expected by agent */
func convertToStringMaps(sources []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(sources))
	for i, source := range sources {
		result[i] = source
	}
	return result
}
