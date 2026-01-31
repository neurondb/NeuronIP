package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/ml"
)

/* NotebooksHandler handles notebook requests */
type NotebooksHandler struct {
	service *ml.NotebookService
}

/* NewNotebooksHandler creates a new notebooks handler */
func NewNotebooksHandler(service *ml.NotebookService) *NotebooksHandler {
	return &NotebooksHandler{service: service}
}

/* CreateNotebook handles POST /api/v1/notebooks */
func (h *NotebooksHandler) CreateNotebook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string     `json:"name"`
		Description     string     `json:"description"`
		OwnerID         string     `json:"owner_id"`
		WorkspaceID     *uuid.UUID `json:"workspace_id,omitempty"`
		WorkflowID      *uuid.UUID `json:"workflow_id,omitempty"`
		DefaultLanguage string     `json:"default_language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.Name == "" || req.OwnerID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name and owner_id are required", nil))
		return
	}
	nb, err := h.service.CreateNotebook(r.Context(), req.Name, req.Description, req.OwnerID, req.WorkspaceID, req.WorkflowID, req.DefaultLanguage)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, nb)
}

/* GetNotebook handles GET /api/v1/notebooks/{id} */
func (h *NotebooksHandler) GetNotebook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notebook ID"))
		return
	}
	nb, err := h.service.GetNotebook(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, nb)
}

/* ListNotebooks handles GET /api/v1/notebooks */
func (h *NotebooksHandler) ListNotebooks(w http.ResponseWriter, r *http.Request) {
	var workspaceID *uuid.UUID
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if id, err := uuid.Parse(wsStr); err == nil {
			workspaceID = &id
		}
	}
	ownerID := r.URL.Query().Get("owner_id")
	list, err := h.service.ListNotebooks(r.Context(), workspaceID, ownerID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* AddCell handles POST /api/v1/notebooks/{id}/cells */
func (h *NotebooksHandler) AddCell(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notebook ID"))
		return
	}
	var req struct {
		Position int    `json:"position"`
		CellType string `json:"cell_type"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.CellType == "" {
		req.CellType = "sql"
	}
	cell, err := h.service.AddCell(r.Context(), id, req.Position, req.CellType, req.Content)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, cell)
}

/* ListCells handles GET /api/v1/notebooks/{id}/cells */
func (h *NotebooksHandler) ListCells(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notebook ID"))
		return
	}
	list, err := h.service.ListCells(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

/* CreateRun handles POST /api/v1/notebooks/{id}/runs */
func (h *NotebooksHandler) CreateRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notebook ID"))
		return
	}
	var req struct {
		TriggeredBy         string     `json:"triggered_by"`
		WorkflowExecutionID *uuid.UUID `json:"workflow_execution_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}
	if req.TriggeredBy == "" {
		req.TriggeredBy = "api"
	}
	run, err := h.service.CreateRun(r.Context(), id, req.TriggeredBy, req.WorkflowExecutionID)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, run)
}

/* ListRuns handles GET /api/v1/notebooks/{id}/runs */
func (h *NotebooksHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid notebook ID"))
		return
	}
	list, err := h.service.ListRuns(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, list)
}
