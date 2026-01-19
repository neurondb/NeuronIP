package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* CatalogHandler handles data catalog requests */
type CatalogHandler struct {
	service *catalog.CatalogService
}

/* NewCatalogHandler creates a new catalog handler */
func NewCatalogHandler(service *catalog.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

/* ListDatasets handles GET /api/v1/catalog/datasets */
func (h *CatalogHandler) ListDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := h.service.ListDatasets(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"datasets": datasets,
		"count":    len(datasets),
	})
}

/* GetDataset handles GET /api/v1/catalog/datasets/{id} */
func (h *CatalogHandler) GetDataset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid dataset ID"))
		return
	}

	dataset, err := h.service.GetDataset(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataset)
}

/* SearchDatasets handles GET /api/v1/catalog/search */
func (h *CatalogHandler) SearchDatasets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		WriteErrorResponse(w, errors.ValidationFailed("query is required", nil))
		return
	}

	datasets, err := h.service.SearchDatasets(r.Context(), query)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"datasets": datasets,
		"count":    len(datasets),
	})
}

/* ListOwners handles GET /api/v1/catalog/owners */
func (h *CatalogHandler) ListOwners(w http.ResponseWriter, r *http.Request) {
	owners, err := h.service.ListOwners(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"owners": owners,
		"count":  len(owners),
	})
}

/* DiscoverDatasets handles POST /api/v1/catalog/discover */
func (h *CatalogHandler) DiscoverDatasets(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	datasets, err := h.service.DiscoverDatasets(r.Context(), req.Tags)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"datasets": datasets,
		"count":    len(datasets),
	})
}
