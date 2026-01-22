package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* SessionService provides enhanced session management */
type SessionService struct {
	queries *db.Queries
}

/* NewSessionService creates a new session service */
func NewSessionService(queries *db.Queries) *SessionService {
	return &SessionService{
		queries: queries,
	}
}

/* CheckConcurrentSessions checks and enforces concurrent session limits */
func (s *SessionService) CheckConcurrentSessions(ctx context.Context, userID uuid.UUID) error {
	// Get session limit for user (default is 5)
	limit := 5
	query := `SELECT max_concurrent_sessions FROM neuronip.session_limits WHERE user_id = $1`
	err := s.queries.DB.QueryRow(ctx, query, userID).Scan(&limit)
	if err != nil {
		// Use default if no limit set
		limit = 5
	}

	// Count active sessions
	countQuery := `SELECT COUNT(*) FROM neuronip.user_sessions WHERE user_id = $1 AND expires_at > NOW()`
	var activeCount int
	err = s.queries.DB.QueryRow(ctx, countQuery, userID).Scan(&activeCount)
	if err != nil {
		return fmt.Errorf("failed to count sessions: %w", err)
	}

	// If at limit, revoke oldest session
	if activeCount >= limit {
		revokeQuery := `
			DELETE FROM neuronip.user_sessions 
			WHERE id = (
				SELECT id FROM neuronip.user_sessions 
				WHERE user_id = $1 AND expires_at > NOW()
				ORDER BY created_at ASC
				LIMIT 1
			)
		`
		_, err = s.queries.DB.Exec(ctx, revokeQuery, userID)
		if err != nil {
			return fmt.Errorf("failed to revoke old session: %w", err)
		}
	}

	return nil
}

/* RevokeSession revokes a specific session */
func (s *SessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.queries.DeleteUserSession(ctx, sessionID)
}

/* RevokeAllUserSessions revokes all sessions for a user */
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM neuronip.user_sessions WHERE user_id = $1`
	_, err := s.queries.DB.Exec(ctx, query, userID)
	return err
}

/* RevokeExpiredSessions revokes all expired sessions */
func (s *SessionService) RevokeExpiredSessions(ctx context.Context) (int, error) {
	query := `DELETE FROM neuronip.user_sessions WHERE expires_at < NOW()`
	result, err := s.queries.DB.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

/* GetUserSessions gets all active sessions for a user */
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*db.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, refresh_token, ip_address, user_agent, 
		       expires_at, created_at
		FROM neuronip.user_sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := s.queries.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*db.UserSession
	for rows.Next() {
		var session db.UserSession
		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
			&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			continue
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

/* SetSessionLimit sets the concurrent session limit for a user */
func (s *SessionService) SetSessionLimit(ctx context.Context, userID uuid.UUID, maxSessions int, timeoutMinutes int) error {
	query := `
		INSERT INTO neuronip.session_limits (user_id, max_concurrent_sessions, session_timeout_minutes)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			max_concurrent_sessions = $2,
			session_timeout_minutes = $3,
			updated_at = NOW()
	`
	_, err := s.queries.DB.Exec(ctx, query, userID, maxSessions, timeoutMinutes)
	return err
}

/* GetSessionLimit gets the session limit for a user */
func (s *SessionService) GetSessionLimit(ctx context.Context, userID uuid.UUID) (maxSessions int, timeoutMinutes int, err error) {
	query := `SELECT max_concurrent_sessions, session_timeout_minutes FROM neuronip.session_limits WHERE user_id = $1`
	err = s.queries.DB.QueryRow(ctx, query, userID).Scan(&maxSessions, &timeoutMinutes)
	if err != nil {
		// Return defaults if not set
		return 5, 1440, nil
	}
	return maxSessions, timeoutMinutes, nil
}

/* RefreshSession refreshes a session token */
func (s *SessionService) RefreshSession(ctx context.Context, refreshToken string) (*db.UserSession, error) {
	// Find session by refresh token
	query := `
		SELECT id, user_id, session_token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM neuronip.user_sessions
		WHERE refresh_token = $1 AND expires_at > NOW()
	`
	var session db.UserSession
	err := s.queries.DB.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Generate new session tokens
	newSessionToken, newRefreshToken, newExpiresAt, err := generateSessionTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Update session
	updateQuery := `
		UPDATE neuronip.user_sessions
		SET session_token = $1, refresh_token = $2, expires_at = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err = s.queries.DB.Exec(ctx, updateQuery, newSessionToken, newRefreshToken, newExpiresAt, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	session.SessionToken = newSessionToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = newExpiresAt

	return &session, nil
}

/* generateSessionTokens generates session and refresh tokens (helper function) */
func generateSessionTokens() (string, string, time.Time, error) {
	// Generate cryptographically secure random tokens
	sessionToken, err := generateSecureToken(32)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate session token: %w", err)
	}
	
	refreshToken, err := generateSecureToken(32)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Set expiration to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)
	
	return sessionToken, refreshToken, expiresAt, nil
}

/* generateSecureToken generates a cryptographically secure random token */
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(bytes), nil
}
