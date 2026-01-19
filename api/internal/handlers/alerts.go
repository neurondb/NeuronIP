package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/alerts"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* AlertsHandler handles alerts requests */
type AlertsHandler struct {
	service *alerts.Service
}

/* NewAlertsHandler creates a new alerts handler */
func NewAlertsHandler(service *alerts.Service) *AlertsHandler {
	return &AlertsHandler{service: service}
}

/* CheckAlerts handles alert checking requests */
func (h *AlertsHandler) CheckAlerts(w http.ResponseWriter, r *http.Request) {
	newAlerts, err := h.service.CheckAlerts(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": newAlerts,
		"count":  len(newAlerts),
	})
}

/* GetAlerts handles alert retrieval requests */
func (h *AlertsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alertsList, err := h.service.GetAlerts(r.Context(), status, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alertsList,
		"count":  len(alertsList),
	})
}

/* ResolveAlertRequest represents alert resolution request */
type ResolveAlertRequest struct {
	Resolution string `json:"resolution"`
}

/* ResolveAlert handles alert resolution requests */
func (h *AlertsHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid alert ID"))
		return
	}

	var req ResolveAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	err = h.service.ResolveAlert(r.Context(), alertID, req.Resolution)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Alert resolved successfully",
		"alert_id": alertID,
	})
}

/* CreateAlertRuleRequest represents alert rule creation request */
type CreateAlertRuleRequest struct {
	Name      string                 `json:"name"`
	RuleType  string                 `json:"rule_type"`
	Threshold float64                `json:"threshold,omitempty"`
	Metric    string                 `json:"metric"`
	Condition string                 `json:"condition"`
	Enabled   bool                   `json:"enabled"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

/* CreateAlertRule handles alert rule creation requests */
func (h *AlertsHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	var req CreateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Name == "" || req.RuleType == "" || req.Metric == "" || req.Condition == "" {
		WriteErrorResponse(w, errors.ValidationFailed("name, rule_type, metric, and condition are required", nil))
		return
	}

	rule := alerts.AlertRule{
		Name:      req.Name,
		RuleType:  req.RuleType,
		Threshold: req.Threshold,
		Metric:    req.Metric,
		Condition: req.Condition,
		Enabled:   req.Enabled,
		Config:    req.Config,
	}

	// Store alert rule
	err := h.service.CreateAlertRule(r.Context(), rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

/* UpdateAlertRuleRequest represents alert rule update request */
type UpdateAlertRuleRequest struct {
	Name      *string                `json:"name,omitempty"`
	RuleType  *string                `json:"rule_type,omitempty"`
	Threshold *float64               `json:"threshold,omitempty"`
	Metric    *string               `json:"metric,omitempty"`
	Condition *string                `json:"condition,omitempty"`
	Enabled   *bool                  `json:"enabled,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

/* UpdateAlertRule handles alert rule update requests */
func (h *AlertsHandler) UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid rule ID"))
		return
	}

	var req UpdateAlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	rule, err := h.service.GetAlertRule(r.Context(), ruleID)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.RuleType != nil {
		rule.RuleType = *req.RuleType
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Metric != nil {
		rule.Metric = *req.Metric
	}
	if req.Condition != nil {
		rule.Condition = *req.Condition
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Config != nil {
		rule.Config = req.Config
	}

	err = h.service.UpdateAlertRule(r.Context(), *rule)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

/* DeleteAlertRule handles alert rule deletion requests */
func (h *AlertsHandler) DeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid rule ID"))
		return
	}

	err = h.service.DeleteAlertRule(r.Context(), ruleID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Alert rule deleted successfully",
		"rule_id": ruleID,
	})
}
