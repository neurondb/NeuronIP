package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/quality"
)

/* QualityHandler handles quality-related requests */
type QualityHandler struct {
	service *quality.Service
}

/* NewQualityHandler creates a new quality handler */
func NewQualityHandler(service *quality.Service) *QualityHandler {
	return &QualityHandler{service: service}
}

/* CreateRule handles POST /api/v1/quality/rules */
func (h *QualityHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule quality.QualityRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		WriteErrorResponse(w, errors.BadRequest("invalid request body"))
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

/* GetRule handles GET /api/v1/quality/rules/{id} */
func (h *QualityHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("invalid rule id"))
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

/* ListRules handles GET /api/v1/quality/rules */
func (h *QualityHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	var enabled *bool
	var ruleType *string

	if val := r.URL.Query().Get("enabled"); val != "" {
		e := val == "true"
		enabled = &e
	}

	if val := r.URL.Query().Get("rule_type"); val != "" {
		ruleType = &val
	}

	rules, err := h.service.ListRules(r.Context(), enabled, ruleType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

/* ExecuteRule handles POST /api/v1/quality/rules/{id}/execute */
func (h *QualityHandler) ExecuteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("invalid rule id"))
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

/* GetQualityScore handles GET /api/v1/quality/scores */
func (h *QualityHandler) GetQualityScore(w http.ResponseWriter, r *http.Request) {
	var connectorID *uuid.UUID
	var schemaName, tableName, columnName *string

	if val := r.URL.Query().Get("connector_id"); val != "" {
		id, err := uuid.Parse(val)
		if err == nil {
			connectorID = &id
		}
	}

	if val := r.URL.Query().Get("schema_name"); val != "" {
		schemaName = &val
	}

	if val := r.URL.Query().Get("table_name"); val != "" {
		tableName = &val
	}

	if val := r.URL.Query().Get("column_name"); val != "" {
		columnName = &val
	}

	score, err := h.service.GetQualityScore(r.Context(), connectorID, schemaName, tableName, columnName)
	if err != nil {
		WriteError(w, err)
		return
	}

	if score == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"score": nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}

/* ListQualityChecks handles GET /api/v1/quality/checks */
func (h *QualityHandler) ListQualityChecks(w http.ResponseWriter, r *http.Request) {
	var ruleID *uuid.UUID
	var status *string
	limit := 100

	if val := r.URL.Query().Get("rule_id"); val != "" {
		id, err := uuid.Parse(val)
		if err == nil {
			ruleID = &id
		}
	}

	if val := r.URL.Query().Get("status"); val != "" {
		status = &val
	}

	if val := r.URL.Query().Get("limit"); val != "" {
		if l, err := strconv.Atoi(val); err == nil {
			limit = l
		}
	}

	checks, err := h.service.ListQualityChecks(r.Context(), ruleID, status, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"checks": checks,
		"count":  len(checks),
	})
}
