package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* AuthEnhancedHandler handles enhanced authentication requests */
type AuthEnhancedHandler struct {
	authService    *auth.AuthService
	oidcService    *auth.OIDCService
	scimService    *auth.SCIMService
	sessionService *auth.SessionService
	twoFactorService *auth.TwoFactorService
}

/* NewAuthEnhancedHandler creates a new enhanced auth handler */
func NewAuthEnhancedHandler(
	authService *auth.AuthService,
	oidcService *auth.OIDCService,
	scimService *auth.SCIMService,
	sessionService *auth.SessionService,
	twoFactorService *auth.TwoFactorService,
) *AuthEnhancedHandler {
	return &AuthEnhancedHandler{
		authService:     authService,
		oidcService:     oidcService,
		scimService:     scimService,
		sessionService:  sessionService,
		twoFactorService: twoFactorService,
	}
}

/* InitiateOIDC handles OIDC SSO initiation */
func (h *AuthEnhancedHandler) InitiateOIDC(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerName := vars["provider"]

	authURL, state, err := h.oidcService.InitiateOIDCFlow(r.Context(), providerName)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

/* HandleOIDCCallback handles OIDC callback */
func (h *AuthEnhancedHandler) HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerName := vars["provider"]

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	user, sessionToken, err := h.oidcService.HandleOIDCCallback(r.Context(), providerName, code, state)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":         user,
		"session_token": sessionToken,
	})
}

/* HandleSCIM handles SCIM requests */
func (h *AuthEnhancedHandler) HandleSCIM(w http.ResponseWriter, r *http.Request) {
	h.scimService.HandleSCIMRequest(w, r)
}

/* GenerateTOTPSecret handles TOTP secret generation */
func (h *AuthEnhancedHandler) GenerateTOTPSecret(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid user ID"))
		return
	}

	secret, err := h.twoFactorService.GenerateTOTPSecret(r.Context(), userID, req.Email)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

/* GetUserSessions handles listing user sessions */
func (h *AuthEnhancedHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
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

	sessions, err := h.sessionService.GetUserSessions(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

/* RevokeSession handles revoking a session */
func (h *AuthEnhancedHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid session ID"))
		return
	}

	if err := h.sessionService.RevokeSession(r.Context(), sessionID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
