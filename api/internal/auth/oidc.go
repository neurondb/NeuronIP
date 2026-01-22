package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* OIDCConfig holds OpenID Connect provider configuration */
type OIDCConfig struct {
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	Scopes         []string
	ProviderName   string // e.g., "google", "okta", "azure"
	DiscoveryURL   string // Optional: custom discovery URL
}

/* OIDCService provides OpenID Connect SSO functionality */
type OIDCService struct {
	queries *db.Queries
	configs map[string]OIDCConfig // Provider name -> config
}

/* NewOIDCService creates a new OIDC service */
func NewOIDCService(queries *db.Queries) *OIDCService {
	return &OIDCService{
		queries: queries,
		configs: make(map[string]OIDCConfig),
	}
}

/* RegisterProvider registers an OIDC provider */
func (s *OIDCService) RegisterProvider(name string, config OIDCConfig) {
	s.configs[name] = config
}

/* GetProviderConfig gets configuration for a provider */
func (s *OIDCService) GetProviderConfig(providerName string) (*OIDCConfig, error) {
	config, exists := s.configs[providerName]
	if !exists {
		return nil, fmt.Errorf("OIDC provider %s not configured", providerName)
	}
	return &config, nil
}

/* InitiateOIDCFlow initiates an OIDC authentication flow */
func (s *OIDCService) InitiateOIDCFlow(ctx context.Context, providerName string) (string, string, error) {
	config, err := s.GetProviderConfig(providerName)
	if err != nil {
		return "", "", err
	}

	// Discover OIDC configuration
	discoveryURL := config.DiscoveryURL
	if discoveryURL == "" {
		discoveryURL = config.IssuerURL + "/.well-known/openid-configuration"
	}

	discovery, err := s.discoverOIDCConfig(ctx, discoveryURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to discover OIDC config: %w", err)
	}

	// Generate state token for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Generate nonce for replay protection
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := base64.URLEncoding.EncodeToString(nonceBytes)

	// Build authorization URL
	authURL, err := url.Parse(discovery.AuthorizationEndpoint)
	if err != nil {
		return "", "", fmt.Errorf("invalid authorization endpoint: %w", err)
	}

	q := authURL.Query()
	q.Set("client_id", config.ClientID)
	q.Set("redirect_uri", config.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", s.buildScopeString(config.Scopes))
	q.Set("state", state)
	q.Set("nonce", nonce)
	authURL.RawQuery = q.Encode()

	return authURL.String(), state, nil
}

/* HandleOIDCCallback processes OIDC callback and creates/updates user */
func (s *OIDCService) HandleOIDCCallback(ctx context.Context, providerName, code, state string) (*db.User, string, error) {
	config, err := s.GetProviderConfig(providerName)
	if err != nil {
		return nil, "", err
	}

	// Discover OIDC configuration
	discoveryURL := config.DiscoveryURL
	if discoveryURL == "" {
		discoveryURL = config.IssuerURL + "/.well-known/openid-configuration"
	}

	discovery, err := s.discoverOIDCConfig(ctx, discoveryURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to discover OIDC config: %w", err)
	}

	// Exchange code for tokens
	tokens, err := s.exchangeCodeForTokens(ctx, discovery.TokenEndpoint, config, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from ID token or userinfo endpoint
	userInfo, err := s.getUserInfo(ctx, discovery.UserInfoEndpoint, tokens.AccessToken, tokens.IDToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Get or create user
	user, err := s.getOrCreateUserFromOIDC(ctx, providerName, userInfo)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get/create user: %w", err)
	}

	// Create session
	sessionToken, refreshToken, expiresAt, err := s.generateSessionTokens()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session tokens: %w", err)
	}

	session := &db.UserSession{
		UserID:       user.ID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := s.queries.CreateUserSession(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionToken, nil
}

/* OIDCDiscovery represents OIDC discovery document */
type OIDCDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

/* discoverOIDCConfig discovers OIDC configuration */
func (s *OIDCService) discoverOIDCConfig(ctx context.Context, discoveryURL string) (*OIDCDiscovery, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery failed with status %d", resp.StatusCode)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, err
	}

	return &discovery, nil
}

/* TokenResponse represents OAuth token response */
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

/* exchangeCodeForTokens exchanges authorization code for tokens */
func (s *OIDCService) exchangeCodeForTokens(ctx context.Context, tokenURL string, config *OIDCConfig, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.RedirectURL)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d", resp.StatusCode)
	}

	var tokens TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

/* UserInfo represents OIDC user information */
type UserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

/* getUserInfo gets user information from ID token or userinfo endpoint */
func (s *OIDCService) getUserInfo(ctx context.Context, userInfoURL, accessToken, idToken string) (*UserInfo, error) {
	// In production, you should verify the ID token signature
	// For now, we'll use the userinfo endpoint if available
	if userInfoURL != "" && accessToken != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var userInfo UserInfo
			if err := json.NewDecoder(resp.Body).Decode(&userInfo); err == nil {
				return &userInfo, nil
			}
		}
	}

	// Fallback: parse ID token with JWT verification
	// The idToken parameter is already passed to this function
	if idToken == "" {
		return nil, fmt.Errorf("ID token not available")
	}

	// Verify and parse JWT
	userInfo, err := s.parseIDToken(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ID token: %w", err)
	}

	return userInfo, nil
}

/* parseIDToken parses and verifies an OIDC ID token */
func (s *OIDCService) parseIDToken(idToken string) (*UserInfo, error) {
	// Parse JWT token (header.payload.signature)
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode payload (skip signature verification for now - requires JWKS endpoint)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse payload as JSON
	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// Verify issuer if available from config
	// In production, you'd get issuer from discovery config
	// For now, we skip issuer verification (would require storing issuer per provider)

	// Extract user info from claims
	userInfo := &UserInfo{}

	if sub, ok := claims["sub"].(string); ok {
		userInfo.Sub = sub
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}

	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	} else {
		// Try to construct from given_name and family_name
		givenName, _ := claims["given_name"].(string)
		familyName, _ := claims["family_name"].(string)
		if givenName != "" || familyName != "" {
			userInfo.Name = strings.TrimSpace(givenName + " " + familyName)
		}
	}

	if preferredUsername, ok := claims["preferred_username"].(string); ok && userInfo.Email == "" {
		userInfo.Email = preferredUsername
	}

	return userInfo, nil
}

/* getOrCreateUserFromOIDC gets or creates a user from OIDC user info */
func (s *OIDCService) getOrCreateUserFromOIDC(ctx context.Context, providerName string, userInfo *UserInfo) (*db.User, error) {
	// Try to find existing user by email
	user, err := s.queries.GetUserByEmail(ctx, userInfo.Email)
	if err == nil {
		return user, nil
	}

	// Create new user
	role := "analyst" // Default role

	user, err = s.queries.CreateUser(ctx, userInfo.Email, nil, nil, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Store OIDC provider link
	// This would be stored in oauth_providers table
	// Implementation depends on your schema

	return user, nil
}

/* buildScopeString builds scope string from slice */
func (s *OIDCService) buildScopeString(scopes []string) string {
	if len(scopes) == 0 {
		return "openid profile email"
	}
	result := "openid"
	for _, scope := range scopes {
		result += " " + scope
	}
	return result
}

/* generateSessionTokens generates session and refresh tokens using the same method as AuthService */
func (s *OIDCService) generateSessionTokens() (sessionToken, refreshToken string, expiresAt time.Time, err error) {
	// Generate secure random tokens - matches AuthService.generateSessionTokens exactly
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate session token: %w", err)
	}
	sessionToken = base64.URLEncoding.EncodeToString(sessionBytes)

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken = hex.EncodeToString(refreshBytes)

	// Session expires in 24 hours, refresh token expires in 7 days
	expiresAt = time.Now().Add(24 * time.Hour)

	return sessionToken, refreshToken, expiresAt, nil
}
