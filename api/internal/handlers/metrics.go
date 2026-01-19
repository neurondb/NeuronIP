package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
)

/* MetricsHandler handles metrics and semantic layer requests */
type MetricsHandler struct {
	service *metrics.MetricsService
}

/* NewMetricsHandler creates a new metrics handler */
func NewMetricsHandler(service *metrics.MetricsService) *MetricsHandler {
	return &MetricsHandler{service: service}
}

/* ListMetrics handles GET /api/v1/metrics */
func (h *MetricsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.service.ListMetrics(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

/* GetMetric handles GET /api/v1/metrics/{id} */
func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	metric, err := h.service.GetMetric(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metric)
}

/* CreateMetric handles POST /api/v1/metrics */
func (h *MetricsHandler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	var req metrics.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name is required", nil))
		return
	}

	metric, err := h.service.CreateMetric(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(metric)
}

/* UpdateMetric handles PUT /api/v1/metrics/{id} */
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req metrics.Metric
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	metric, err := h.service.UpdateMetric(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metric)
}

/* DeleteMetric handles DELETE /api/v1/metrics/{id} */
func (h *MetricsHandler) DeleteMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	if err := h.service.DeleteMetric(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* SearchMetrics handles POST /api/v1/metrics/search */
func (h *MetricsHandler) SearchMetrics(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	results, err := h.service.SearchMetrics(r.Context(), req.Query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
