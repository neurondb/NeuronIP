package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* DSARHandler handles DSAR requests */
type DSARHandler struct {
	service *compliance.DSARService
}

/* NewDSARHandler creates a new DSAR handler */
func NewDSARHandler(service *compliance.DSARService) *DSARHandler {
	return &DSARHandler{service: service}
}

/* CreateDSARRequest handles POST /api/v1/dsar/requests */
func (h *DSARHandler) CreateDSARRequest(w http.ResponseWriter, r *http.Request) {
	var req compliance.DSARRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreateDSARRequest(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetDSARRequest handles GET /api/v1/dsar/requests/{id} */
func (h *DSARHandler) GetDSARRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid DSAR request ID"))
		return
	}

	req, err := h.service.GetDSARRequest(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

/* ListDSARRequests handles GET /api/v1/dsar/requests */
func (h *DSARHandler) ListDSARRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 100

	requests, err := h.service.ListDSARRequests(r.Context(), status, limit)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"count":    len(requests),
	})
}

/* CompleteDSARRequest handles POST /api/v1/dsar/requests/{id}/complete */
func (h *DSARHandler) CompleteDSARRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid DSAR request ID"))
		return
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&responseData); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.CompleteDSARRequest(r.Context(), id, responseData); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "completed",
		"request_id": id,
	})
}
