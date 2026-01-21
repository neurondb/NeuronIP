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

/* ListWorkflows handles GET /api/v1/workflows */
func (h *WorkflowHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	workflows, err := h.service.ListWorkflows(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflows": workflows,
	})
}

/* CreateWorkflowVersion handles POST /api/v1/workflows/{id}/versions */
func (h *WorkflowHandler) CreateWorkflowVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	var req struct {
		Version string                 `json:"version"`
		Changes map[string]interface{} `json:"changes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Version == "" {
		req.Version = "1.0.0"
	}

	workflow, err := h.service.CreateWorkflowVersion(r.Context(), id, req.Version, req.Changes)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workflow)
}

/* ScheduleWorkflow handles POST /api/v1/workflows/{id}/schedule */
func (h *WorkflowHandler) ScheduleWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	var req workflows.ScheduleConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ScheduleWorkflow(r.Context(), id, req); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "scheduled",
		"message": "Workflow scheduled successfully",
	})
}

/* RecoverWorkflowExecution handles POST /api/v1/workflows/executions/{id}/recover */
func (h *WorkflowHandler) RecoverWorkflowExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid execution ID"))
		return
	}

	var req struct {
		RetryFromStep *string `json:"retry_from_step,omitempty"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := h.service.RecoverWorkflowExecution(r.Context(), id, req.RetryFromStep)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetWorkflowExecutionStatus handles GET /api/v1/workflows/executions/{id}/status */
func (h *WorkflowHandler) GetWorkflowExecutionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid execution ID"))
		return
	}

	status, err := h.service.GetWorkflowExecutionStatus(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

/* GetWorkflowMonitoring handles GET /api/v1/workflows/{id}/monitoring */
func (h *WorkflowHandler) GetWorkflowMonitoring(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "24h"
	}

	monitoring, err := h.service.GetWorkflowMonitoring(r.Context(), id, timeRange)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitoring)
}

/* CreateWorkflow handles POST /api/v1/workflows */
func (h *WorkflowHandler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var workflow workflows.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateWorkflow(r.Context(), workflow)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* UpdateWorkflow handles PUT /api/v1/workflows/{id} */
func (h *WorkflowHandler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	var workflow workflows.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	updated, err := h.service.UpdateWorkflow(r.Context(), id, workflow)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

/* DeleteWorkflow handles DELETE /api/v1/workflows/{id} */
func (h *WorkflowHandler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	if err := h.service.DeleteWorkflow(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "deleted",
		"message": "Workflow deleted successfully",
	})
}

/* GetWorkflowVersions handles GET /api/v1/workflows/{id}/versions */
func (h *WorkflowHandler) GetWorkflowVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	versions, err := h.service.GetWorkflowVersions(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": versions,
	})
}

/* GetWorkflowVersion handles GET /api/v1/workflows/{id}/versions/{version_id} */
func (h *WorkflowHandler) GetWorkflowVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	versionID, err := uuid.Parse(vars["version_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid version ID"))
		return
	}

	version, err := h.service.GetWorkflowVersion(r.Context(), id, versionID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

/* GetScheduledWorkflows handles GET /api/v1/workflows/{id}/schedules */
func (h *WorkflowHandler) GetScheduledWorkflows(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	schedules, err := h.service.GetScheduledWorkflows(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules": schedules,
	})
}

/* CancelScheduledWorkflow handles POST /api/v1/workflows/{id}/schedules/{schedule_id}/cancel */
func (h *WorkflowHandler) CancelScheduledWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	scheduleID, err := uuid.Parse(vars["schedule_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid schedule ID"))
		return
	}

	if err := h.service.CancelScheduledWorkflow(r.Context(), id, scheduleID); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "cancelled",
		"message": "Schedule cancelled successfully",
	})
}

/* GetWorkflowExecutionLogs handles GET /api/v1/workflows/executions/{id}/logs */
func (h *WorkflowHandler) GetWorkflowExecutionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid execution ID"))
		return
	}

	logs, err := h.service.GetWorkflowExecutionLogs(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

/* GetWorkflowExecutionMetrics handles GET /api/v1/workflows/executions/{id}/metrics */
func (h *WorkflowHandler) GetWorkflowExecutionMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid execution ID"))
		return
	}

	metrics, err := h.service.GetWorkflowExecutionMetrics(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metrics,
	})
}

/* GetWorkflowExecutionDecisions handles GET /api/v1/workflows/executions/{id}/decisions */
func (h *WorkflowHandler) GetWorkflowExecutionDecisions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid execution ID"))
		return
	}

	decisions, err := h.service.GetWorkflowExecutionDecisions(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"decisions": decisions,
	})
}
