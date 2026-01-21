package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/connectors"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ConnectorHandler handles connector requests */
type ConnectorHandler struct {
	service *connectors.ConnectorService
}

/* NewConnectorHandler creates a new connector handler */
func NewConnectorHandler(service *connectors.ConnectorService) *ConnectorHandler {
	return &ConnectorHandler{service: service}
}

/* CreateConnector creates a new connector */
func (h *ConnectorHandler) CreateConnector(w http.ResponseWriter, r *http.Request) {
	var connector connectors.DataSourceConnector
	if err := json.NewDecoder(r.Body).Decode(&connector); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateConnector(r.Context(), connector)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetConnector retrieves a connector */
func (h *ConnectorHandler) GetConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	connector, err := h.service.GetConnector(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connector)
}

/* ListConnectors lists all connectors */
func (h *ConnectorHandler) ListConnectors(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"
	
	connectorList, err := h.service.ListConnectors(r.Context(), enabledOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connectorList)
}

/* SyncConnector synchronizes a connector */
func (h *ConnectorHandler) SyncConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid connector ID"))
		return
	}

	var req struct {
		SyncType string `json:"sync_type"` // full, incremental, schema_only
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.SyncType = "full" // Default
	}

	// Run sync in background
	go func() {
		h.service.SyncConnector(r.Context(), id, req.SyncType)
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "sync_started",
		"message": "Connector sync started",
	})
}
