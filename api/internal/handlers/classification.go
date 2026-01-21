package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/classification"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ClassificationHandler handles classification requests */
type ClassificationHandler struct {
	service *classification.Service
}

/* NewClassificationHandler creates a new classification handler */
func NewClassificationHandler(service *classification.Service) *ClassificationHandler {
	return &ClassificationHandler{service: service}
}

/* ClassifyColumn classifies a column */
func (h *ClassificationHandler) ClassifyColumn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	schemaName := vars["schema_name"]
	tableName := vars["table_name"]
	columnName := vars["column_name"]

	result, err := h.service.ClassifyColumn(r.Context(), connectorID, schemaName, tableName, columnName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* ClassifyConnector classifies all columns in a connector */
func (h *ClassificationHandler) ClassifyConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	if err := h.service.ClassifyConnector(r.Context(), connectorID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "Classification started for connector",
	})
}

/* CreateClassificationRule creates a classification rule */
func (h *ClassificationHandler) CreateClassificationRule(w http.ResponseWriter, r *http.Request) {
	var rule classification.ClassificationRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateRule(r.Context(), rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
