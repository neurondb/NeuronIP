package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/analytics"
)

/* AnalyticsHandler handles analytics requests */
type AnalyticsHandler struct {
	service *analytics.Service
}

/* NewAnalyticsHandler creates a new analytics handler */
func NewAnalyticsHandler(service *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

/* GetSearchAnalytics handles search analytics requests */
func (h *AnalyticsHandler) GetSearchAnalytics(w http.ResponseWriter, r *http.Request) {
	query := h.buildAnalyticsQuery(r)

	result, err := h.service.GetSearchAnalytics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetWarehouseAnalytics handles warehouse analytics requests */
func (h *AnalyticsHandler) GetWarehouseAnalytics(w http.ResponseWriter, r *http.Request) {
	query := h.buildAnalyticsQuery(r)

	result, err := h.service.GetWarehouseAnalytics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetWorkflowAnalytics handles workflow analytics requests */
func (h *AnalyticsHandler) GetWorkflowAnalytics(w http.ResponseWriter, r *http.Request) {
	query := h.buildAnalyticsQuery(r)

	result, err := h.service.GetWorkflowAnalytics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetComplianceAnalytics handles compliance analytics requests */
func (h *AnalyticsHandler) GetComplianceAnalytics(w http.ResponseWriter, r *http.Request) {
	query := h.buildAnalyticsQuery(r)

	result, err := h.service.GetComplianceAnalytics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetRetrievalQuality handles retrieval quality metrics requests */
func (h *AnalyticsHandler) GetRetrievalQuality(w http.ResponseWriter, r *http.Request) {
	query := h.buildAnalyticsQuery(r)

	result, err := h.service.GetRetrievalQualityMetrics(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* buildAnalyticsQuery builds analytics query from HTTP request */
func (h *AnalyticsHandler) buildAnalyticsQuery(r *http.Request) analytics.AnalyticsQuery {
	query := analytics.AnalyticsQuery{}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			query.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			query.EndDate = &endDate
		}
	}

	query.EntityType = r.URL.Query().Get("entity_type")
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		query.UserID = &userID
	}

	return query
}
