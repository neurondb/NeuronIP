package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/versioning"
)

/* VersioningHandler handles versioning requests */
type VersioningHandler struct {
	service *versioning.VersioningService
}

/* NewVersioningHandler creates a new versioning handler */
func NewVersioningHandler(service *versioning.VersioningService) *VersioningHandler {
	return &VersioningHandler{service: service}
}

/* ListVersions handles GET /api/v1/versions/{resource_type}/{resource_id} */
func (h *VersioningHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["resource_type"]
	resourceIDStr := vars["resource_id"]

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid resource_id"))
		return
	}

	versions, err := h.service.ListVersions(r.Context(), resourceType, resourceID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": versions,
		"count":    len(versions),
	})
}

/* CreateVersion handles POST /api/v1/versions/create */
func (h *VersioningHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	var req versioning.Version
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.ResourceType == "" || req.VersionNumber == "" {
		WriteErrorResponse(w, errors.ValidationFailed("resource_type and version_number are required", nil))
		return
	}

	version, err := h.service.CreateVersion(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(version)
}

/* RollbackVersion handles POST /api/v1/versions/{id}/rollback */
func (h *VersioningHandler) RollbackVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid version ID"))
		return
	}

	if err := h.service.RollbackVersion(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "rolled_back",
		"id":     id,
	})
}

/* GetVersion handles GET /api/v1/versions/{id} */
func (h *VersioningHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid version ID"))
		return
	}

	version, err := h.service.GetVersion(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

/* GetVersionHistory handles GET /api/v1/versions/{id}/history */
func (h *VersioningHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid version ID"))
		return
	}

	history, err := h.service.GetVersionHistory(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}
