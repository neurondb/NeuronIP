package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ComplianceHandler handles compliance requests */
type ComplianceHandler struct {
	service        *compliance.Service
	anomalyService *compliance.AnomalyService
}

/* NewComplianceHandler creates a new compliance handler */
func NewComplianceHandler(service *compliance.Service, anomalyService *compliance.AnomalyService) *ComplianceHandler {
	return &ComplianceHandler{
		service:        service,
		anomalyService: anomalyService,
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
