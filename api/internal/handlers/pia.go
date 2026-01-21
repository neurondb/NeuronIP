package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* PIAHandler handles PIA requests */
type PIAHandler struct {
	service *compliance.PIAService
}

/* NewPIAHandler creates a new PIA handler */
func NewPIAHandler(service *compliance.PIAService) *PIAHandler {
	return &PIAHandler{service: service}
}

/* CreatePIARequest handles POST /api/v1/pia/requests */
func (h *PIAHandler) CreatePIARequest(w http.ResponseWriter, r *http.Request) {
	var req compliance.PIARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.service.CreatePIARequest(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetPIARequest handles GET /api/v1/pia/requests/{id} */
func (h *PIAHandler) GetPIARequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid PIA request ID"))
		return
	}

	req, err := h.service.GetPIARequest(r.Context(), id)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

/* SubmitPIARequest handles POST /api/v1/pia/requests/{id}/submit */
func (h *PIAHandler) SubmitPIARequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid PIA request ID"))
		return
	}

	if err := h.service.SubmitPIARequest(r.Context(), id); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "submitted",
		"request_id": id,
	})
}

/* ReviewPIARequest handles POST /api/v1/pia/requests/{id}/review */
func (h *PIAHandler) ReviewPIARequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid PIA request ID"))
		return
	}

	var review struct {
		ReviewerID string `json:"reviewer_id"`
		Approved   bool   `json:"approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ReviewPIARequest(r.Context(), id, review.ReviewerID, review.Approved); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "reviewed",
		"request_id": id,
		"approved":   review.Approved,
	})
}
