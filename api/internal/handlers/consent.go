package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ConsentHandler handles consent requests */
type ConsentHandler struct {
	service *compliance.ConsentService
}

/* NewConsentHandler creates a new consent handler */
func NewConsentHandler(service *compliance.ConsentService) *ConsentHandler {
	return &ConsentHandler{service: service}
}

/* RecordConsent handles POST /api/v1/consent */
func (h *ConsentHandler) RecordConsent(w http.ResponseWriter, r *http.Request) {
	var consent compliance.ConsentRecord
	if err := json.NewDecoder(r.Body).Decode(&consent); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	recorded, err := h.service.RecordConsent(r.Context(), consent)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(recorded)
}

/* WithdrawConsent handles POST /api/v1/consent/withdraw */
func (h *ConsentHandler) WithdrawConsent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SubjectID   string `json:"subject_id"`
		ConsentType string `json:"consent_type"`
		Purpose     string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.WithdrawConsent(r.Context(), req.SubjectID, req.ConsentType, req.Purpose); err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "withdrawn",
	})
}

/* CheckConsent handles GET /api/v1/consent/{subject_id} */
func (h *ConsentHandler) CheckConsent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subjectID := vars["subject_id"]
	consentType := r.URL.Query().Get("type")
	purpose := r.URL.Query().Get("purpose")

	if consentType == "" || purpose == "" {
		WriteErrorResponse(w, errors.BadRequest("consent_type and purpose are required"))
		return
	}

	consented, err := h.service.CheckConsent(r.Context(), subjectID, consentType, purpose)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subject_id":   subjectID,
		"consent_type": consentType,
		"purpose":      purpose,
		"consented":    consented,
	})
}

/* GetSubjectConsents handles GET /api/v1/consent/subject/{subject_id} */
func (h *ConsentHandler) GetSubjectConsents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subjectID := vars["subject_id"]

	consents, err := h.service.GetSubjectConsents(r.Context(), subjectID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"consents": consents,
		"count":    len(consents),
	})
}
