package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/dataquality"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* DataQualityHandler handles data quality requests */
type DataQualityHandler struct {
	service *dataquality.Service
}

/* NewDataQualityHandler creates a new data quality handler */
func NewDataQualityHandler(service *dataquality.Service) *DataQualityHandler {
	return &DataQualityHandler{service: service}
}

/* CreateRule creates a new quality rule */
func (h *DataQualityHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule dataquality.QualityRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateRule(r.Context(), rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetRule retrieves a quality rule */
func (h *DataQualityHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid rule ID"))
		return
	}

	rule, err := h.service.GetRule(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

/* ExecuteRule executes a quality rule */
func (h *DataQualityHandler) ExecuteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid rule ID"))
		return
	}

	check, err := h.service.ExecuteRule(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(check)
}

/* GetDashboard gets quality dashboard data */
func (h *DataQualityHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.service.GetQualityDashboard(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

/* GetTrends gets quality trends */
func (h *DataQualityHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "overall"
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		fmt.Sscanf(daysStr, "%d", &days)
	}

	trends, err := h.service.GetQualityTrends(r.Context(), level, days)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}
