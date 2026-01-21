package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/profiling"
)

/* ProfilingHandler handles data profiling requests */
type ProfilingHandler struct {
	service *profiling.Service
}

/* NewProfilingHandler creates a new profiling handler */
func NewProfilingHandler(service *profiling.Service) *ProfilingHandler {
	return &ProfilingHandler{service: service}
}

/* ProfileTable profiles a table */
func (h *ProfilingHandler) ProfileTable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	schemaName := vars["schema_name"]
	tableName := vars["table_name"]

	req := profiling.ProfileRequest{
		ConnectorID: connectorID,
		SchemaName:  schemaName,
		TableName:   tableName,
	}

	result, err := h.service.ProfileTable(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* ProfileColumn profiles a column */
func (h *ProfilingHandler) ProfileColumn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	schemaName := vars["schema_name"]
	tableName := vars["table_name"]
	columnName := vars["column_name"]

	req := profiling.ProfileRequest{
		ConnectorID: connectorID,
		SchemaName:  schemaName,
		TableName:   tableName,
		ColumnName:  &columnName,
	}

	result, err := h.service.ProfileColumn(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
