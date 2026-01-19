package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/workflows"
)

/* WorkflowHandler handles workflow requests */
type WorkflowHandler struct {
	service *workflows.Service
}

/* NewWorkflowHandler creates a new workflow handler */
func NewWorkflowHandler(service *workflows.Service) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

/* ExecuteWorkflowRequest represents workflow execution request */
type ExecuteWorkflowRequest struct {
	Input map[string]interface{} `json:"input"`
}

/* ExecuteWorkflow handles workflow execution requests */
func (h *WorkflowHandler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	var req ExecuteWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Input == nil {
		req.Input = make(map[string]interface{})
	}

	result, err := h.service.ExecuteWorkflow(r.Context(), workflowID, req.Input)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetWorkflow handles workflow retrieval requests */
func (h *WorkflowHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	workflow, err := h.service.GetWorkflow(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workflow)
}
