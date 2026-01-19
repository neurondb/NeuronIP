package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* OAuthConfig holds OAuth provider configuration */
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

/* OAuthService provides OAuth authentication services */
type OAuthService struct {
	queries  *db.Queries
	google   OAuthConfig
	github   OAuthConfig
	microsoft OAuthConfig
}

/* NewOAuthService creates a new OAuth service */
func NewOAuthService(queries *db.Queries, google, github, microsoft OAuthConfig) *OAuthService {
	return &OAuthService{
		queries:    queries,
		google:     google,
		github:     github,
		microsoft:  microsoft,
	}
}

/* InitiateOAuth initiates an OAuth flow and returns the authorization URL */
func (s *OAuthService) InitiateOAuth(ctx context.Context, provider string, state string) (string, error) {
	var config OAuthConfig
	var authURL string

	switch provider {
	case "google":
		config = s.google
		authURL = "https://accounts.google.com/o/oauth2/v2/auth"
	case "github":
		config = s.github
		authURL = "https://github.com/login/oauth/authorize"
	case "microsoft":
		config = s.microsoft
		authURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	default:
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	if config.ClientID == "" {
		return "", fmt.Errorf("OAuth provider %s not configured", provider)
	}

	// Build authorization URL
	u, err := url.Parse(authURL)
	if err != nil {
		return "", fmt.Errorf("invalid OAuth URL: %w", err)
	}

	q := u.Query()
	q.Set("client_id", config.ClientID)
	q.Set("redirect_uri", config.RedirectURL)
	q.Set("response_type", "code")
	if len(config.Scopes) > 0 {
		var scopeStr string
		if provider == "github" {
			scopeStr = config.Scopes[0]
			for i := 1; i < len(config.Scopes); i++ {
				scopeStr += " " + config.Scopes[i]
			}
		} else {
			scopeStr = config.Scopes[0]
			for i := 1; i < len(config.Scopes); i++ {
				scopeStr += " " + config.Scopes[i]
			}
		}
		q.Set("scope", scopeStr)
	}
	if state == "" {
		// Generate state token
		stateBytes := make([]byte, 32)
		if _, err := rand.Read(stateBytes); err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
		state = base64.URLEncoding.EncodeToString(stateBytes)
	}
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

/* HandleOAuthCallback processes OAuth callback and creates/updates user */
func (s *OAuthService) HandleOAuthCallback(ctx context.Context, provider, code, state string) (*db.User, *db.OAuthProvider, error) {
	var config OAuthConfig
	var tokenURL string

	switch provider {
	case "google":
		config = s.google
		tokenURL = "https://oauth2.googleapis.com/token"
	case "github":
		config = s.github
		tokenURL = "https://github.com/login/oauth/access_token"
	case "microsoft":
		config = s.microsoft
		tokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	default:
		return nil, nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Exchange code for tokens
	accessToken, refreshToken, providerUserID, expiresAt, err := s.exchangeCodeForTokens(ctx, provider, code, config, tokenURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get or create OAuth provider record
	oauthProvider, err := s.getOrCreateOAuthProvider(ctx, provider, providerUserID, accessToken, refreshToken, expiresAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get/create OAuth provider: %w", err)
	}

	// Get or create user
	user, err := s.queries.GetUserByID(ctx, oauthProvider.UserID)
	if err != nil {
		// User doesn't exist, create one
		email := providerUserID // Fallback, should get from provider API
		user, err = s.queries.CreateUser(ctx, email, nil, nil, "analyst")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
		// Update OAuth provider with user ID if it wasn't set
		if oauthProvider.UserID != user.ID {
			oauthProvider.UserID = user.ID
		}
	}

	return user, oauthProvider, nil
}

/* exchangeCodeForTokens exchanges authorization code for access token */
func (s *OAuthService) exchangeCodeForTokens(ctx context.Context, provider, code string, config OAuthConfig, tokenURL string) (accessToken, refreshToken, providerUserID string, expiresAt *time.Time, err error) {
	// Build token request
	values := url.Values{}
	values.Set("client_id", config.ClientID)
	values.Set("client_secret", config.ClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", config.RedirectURL)

	if provider == "microsoft" {
		values.Set("grant_type", "authorization_code")
	} else if provider == "github" {
		values.Set("grant_type", "authorization_code")
	} else {
		values.Set("grant_type", "authorization_code")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return "", "", "", nil, err
	}

	if provider == "github" {
		req.Header.Set("Accept", "application/json")
		req.URL.RawQuery = values.Encode()
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.URL.RawQuery = values.Encode()
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", nil, err
	}
	defer resp.Body.Close()

	var tokenData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return "", "", "", nil, err
	}

	accessToken, _ = tokenData["access_token"].(string)
	if accessToken == "" {
		return "", "", "", nil, fmt.Errorf("no access token in response")
	}

	if rt, ok := tokenData["refresh_token"].(string); ok {
		refreshToken = rt
	}

	// Get user info from provider
	providerUserID, err = s.getProviderUserID(ctx, provider, accessToken)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if exp, ok := tokenData["expires_in"].(float64); ok {
		expTime := time.Now().Add(time.Duration(exp) * time.Second)
		expiresAt = &expTime
	}

	return accessToken, refreshToken, providerUserID, expiresAt, nil
}

/* getProviderUserID gets user ID from OAuth provider */
func (s *OAuthService) getProviderUserID(ctx context.Context, provider, accessToken string) (string, error) {
	var userInfoURL string

	switch provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userInfoURL = "https://api.github.com/user"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	default:
		return "", fmt.Errorf("unsupported provider")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var userData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		return "", err
	}

	var userID string
	switch provider {
	case "google":
		userID, _ = userData["id"].(string)
	case "github":
		if id, ok := userData["id"].(float64); ok {
			userID = fmt.Sprintf("%.0f", id)
		}
	case "microsoft":
		userID, _ = userData["id"].(string)
	}

	if userID == "" {
		return "", fmt.Errorf("could not extract user ID from provider response")
	}

	return userID, nil
}

/* getOrCreateOAuthProvider gets or creates an OAuth provider record */
func (s *OAuthService) getOrCreateOAuthProvider(ctx context.Context, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) (*db.OAuthProvider, error) {
	// Try to find existing OAuth provider
	query := `SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at, updated_at 
	          FROM neuronip.oauth_providers WHERE provider = $1 AND provider_user_id = $2`
	
	var oauthProvider db.OAuthProvider
	err := s.queries.DB.QueryRow(ctx, query, provider, providerUserID).Scan(
		&oauthProvider.ID, &oauthProvider.UserID, &oauthProvider.Provider,
		&oauthProvider.ProviderUserID, &oauthProvider.AccessToken, &oauthProvider.RefreshToken,
		&oauthProvider.ExpiresAt, &oauthProvider.CreatedAt, &oauthProvider.UpdatedAt,
	)

	if err == nil {
		// Update tokens
		updateQuery := `UPDATE neuronip.oauth_providers SET access_token = $1, refresh_token = $2, expires_at = $3, updated_at = NOW() WHERE id = $4`
		_, err = s.queries.DB.Exec(ctx, updateQuery, accessToken, refreshToken, expiresAt, oauthProvider.ID)
		if err != nil {
			return nil, err
		}
		oauthProvider.AccessToken = &accessToken
		oauthProvider.RefreshToken = &refreshToken
		oauthProvider.ExpiresAt = expiresAt
		return &oauthProvider, nil
	}

	// Create new OAuth provider (will need user ID later)
	insertQuery := `INSERT INTO neuronip.oauth_providers (user_id, provider, provider_user_id, access_token, refresh_token, expires_at) 
	                VALUES (gen_random_uuid(), $1, $2, $3, $4, $5) 
	                RETURNING id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at, updated_at`
	
	tempUserID := uuid.New()
	err = s.queries.DB.QueryRow(ctx, insertQuery, provider, providerUserID, accessToken, refreshToken, expiresAt, tempUserID).Scan(
		&oauthProvider.ID, &oauthProvider.UserID, &oauthProvider.Provider,
		&oauthProvider.ProviderUserID, &oauthProvider.AccessToken, &oauthProvider.RefreshToken,
		&oauthProvider.ExpiresAt, &oauthProvider.CreatedAt, &oauthProvider.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &oauthProvider, nil
}
