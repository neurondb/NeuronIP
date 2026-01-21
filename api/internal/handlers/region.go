package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/tenancy"
)

/* RegionHandler handles region requests */
type RegionHandler struct {
	service *tenancy.RegionService
}

/* NewRegionHandler creates a new region handler */
func NewRegionHandler(service *tenancy.RegionService) *RegionHandler {
	return &RegionHandler{service: service}
}

/* CreateRegion handles POST /api/v1/regions */
func (h *RegionHandler) CreateRegion(w http.ResponseWriter, r *http.Request) {
	var region tenancy.Region
	if err := json.NewDecoder(r.Body).Decode(&region); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateRegion(r.Context(), region)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetRegion handles GET /api/v1/regions/{id} */
func (h *RegionHandler) GetRegion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid region ID"))
		return
	}

	region, err := h.service.GetRegion(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(region)
}

/* ListRegions handles GET /api/v1/regions */
func (h *RegionHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	regions, err := h.service.ListRegions(r.Context(), activeOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"regions": regions,
		"count":   len(regions),
	})
}

/* CheckRegionHealth handles GET /api/v1/regions/{id}/health */
func (h *RegionHandler) CheckRegionHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid region ID"))
		return
	}

	status, err := h.service.CheckRegionHealth(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"region_id": id,
		"status":    status,
	})
}

/* FailoverToRegion handles POST /api/v1/regions/{id}/failover */
func (h *RegionHandler) FailoverToRegion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid region ID"))
		return
	}

	if err := h.service.FailoverToRegion(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "failover_initiated",
		"region_id": id,
	})
}
