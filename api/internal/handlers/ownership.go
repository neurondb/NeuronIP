package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/ownership"
)

/* OwnershipHandler handles ownership requests */
type OwnershipHandler struct {
	service *ownership.Service
}

/* NewOwnershipHandler creates a new ownership handler */
func NewOwnershipHandler(service *ownership.Service) *OwnershipHandler {
	return &OwnershipHandler{service: service}
}

/* AssignOwnership assigns ownership to a resource */
func (h *OwnershipHandler) AssignOwnership(w http.ResponseWriter, r *http.Request) {
	var own ownership.Ownership
	if err := json.NewDecoder(r.Body).Decode(&own); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Get user ID from context
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		own.AssignedBy = &userID
	}

	assigned, err := h.service.AssignOwnership(r.Context(), own)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assigned)
}

/* GetOwnership retrieves ownership for a resource */
func (h *OwnershipHandler) GetOwnership(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["resource_type"]
	resourceID, err := uuid.Parse(vars["resource_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	own, err := h.service.GetOwnership(r.Context(), resourceType, resourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(own)
}

/* ListOwnershipByOwner lists resources owned by a user/team/org */
func (h *OwnershipHandler) ListOwnershipByOwner(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	ownerType := r.URL.Query().Get("owner_type")
	if ownerID == "" || ownerType == "" {
		WriteErrorResponse(w, errors.BadRequest("owner_id and owner_type required"))
		return
	}

	ownerships, err := h.service.ListOwnershipByOwner(r.Context(), ownerID, ownerType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ownerships)
}

/* RemoveOwnership removes ownership from a resource */
func (h *OwnershipHandler) RemoveOwnership(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["resource_type"]
	resourceID, err := uuid.Parse(vars["resource_id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource ID"))
		return
	}

	if err := h.service.RemoveOwnership(r.Context(), resourceType, resourceID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
