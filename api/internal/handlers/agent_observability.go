package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* AgentObservabilityHandler handles agent observability requests */
type AgentObservabilityHandler struct {
	tracingService        *agent.TracingService
	evidenceTracker       *agent.EvidenceTracker
	hallucinationDetector  *agent.HallucinationDetector
	auditTrailService     *agent.AuditTrailService
}

/* NewAgentObservabilityHandler creates a new agent observability handler */
func NewAgentObservabilityHandler(
	tracingService *agent.TracingService,
	evidenceTracker *agent.EvidenceTracker,
	hallucinationDetector *agent.HallucinationDetector,
	auditTrailService *agent.AuditTrailService,
) *AgentObservabilityHandler {
	return &AgentObservabilityHandler{
		tracingService:       tracingService,
		evidenceTracker:      evidenceTracker,
		hallucinationDetector: hallucinationDetector,
		auditTrailService:    auditTrailService,
	}
}

/* GetTraces handles GET /api/v1/observability/agent/traces */
func (h *AgentObservabilityHandler) GetTraces(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	var agentIDPtr *string
	if agentID != "" {
		agentIDPtr = &agentID
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	traces, err := h.tracingService.ListTraces(r.Context(), agentIDPtr, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(traces)
}

/* GetTrace handles GET /api/v1/observability/agent/traces/{id} */
func (h *AgentObservabilityHandler) GetTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid trace ID"))
		return
	}

	trace, err := h.tracingService.GetTrace(r.Context(), traceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

/* GetEvidenceCoverage handles GET /api/v1/observability/agent/traces/{id}/evidence */
func (h *AgentObservabilityHandler) GetEvidenceCoverage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid trace ID"))
		return
	}

	coverage, err := h.evidenceTracker.GetCoverage(r.Context(), traceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coverage)
}

/* GetHallucinationRisk handles GET /api/v1/observability/agent/traces/{id}/hallucination */
func (h *AgentObservabilityHandler) GetHallucinationRisk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid trace ID"))
		return
	}

	risk, err := h.hallucinationDetector.GetRisk(r.Context(), traceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risk)
}

/* GetAuditTrail handles GET /api/v1/observability/agent/traces/{id}/audit */
func (h *AgentObservabilityHandler) GetAuditTrail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid trace ID"))
		return
	}

	auditTrail, err := h.auditTrailService.GetAuditTrail(r.Context(), traceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditTrail)
}
