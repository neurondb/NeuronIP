package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/session"
)

/* AuthEnhancedHandler handles enhanced authentication requests */
type AuthEnhancedHandler struct {
	authService     *auth.AuthService
	oidcService     *auth.OIDCService
	scimService     *auth.SCIMService
	sessionService  *auth.SessionService
	twoFactorService *auth.TwoFactorService
	sessionManager  *session.Manager
}

/* NewAuthEnhancedHandler creates a new enhanced auth handler */
func NewAuthEnhancedHandler(
	authService *auth.AuthService,
	oidcService *auth.OIDCService,
	scimService *auth.SCIMService,
	sessionService *auth.SessionService,
	twoFactorService *auth.TwoFactorService,
	sessionManager *session.Manager,
) *AuthEnhancedHandler {
	return &AuthEnhancedHandler{
		authService:      authService,
		oidcService:      oidcService,
		scimService:      scimService,
		sessionService:   sessionService,
		twoFactorService: twoFactorService,
		sessionManager:   sessionManager,
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

/* Login handles user login with username/password */
func (h *AuthEnhancedHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database,omitempty"` // neuronip or neuronai-demo
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteErrorResponse(w, errors.BadRequest("Username and password are required"))
		return
	}

	// Default to neuronip
	if req.Database == "" {
		req.Database = "neuronip"
	}
	if req.Database != "neuronip" && req.Database != "neuronai-demo" {
		WriteErrorResponse(w, errors.BadRequest("Database must be 'neuronip' or 'neuronai-demo'"))
		return
	}

	// Get IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Authenticate user
	user, sess, refreshToken, err := h.authService.LoginWithUsername(r.Context(), req.Username, req.Password, req.Database, ipAddress, userAgent)
	if err != nil {
		WriteErrorResponse(w, errors.Unauthorized(err.Error()))
		return
	}

	// Set cookies
	h.sessionManager.SetCookies(w, sess.ID, refreshToken)

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": req.Username,
			"name":     user.Name,
			"role":     user.Role,
		},
		"database": req.Database,
		"token":    sess.ID, // For backward compatibility
	})
}

/* Register handles user registration with username/password */
func (h *AuthEnhancedHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database,omitempty"` // neuronip or neuronai-demo
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteErrorResponse(w, errors.BadRequest("Username and password are required"))
		return
	}

	// Validate username length
	if len(req.Username) < 3 {
		WriteErrorResponse(w, errors.BadRequest("Username must be at least 3 characters"))
		return
	}
	if len(req.Username) > 100 {
		WriteErrorResponse(w, errors.BadRequest("Username must be less than 100 characters"))
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		WriteErrorResponse(w, errors.BadRequest("Password must be at least 6 characters"))
		return
	}
	if len(req.Password) > 1000 {
		WriteErrorResponse(w, errors.BadRequest("Password too long"))
		return
	}

	// Default to neuronip
	if req.Database == "" {
		req.Database = "neuronip"
	}
	if err := session.ValidateDatabaseName(req.Database); err != nil {
		WriteErrorResponse(w, errors.BadRequest(err.Error()))
		return
	}

	// Get IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// Register user
	user, sess, refreshToken, err := h.authService.RegisterWithUsername(r.Context(), req.Username, req.Password, req.Database, ipAddress, userAgent)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest(err.Error()))
		return
	}

	// Set cookies
	h.sessionManager.SetCookies(w, sess.ID, refreshToken)

	// Return user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"email":    user.Email,
			"username": req.Username,
			"name":     user.Name,
			"role":     user.Role,
		},
		"database": req.Database,
		"token":    sess.ID, // For backward compatibility
	})
}

/* GetCurrentUser returns the current authenticated user */
func (h *AuthEnhancedHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetCurrentUser(r.Context())
	if err != nil {
		WriteErrorResponse(w, errors.Unauthorized("Not authenticated"))
		return
	}

	// Get database from session
	database := "neuronip"
	if sess, ok := session.GetSessionFromContext(r.Context()); ok {
		database = sess.Database
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":     user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"role":   user.Role,
		},
		"database": database,
	})
}

/* Logout handles user logout */
func (h *AuthEnhancedHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.GetSessionFromContext(r.Context())
	if !ok {
		// No session, just clear cookies
		h.sessionManager.ClearCookies(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Revoke session
	if err := h.sessionManager.RevokeSession(r.Context(), sess.ID); err != nil {
		WriteError(w, err)
		return
	}

	// Clear cookies
	h.sessionManager.ClearCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

/* RefreshToken handles refresh token rotation */
func (h *AuthEnhancedHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := h.sessionManager.GetRefreshTokenFromRequest(r)
	if refreshToken == "" {
		WriteErrorResponse(w, errors.Unauthorized("Refresh token required"))
		return
	}

	// Refresh session
	sess, newAccessToken, newRefreshToken, err := h.sessionManager.RefreshSession(r.Context(), refreshToken)
	if err != nil {
		WriteErrorResponse(w, errors.Unauthorized(err.Error()))
		return
	}

	// Set new cookies
	h.sessionManager.SetCookies(w, newAccessToken, newRefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"database":      sess.Database,
	})
}

/* Helper function to get client IP */
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			// Remove port if present
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			return ip
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		// Remove port if present
		if idx := strings.LastIndex(realIP, ":"); idx != -1 {
			realIP = realIP[:idx]
		}
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
