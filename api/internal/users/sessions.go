package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* SessionService provides session management functionality */
type SessionService struct {
	queries *db.Queries
}

/* NewSessionService creates a new session service */
func NewSessionService(queries *db.Queries) *SessionService {
	return &SessionService{
		queries: queries,
	}
}

/* GetUserSessions lists all active sessions for a user */
func (s *SessionService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]db.UserSession, error) {
	return s.queries.ListUserSessions(ctx, userID)
}

/* RevokeSession revokes a specific session */
func (s *SessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.queries.DeleteUserSession(ctx, sessionID)
}

/* RevokeAllSessions revokes all sessions for a user */
func (s *SessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.queries.DeleteUserSessionsByUserID(ctx, userID)
}
