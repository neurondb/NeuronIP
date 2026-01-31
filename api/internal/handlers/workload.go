package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/execution"
)

/* WorkloadHandler handles workload queue requests */
type WorkloadHandler struct {
	service *execution.WorkloadService
}

/* NewWorkloadHandler creates a new workload handler */
func NewWorkloadHandler(service *execution.WorkloadService) *WorkloadHandler {
	return &WorkloadHandler{service: service}
}

/* ListQueues handles GET /api/v1/warehouse/workload/queues */
func (h *WorkloadHandler) ListQueues(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	list, err := h.service.ListQueues(r.Context(), workspaceID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* CreateQueue handles POST /api/v1/warehouse/workload/queues */
func (h *WorkloadHandler) CreateQueue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name                string     `json:"name"`
		Description         string     `json:"description"`
		Priority            int        `json:"priority"`
		MaxConcurrency      int        `json:"max_concurrency"`
		QueryTimeoutSeconds int        `json:"query_timeout_seconds"`
		WorkspaceID         *uuid.UUID `json:"workspace_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}
	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 10
	}
	if req.QueryTimeoutSeconds <= 0 {
		req.QueryTimeoutSeconds = 300
	}
	cfg, err := h.service.CreateQueue(r.Context(), req.Name, req.Description, req.Priority, req.MaxConcurrency, req.QueryTimeoutSeconds, req.WorkspaceID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, cfg)
}

/* GetQueueByName handles GET /api/v1/warehouse/workload/queues/{name} */
func (h *WorkloadHandler) GetQueueByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	cfg, err := h.service.GetQueueByName(r.Context(), name, workspaceID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, cfg)
}
