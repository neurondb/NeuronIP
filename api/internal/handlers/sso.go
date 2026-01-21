package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* SSOHandler handles SSO authentication requests */
type SSOHandler struct {
	ssoService *auth.SSOService
}

/* NewSSOHandler creates a new SSO handler */
func NewSSOHandler(ssoService *auth.SSOService) *SSOHandler {
	return &SSOHandler{ssoService: ssoService}
}

/* CreateProvider creates a new SSO provider */
func (h *SSOHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var provider auth.SSOProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid request body"))
		return
	}

	created, err := h.ssoService.CreateProvider(r.Context(), provider)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

/* GetProvider retrieves an SSO provider */
func (h *SSOHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid provider ID"))
		return
	}

	provider, err := h.ssoService.GetProvider(r.Context(), id)
	if err != nil {
		WriteErrorResponse(w, errors.NotFound("Provider not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

/* ListProviders lists all SSO providers */
func (h *SSOHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"
	
	providers, err := h.ssoService.ListProviders(r.Context(), enabledOnly)
	if err != nil {
		WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

/* InitiateSSO initiates SSO flow */
func (h *SSOHandler) InitiateSSO(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID, err := uuid.Parse(vars["id"])
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid provider ID"))
		return
	}

	provider, err := h.ssoService.GetProvider(r.Context(), providerID)
	if err != nil {
		WriteErrorResponse(w, errors.NotFound("Provider not found"))
		return
	}

	relayState := r.URL.Query().Get("relay_state")
	state := r.URL.Query().Get("state")

	var redirectURL string
	if provider.ProviderType == "saml" {
		redirectURL, err = h.ssoService.InitiateSAML(r.Context(), providerID, relayState)
	} else if provider.ProviderType == "oauth2" || provider.ProviderType == "oidc" {
		redirectURL, err = h.ssoService.InitiateOAuth2(r.Context(), providerID, state)
	} else {
		WriteErrorResponse(w, errors.BadRequest("Unsupported provider type"))
		return
	}

	if err != nil {
		WriteError(w, err)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

/* SSOCallback handles SSO callback */
func (h *SSOHandler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	providerIDStr := r.URL.Query().Get("provider_id")
	if providerIDStr == "" {
		WriteErrorResponse(w, errors.BadRequest("Provider ID required"))
		return
	}

	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		WriteErrorResponse(w, errors.BadRequest("Invalid provider ID"))
		return
	}

	provider, err := h.ssoService.GetProvider(r.Context(), providerID)
	if err != nil {
		WriteErrorResponse(w, errors.NotFound("Provider not found"))
		return
	}

	var user *auth.SSOUser
	if provider.ProviderType == "saml" {
		samlResponse := r.FormValue("SAMLResponse")
		relayState := r.FormValue("RelayState")
		user, err = h.ssoService.ProcessSAMLCallback(r.Context(), providerID, samlResponse, relayState)
	} else if provider.ProviderType == "oauth2" || provider.ProviderType == "oidc" {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		user, err = h.ssoService.ProcessOAuth2Callback(r.Context(), providerID, code, state)
	} else {
		WriteErrorResponse(w, errors.BadRequest("Unsupported provider type"))
		return
	}

	if err != nil {
		WriteError(w, err)
		return
	}

	// Return user info and session token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       user.UserID,
		"session_token": user.SessionToken,
		"email":         user.Email,
		"name":          user.Name,
	})
}

/* ValidateSession validates an SSO session */
func (h *SSOHandler) ValidateSession(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		WriteErrorResponse(w, errors.Unauthorized("Session token required"))
		return
	}

	session, err := h.ssoService.ValidateSession(r.Context(), sessionToken)
	if err != nil {
		WriteErrorResponse(w, errors.Unauthorized("Invalid session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
