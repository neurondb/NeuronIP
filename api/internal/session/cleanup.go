package session

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

/* CleanupService handles cleanup of expired sessions and tokens */
type CleanupService struct {
	db     *pgxpool.Pool
	ticker *time.Ticker
	done   chan bool
}

/* NewCleanupService creates a new cleanup service */
func NewCleanupService(db *pgxpool.Pool, interval time.Duration) *CleanupService {
	return &CleanupService{
		db:     db,
		ticker: time.NewTicker(interval),
		done:   make(chan bool),
	}
}

/* Start starts the cleanup service */
func (c *CleanupService) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.cleanupExpiredSessions(ctx)
				c.cleanupExpiredTokens(ctx)
			case <-c.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

/* Stop stops the cleanup service */
func (c *CleanupService) Stop() {
	c.ticker.Stop()
	close(c.done)
}

/* cleanupExpiredSessions removes expired sessions */
func (c *CleanupService) cleanupExpiredSessions(ctx context.Context) {
	// Revoke sessions that haven't been seen in 30 days
	query := `
		UPDATE neuronip.sessions 
		SET revoked_at = NOW() 
		WHERE revoked_at IS NULL 
		AND last_seen_at < NOW() - INTERVAL '30 days'
	`
	_, err := c.db.Exec(ctx, query)
	if err != nil {
		// Log error but don't fail
		return
	}
}

/* cleanupExpiredTokens removes expired refresh tokens */
func (c *CleanupService) cleanupExpiredTokens(ctx context.Context) {
	// Delete expired refresh tokens (older than 7 days past expiration)
	query := `
		DELETE FROM neuronip.refresh_tokens 
		WHERE expires_at < NOW() - INTERVAL '7 days'
	`
	_, err := c.db.Exec(ctx, query)
	if err != nil {
		// Log error but don't fail
		return
	}
}

/* CleanupExpiredSessions manually triggers cleanup of expired sessions */
func (m *Manager) CleanupExpiredSessions(ctx context.Context) (int, error) {
	query := `
		UPDATE neuronip.sessions 
		SET revoked_at = NOW() 
		WHERE revoked_at IS NULL 
		AND last_seen_at < NOW() - INTERVAL '30 days'
		RETURNING id
	`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}

/* CleanupExpiredTokens manually triggers cleanup of expired tokens */
func (m *Manager) CleanupExpiredTokens(ctx context.Context) (int, error) {
	query := `
		DELETE FROM neuronip.refresh_tokens 
		WHERE expires_at < NOW() - INTERVAL '7 days'
		RETURNING id
	`
	rows, err := m.db.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	return count, nil
}
