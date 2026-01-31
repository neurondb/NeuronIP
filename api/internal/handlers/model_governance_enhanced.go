package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/models"
)

/* ModelGovernanceEnhancedHandler handles model governance requests */
type ModelGovernanceEnhancedHandler struct {
	governanceService *models.ModelGovernanceService
	monitoringService *models.ModelMonitoringService
}

/* NewModelGovernanceEnhancedHandler creates a new model governance enhanced handler */
func NewModelGovernanceEnhancedHandler(
	governanceService *models.ModelGovernanceService,
	monitoringService *models.ModelMonitoringService,
) *ModelGovernanceEnhancedHandler {
	return &ModelGovernanceEnhancedHandler{
		governanceService: governanceService,
		monitoringService: monitoringService,
	}
}

/* CreateApprovalWorkflow handles POST /api/v1/models/governance/approval-workflows */
func (h *ModelGovernanceEnhancedHandler) CreateApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	var workflow models.ApprovalWorkflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	createdWorkflow, err := h.governanceService.CreateApprovalWorkflow(r.Context(), workflow)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdWorkflow)
}

/* CheckCompliance handles POST /api/v1/models/{id}/compliance/check */
func (h *ModelGovernanceEnhancedHandler) CheckCompliance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	check, err := h.governanceService.CheckCompliance(r.Context(), modelID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(check)
}

/* RecordPrediction handles POST /api/v1/models/{id}/predictions */
func (h *ModelGovernanceEnhancedHandler) RecordPrediction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	var prediction models.ModelPrediction
	if err := json.NewDecoder(r.Body).Decode(&prediction); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	prediction.ModelID = modelID
	if err := h.monitoringService.RecordPrediction(r.Context(), prediction); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "recorded",
	})
}

/* DetectDrift handles POST /api/v1/models/{id}/drift/detect */
func (h *ModelGovernanceEnhancedHandler) DetectDrift(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid model ID"))
		return
	}

	detection, err := h.monitoringService.DetectDataDrift(r.Context(), modelID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detection)
}
