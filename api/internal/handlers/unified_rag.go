package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/rag"
)

/* UnifiedRAGHandler handles unified RAG service requests */
type UnifiedRAGHandler struct {
	service *rag.UnifiedRAGService
}

/* NewUnifiedRAGHandler creates a new unified RAG handler */
func NewUnifiedRAGHandler(service *rag.UnifiedRAGService) *UnifiedRAGHandler {
	return &UnifiedRAGHandler{service: service}
}

/* PerformRAGRequest represents a RAG request */
type PerformRAGRequest struct {
	Query        string  `json:"query"`
	CollectionID *string `json:"collection_id,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	UseReranking bool   `json:"use_reranking,omitempty"`
	RerankMethod string `json:"rerank_method,omitempty"`
}

/* PerformRAG handles POST /api/v1/rag/query */
func (h *UnifiedRAGHandler) PerformRAG(w http.ResponseWriter, r *http.Request) {
	var req PerformRAGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}

	ragReq := rag.RAGRequest{
		Query:        req.Query,
		CollectionID: req.CollectionID,
		Limit:        req.Limit,
		UseReranking: req.UseReranking,
		RerankMethod: req.RerankMethod,
	}

	result, err := h.service.ExecuteRAGPipeline(r.Context(), ragReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* PerformRAGStream handles POST /api/v1/rag/query/stream */
func (h *UnifiedRAGHandler) PerformRAGStream(w http.ResponseWriter, r *http.Request) {
	var req PerformRAGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Set up streaming response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ragReq := rag.RAGRequest{
		Query:        req.Query,
		CollectionID: req.CollectionID,
		Limit:        req.Limit,
		UseReranking: req.UseReranking,
		RerankMethod: req.RerankMethod,
	}

	// For now, perform regular RAG and stream the result
	// In a full implementation, this would stream intermediate results
	result, err := h.service.ExecuteRAGPipeline(r.Context(), ragReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Stream the result as JSON
	jsonData, _ := json.Marshal(result)
	w.Write([]byte("data: " + string(jsonData) + "\n\n"))
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

/* GetRAGStatus handles GET /api/v1/rag/status */
func (h *UnifiedRAGHandler) GetRAGStatus(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	collectionIDStr := queryParams.Get("collection_id")
	limitStr := queryParams.Get("limit")

	var collectionID *string
	if collectionIDStr != "" {
		collectionID = &collectionIDStr
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	status := map[string]interface{}{
		"collection_id": collectionID,
		"limit":         limit,
		"status":        "ready",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
