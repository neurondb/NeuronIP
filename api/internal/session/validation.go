package session

import (
	"fmt"
	"strings"
	"time"
)

/* ValidateSessionID validates a session ID format */
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}
	if len(sessionID) < 32 {
		return fmt.Errorf("session ID too short")
	}
	if len(sessionID) > 128 {
		return fmt.Errorf("session ID too long")
	}
	return nil
}

/* ValidateRefreshToken validates a refresh token format */
func ValidateRefreshToken(token string) error {
	if token == "" {
		return fmt.Errorf("refresh token cannot be empty")
	}
	if len(token) < 32 {
		return fmt.Errorf("refresh token too short")
	}
	return nil
}

/* ValidateDatabaseName validates database name */
func ValidateDatabaseName(database string) error {
	validDatabases := []string{"neuronip", "neuronai-demo"}
	for _, valid := range validDatabases {
		if database == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid database name: %s (must be one of: %s)", database, strings.Join(validDatabases, ", "))
}

/* ValidateUserID validates user ID format */
func ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if len(userID) < 32 {
		return fmt.Errorf("user ID too short")
	}
	return nil
}

/* IsSessionExpired checks if a session is expired based on last seen time */
func IsSessionExpired(lastSeen time.Time, maxAge time.Duration) bool {
	return time.Since(lastSeen) > maxAge
}

/* SanitizeUserAgent sanitizes user agent string */
func SanitizeUserAgent(userAgent string) string {
	if len(userAgent) > 500 {
		return userAgent[:500]
	}
	return userAgent
}

/* SanitizeIP sanitizes IP address */
func SanitizeIP(ip string) string {
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	// Limit length
	if len(ip) > 45 { // IPv6 max length
		return ip[:45]
	}
	return ip
}
