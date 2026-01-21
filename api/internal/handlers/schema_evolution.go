package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* SchemaEvolutionHandler handles schema evolution requests */
type SchemaEvolutionHandler struct {
	schemaEvolutionService *catalog.SchemaEvolutionService
}

/* NewSchemaEvolutionHandler creates a new schema evolution handler */
func NewSchemaEvolutionHandler(schemaEvolutionService *catalog.SchemaEvolutionService) *SchemaEvolutionHandler {
	return &SchemaEvolutionHandler{schemaEvolutionService: schemaEvolutionService}
}

/* TrackSchemaEvolution handles POST /api/v1/catalog/schema-evolution/track */
func (h *SchemaEvolutionHandler) TrackSchemaEvolution(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConnectorID   uuid.UUID              `json:"connector_id"`
		SchemaName    string                 `json:"schema_name"`
		TableName     string                 `json:"table_name"`
		CurrentSchema map[string]interface{} `json:"current_schema"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	version, err := h.schemaEvolutionService.TrackSchemaEvolution(r.Context(),
		req.ConnectorID, req.SchemaName, req.TableName, req.CurrentSchema)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

/* GetSchemaHistory handles GET /api/v1/catalog/schema-evolution/{connector_id}/{schema_name}/{table_name}/history */
func (h *SchemaEvolutionHandler) GetSchemaHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	schemaName := vars["schema_name"]
	tableName := vars["table_name"]

	history, err := h.schemaEvolutionService.GetSchemaHistory(r.Context(), connectorID, schemaName, tableName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
