package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/observability"
)

/* ObservabilityHandler handles observability requests */
type ObservabilityHandler struct {
	service              *observability.ObservabilityService
	retrievalService     *observability.RetrievalMetricsService
	hallucinationService *observability.HallucinationDetectionService
}

/* NewObservabilityHandler creates a new observability handler */
func NewObservabilityHandler(pool *pgxpool.Pool) *ObservabilityHandler {
	return &ObservabilityHandler{
		service:              observability.NewObservabilityService(pool),
		retrievalService:     observability.NewRetrievalMetricsService(pool),
		hallucinationService: observability.NewHallucinationDetectionService(pool),
	}
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

/* GetRealTimeMetrics handles GET /api/v1/observability/realtime */
func (h *ObservabilityHandler) GetRealTimeMetrics(w http.ResponseWriter, r *http.Request) {
	timeWindow := r.URL.Query().Get("window")
	if timeWindow == "" {
		timeWindow = "5m"
	}

	metrics, err := h.service.GetRealTimeMetrics(r.Context(), timeWindow)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetLogStream handles GET /api/v1/observability/logs/stream */
func (h *ObservabilityHandler) GetLogStream(w http.ResponseWriter, r *http.Request) {
	logType := r.URL.Query().Get("log_type")
	level := r.URL.Query().Get("level")
	limit := 1000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	sinceStr := r.URL.Query().Get("since")
	since := time.Now().Add(-1 * time.Hour) // Default to 1 hour ago
	if sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsed
		}
	}

	logs, err := h.service.GetLogStream(r.Context(), logType, level, since, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

/* GetPerformanceBenchmark handles GET /api/v1/observability/benchmark */
func (h *ObservabilityHandler) GetPerformanceBenchmark(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metric_type")
	if metricType == "" {
		metricType = "query"
	}

	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "24h"
	}

	benchmark, err := h.service.GetPerformanceBenchmark(r.Context(), metricType, timeRange)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(benchmark)
}

/* GetCostBreakdown handles GET /api/v1/observability/cost/breakdown */
func (h *ObservabilityHandler) GetCostBreakdown(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "category"
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid start_time format"))
			return
		}
	} else {
		startTime = time.Now().Add(-30 * 24 * time.Hour) // Default to 30 days ago
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid end_time format"))
			return
		}
	} else {
		endTime = time.Now()
	}

	var userID *string
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID = &userIDStr
	}

	breakdown, err := h.service.GetCostBreakdown(r.Context(), userID, startTime, endTime, groupBy)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(breakdown)
}

/* GetAgentExecutionLogs handles GET /api/v1/observability/agents/{agent_id}/logs */
func (h *ObservabilityHandler) GetAgentExecutionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	var agentRunID *uuid.UUID
	if runIDStr := r.URL.Query().Get("agent_run_id"); runIDStr != "" {
		if id, err := uuid.Parse(runIDStr); err == nil {
			agentRunID = &id
		}
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := h.service.GetAgentExecutionLogs(r.Context(), &agentID, agentRunID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

/* GetRetrievalMetrics handles GET /api/v1/observability/retrieval/metrics */
func (h *ObservabilityHandler) GetRetrievalMetrics(w http.ResponseWriter, r *http.Request) {
	var queryID *uuid.UUID
	if queryIDStr := r.URL.Query().Get("query_id"); queryIDStr != "" {
		if id, err := uuid.Parse(queryIDStr); err == nil {
			queryID = &id
		}
	}

	var agentRunID *uuid.UUID
	if runIDStr := r.URL.Query().Get("agent_run_id"); runIDStr != "" {
		if id, err := uuid.Parse(runIDStr); err == nil {
			agentRunID = &id
		}
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	metrics, err := h.retrievalService.GetRetrievalMetrics(r.Context(), queryID, agentRunID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetRetrievalStats handles GET /api/v1/observability/retrieval/stats */
func (h *ObservabilityHandler) GetRetrievalStats(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "24h"
	}

	stats, err := h.retrievalService.GetRetrievalStats(r.Context(), timeRange)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

/* GetHallucinationSignals handles GET /api/v1/observability/hallucination/signals */
func (h *ObservabilityHandler) GetHallucinationSignals(w http.ResponseWriter, r *http.Request) {
	var queryID *uuid.UUID
	if queryIDStr := r.URL.Query().Get("query_id"); queryIDStr != "" {
		if id, err := uuid.Parse(queryIDStr); err == nil {
			queryID = &id
		}
	}

	var agentRunID *uuid.UUID
	if runIDStr := r.URL.Query().Get("agent_run_id"); runIDStr != "" {
		if id, err := uuid.Parse(runIDStr); err == nil {
			agentRunID = &id
		}
	}

	riskLevel := r.URL.Query().Get("risk_level")
	var riskLevelPtr *string
	if riskLevel != "" {
		riskLevelPtr = &riskLevel
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	signals, err := h.hallucinationService.GetHallucinationSignals(r.Context(), queryID, agentRunID, riskLevelPtr, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signals)
}

/* GetHallucinationStats handles GET /api/v1/observability/hallucination/stats */
func (h *ObservabilityHandler) GetHallucinationStats(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "24h"
	}

	stats, err := h.hallucinationService.GetHallucinationStats(r.Context(), timeRange)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

/* GetQueryCost handles GET /api/v1/observability/queries/{id}/cost */
func (h *ObservabilityHandler) GetQueryCost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid query ID"))
		return
	}

	cost, err := h.service.GetQueryCost(r.Context(), queryID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}

/* GetAgentRunCost handles GET /api/v1/observability/agents/runs/{id}/cost */
func (h *ObservabilityHandler) GetAgentRunCost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentRunID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid agent run ID"))
		return
	}

	cost, err := h.service.GetAgentRunCost(r.Context(), agentRunID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cost)
}
