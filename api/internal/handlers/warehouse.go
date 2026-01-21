package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* WarehouseHandler handles warehouse query requests */
type WarehouseHandler struct {
	service *warehouse.Service
}

/* NewWarehouseHandler creates a new warehouse handler */
func NewWarehouseHandler(service *warehouse.Service) *WarehouseHandler {
	return &WarehouseHandler{service: service}
}

/* QueryRequest represents a warehouse query request */
type QueryRequest struct {
	Query         string                 `json:"query"`
	SchemaID      *uuid.UUID             `json:"schema_id,omitempty"`
	SemanticQuery *string                `json:"semantic_query,omitempty"`
	SQLFilters    map[string]interface{} `json:"sql_filters,omitempty"`
}

/* Query handles warehouse query execution requests */
func (h *WarehouseHandler) Query(w http.ResponseWriter, r *http.Request) {
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	// Extract userID from context
	var userID *string
	if uid, ok := auth.GetUserIDFromContext(r.Context()); ok {
		userID = &uid
	}

	warehouseReq := warehouse.QueryRequest{
		Query:         req.Query,
		SchemaID:      req.SchemaID,
		UserID:        userID,
		SemanticQuery: req.SemanticQuery,
		SQLFilters:    req.SQLFilters,
	}

	result, err := h.service.ExecuteQuery(r.Context(), warehouseReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Increment business metric
	metrics.IncrementWarehouseQuery("completed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetQuery handles warehouse query retrieval requests */
func (h *WarehouseHandler) GetQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid query ID"))
		return
	}

	query, err := h.service.GetQuery(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(query)
}

/* ListSchemasRequest represents a list schemas request */
type ListSchemasRequest struct {
	// No specific filters for now
}

/* ListSchemas handles schema listing requests */
func (h *WarehouseHandler) ListSchemas(w http.ResponseWriter, r *http.Request) {
	schemas, err := h.service.ListSchemas(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schemas": schemas,
		"count":   len(schemas),
	})
}

/* CreateSchemaRequest represents a schema creation request */
type CreateSchemaRequest struct {
	SchemaName   string                 `json:"schema_name"`
	DatabaseName string                 `json:"database_name"`
	Description  *string                `json:"description,omitempty"`
	Tables       []map[string]interface{} `json:"tables,omitempty"`
}

/* CreateSchema handles schema creation requests */
func (h *WarehouseHandler) CreateSchema(w http.ResponseWriter, r *http.Request) {
	var req CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.SchemaName == "" || req.DatabaseName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("schema_name and database_name are required", nil))
		return
	}

	if req.Tables == nil {
		req.Tables = []map[string]interface{}{}
	}

	schema, err := h.service.CreateSchema(r.Context(), req.SchemaName, req.DatabaseName, req.Description, req.Tables)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schema)
}

/* GetSchema handles schema retrieval requests */
func (h *WarehouseHandler) GetSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid schema ID"))
		return
	}

	schema, err := h.service.GetSchema(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

/* GetQueryHistory handles query history retrieval requests */
func (h *WarehouseHandler) GetQueryHistory(w http.ResponseWriter, r *http.Request) {
	var userID string
	if uid, ok := auth.GetUserIDFromContext(r.Context()); ok {
		userID = uid
	} else {
		WriteErrorResponse(w, errors.Unauthorized("User ID required"))
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	history, err := h.service.GetQueryHistory(r.Context(), userID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

/* GetQueryOptimization handles query optimization suggestion requests */
func (h *WarehouseHandler) GetQueryOptimization(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.SQL == "" {
		WriteErrorResponse(w, errors.ValidationFailed("sql is required", nil))
		return
	}

	suggestions, err := h.service.GetQueryOptimizationSuggestions(r.Context(), req.SQL)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}