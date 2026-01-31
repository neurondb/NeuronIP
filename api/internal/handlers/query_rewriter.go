package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
)

/* QueryRewriterHandler handles query rewriting requests */
type QueryRewriterHandler struct {
	queryRewriter *semantic.QueryRewriter
}

/* NewQueryRewriterHandler creates a new query rewriter handler */
func NewQueryRewriterHandler(queryRewriter *semantic.QueryRewriter) *QueryRewriterHandler {
	return &QueryRewriterHandler{queryRewriter: queryRewriter}
}

/* RewriteQueryRequest represents a query rewrite request */
type RewriteQueryRequest struct {
	Query     string `json:"query"`
	QueryType string `json:"query_type"` // "sql" or "nl"
}

/* RewriteQuery handles query rewriting */
func (h *QueryRewriterHandler) RewriteQuery(w http.ResponseWriter, r *http.Request) {
	var req RewriteQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query is required", nil))
		return
	}

	if req.QueryType == "" {
		req.QueryType = "sql"
	}

	rewrittenQuery, err := h.queryRewriter.RewriteQuery(r.Context(), req.Query, req.QueryType)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Get semantics
	semantics, _ := h.queryRewriter.GetQuerySemantics(r.Context(), req.Query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"original_query":  req.Query,
		"rewritten_query": rewrittenQuery,
		"semantics":       semantics,
	})
}

/* GetQuerySemantics handles semantics extraction */
func (h *QueryRewriterHandler) GetQuerySemantics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query parameter is required", nil))
		return
	}

	semantics, err := h.queryRewriter.GetQuerySemantics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semantics)
}
