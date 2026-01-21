package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* APIKeyEnhancedHandler handles enhanced API key requests */
type APIKeyEnhancedHandler struct {
	service *auth.APIKeyService
	queries *db.Queries
}

/* NewAPIKeyEnhancedHandler creates a new enhanced API key handler */
func NewAPIKeyEnhancedHandler(service *auth.APIKeyService, queries *db.Queries) *APIKeyEnhancedHandler {
	return &APIKeyEnhancedHandler{
		service: service,
		queries: queries,
	}
}

/* CreateAPIKey handles creating an API key with enhanced features */
func (h *APIKeyEnhancedHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		WriteErrorResponse(w, errors.Unauthorized("User not authenticated"))
		return
	}

	var req struct {
		Name                string    `json:"name"`
		Scopes              []string  `json:"scopes"`
		RateLimit           int       `json:"rate_limit"`
		ExpiresAt           *time.Time `json:"expires_at"`
		RotationEnabled     bool      `json:"rotation_enabled"`
		RotationIntervalDays int      `json:"rotation_interval_days"`
		Tags                []string  `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	apiKeyReq := auth.CreateAPIKeyRequest{
		Name:                req.Name,
		UserID:              &userID,
		Scopes:              req.Scopes,
		RateLimit:           req.RateLimit,
		ExpiresAt:           req.ExpiresAt,
		RotationEnabled:     req.RotationEnabled,
		RotationIntervalDays: req.RotationIntervalDays,
		Tags:                req.Tags,
	}

	created, err := h.service.CreateAPIKey(r.Context(), apiKeyReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* RotateAPIKey handles rotating an API key */
func (h *APIKeyEnhancedHandler) RotateAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid API key ID"))
		return
	}

	var req struct {
		RotationType string `json:"rotation_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.RotationType = "manual"
	}

	rotated, err := h.service.RotateAPIKey(r.Context(), keyID, req.RotationType)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rotated)
}

/* GetUsageAnalytics handles getting API key usage analytics */
func (h *APIKeyEnhancedHandler) GetUsageAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid API key ID"))
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startDate := time.Now().AddDate(0, 0, -30) // Default: last 30 days
	endDate := time.Now()

	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	analytics, err := h.service.GetUsageAnalytics(r.Context(), keyID, startDate, endDate)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

/* RevokeAPIKey handles revoking an API key */
func (h *APIKeyEnhancedHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid API key ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.RevokeAPIKey(r.Context(), keyID, req.Reason); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
