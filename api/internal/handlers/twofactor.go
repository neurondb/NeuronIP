package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* TwoFactorHandler handles 2FA requests */
type TwoFactorHandler struct {
	service *auth.TwoFactorService
	queries *db.Queries
}

/* NewTwoFactorHandler creates a new 2FA handler */
func NewTwoFactorHandler(service *auth.TwoFactorService, queries *db.Queries) *TwoFactorHandler {
	return &TwoFactorHandler{
		service: service,
		queries: queries,
	}
}

/* Setup2FA handles 2FA setup */
func (h *TwoFactorHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	secret, err := h.service.GenerateTOTPSecret(r.Context(), userID, user.Email)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

/* Verify2FA handles 2FA verification */
func (h *TwoFactorHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	valid, err := h.service.VerifyTOTP(req.Secret, req.Code)
	if err != nil || !valid {
		WriteErrorResponse(w, errors.ValidationFailed("Invalid 2FA code", nil))
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* Enable2FA handles enabling 2FA */
func (h *TwoFactorHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	var req struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.Enable2FA(r.Context(), userID, req.Secret); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

/* Disable2FA handles disabling 2FA */
func (h *TwoFactorHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	if err := h.service.Disable2FA(r.Context(), userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
