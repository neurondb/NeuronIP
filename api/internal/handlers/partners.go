package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/partners"
)

/* PartnerHandler handles partner management requests */
type PartnerHandler struct {
	service *partners.PartnerService
}

/* NewPartnerHandler creates a new partner handler */
func NewPartnerHandler(service *partners.PartnerService) *PartnerHandler {
	return &PartnerHandler{service: service}
}

/* RegisterPartner handles POST /api/v1/partners */
func (h *PartnerHandler) RegisterPartner(w http.ResponseWriter, r *http.Request) {
	var partner partners.Partner
	if err := json.NewDecoder(r.Body).Decode(&partner); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	registeredPartner, err := h.service.RegisterPartner(r.Context(), partner)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registeredPartner)
}

/* GetPartner handles GET /api/v1/partners/{id} */
func (h *PartnerHandler) GetPartner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	partnerID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid partner ID"))
		return
	}

	// Implementation would get partner by ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": partnerID,
	})
}
