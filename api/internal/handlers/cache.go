package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* CacheHandler handles query cache requests */
type CacheHandler struct {
	service *warehouse.CacheService
}

/* NewCacheHandler creates a new cache handler */
func NewCacheHandler(service *warehouse.CacheService) *CacheHandler {
	return &CacheHandler{service: service}
}

/* GetCachedResult handles getting cached query results */
func (h *CacheHandler) GetCachedResult(w http.ResponseWriter, r *http.Request) {
	cacheKey := r.URL.Query().Get("cache_key")
	if cacheKey == "" {
		WriteErrorResponse(w, errors.BadRequest("cache_key is required"))
		return
	}

	entry, err := h.service.GetCachedResult(r.Context(), cacheKey)
	if err != nil {
		WriteErrorResponse(w, errors.NotFound("Cache entry not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

/* InvalidateCache handles cache invalidation */
func (h *CacheHandler) InvalidateCache(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	rule := warehouse.InvalidationRule{
		Type:  req.Type,
		Value: req.Value,
	}

	if err := h.service.InvalidateCache(r.Context(), rule); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* GetCacheStats handles getting cache statistics */
func (h *CacheHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetCacheStats(r.Context())
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
