package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
)

/* MetricsEnhancedHandler handles enhanced metrics requests */
type MetricsEnhancedHandler struct {
	collector *metrics.MetricsCollector
}

/* NewMetricsEnhancedHandler creates a new enhanced metrics handler */
func NewMetricsEnhancedHandler(collector *metrics.MetricsCollector) *MetricsEnhancedHandler {
	return &MetricsEnhancedHandler{collector: collector}
}

/* GetLatencyMetrics handles getting latency metrics */
func (h *MetricsEnhancedHandler) GetLatencyMetrics(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		WriteErrorResponse(w, errors.BadRequest("endpoint is required"))
		return
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	latencyMetrics, err := h.collector.GetLatencyMetrics(r.Context(), endpoint, startTime, endTime)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latencyMetrics)
}

/* GetErrorRate handles getting error rates */
func (h *MetricsEnhancedHandler) GetErrorRate(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Query().Get("endpoint")
	if endpoint == "" {
		WriteErrorResponse(w, errors.BadRequest("endpoint is required"))
		return
	}

	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	errorRate, err := h.collector.GetErrorRate(r.Context(), endpoint, startTime, endTime)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{
		"error_rate": errorRate,
	})
}

/* GetTokenUsage handles getting token usage metrics */
func (h *MetricsEnhancedHandler) GetTokenUsage(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	usage, err := h.collector.GetTokenUsage(r.Context(), userIDPtr, startTime, endTime)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

/* GetEmbeddingCost handles getting embedding cost metrics */
func (h *MetricsEnhancedHandler) GetEmbeddingCost(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if startTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = parsed
		}
	}
	if endTimeStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = parsed
		}
	}

	costs, err := h.collector.GetEmbeddingCost(r.Context(), startTime, endTime)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costs)
}
