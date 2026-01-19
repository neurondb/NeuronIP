package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/observability"
)

/* ObservabilityHandler handles observability requests */
type ObservabilityHandler struct {
	service *observability.ObservabilityService
}

/* NewObservabilityHandler creates a new observability handler */
func NewObservabilityHandler(service *observability.ObservabilityService) *ObservabilityHandler {
	return &ObservabilityHandler{service: service}
}

/* GetQueryPerformance handles GET /api/v1/observability/queries/performance */
func (h *ObservabilityHandler) GetQueryPerformance(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	perf, err := h.service.GetQueryPerformance(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perf)
}

/* GetSystemLogs handles GET /api/v1/observability/logs */
func (h *ObservabilityHandler) GetSystemLogs(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("log_type")
	level := r.URL.Query().Get("level")
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	logs, err := h.service.GetSystemLogs(r.Context(), logType, level, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

/* GetSystemMetrics handles GET /api/v1/observability/metrics */
func (h *ObservabilityHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.GetSystemMetrics(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetAgentLogs handles GET /api/v1/observability/agent-logs */
func (h *ObservabilityHandler) GetAgentLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	logs, err := h.service.GetAgentLogs(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

/* GetWorkflowLogs handles GET /api/v1/observability/workflow-logs */
func (h *ObservabilityHandler) GetWorkflowLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	logs, err := h.service.GetWorkflowLogs(r.Context(), limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
