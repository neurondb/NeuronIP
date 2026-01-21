package handlers

import (
	"encoding/json"
	"net/http"

	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
)

/* MetricsHandler handles metrics and semantic layer requests */
type MetricsHandler struct {
	service *metrics.MetricsService
	catalogService *catalog.MetricsService
}

/* NewMetricsHandler creates a new metrics handler */
func NewMetricsHandler(service *metrics.MetricsService, catalogService *catalog.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		service: service,
		catalogService: catalogService,
	}
}

/* ListMetrics handles GET /api/v1/metrics */
func (h *MetricsHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	
	var category *string
	if c := r.URL.Query().Get("category"); c != "" {
		category = &c
	}
	
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	
	metrics, err := h.catalogService.ListMetrics(r.Context(), status, category, limit)
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

	metric, err := h.catalogService.GetMetric(r.Context(), id)
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

	// Convert metrics.Metric to catalog.Metric
	catalogMetric := catalog.Metric{
		Name:          req.Name,
		DisplayName:   req.Name,
		Description:   req.BusinessTerm,
		SQLExpression: req.Definition,
		MetricType:    "custom",
		Status:        "draft",
		Version:       "1.0.0",
	}
	if req.KPIType != nil {
		catalogMetric.Category = req.KPIType
	}
	
	metric, err := h.catalogService.CreateMetric(r.Context(), catalogMetric)
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

	if h.service == nil {
		WriteErrorResponse(w, errors.InternalServer("Metrics service not initialized"))
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

/* CalculateMetric handles POST /api/v1/metrics/{id}/calculate */
func (h *MetricsHandler) CalculateMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		Filters map[string]interface{} `json:"filters,omitempty"`
	}
	if r.Body != nil && r.ContentLength > 0 {
		json.NewDecoder(r.Body).Decode(&req)
	}

	if h.service == nil {
		WriteErrorResponse(w, errors.InternalServer("Metrics service not initialized"))
		return
	}

	result, err := h.service.CalculateMetric(r.Context(), id, req.Filters)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"metric_id": id,
		"value": result,
	})
}

/* GetMetricLineage handles GET /api/v1/metrics/{id}/lineage */
func (h *MetricsHandler) GetMetricLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	if h.service == nil {
		WriteErrorResponse(w, errors.InternalServer("Metrics service not initialized"))
		return
	}

	lineage, err := h.service.GetMetricLineage(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineage)
}

/* AddMetricLineage handles POST /api/v1/metrics/{id}/lineage */
func (h *MetricsHandler) AddMetricLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		DependsOnMetricID *uuid.UUID `json:"depends_on_metric_id,omitempty"`
		DependsOnTable    *string    `json:"depends_on_table,omitempty"`
		DependsOnColumn   *string    `json:"depends_on_column,omitempty"`
		RelationshipType  string     `json:"relationship_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.RelationshipType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("relationship_type is required", nil))
		return
	}

	if h.service == nil {
		WriteErrorResponse(w, errors.InternalServer("Metrics service not initialized"))
		return
	}

	err = h.service.AddMetricLineage(r.Context(), id, req.DependsOnMetricID, req.DependsOnTable, req.DependsOnColumn, req.RelationshipType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

/* ApproveMetric handles POST /api/v1/metrics/{id}/approve */
func (h *MetricsHandler) ApproveMetric(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid metric ID"))
		return
	}

	var req struct {
		ApprovedBy string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ApprovedBy == "" {
		WriteErrorResponse(w, errors.ValidationFailed("approved_by is required", nil))
		return
	}

	// Use catalog service for approval
	err = h.catalogService.ApproveMetric(r.Context(), id, req.ApprovedBy)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* DiscoverMetrics handles POST /api/v1/metrics/discover */
func (h *MetricsHandler) DiscoverMetrics(w http.ResponseWriter, r *http.Request) {
	var req metrics.MetricDiscoveryCriteria
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if h.service == nil {
		WriteErrorResponse(w, errors.InternalServer("Metrics service not initialized"))
		return
	}

	results, err := h.service.DiscoverMetrics(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
