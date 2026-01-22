package session

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* RateLimitService handles rate limiting for authentication attempts */
type RateLimitService struct {
	db *pgxpool.Pool
}

/* NewRateLimitService creates a new rate limit service */
func NewRateLimitService(db *pgxpool.Pool) *RateLimitService {
	return &RateLimitService{db: db}
}

/* CheckRateLimit checks if a user/IP has exceeded rate limits */
func (r *RateLimitService) CheckRateLimit(ctx context.Context, identifier string, maxAttempts int, window time.Duration) (bool, error) {
	// Check recent failed attempts
	query := `
		SELECT COUNT(*) 
		FROM neuronip.failed_login_attempts 
		WHERE identifier = $1 AND attempted_at > NOW() - $2::interval
	`
	var count int
	err := r.db.QueryRow(ctx, query, identifier, window).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return count < maxAttempts, nil
}

/* RecordFailedAttempt records a failed login attempt */
func (r *RateLimitService) RecordFailedAttempt(ctx context.Context, identifier, ipAddress string) error {
	query := `
		INSERT INTO neuronip.failed_login_attempts (identifier, ip_address, attempted_at)
		VALUES ($1, $2, NOW())
	`
	_, err := r.db.Exec(ctx, query, identifier, ipAddress)
	return err
}

/* ClearFailedAttempts clears failed attempts for an identifier */
func (r *RateLimitService) ClearFailedAttempts(ctx context.Context, identifier string) error {
	query := `DELETE FROM neuronip.failed_login_attempts WHERE identifier = $1`
	_, err := r.db.Exec(ctx, query, identifier)
	return err
}
