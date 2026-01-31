package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/billing"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* CostAnalyticsHandler handles cost analytics requests */
type CostAnalyticsHandler struct {
	costService  *billing.CostTrackingService
	budgetService *billing.BudgetService
}

/* NewCostAnalyticsHandler creates a new cost analytics handler */
func NewCostAnalyticsHandler(
	costService *billing.CostTrackingService,
	budgetService *billing.BudgetService,
) *CostAnalyticsHandler {
	return &CostAnalyticsHandler{
		costService:  costService,
		budgetService: budgetService,
	}
}

/* GetCostSummary handles GET /api/v1/cost/summary */
func (h *CostAnalyticsHandler) GetCostSummary(w http.ResponseWriter, r *http.Request) {
	// Parse time range
	startTime := parseTime(r.URL.Query().Get("start_time"))
	endTime := parseTime(r.URL.Query().Get("end_time"))
	if endTime.IsZero() {
		endTime = time.Now()
	}
	if startTime.IsZero() {
		startTime = endTime.AddDate(0, 0, -30) // Default to last 30 days
	}

	userID := r.URL.Query().Get("user_id")
	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	summary, err := h.costService.GetCostSummary(r.Context(), userIDPtr, startTime, endTime)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

/* CreateBudget handles POST /api/v1/cost/budgets */
func (h *CostAnalyticsHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var budget billing.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdBudget, err := h.budgetService.CreateBudget(r.Context(), budget)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBudget)
}

/* GetBudgetStatus handles GET /api/v1/cost/budgets/{id}/status */
func (h *CostAnalyticsHandler) GetBudgetStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	budgetID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid budget ID"))
		return
	}

	status, err := h.budgetService.CheckBudgetStatus(r.Context(), budgetID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

/* RecordCost handles POST /api/v1/cost/record */
func (h *CostAnalyticsHandler) RecordCost(w http.ResponseWriter, r *http.Request) {
	var cost billing.CostRecord
	if err := json.NewDecoder(r.Body).Decode(&cost); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.costService.RecordCost(r.Context(), cost); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "recorded",
	})
}
