package session

import "fmt"

/* Session errors */
var (
	ErrSessionNotFound      = fmt.Errorf("session not found")
	ErrSessionRevoked      = fmt.Errorf("session revoked")
	ErrSessionExpired      = fmt.Errorf("session expired")
	ErrInvalidSessionID    = fmt.Errorf("invalid session ID")
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
	ErrTokenReuse          = fmt.Errorf("refresh token reuse detected")
	ErrDatabaseInvalid     = fmt.Errorf("invalid database name")
)
