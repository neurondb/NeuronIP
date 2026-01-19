package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
)

/* LineageHandler handles data lineage requests */
type LineageHandler struct {
	service *lineage.LineageService
}

/* NewLineageHandler creates a new lineage handler */
func NewLineageHandler(service *lineage.LineageService) *LineageHandler {
	return &LineageHandler{service: service}
}

/* GetLineage handles GET /api/v1/lineage/{resource_type}/{resource_id} */
func (h *LineageHandler) GetLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["resource_type"]
	resourceID := vars["resource_id"]

	if resourceType == "" || resourceID == "" {
		WriteErrorResponse(w, errors.BadRequest("resource_type and resource_id are required"))
		return
	}

	graph, err := h.service.GetLineage(r.Context(), resourceType, resourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

/* TrackTransformation handles POST /api/v1/lineage/track */
func (h *LineageHandler) TrackTransformation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceID      uuid.UUID              `json:"source_id"`
		TargetID      uuid.UUID              `json:"target_id"`
		EdgeType      string                 `json:"edge_type"`
		Transformation map[string]interface{} `json:"transformation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.TrackTransformation(r.Context(), req.SourceID, req.TargetID, req.EdgeType, req.Transformation); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "tracked",
	})
}

/* GetImpactAnalysis handles GET /api/v1/lineage/impact/{resource_id} */
func (h *LineageHandler) GetImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceID := vars["resource_id"]

	if resourceID == "" {
		WriteErrorResponse(w, errors.BadRequest("resource_id is required"))
		return
	}

	nodes, err := h.service.GetImpactAnalysis(r.Context(), resourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"impacted_nodes": nodes,
		"count":          len(nodes),
	})
}

/* GetFullGraph handles GET /api/v1/lineage/graph */
func (h *LineageHandler) GetFullGraph(w http.ResponseWriter, r *http.Request) {
	graph, err := h.service.GetFullGraph(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}
