package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/email"
)

/* AuthHandler handles authentication requests */
type AuthHandler struct {
	authService  *auth.AuthService
	oauthService *auth.OAuthService
	emailService *email.Service
}

/* NewAuthHandler creates a new auth handler */
func NewAuthHandler(authService *auth.AuthService, oauthService *auth.OAuthService, emailService *email.Service) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		oauthService: oauthService,
		emailService: emailService,
	}
}

/* RegisterRequest represents registration request */
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

/* LoginRequest represents login request */
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

/* AuthResponse represents auth response */
type AuthResponse struct {
	User         interface{} `json:"user"`
	SessionToken string      `json:"session_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    string      `json:"expires_at"`
}

/* Register handles user registration */
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteErrorResponse(w, errors.ValidationFailed("email and password are required", nil))
		return
	}

	authReq := auth.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	response, err := h.authService.RegisterUser(r.Context(), authReq)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Send verification email
	go h.emailService.SendVerificationEmail(r.Context(), response.User.ID, req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		User:         response.User,
		SessionToken: response.SessionToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

/* Login handles user login */
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteErrorResponse(w, errors.ValidationFailed("email and password are required", nil))
		return
	}

	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	authReq := auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		TOTPCode: req.TOTPCode,
	}

	response, err := h.authService.LoginUser(r.Context(), authReq, ipAddress, userAgent)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		User:         response.User,
		SessionToken: response.SessionToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

/* Logout handles user logout */
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Session ID would be extracted from token in middleware
	// For now, return success
	w.WriteHeader(http.StatusOK)
}

/* Refresh handles token refresh */
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.RefreshToken == "" {
		WriteErrorResponse(w, errors.ValidationFailed("refresh_token is required", nil))
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		User:         response.User,
		SessionToken: response.SessionToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

/* InitiateOAuth initiates OAuth flow */
func (h *AuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	authURL, err := h.oauthService.InitiateOAuth(r.Context(), provider, "")
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"auth_url": authURL,
	})
}

/* OAuthCallback handles OAuth callback */
func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		WriteErrorResponse(w, errors.BadRequest("code parameter required"))
		return
	}

	user, _, err := h.oauthService.HandleOAuthCallback(r.Context(), provider, code, state)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}

/* VerifyEmail handles email verification */
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	userID, err := h.emailService.VerifyEmailToken(r.Context(), req.Token)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"verified": true,
	})
}

/* RequestPasswordReset handles password reset request */
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Get user by email would be needed here
	// For now, return success (don't reveal if email exists)
	w.WriteHeader(http.StatusOK)
}

/* ConfirmPasswordReset handles password reset confirmation */
func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		WriteErrorResponse(w, errors.ValidationFailed("token and new_password are required", nil))
		return
	}

	userID, err := h.emailService.VerifyPasswordResetToken(r.Context(), req.Token)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Update password would be done here
	_ = userID
	_ = req.NewPassword

	w.WriteHeader(http.StatusOK)
}
