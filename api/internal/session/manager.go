package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Context keys for session and user ID */
type contextKey string

const (
	sessionKey contextKey = "session"
	userIDKey  contextKey = "userID"
	databaseKey contextKey = "database" // Store selected database
)

/* Session represents a user session */
type Session struct {
	ID            string
	UserID        string
	Database      string // Selected database (neuronip or neuronai-demo)
	CreatedAt     time.Time
	LastSeenAt    time.Time
	RevokedAt     *time.Time
	UserAgentHash string
	IPHash        string
}

/* RefreshToken represents a refresh token */
type RefreshToken struct {
	ID          string
	SessionID   string
	TokenHash   string
	RotatedFrom *string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
}

/* Manager manages sessions and refresh tokens */
type Manager struct {
	db             *pgxpool.Pool
	accessTTL      time.Duration
	refreshTTL     time.Duration
	cookieDomain   string
	cookieSecure   bool
	cookieSameSite http.SameSite
}

/* NewManager creates a new session manager */
func NewManager(db *pgxpool.Pool, accessTTL, refreshTTL time.Duration, cookieDomain string, cookieSecure bool, cookieSameSite string) *Manager {
	sameSite := http.SameSiteLaxMode
	switch cookieSameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "None":
		sameSite = http.SameSiteNoneMode
	}

	return &Manager{
		db:             db,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		cookieDomain:   cookieDomain,
		cookieSecure:   cookieSecure,
		cookieSameSite: sameSite,
	}
}

/* CreateSession creates a new session and refresh token */
func (m *Manager) CreateSession(ctx context.Context, userID, database, userAgent, ip string) (*Session, string, error) {
	// Validate inputs
	if err := ValidateUserID(userID); err != nil {
		return nil, "", fmt.Errorf("invalid user ID: %w", err)
	}
	if err := ValidateDatabaseName(database); err != nil {
		return nil, "", err
	}

	// Sanitize inputs
	userAgent = SanitizeUserAgent(userAgent)
	ip = SanitizeIP(ip)

	/* Hash user agent and IP for privacy */
	userAgentHash := hashString(userAgent)
	ipHash := hashString(ip)

	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:            sessionID,
		UserID:        userID,
		Database:      database,
		CreatedAt:     now,
		LastSeenAt:    now,
		UserAgentHash: userAgentHash,
		IPHash:        ipHash,
	}

	/* Create session in database */
	query := `
		INSERT INTO neuronip.sessions (id, user_id, database_name, created_at, last_seen_at, user_agent_hash, ip_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := m.db.Exec(ctx, query,
		session.ID, session.UserID, session.Database, session.CreatedAt, session.LastSeenAt,
		session.UserAgentHash, session.IPHash)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	/* Generate refresh token */
	refreshToken, tokenString, err := m.generateRefreshToken(ctx, sessionID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	/* Store refresh token */
	refreshQuery := `
		INSERT INTO neuronip.refresh_tokens (id, session_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = m.db.Exec(ctx, refreshQuery,
		refreshToken.ID, refreshToken.SessionID, refreshToken.TokenHash,
		refreshToken.ExpiresAt, refreshToken.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return session, tokenString, nil
}

/* ValidateSession validates a session ID */
func (m *Manager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	// Validate session ID format
	if err := ValidateSessionID(sessionID); err != nil {
		return nil, err
	}

	var session Session
	var revokedAt *time.Time

	query := `
		SELECT id, user_id, database_name, created_at, last_seen_at, revoked_at, user_agent_hash, ip_hash
		FROM neuronip.sessions
		WHERE id = $1 AND revoked_at IS NULL
	`
	err := m.db.QueryRow(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.Database, &session.CreatedAt, &session.LastSeenAt,
		&revokedAt, &session.UserAgentHash, &session.IPHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session not found or revoked")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	if revokedAt != nil {
		session.RevokedAt = revokedAt
	}

	/* Update last seen */
	updateQuery := `UPDATE neuronip.sessions SET last_seen_at = $1 WHERE id = $2`
	m.db.Exec(ctx, updateQuery, time.Now(), sessionID)

	return &session, nil
}

/* RefreshSession rotates refresh token and returns new access/refresh tokens */
func (m *Manager) RefreshSession(ctx context.Context, refreshTokenString string) (*Session, string, string, error) {
	// Validate refresh token format
	if err := ValidateRefreshToken(refreshTokenString); err != nil {
		return nil, "", "", err
	}

	/* Find refresh token by hash */
	refreshTokenHash := hashToken(refreshTokenString)

	var refreshToken RefreshToken
	var revokedAt *time.Time
	var rotatedFrom *string

	query := `
		SELECT id, session_id, token_hash, rotated_from, expires_at, revoked_at, created_at
		FROM neuronip.refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`
	err := m.db.QueryRow(ctx, query, refreshTokenHash).Scan(
		&refreshToken.ID, &refreshToken.SessionID, &refreshToken.TokenHash,
		&rotatedFrom, &refreshToken.ExpiresAt, &revokedAt, &refreshToken.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", "", fmt.Errorf("invalid or expired refresh token")
		}
		return nil, "", "", fmt.Errorf("failed to validate refresh token: %w", err)
	}

	if revokedAt != nil {
		/* Token reuse detected - revoke all tokens for this session */
		m.RevokeSession(ctx, refreshToken.SessionID)
		return nil, "", "", fmt.Errorf("refresh token reuse detected - session revoked")
	}

	/* Get session */
	session, err := m.ValidateSession(ctx, refreshToken.SessionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("session invalid: %w", err)
	}

	/* Revoke old refresh token */
	revokeQuery := `UPDATE neuronip.refresh_tokens SET revoked_at = $1 WHERE id = $2`
	m.db.Exec(ctx, revokeQuery, time.Now(), refreshToken.ID)

	/* Generate new refresh token (rotation) */
	newRefreshToken, newRefreshTokenString, err := m.generateRefreshToken(ctx, session.ID, &refreshToken.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	/* Store new refresh token */
	insertQuery := `
		INSERT INTO neuronip.refresh_tokens (id, session_id, token_hash, rotated_from, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = m.db.Exec(ctx, insertQuery,
		newRefreshToken.ID, newRefreshToken.SessionID, newRefreshToken.TokenHash,
		newRefreshToken.RotatedFrom, newRefreshToken.ExpiresAt, newRefreshToken.CreatedAt)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to store new refresh token: %w", err)
	}

	/* Generate new access token (session ID) */
	accessToken := session.ID

	return session, accessToken, newRefreshTokenString, nil
}

/* RevokeSession revokes a session and all its refresh tokens */
func (m *Manager) RevokeSession(ctx context.Context, sessionID string) error {
	now := time.Now()

	/* Revoke session */
	sessionQuery := `UPDATE neuronip.sessions SET revoked_at = $1 WHERE id = $2`
	_, err := m.db.Exec(ctx, sessionQuery, now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	/* Revoke all refresh tokens for this session */
	tokenQuery := `UPDATE neuronip.refresh_tokens SET revoked_at = $1 WHERE session_id = $2 AND revoked_at IS NULL`
	_, err = m.db.Exec(ctx, tokenQuery, now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

/* SetCookies sets access and refresh token cookies */
func (m *Manager) SetCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	/* Access token cookie (short-lived) */
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   int(m.accessTTL.Seconds()),
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
	})

	/* Refresh token cookie (long-lived) */
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   int(m.refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
	})
}

/* ClearCookies clears session cookies */
func (m *Manager) ClearCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Domain:   m.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   m.cookieSecure,
		SameSite: m.cookieSameSite,
	})
}

/* GetAccessTokenFromRequest extracts access token from cookie */
func (m *Manager) GetAccessTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

/* GetRefreshTokenFromRequest extracts refresh token from cookie */
func (m *Manager) GetRefreshTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

/* Helper functions */

func (m *Manager) generateRefreshToken(ctx context.Context, sessionID string, rotatedFrom *string) (*RefreshToken, string, error) {
	/* Generate random token */
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)
	tokenHash := hashToken(tokenString)

	now := time.Now()
	token := &RefreshToken{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		TokenHash:   tokenHash,
		RotatedFrom: rotatedFrom,
		ExpiresAt:   now.Add(m.refreshTTL),
		CreatedAt:   now,
	}

	return token, tokenString, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(h[:])
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return base64.URLEncoding.EncodeToString(h[:])
}

/* GetSessionFromContext gets session from context */
func GetSessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionKey).(*Session)
	return session, ok
}

/* GetUserIDFromContext gets user ID from context */
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

/* GetDatabaseFromContext gets database from context */
func GetDatabaseFromContext(ctx context.Context) (string, bool) {
	database, ok := ctx.Value(databaseKey).(string)
	return database, ok
}
