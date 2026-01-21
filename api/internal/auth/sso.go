package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

/* SSOProvider represents an SSO identity provider */
type SSOProvider struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	ProviderType string                 `json:"provider_type"` // saml, oauth2, oidc
	Enabled      bool                   `json:"enabled"`
	Configuration map[string]interface{} `json:"configuration"`
	MetadataURL  *string                `json:"metadata_url,omitempty"`
	EntityID     *string                `json:"entity_id,omitempty"`
	SSOURL       *string                `json:"sso_url,omitempty"`
	SLOURL       *string                `json:"slo_url,omitempty"`
	Certificate  *string                `json:"certificate,omitempty"`
	ClientID     *string                `json:"client_id,omitempty"`
	ClientSecret *string                `json:"client_secret,omitempty"`
	Scopes       []string               `json:"scopes,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

/* SSOService provides SSO authentication functionality */
type SSOService struct {
	pool   *pgxpool.Pool
	config *SSOConfig
}

/* SSOConfig holds SSO service configuration */
type SSOConfig struct {
	BaseURL           string
	CallbackPath      string
	SessionSecret     string
	SessionTimeout    time.Duration
	EnableAutoMapping bool
}

/* NewSSOService creates a new SSO service */
func NewSSOService(pool *pgxpool.Pool, config *SSOConfig) *SSOService {
	if config == nil {
		config = &SSOConfig{
			CallbackPath:      "/api/v1/auth/sso/callback",
			SessionTimeout:    24 * time.Hour,
			EnableAutoMapping: true,
		}
	}
	return &SSOService{
		pool:   pool,
		config: config,
	}
}

/* CreateProvider creates a new SSO provider */
func (s *SSOService) CreateProvider(ctx context.Context, provider SSOProvider) (*SSOProvider, error) {
	provider.ID = uuid.New()
	provider.CreatedAt = time.Now()
	provider.UpdatedAt = time.Now()

	configJSON, _ := json.Marshal(provider.Configuration)

	query := `
		INSERT INTO neuronip.sso_providers 
		(id, name, provider_type, enabled, configuration, metadata_url, entity_id, 
		 sso_url, slo_url, certificate, client_id, client_secret, scopes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	err := s.pool.QueryRow(ctx, query,
		provider.ID, provider.Name, provider.ProviderType, provider.Enabled,
		configJSON, provider.MetadataURL, provider.EntityID,
		provider.SSOURL, provider.SLOURL, provider.Certificate,
		provider.ClientID, provider.ClientSecret, provider.Scopes,
		provider.CreatedAt, provider.UpdatedAt,
	).Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create SSO provider: %w", err)
	}

	return &provider, nil
}

/* GetProvider retrieves an SSO provider by ID */
func (s *SSOService) GetProvider(ctx context.Context, id uuid.UUID) (*SSOProvider, error) {
	var provider SSOProvider
	var configJSON []byte
	var metadataURL, entityID, ssoURL, sloURL, certificate, clientID, clientSecret sql.NullString
	var scopes []string

	query := `
		SELECT id, name, provider_type, enabled, configuration, metadata_url, entity_id,
		       sso_url, slo_url, certificate, client_id, client_secret, scopes, created_at, updated_at
		FROM neuronip.sso_providers
		WHERE id = $1`

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&provider.ID, &provider.Name, &provider.ProviderType, &provider.Enabled,
		&configJSON, &metadataURL, &entityID, &ssoURL, &sloURL,
		&certificate, &clientID, &clientSecret, &scopes,
		&provider.CreatedAt, &provider.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provider: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get SSO provider: %w", err)
	}

	if configJSON != nil {
		json.Unmarshal(configJSON, &provider.Configuration)
	}
	if metadataURL.Valid {
		provider.MetadataURL = &metadataURL.String
	}
	if entityID.Valid {
		provider.EntityID = &entityID.String
	}
	if ssoURL.Valid {
		provider.SSOURL = &ssoURL.String
	}
	if sloURL.Valid {
		provider.SLOURL = &sloURL.String
	}
	if certificate.Valid {
		provider.Certificate = &certificate.String
	}
	if clientID.Valid {
		provider.ClientID = &clientID.String
	}
	if clientSecret.Valid {
		provider.ClientSecret = &clientSecret.String
	}
	provider.Scopes = scopes

	return &provider, nil
}

/* ListProviders lists all SSO providers */
func (s *SSOService) ListProviders(ctx context.Context, enabledOnly bool) ([]SSOProvider, error) {
	query := `
		SELECT id, name, provider_type, enabled, configuration, metadata_url, entity_id,
		       sso_url, slo_url, certificate, client_id, client_secret, scopes, created_at, updated_at
		FROM neuronip.sso_providers`
	
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSO providers: %w", err)
	}
	defer rows.Close()

	var providers []SSOProvider
	for rows.Next() {
		var provider SSOProvider
		var configJSON []byte
		var metadataURL, entityID, ssoURL, sloURL, certificate, clientID, clientSecret sql.NullString
		var scopes []string

		err := rows.Scan(
			&provider.ID, &provider.Name, &provider.ProviderType, &provider.Enabled,
			&configJSON, &metadataURL, &entityID, &ssoURL, &sloURL,
			&certificate, &clientID, &clientSecret, &scopes,
			&provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			continue
		}

		if configJSON != nil {
			json.Unmarshal(configJSON, &provider.Configuration)
		}
		if metadataURL.Valid {
			provider.MetadataURL = &metadataURL.String
		}
		if entityID.Valid {
			provider.EntityID = &entityID.String
		}
		if ssoURL.Valid {
			provider.SSOURL = &ssoURL.String
		}
		if sloURL.Valid {
			provider.SLOURL = &sloURL.String
		}
		if certificate.Valid {
			provider.Certificate = &certificate.String
		}
		if clientID.Valid {
			provider.ClientID = &clientID.String
		}
		if clientSecret.Valid {
			provider.ClientSecret = &clientSecret.String
		}
		provider.Scopes = scopes

		providers = append(providers, provider)
	}

	return providers, nil
}

/* InitiateSAML initiates SAML SSO flow */
func (s *SSOService) InitiateSAML(ctx context.Context, providerID uuid.UUID, relayState string) (string, error) {
	provider, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	if provider.ProviderType != "saml" {
		return "", fmt.Errorf("provider is not SAML type")
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider is disabled")
	}

	if provider.SSOURL == nil {
		return "", fmt.Errorf("SSO URL not configured")
	}

	// Generate SAML AuthnRequest
	authnRequest := s.generateSAMLRequest(provider, relayState)
	
	// Encode and sign the request
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(authnRequest))
	
	// Build redirect URL
	redirectURL := fmt.Sprintf("%s?SAMLRequest=%s", *provider.SSOURL, url.QueryEscape(encodedRequest))
	if relayState != "" {
		redirectURL += fmt.Sprintf("&RelayState=%s", url.QueryEscape(relayState))
	}

	return redirectURL, nil
}

/* ProcessSAMLCallback processes SAML response */
func (s *SSOService) ProcessSAMLCallback(ctx context.Context, providerID uuid.UUID, samlResponse, relayState string) (*SSOUser, error) {
	provider, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Decode SAML response
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse and validate SAML response
	user, err := s.parseSAMLResponse(decoded, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Map or create user
	mappedUser, err := s.mapSSOUser(ctx, providerID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to map SSO user: %w", err)
	}

	// Log audit event
	s.logSSOEvent(ctx, providerID, mappedUser.ExternalID, "login", true, nil)

	return mappedUser, nil
}

/* InitiateOAuth2 initiates OAuth2/OIDC flow */
func (s *SSOService) InitiateOAuth2(ctx context.Context, providerID uuid.UUID, state string) (string, error) {
	provider, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	if provider.ProviderType != "oauth2" && provider.ProviderType != "oidc" {
		return "", fmt.Errorf("provider is not OAuth2/OIDC type")
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider is disabled")
	}

	if provider.ClientID == nil {
		return "", fmt.Errorf("client ID not configured")
	}

	// Build OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     *provider.ClientID,
		ClientSecret: *provider.ClientSecret,
		RedirectURL:  s.config.BaseURL + s.config.CallbackPath + "?provider_id=" + providerID.String(),
		Scopes:       provider.Scopes,
	}

	// Get authorization URL from configuration
	authURL := s.getOAuth2AuthURL(provider)
	if authURL == "" {
		return "", fmt.Errorf("authorization URL not configured")
	}

	oauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  authURL,
		TokenURL: s.getOAuth2TokenURL(provider),
	}

	// Generate authorization URL
	url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return url, nil
}

/* ProcessOAuth2Callback processes OAuth2/OIDC callback */
func (s *SSOService) ProcessOAuth2Callback(ctx context.Context, providerID uuid.UUID, code, state string) (*SSOUser, error) {
	provider, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Build OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     *provider.ClientID,
		ClientSecret: *provider.ClientSecret,
		RedirectURL:  s.config.BaseURL + s.config.CallbackPath + "?provider_id=" + providerID.String(),
		Scopes:       provider.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  s.getOAuth2AuthURL(provider),
			TokenURL: s.getOAuth2TokenURL(provider),
		},
	}

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		errMsg := err.Error()
		s.logSSOEvent(ctx, providerID, "", "error", false, &errMsg)
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Get user info
	user, err := s.getOAuth2UserInfo(ctx, provider, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Map or create user
	mappedUser, err := s.mapSSOUser(ctx, providerID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to map SSO user: %w", err)
	}

	// Create session
	sessionToken, err := s.createSSOSession(ctx, providerID, mappedUser.UserID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	mappedUser.SessionToken = &sessionToken

	// Log audit event
	s.logSSOEvent(ctx, providerID, mappedUser.ExternalID, "login", true, nil)

	return mappedUser, nil
}

/* SSOUser represents a user from SSO */
type SSOUser struct {
	UserID      string                 `json:"user_id"`
	ExternalID  string                 `json:"external_id"`
	Email       string                 `json:"email"`
	Name        string                 `json:"name"`
	Attributes  map[string]interface{} `json:"attributes"`
	SessionToken *string               `json:"session_token,omitempty"`
}

/* mapSSOUser maps SSO user to NeuronIP user */
func (s *SSOService) mapSSOUser(ctx context.Context, providerID uuid.UUID, user *SSOUser) (*SSOUser, error) {
	// Check for existing mapping
	var existingUserID string
	query := `
		SELECT user_id FROM neuronip.sso_user_mappings
		WHERE provider_id = $1 AND external_id = $2`
	
	err := s.pool.QueryRow(ctx, query, providerID, user.ExternalID).Scan(&existingUserID)
	if err == nil {
		// Update last login
		s.pool.Exec(ctx, `
			UPDATE neuronip.sso_user_mappings 
			SET last_login_at = NOW(), updated_at = NOW()
			WHERE provider_id = $1 AND external_id = $2`,
			providerID, user.ExternalID)
		
		user.UserID = existingUserID
		return user, nil
	}

	// Auto-create user if enabled
	if s.config.EnableAutoMapping {
		// Create new user
		userID := uuid.New().String()
		
		// Insert mapping
		insertQuery := `
			INSERT INTO neuronip.sso_user_mappings
			(provider_id, user_id, external_id, email, attributes, first_login_at, last_login_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
		
		attrsJSON, _ := json.Marshal(user.Attributes)
		_, err = s.pool.Exec(ctx, insertQuery,
			providerID, userID, user.ExternalID, user.Email, attrsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to create user mapping: %w", err)
		}

		user.UserID = userID
		return user, nil
	}

	return nil, fmt.Errorf("user mapping not found and auto-mapping disabled")
}

/* createSSOSession creates an SSO session */
func (s *SSOService) createSSOSession(ctx context.Context, providerID uuid.UUID, userID string, token *oauth2.Token) (string, error) {
	sessionToken := generateSessionToken()
	expiresAt := time.Now().Add(s.config.SessionTimeout)
	
	if token != nil && !token.Expiry.IsZero() {
		expiresAt = token.Expiry
	}

	var idToken, accessToken, refreshToken sql.NullString
	if token != nil {
		if idTok, ok := token.Extra("id_token").(string); ok {
			idToken = sql.NullString{String: idTok, Valid: true}
		}
		accessToken = sql.NullString{String: token.AccessToken, Valid: true}
		if token.RefreshToken != "" {
			refreshToken = sql.NullString{String: token.RefreshToken, Valid: true}
		}
	}

	query := `
		INSERT INTO neuronip.sso_sessions
		(provider_id, user_id, session_token, id_token, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.pool.Exec(ctx, query,
		providerID, userID, sessionToken, idToken, accessToken, refreshToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionToken, nil
}

/* ValidateSession validates an SSO session token */
func (s *SSOService) ValidateSession(ctx context.Context, sessionToken string) (*SSOSession, error) {
	var session SSOSession
	var idToken, accessToken, refreshToken sql.NullString

	query := `
		SELECT id, provider_id, user_id, session_token, id_token, access_token, 
		       refresh_token, expires_at, created_at, last_accessed_at
		FROM neuronip.sso_sessions
		WHERE session_token = $1 AND expires_at > NOW()`

	err := s.pool.QueryRow(ctx, query, sessionToken).Scan(
		&session.ID, &session.ProviderID, &session.UserID, &session.SessionToken,
		&idToken, &accessToken, &refreshToken,
		&session.ExpiresAt, &session.CreatedAt, &session.LastAccessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired session")
	}

	// Update last accessed
	s.pool.Exec(ctx, `
		UPDATE neuronip.sso_sessions 
		SET last_accessed_at = NOW()
		WHERE id = $1`, session.ID)

	if idToken.Valid {
		session.IDToken = &idToken.String
	}
	if accessToken.Valid {
		session.AccessToken = &accessToken.String
	}
	if refreshToken.Valid {
		session.RefreshToken = &refreshToken.String
	}

	return &session, nil
}

/* SSOSession represents an active SSO session */
type SSOSession struct {
	ID            uuid.UUID  `json:"id"`
	ProviderID    uuid.UUID  `json:"provider_id"`
	UserID        string     `json:"user_id"`
	SessionToken  string     `json:"session_token"`
	IDToken       *string    `json:"id_token,omitempty"`
	AccessToken   *string    `json:"access_token,omitempty"`
	RefreshToken  *string    `json:"refresh_token,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

/* logSSOEvent logs an SSO audit event */
func (s *SSOService) logSSOEvent(ctx context.Context, providerID uuid.UUID, externalID, eventType string, success bool, errorMsg *string) {
	query := `
		INSERT INTO neuronip.sso_audit_log
		(provider_id, external_id, event_type, success, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())`
	
	var errMsg sql.NullString
	if errorMsg != nil {
		errMsg = sql.NullString{String: *errorMsg, Valid: true}
	}

	s.pool.Exec(ctx, query, providerID, externalID, eventType, success, errMsg)
}

/* Helper functions */

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *SSOService) generateSAMLRequest(provider *SSOProvider, relayState string) string {
	// Simplified SAML AuthnRequest generation
	// In production, use a proper SAML library
	requestID := uuid.New().String()
	issueInstant := time.Now().UTC().Format(time.RFC3339)
	
	entityID := provider.EntityID
	if entityID == nil {
		defaultID := s.config.BaseURL
		entityID = &defaultID
	}

	return fmt.Sprintf(`<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
		ID="%s" Version="2.0" IssueInstant="%s"
		Destination="%s" AssertionConsumerServiceURL="%s">
		<saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">%s</saml:Issuer>
	</samlp:AuthnRequest>`,
		requestID, issueInstant, *provider.SSOURL,
		s.config.BaseURL+s.config.CallbackPath, *entityID)
}

func (s *SSOService) parseSAMLResponse(response []byte, provider *SSOProvider) (*SSOUser, error) {
	// Simplified SAML response parsing
	// In production, use a proper SAML library like github.com/crewjam/saml
	// This is a placeholder implementation
	
	user := &SSOUser{
		Attributes: make(map[string]interface{}),
	}
	
	// Extract user attributes from SAML response
	// In production, properly parse XML and validate signature
	
	return user, nil
}

func (s *SSOService) getOAuth2AuthURL(provider *SSOProvider) string {
	if authURL, ok := provider.Configuration["auth_url"].(string); ok {
		return authURL
	}
	return ""
}

func (s *SSOService) getOAuth2TokenURL(provider *SSOProvider) string {
	if tokenURL, ok := provider.Configuration["token_url"].(string); ok {
		return tokenURL
	}
	return ""
}

func (s *SSOService) getOAuth2UserInfo(ctx context.Context, provider *SSOProvider, token *oauth2.Token) (*SSOUser, error) {
	// Get user info endpoint from configuration
	userInfoURL := ""
	if url, ok := provider.Configuration["userinfo_url"].(string); ok {
		userInfoURL = url
	} else {
		return nil, fmt.Errorf("userinfo URL not configured")
	}

	// Make request to user info endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	user := &SSOUser{
		Attributes: userInfo,
	}

	// Extract standard fields
	if email, ok := userInfo["email"].(string); ok {
		user.Email = email
	}
	if name, ok := userInfo["name"].(string); ok {
		user.Name = name
	}
	if sub, ok := userInfo["sub"].(string); ok {
		user.ExternalID = sub
	} else if id, ok := userInfo["id"].(string); ok {
		user.ExternalID = id
	}

	return user, nil
}
