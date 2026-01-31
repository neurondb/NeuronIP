package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/governance"
)

/* DecisionDashboardsHandler handles decision dashboard requests */
type DecisionDashboardsHandler struct {
	service *governance.DecisionDashboardService
}

/* NewDecisionDashboardsHandler creates a new decision dashboards handler */
func NewDecisionDashboardsHandler(service *governance.DecisionDashboardService) *DecisionDashboardsHandler {
	return &DecisionDashboardsHandler{service: service}
}

/* Create handles POST /api/v1/decision-dashboards */
func (h *DecisionDashboardsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name                string                   `json:"name"`
		Description         string                   `json:"description"`
		OwnerID             string                   `json:"owner_id"`
		WorkspaceID         *uuid.UUID               `json:"workspace_id,omitempty"`
		Layout              []map[string]interface{} `json:"layout"`
		MetricIDs           []uuid.UUID              `json:"metric_ids,omitempty"`
		EvidenceSources     []map[string]interface{} `json:"evidence_sources,omitempty"`
		LineageResourceType *string                  `json:"lineage_resource_type,omitempty"`
		LineageResourceID   *uuid.UUID               `json:"lineage_resource_id,omitempty"`
		WorkflowIDs         []uuid.UUID              `json:"workflow_ids,omitempty"`
		ApprovalWorkflowID  *uuid.UUID               `json:"approval_workflow_id,omitempty"`
		Visibility          string                   `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.OwnerID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name and owner_id are required", nil))
		return
	}
	if req.Layout == nil {
		req.Layout = []map[string]interface{}{}
	}
	d, err := h.service.Create(r.Context(), req.Name, req.Description, req.OwnerID, req.WorkspaceID, req.Layout, req.MetricIDs, req.EvidenceSources, req.LineageResourceType, req.LineageResourceID, req.WorkflowIDs, req.ApprovalWorkflowID, req.Visibility)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, d)
}

/* Get handles GET /api/v1/decision-dashboards/{id} */
func (h *DecisionDashboardsHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid dashboard ID"))
		return
	}
	d, err := h.service.Get(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, d)
}

/* List handles GET /api/v1/decision-dashboards */
func (h *DecisionDashboardsHandler) List(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	ownerID := r.URL.Query().Get("owner_id")
	list, err := h.service.List(r.Context(), workspaceID, ownerID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* RecordRun handles POST /api/v1/decision-dashboards/{id}/runs */
func (h *DecisionDashboardsHandler) RecordRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid dashboard ID"))
		return
	}
	var req struct {
		TriggeredBy string                 `json:"triggered_by"`
		Snapshot    map[string]interface{} `json:"snapshot"`
		Status      string                 `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Snapshot == nil {
		req.Snapshot = map[string]interface{}{}
	}
	run, err := h.service.RecordRun(r.Context(), id, req.TriggeredBy, req.Snapshot, req.Status)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, run)
}
