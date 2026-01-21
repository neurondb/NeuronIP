package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/reporting"
)

/* ReportingHandler handles reporting requests */
type ReportingHandler struct {
	reportingService *reporting.ReportingService
}

/* NewReportingHandler creates a new reporting handler */
func NewReportingHandler(reportingService *reporting.ReportingService) *ReportingHandler {
	return &ReportingHandler{reportingService: reportingService}
}

/* CreateReport handles POST /api/v1/reports */
func (h *ReportingHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var report reporting.Report
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.reportingService.CreateReport(r.Context(), report)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GenerateReport handles POST /api/v1/reports/{id}/generate */
func (h *ReportingHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid report ID"))
		return
	}

	result, err := h.reportingService.GenerateReport(r.Context(), reportID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* GetReport handles GET /api/v1/reports/{id} */
func (h *ReportingHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid report ID"))
		return
	}

	report, err := h.reportingService.GetReport(r.Context(), reportID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

/* ListReports handles GET /api/v1/reports */
func (h *ReportingHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	limit := 50

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	reports, err := h.reportingService.ListReports(r.Context(), categoryPtr, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}
