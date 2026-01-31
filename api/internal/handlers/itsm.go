package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/itsm"
)

/* ITSMHandler handles ITSM (incidents, changes, runbooks) requests */
type ITSMHandler struct {
	service *itsm.Service
}

/* NewITSMHandler creates a new ITSM handler */
func NewITSMHandler(service *itsm.Service) *ITSMHandler {
	return &ITSMHandler{service: service}
}

/* CreateIncident handles POST /api/v1/itsm/incidents */
func (h *ITSMHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Priority    string     `json:"priority"`
		RequesterID string     `json:"requester_id"`
		AssigneeID  *string    `json:"assignee_id,omitempty"`
		RunbookID   *uuid.UUID `json:"runbook_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Title == "" || req.RequesterID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("title and requester_id are required", nil))
		return
	}
	inc, err := h.service.CreateIncident(r.Context(), req.Title, req.Description, req.Priority, req.RequesterID, req.AssigneeID, req.RunbookID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, inc)
}

/* ListIncidents handles GET /api/v1/itsm/incidents */
func (h *ITSMHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	assigneeID := r.URL.Query().Get("assignee_id")
	list, err := h.service.ListIncidents(r.Context(), status, assigneeID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* CreateChange handles POST /api/v1/itsm/changes */
func (h *ITSMHandler) CreateChange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title          string     `json:"title"`
		Description    string     `json:"description"`
		ChangeType     string     `json:"change_type"`
		RequesterID    string     `json:"requester_id"`
		ScheduledStart *time.Time `json:"scheduled_start,omitempty"`
		ScheduledEnd   *time.Time `json:"scheduled_end,omitempty"`
		RunbookID      *uuid.UUID `json:"runbook_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Title == "" || req.RequesterID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("title and requester_id are required", nil))
		return
	}
	chg, err := h.service.CreateChange(r.Context(), req.Title, req.Description, req.ChangeType, req.RequesterID, req.ScheduledStart, req.ScheduledEnd, req.RunbookID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, chg)
}

/* ListChanges handles GET /api/v1/itsm/changes */
func (h *ITSMHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	list, err := h.service.ListChanges(r.Context(), status)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* CreateRunbook handles POST /api/v1/itsm/runbooks */
func (h *ITSMHandler) CreateRunbook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string                   `json:"name"`
		Description       string                   `json:"description"`
		WorkflowID        uuid.UUID                `json:"workflow_id"`
		TriggerConditions []map[string]interface{} `json:"trigger_conditions,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.WorkflowID == uuid.Nil {
		WriteErrorResponse(w, errors.ValidationFailed("name and workflow_id are required", nil))
		return
	}
	rb, err := h.service.CreateRunbook(r.Context(), req.Name, req.Description, req.WorkflowID, req.TriggerConditions)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, rb)
}

/* ListRunbooks handles GET /api/v1/itsm/runbooks */
func (h *ITSMHandler) ListRunbooks(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.ListRunbooks(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}
