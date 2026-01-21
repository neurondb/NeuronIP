package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
)

/* SavedSearchHandler handles saved search requests */
type SavedSearchHandler struct {
	service         *warehouse.SavedSearchService
	warehouseService *warehouse.Service
}

/* NewSavedSearchHandler creates a new saved search handler */
func NewSavedSearchHandler(service *warehouse.SavedSearchService, warehouseService *warehouse.Service) *SavedSearchHandler {
	return &SavedSearchHandler{
		service:         service,
		warehouseService: warehouseService,
	}
}

/* CreateSavedSearch handles creating a saved search */
func (h *SavedSearchHandler) CreateSavedSearch(w http.ResponseWriter, r *http.Request) {
	var search warehouse.SavedSearch
	if err := json.NewDecoder(r.Body).Decode(&search); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateSavedSearch(r.Context(), search)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetSavedSearch handles retrieving a saved search */
func (h *SavedSearchHandler) GetSavedSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	searchID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid search ID"))
		return
	}

	search, err := h.service.GetSavedSearch(r.Context(), searchID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(search)
}

/* ListSavedSearches handles listing saved searches */
func (h *SavedSearchHandler) ListSavedSearches(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	publicOnly := r.URL.Query().Get("public_only") == "true"

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	searches, err := h.service.ListSavedSearches(r.Context(), userIDPtr, publicOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}

/* ExecuteSavedSearch handles executing a saved search */
func (h *SavedSearchHandler) ExecuteSavedSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	searchID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid search ID"))
		return
	}

	result, err := h.service.ExecuteSavedSearch(r.Context(), searchID, h.warehouseService)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/* UpdateSavedSearch handles updating a saved search */
func (h *SavedSearchHandler) UpdateSavedSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	searchID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid search ID"))
		return
	}

	var search warehouse.SavedSearch
	if err := json.NewDecoder(r.Body).Decode(&search); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateSavedSearch(r.Context(), searchID, search); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* DeleteSavedSearch handles deleting a saved search */
func (h *SavedSearchHandler) DeleteSavedSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	searchID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid search ID"))
		return
	}

	if err := h.service.DeleteSavedSearch(r.Context(), searchID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
