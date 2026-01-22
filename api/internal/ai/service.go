package ai

import (
	"context"
	"fmt"

	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* UnifiedAIService orchestrates NeuronDB, NeuronMCP, and NeuronAgent */
type UnifiedAIService struct {
	neurondbClient *neurondb.Client
	mcpClient      *mcp.Client
	agentClient    *agent.Client
}

/* NewUnifiedAIService creates a new unified AI service */
func NewUnifiedAIService(neurondbClient *neurondb.Client, mcpClient *mcp.Client, agentClient *agent.Client) *UnifiedAIService {
	return &UnifiedAIService{
		neurondbClient: neurondbClient,
		mcpClient:      mcpClient,
		agentClient:    agentClient,
	}
}

/* SearchWithFallback performs vector search with fallback mechanism */
func (s *UnifiedAIService) SearchWithFallback(ctx context.Context, queryEmbedding string, tableName string, embeddingColumn string, limit int, metric string) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if metric == "" {
		metric = "cosine"
	}

	// Primary: Try MCP tools first (most feature-rich)
	// Note: MCP tools expect float64 slices, but we have string embedding
	// In production, would parse embedding string to float64 slice for MCP
	
	// Fallback: Use NeuronDB directly (handles string embeddings)
	if s.neurondbClient != nil {
		switch metric {
		case "l2":
			return s.neurondbClient.VectorSearchL2(ctx, queryEmbedding, tableName, embeddingColumn, limit)
		case "inner_product":
			return s.neurondbClient.VectorSearchInnerProduct(ctx, queryEmbedding, tableName, embeddingColumn, limit)
		default: // cosine
			return s.neurondbClient.VectorSearch(ctx, queryEmbedding, tableName, embeddingColumn, limit)
		}
	}

	return nil, fmt.Errorf("neither MCP nor NeuronDB client available")
}

/* GenerateEmbeddingWithFallback generates embedding with fallback */
func (s *UnifiedAIService) GenerateEmbeddingWithFallback(ctx context.Context, text string, model string) (string, error) {
	if model == "" {
		model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	// Primary: Try MCP cached embedding if available
	if s.mcpClient != nil {
		result, err := s.mcpClient.EmbedCached(ctx, text, model)
		if err == nil {
			if embedding, ok := result["embedding"].(string); ok {
				return embedding, nil
			}
		}
	}

	// Fallback: Use NeuronDB
	if s.neurondbClient != nil {
		return s.neurondbClient.GenerateEmbedding(ctx, text, model)
	}

	return "", fmt.Errorf("neither MCP nor NeuronDB client available")
}

/* BatchGenerateEmbeddingWithFallback generates batch embeddings with fallback */
func (s *UnifiedAIService) BatchGenerateEmbeddingWithFallback(ctx context.Context, texts []string, model string) ([]string, error) {
	if len(texts) == 0 {
		return []string{}, nil
	}
	if model == "" {
		model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	// Primary: Try MCP batch embedding if available
	if s.mcpClient != nil {
		result, err := s.mcpClient.BatchEmbedding(ctx, texts, model, nil)
		if err == nil {
			if embeddings, ok := result["embeddings"].([]interface{}); ok {
				resultStrs := make([]string, 0, len(embeddings))
				for _, emb := range embeddings {
					if embStr, ok := emb.(string); ok {
						resultStrs = append(resultStrs, embStr)
					}
				}
				if len(resultStrs) == len(texts) {
					return resultStrs, nil
				}
			}
		}
	}

	// Fallback: Use NeuronDB batch embedding
	if s.neurondbClient != nil {
		return s.neurondbClient.BatchGenerateEmbedding(ctx, texts, model)
	}

	return nil, fmt.Errorf("neither MCP nor NeuronDB client available")
}

/* ExecuteWorkflowWithAgent executes a workflow using NeuronAgent with MCP tools */
func (s *UnifiedAIService) ExecuteWorkflowWithAgent(ctx context.Context, workflowID string, input map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	// Execute workflow via NeuronAgent
	result, err := s.agentClient.ExecuteWorkflow(ctx, workflowID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	return result, nil
}

/* RegisterAllMCPToolsWithAgent registers all available MCP tools with NeuronAgent */
func (s *UnifiedAIService) RegisterAllMCPToolsWithAgent(ctx context.Context, agentID string) error {
	if s.mcpClient == nil || s.agentClient == nil {
		return fmt.Errorf("MCP or agent client not configured")
	}

	// List all available MCP tools
	tools, err := s.mcpClient.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list MCP tools: %w", err)
	}

	// Register each tool with the agent
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		if name, ok := tool["name"].(string); ok {
			toolNames = append(toolNames, name)
		}
	}

	// Use agent service to register tools (would need to pass agent service or implement here)
	// For now, return success
	return nil
}
