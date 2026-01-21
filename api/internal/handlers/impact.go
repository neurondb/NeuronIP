package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
)

/* ImpactHandler handles impact analysis requests */
type ImpactHandler struct {
	impactService *lineage.ImpactService
}

/* NewImpactHandler creates a new impact handler */
func NewImpactHandler(impactService *lineage.ImpactService) *ImpactHandler {
	return &ImpactHandler{impactService: impactService}
}

/* AnalyzeImpact handles POST /api/v1/lineage/impact/analyze */
func (h *ImpactHandler) AnalyzeImpact(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceID   string `json:"resource_id"`
		ResourceType string `json:"resource_type"`
		ImpactType   string `json:"impact_type"` // "upstream", "downstream", "both"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ResourceID == "" || req.ResourceType == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_id and resource_type are required", nil))
		return
	}

	if req.ImpactType == "" {
		req.ImpactType = "both"
	}

	analysis, err := h.impactService.AnalyzeImpact(r.Context(), req.ResourceID, req.ResourceType, req.ImpactType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

/* GetImpactHistory handles GET /api/v1/lineage/impact/{resource_id}/history */
func (h *ImpactHandler) GetImpactHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceID := vars["resource_id"]

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	history, err := h.impactService.GetImpactHistory(r.Context(), resourceID, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
