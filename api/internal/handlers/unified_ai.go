package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/ai"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* UnifiedAIHandler handles unified AI service requests */
type UnifiedAIHandler struct {
	service *ai.UnifiedAIService
}

/* NewUnifiedAIHandler creates a new unified AI handler */
func NewUnifiedAIHandler(service *ai.UnifiedAIService) *UnifiedAIHandler {
	return &UnifiedAIHandler{service: service}
}

/* GenerateEmbeddingRequest represents an embedding generation request */
type GenerateEmbeddingRequest struct {
	Text  string `json:"text"`
	Model string `json:"model,omitempty"`
}

/* GenerateEmbeddingResponse represents an embedding generation response */
type GenerateEmbeddingResponse struct {
	Embedding string `json:"embedding"`
	Model     string `json:"model"`
}

/* GenerateEmbedding handles POST /api/v1/ai/embedding */
func (h *UnifiedAIHandler) GenerateEmbedding(w http.ResponseWriter, r *http.Request) {
	var req GenerateEmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Text == "" {
		WriteErrorResponse(w, errors.ValidationFailed("text is required", nil))
		return
	}

	if req.Model == "" {
		req.Model = "sentence-transformers/all-MiniLM-L6-v2"
	}

	embedding, err := h.service.GenerateEmbeddingWithFallback(r.Context(), req.Text, req.Model)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateEmbeddingResponse{
		Embedding: embedding,
		Model:     req.Model,
	})
}

/* ExecuteWorkflowRequest represents a workflow execution request */
type ExecuteWorkflowRequest struct {
	WorkflowID string                 `json:"workflow_id"`
	Input      map[string]interface{} `json:"input"`
}

/* ExecuteWorkflow handles POST /api/v1/ai/workflow */
func (h *UnifiedAIHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	var req ExecuteWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.WorkflowID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("workflow_id is required", nil))
		return
	}

	result, err := h.service.ExecuteWorkflowWithAgent(r.Context(), req.WorkflowID, req.Input)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* RegisterToolsRequest represents a tool registration request */
type RegisterToolsRequest struct {
	AgentID string `json:"agent_id"`
}

/* RegisterTools handles POST /api/v1/ai/register-tools */
func (h *UnifiedAIHandler) RegisterTools(w http.ResponseWriter, r *http.Request) {
	var req RegisterToolsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.AgentID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("agent_id is required", nil))
		return
	}

	err := h.service.RegisterAllMCPToolsWithAgent(r.Context(), req.AgentID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
