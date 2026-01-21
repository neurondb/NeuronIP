package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ComplianceHandler handles compliance requests */
type ComplianceHandler struct {
	service        *compliance.Service
	anomalyService *compliance.AnomalyService
	policyService  *compliance.PolicyService
}

/* NewComplianceHandler creates a new compliance handler */
func NewComplianceHandler(service *compliance.Service, anomalyService *compliance.AnomalyService, policyService *compliance.PolicyService) *ComplianceHandler {
	return &ComplianceHandler{
		service:        service,
		anomalyService: anomalyService,
		policyService:  policyService,
	}
}

/* CheckComplianceRequest represents compliance check request */
type CheckComplianceRequest struct {
	EntityType    string                 `json:"entity_type"`
	EntityID      string                 `json:"entity_id"`
	EntityContent string                 `json:"entity_content"`
}

/* CheckCompliance handles compliance checking requests */
func (h *ComplianceHandler) CheckCompliance(w http.ResponseWriter, r *http.Request) {
	var req CheckComplianceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.EntityType == "" || req.EntityID == "" {
		WriteErrorResponse(w, errors.ValidationFailed("entity_type and entity_id are required", nil))
		return
	}

	matches, err := h.service.CheckCompliance(r.Context(), req.EntityType, req.EntityID, req.EntityContent)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matches": matches,
		"count":   len(matches),
	})
}

/* GetAnomalyDetections handles anomaly detection retrieval requests */
func (h *ComplianceHandler) GetAnomalyDetections(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityID := r.URL.Query().Get("entity_id")
	status := r.URL.Query().Get("status")

	detections, err := h.anomalyService.GetAnomalyDetections(r.Context(), entityType, entityID, status, 10)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"detections": detections,
		"count":      len(detections),
	})
}

/* ListPolicies handles GET /api/v1/compliance/policies */
func (h *ComplianceHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policyType := r.URL.Query().Get("policy_type")
	enabledStr := r.URL.Query().Get("enabled")

	var policyTypePtr *string
	if policyType != "" {
		policyTypePtr = &policyType
	}

	var enabledPtr *bool
	if enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			enabledPtr = &enabled
		}
	}

	policies, err := h.policyService.ListPolicies(r.Context(), policyTypePtr, enabledPtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policies)
}

/* GetPolicy handles GET /api/v1/compliance/policies/{id} */
func (h *ComplianceHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}

	policy, err := h.policyService.GetPolicy(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

/* CreatePolicy handles POST /api/v1/compliance/policies */
func (h *ComplianceHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req compliance.Policy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.PolicyName == "" {
		WriteErrorResponse(w, errors.ValidationFailed("policy_name is required", nil))
		return
	}
	if req.PolicyType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("policy_type is required", nil))
		return
	}

	policy, err := h.policyService.CreatePolicy(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

/* UpdatePolicy handles PUT /api/v1/compliance/policies/{id} */
func (h *ComplianceHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}

	var req compliance.Policy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	policy, err := h.policyService.UpdatePolicy(r.Context(), id, req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

/* DeletePolicy handles DELETE /api/v1/compliance/policies/{id} */
func (h *ComplianceHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid policy ID"))
		return
	}

	if err := h.policyService.DeletePolicy(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* GetComplianceReport handles GET /api/v1/compliance/report */
func (h *ComplianceHandler) GetComplianceReport(w http.ResponseWriter, r *http.Request) {
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	entityType := r.URL.Query().Get("entity_type")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid start_time format"))
			return
		}
	} else {
		startTime = time.Now().Add(-30 * 24 * time.Hour) // Default to 30 days ago
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			WriteErrorResponse(w, errors.BadRequest("Invalid end_time format"))
			return
		}
	} else {
		endTime = time.Now()
	}

	var entityTypePtr *string
	if entityType != "" {
		entityTypePtr = &entityType
	}

	report, err := h.policyService.GetComplianceReport(r.Context(), startTime, endTime, entityTypePtr)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
