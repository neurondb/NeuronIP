package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
)

/* SemanticHandler handles semantic search requests */
type SemanticHandler struct {
	service *semantic.Service
}

/* NewSemanticHandler creates a new semantic handler */
func NewSemanticHandler(service *semantic.Service) *SemanticHandler {
	return &SemanticHandler{service: service}
}

/* SearchRequest represents the search request body */
type SearchRequest struct {
	Query        string     `json:"query"`
	CollectionID *uuid.UUID `json:"collection_id,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Threshold    float64    `json:"threshold,omitempty"`
}

/* Search handles semantic search requests */
func (h *SemanticHandler) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponseWithContext(w, r, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponseWithContext(w, r, errors.ValidationFailed("Query is required", nil))
		return
	}

	results, err := h.service.Search(r.Context(), semantic.SearchRequest{
		Query: req.Query,
		CollectionID: req.CollectionID,
		Limit:        req.Limit,
		Threshold:    req.Threshold,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	// Increment business metric
	metrics.IncrementSemanticSearches()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

/* CreateDocumentRequest represents document creation request with optional chunking config */
type CreateDocumentRequest struct {
	Document      db.KnowledgeDocument       `json:"document"`
	ChunkingConfig *semantic.ChunkingConfig  `json:"chunking_config,omitempty"`
}

/* CreateDocument handles document creation requests */
func (h *SemanticHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Try to decode as just document for backward compatibility
		// Note: This won't work since body is consumed, so we need to handle differently
		// For now, assume the request format is correct
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	doc := req.Document

	if doc.Title == "" || doc.Content == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Title and content are required", nil))
		return
	}

	if doc.ContentType == "" {
		doc.ContentType = "document"
	}

	// Extract userID from context
	var userID *string
	if uid, ok := auth.GetUserIDFromContext(r.Context()); ok {
		userID = &uid
	}

	if err := h.service.CreateDocument(r.Context(), &doc, req.ChunkingConfig, userID); err != nil {
		WriteError(w, err)
		return
	}

	// Increment business metric
	metrics.IncrementDocumentsCreated()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

/* RAGRequest represents a RAG pipeline request body */
type RAGRequest struct {
	Query        string     `json:"query"`
	CollectionID *uuid.UUID `json:"collection_id,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Threshold    float64    `json:"threshold,omitempty"`
	MaxContext   int        `json:"max_context,omitempty"`
}

/* RAG handles RAG pipeline requests */
func (h *SemanticHandler) RAG(w http.ResponseWriter, r *http.Request) {
	var req RAGRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("Query is required", nil))
		return
	}

	result, err := h.service.RAG(r.Context(), semantic.RAGRequest{
		Query:        req.Query,
		CollectionID: req.CollectionID,
		Limit:        req.Limit,
		Threshold:    req.Threshold,
		MaxContext:   req.MaxContext,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	// Increment business metric
	metrics.IncrementSemanticSearches()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetCollection handles collection retrieval */
func (h *SemanticHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid collection ID"))
		return
	}

	collection, err := h.service.GetCollection(r.Context(), id)
	if err != nil {
		// Check if it's a not found error
		var apiErr *errors.APIError
		if errors.IsAPIError(err) {
			apiErr = errors.AsAPIError(err)
			if apiErr.Code == errors.ErrCodeNotFound {
				WriteErrorResponse(w, errors.NotFound("Collection"))
				return
			}
		}
		// Check for pgx.ErrNoRows in wrapped errors or "not found" in error message
		if err == pgx.ErrNoRows || strings.Contains(err.Error(), "not found") {
			WriteErrorResponse(w, errors.NotFound("Collection"))
			return
		}
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collection)
}

/* UpdateDocumentRequest represents document update request */
type UpdateDocumentRequest struct {
	Document      db.KnowledgeDocument       `json:"document"`
	ChunkingConfig *semantic.ChunkingConfig  `json:"chunking_config,omitempty"`
}

/* UpdateDocument handles document update requests */
func (h *SemanticHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid document ID"))
		return
	}

	var req UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Extract userID from context
	var userID *string
	if uid, ok := auth.GetUserIDFromContext(r.Context()); ok {
		userID = &uid
	}

	if err := h.service.UpdateDocument(r.Context(), docID, &req.Document, req.ChunkingConfig, userID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Document updated successfully",
		"document_id": docID,
	})
}
