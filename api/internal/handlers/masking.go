package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/masking"
)

/* MaskingHandler handles masking requests */
type MaskingHandler struct {
	service *masking.MaskingService
}

/* NewMaskingHandler creates a new masking handler */
func NewMaskingHandler(service *masking.MaskingService) *MaskingHandler {
	return &MaskingHandler{service: service}
}

/* CreateMaskingPolicy handles POST /api/v1/masking/policies */
func (h *MaskingHandler) CreateMaskingPolicy(w http.ResponseWriter, r *http.Request) {
	var policy masking.MaskingPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateMaskingPolicy(r.Context(), policy)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetMaskingPolicy handles GET /api/v1/masking/policies */
func (h *MaskingHandler) GetMaskingPolicy(w http.ResponseWriter, r *http.Request) {
	connectorIDStr := r.URL.Query().Get("connector_id")
	schemaName := r.URL.Query().Get("schema_name")
	tableName := r.URL.Query().Get("table_name")
	columnName := r.URL.Query().Get("column_name")

	if schemaName == "" || tableName == "" || columnName == "" {
		WriteErrorResponse(w, errors.BadRequest("schema_name, table_name, and column_name are required"))
		return
	}

	var connectorID *uuid.UUID
	if connectorIDStr != "" {
		id, err := uuid.Parse(connectorIDStr)
		if err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid connector_id"))
			return
		}
		connectorID = &id
	}

	policy, err := h.service.GetMaskingPolicy(r.Context(), connectorID, schemaName, tableName, columnName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

/* ApplyMasking handles POST /api/v1/masking/apply */
func (h *MaskingHandler) ApplyMasking(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserRole    string      `json:"user_role"`
		ConnectorID *uuid.UUID  `json:"connector_id,omitempty"`
		SchemaName  string      `json:"schema_name"`
		TableName   string      `json:"table_name"`
		ColumnName  string      `json:"column_name"`
		Value       interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	masked, err := h.service.ApplyMasking(r.Context(), req.UserRole, req.ConnectorID, req.SchemaName, req.TableName, req.ColumnName, req.Value)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"original_value": req.Value,
		"masked_value":   masked,
	})
}
