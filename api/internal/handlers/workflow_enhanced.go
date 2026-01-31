package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/workflows"
)

/* WorkflowEnhancedHandler handles enhanced workflow requests */
type WorkflowEnhancedHandler struct {
	orchestrator    *workflows.Orchestrator
	workflowService *workflows.Service
}

/* NewWorkflowEnhancedHandler creates a new enhanced workflow handler */
func NewWorkflowEnhancedHandler(
	orchestrator *workflows.Orchestrator,
	workflowService *workflows.Service,
) *WorkflowEnhancedHandler {
	return &WorkflowEnhancedHandler{
		orchestrator:    orchestrator,
		workflowService: workflowService,
	}
}

/* ExecuteDistributedWorkflow handles POST /api/v1/workflows/{id}/execute-distributed */
func (h *WorkflowEnhancedHandler) ExecuteDistributedWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	execution, err := h.orchestrator.ExecuteDistributedWorkflow(r.Context(), workflowID, input)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(execution)
}

/* GetWorkflowMetrics handles GET /api/v1/workflows/{id}/metrics */
func (h *WorkflowEnhancedHandler) GetWorkflowMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	// Parse time range from query params
	startTime := parseTime(r.URL.Query().Get("start_time"))
	endTime := parseTime(r.URL.Query().Get("end_time"))
	if endTime.IsZero() {
		endTime = time.Now()
	}
	if startTime.IsZero() {
		startTime = endTime.AddDate(0, 0, -30) // Default to last 30 days
	}

	metrics := map[string]interface{}{
		"workflow_id":  workflowID,
		"start_time":   startTime,
		"end_time":     endTime,
		"total_runs":   0,
		"success_rate": 0.0,
	}
	if h.workflowService != nil {
		stats, err := h.workflowService.GetWorkflowRunStats(r.Context(), workflowID, startTime, endTime)
		if err == nil {
			metrics["total_runs"] = stats.TotalRuns
			metrics["success_rate"] = stats.SuccessRate
			metrics["success_count"] = stats.SuccessCount
			metrics["failed_count"] = stats.FailedCount
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetWorkflowCost handles GET /api/v1/workflows/{id}/cost */
func (h *WorkflowEnhancedHandler) GetWorkflowCost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid workflow ID"))
		return
	}

	costResult, err := h.workflowService.GetWorkflowCost(r.Context(), workflowID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflow_id":     costResult.WorkflowID,
		"cost":            costResult.TotalCost,
		"total_runs":      costResult.TotalRuns,
		"cost_per_run":    costResult.CostPerRun,
		"duration_ms_sum": costResult.DurationMsSum,
	})
}
