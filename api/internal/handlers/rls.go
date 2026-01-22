package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* RLSHandler handles row-level security policy requests */
type RLSHandler struct {
	rowSecurityService *auth.RowSecurityService
}

/* NewRLSHandler creates a new RLS handler */
func NewRLSHandler(queries *db.Queries) *RLSHandler {
	return &RLSHandler{
		rowSecurityService: auth.NewRowSecurityService(queries),
	}
}

/* CreateRLSPolicy handles POST /api/v1/governance/rls/policies */
func (h *RLSHandler) CreateRLSPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConnectorID      *uuid.UUID             `json:"connector_id,omitempty"`
		SchemaName       string                 `json:"schema_name"`
		TableName        string                 `json:"table_name"`
		PolicyName       string                 `json:"policy_name"`
		FilterExpression string                 `json:"filter_expression"`
		UserRoles        []string               `json:"user_roles"`
		Metadata         map[string]interface{} `json:"metadata,omitempty"`
		Enabled          bool                   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.SchemaName == "" || req.TableName == "" || req.PolicyName == "" || req.FilterExpression == "" {
		WriteErrorResponse(w, errors.ValidationFailed("schema_name, table_name, policy_name, and filter_expression are required", nil))
		return
	}

	policy := auth.RowSecurityPolicy{
		ConnectorID:      req.ConnectorID,
		SchemaName:       req.SchemaName,
		TableName:        req.TableName,
		PolicyName:       req.PolicyName,
		FilterExpression: req.FilterExpression,
		UserRoles:        req.UserRoles,
		Metadata:         req.Metadata,
		Enabled:          req.Enabled,
	}

	created, err := h.rowSecurityService.CreateRowPolicy(r.Context(), policy)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetRLSPolicies handles GET /api/v1/governance/rls/policies */
func (h *RLSHandler) GetRLSPolicies(w http.ResponseWriter, r *http.Request) {
	var connectorID *uuid.UUID
	if connIDStr := r.URL.Query().Get("connector_id"); connIDStr != "" {
		if id, err := uuid.Parse(connIDStr); err == nil {
			connectorID = &id
		}
	}

	schemaName := r.URL.Query().Get("schema_name")
	tableName := r.URL.Query().Get("table_name")

	if schemaName == "" || tableName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("schema_name and table_name are required", nil))
		return
	}

	policies, err := h.rowSecurityService.GetRowPolicies(r.Context(), connectorID, schemaName, tableName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}
