package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
)

/* EndToEndHandler handles end-to-end lineage requests */
type EndToEndHandler struct {
	endToEndService *lineage.EndToEndService
}

/* NewEndToEndHandler creates a new end-to-end handler */
func NewEndToEndHandler(endToEndService *lineage.EndToEndService) *EndToEndHandler {
	return &EndToEndHandler{endToEndService: endToEndService}
}

/* GetEndToEndLineage handles GET /api/v1/lineage/end-to-end */
func (h *EndToEndHandler) GetEndToEndLineage(w http.ResponseWriter, r *http.Request) {
	sourceSystem := r.URL.Query().Get("source_system")
	targetSystem := r.URL.Query().Get("target_system")
	sourceResourceID := r.URL.Query().Get("source_resource_id")

	if sourceSystem == "" || targetSystem == "" || sourceResourceID == "" {
		WriteErrorResponse(w, errors.BadRequest("source_system, target_system, and source_resource_id are required"))
		return
	}

	path, err := h.endToEndService.GetEndToEndLineage(r.Context(), sourceSystem, targetSystem, sourceResourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(path)
}

/* TrackCrossSystemLineage handles POST /api/v1/lineage/cross-system */
func (h *EndToEndHandler) TrackCrossSystemLineage(w http.ResponseWriter, r *http.Request) {
	var crossSystemLineage lineage.CrossSystemLineage
	if err := json.NewDecoder(r.Body).Decode(&crossSystemLineage); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.endToEndService.TrackCrossSystemLineage(r.Context(), crossSystemLineage); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "tracked",
		"id":     crossSystemLineage.ID,
	})
}

/* GetCrossSystemLineage handles GET /api/v1/lineage/cross-system/{system} */
func (h *EndToEndHandler) GetCrossSystemLineage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	system := vars["system"]

	crossSystemLineage, err := h.endToEndService.GetCrossSystemLineage(r.Context(), system)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(crossSystemLineage)
}

/* FindLineagePaths handles POST /api/v1/lineage/paths */
func (h *EndToEndHandler) FindLineagePaths(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceResourceID string `json:"source_resource_id"`
		TargetResourceID string `json:"target_resource_id"`
		MaxDepth         int    `json:"max_depth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.MaxDepth == 0 {
		req.MaxDepth = 5
	}

	paths, err := h.endToEndService.FindLineagePaths(r.Context(), req.SourceResourceID, req.TargetResourceID, req.MaxDepth)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paths)
}
