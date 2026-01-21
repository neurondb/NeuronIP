package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
)

/* ColumnLineageHandler handles column-level lineage requests */
type ColumnLineageHandler struct {
	service *lineage.ColumnLineageService
}

/* NewColumnLineageHandler creates a new column lineage handler */
func NewColumnLineageHandler(service *lineage.ColumnLineageService) *ColumnLineageHandler {
	return &ColumnLineageHandler{service: service}
}

/* GetColumnLineage handles GET /api/v1/lineage/columns/{connector_id}/{schema_name}/{table_name}/{column_name} */
func (h *ColumnLineageHandler) GetColumnLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorIDStr := vars["connector_id"]
	schemaName := vars["schema_name"]
	tableName := vars["table_name"]
	columnName := vars["column_name"]

	if connectorIDStr == "" || schemaName == "" || tableName == "" || columnName == "" {
		WriteErrorResponse(w, errors.BadRequest("connector_id, schema_name, table_name, and column_name are required"))
		return
	}

	connectorID, err := uuid.Parse(connectorIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("invalid connector_id format"))
		return
	}

	graph, err := h.service.GetColumnLineage(r.Context(), connectorID, schemaName, tableName, columnName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

/* TrackColumnLineage handles POST /api/v1/lineage/columns/track */
func (h *ColumnLineageHandler) TrackColumnLineage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceConnectorID *uuid.UUID             `json:"source_connector_id,omitempty"`
		SourceSchemaName  string                 `json:"source_schema_name"`
		SourceTableName   string                 `json:"source_table_name"`
		SourceColumnName  string                 `json:"source_column_name"`
		TargetConnectorID *uuid.UUID             `json:"target_connector_id,omitempty"`
		TargetSchemaName  string                 `json:"target_schema_name"`
		TargetTableName   string                 `json:"target_table_name"`
		TargetColumnName  string                 `json:"target_column_name"`
		EdgeType          string                 `json:"edge_type"`
		Transformation    map[string]interface{} `json:"transformation,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.SourceSchemaName == "" || req.SourceTableName == "" || req.SourceColumnName == "" ||
		req.TargetSchemaName == "" || req.TargetTableName == "" || req.TargetColumnName == "" ||
		req.EdgeType == "" {
		WriteErrorResponse(w, errors.BadRequest("source and target schema/table/column names and edge_type are required"))
		return
	}

	edge, err := h.service.TrackColumnLineage(r.Context(),
		req.SourceConnectorID, req.SourceSchemaName, req.SourceTableName, req.SourceColumnName,
		req.TargetConnectorID, req.TargetSchemaName, req.TargetTableName, req.TargetColumnName,
		req.EdgeType, req.Transformation)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "tracked",
		"edge":   edge,
	})
}

/* CreateColumnNode handles POST /api/v1/lineage/columns/nodes */
func (h *ColumnLineageHandler) CreateColumnNode(w http.ResponseWriter, r *http.Request) {
	var node lineage.ColumnLineageNode
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if node.SchemaName == "" || node.TableName == "" || node.ColumnName == "" {
		WriteErrorResponse(w, errors.BadRequest("schema_name, table_name, and column_name are required"))
		return
	}

	createdNode, err := h.service.CreateColumnNode(r.Context(), node)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdNode)
}
