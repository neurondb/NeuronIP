package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/billing"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* BillingHandler handles billing and usage requests */
type BillingHandler struct {
	service *billing.BillingService
}

/* NewBillingHandler creates a new billing handler */
func NewBillingHandler(service *billing.BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

/* GetUsageMetrics handles GET /api/v1/billing/usage */
func (h *BillingHandler) GetUsageMetrics(w http.ResponseWriter, r *http.Request) {
	metricType := r.URL.Query().Get("metric_type")
	userID := r.URL.Query().Get("user_id")
	limit := 100

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	metrics, err := h.service.GetUsageMetrics(r.Context(), metricType, userID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

/* GetDetailedMetrics handles GET /api/v1/billing/metrics */
func (h *BillingHandler) GetDetailedMetrics(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	periodStart := time.Now().AddDate(0, 0, -30) // Default to last 30 days
	periodEnd := time.Now()

	if startStr := r.URL.Query().Get("period_start"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			periodStart = parsed
		}
	}

	if endStr := r.URL.Query().Get("period_end"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			periodEnd = parsed
		}
	}

	metrics, err := h.service.GetDetailedMetrics(r.Context(), userID, periodStart, periodEnd)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetDashboardData handles GET /api/v1/billing/dashboard */
func (h *BillingHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetDashboardData(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

/* TrackUsage handles POST /api/v1/billing/track */
func (h *BillingHandler) TrackUsage(w http.ResponseWriter, r *http.Request) {
	var req billing.UsageMetric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.MetricType == "" || req.MetricName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("metric_type and metric_name are required", nil))
		return
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	if err := h.service.TrackUsage(r.Context(), req); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "tracked",
	})
}
