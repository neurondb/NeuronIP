package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* LockManager manages distributed locks */
type LockManager struct {
	pool    *pgxpool.Pool
	nodeID  string
	cleanup *time.Ticker
}

/* NewLockManager creates a new lock manager */
func NewLockManager(pool *pgxpool.Pool, nodeID string) *LockManager {
	lm := &LockManager{
		pool:   pool,
		nodeID: nodeID,
	}

	// Start cleanup goroutine
	lm.cleanup = time.NewTicker(1 * time.Minute)
	go lm.cleanupExpiredLocks(context.Background())

	return lm
}

/* AcquireLock acquires a distributed lock */
func (lm *LockManager) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (*Lock, error) {
	lockID := uuid.New()
	expiresAt := time.Now().Add(ttl)

	// Try to acquire lock
	query := `
		INSERT INTO neuronip.distributed_locks (id, lock_key, lock_holder, expires_at, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (lock_key) DO UPDATE SET
			lock_holder = CASE 
				WHEN expires_at < NOW() THEN EXCLUDED.lock_holder
				ELSE distributed_locks.lock_holder
			END,
			expires_at = CASE 
				WHEN expires_at < NOW() THEN EXCLUDED.expires_at
				ELSE distributed_locks.expires_at
			END,
			metadata = EXCLUDED.metadata
		RETURNING id, lock_key, lock_holder, expires_at, metadata, created_at`

	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"node_id": lm.nodeID,
		"acquired_at": time.Now(),
	})

	var lock Lock
	err := lm.pool.QueryRow(ctx, query, lockID, lockKey, lm.nodeID, expiresAt, metadataJSON).Scan(
		&lock.ID, &lock.LockKey, &lock.LockHolder, &lock.ExpiresAt, &lock.Metadata, &lock.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Check if we actually got the lock
	if lock.LockHolder != lm.nodeID {
		return nil, fmt.Errorf("lock already held by: %s", lock.LockHolder)
	}

	return &lock, nil
}

/* ReleaseLock releases a distributed lock */
func (lm *LockManager) ReleaseLock(ctx context.Context, lockKey string) error {
	query := `
		DELETE FROM neuronip.distributed_locks
		WHERE lock_key = $1 AND lock_holder = $2`

	result, err := lm.pool.Exec(ctx, query, lockKey, lm.nodeID)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("lock not found or not held by this node")
	}

	return nil
}

/* RenewLock renews a lock's expiration time */
func (lm *LockManager) RenewLock(ctx context.Context, lockKey string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	query := `
		UPDATE neuronip.distributed_locks
		SET expires_at = $1
		WHERE lock_key = $2 AND lock_holder = $3`

	result, err := lm.pool.Exec(ctx, query, expiresAt, lockKey, lm.nodeID)
	if err != nil {
		return fmt.Errorf("failed to renew lock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("lock not found or not held by this node")
	}

	return nil
}

/* IsLocked checks if a lock is currently held */
func (lm *LockManager) IsLocked(ctx context.Context, lockKey string) (bool, *Lock, error) {
	query := `
		SELECT id, lock_key, lock_holder, expires_at, metadata, created_at
		FROM neuronip.distributed_locks
		WHERE lock_key = $1 AND expires_at > NOW()`

	var lock Lock
	err := lm.pool.QueryRow(ctx, query, lockKey).Scan(
		&lock.ID, &lock.LockKey, &lock.LockHolder, &lock.ExpiresAt, &lock.Metadata, &lock.CreatedAt,
	)
	if err != nil {
		return false, nil, nil // Lock not held
	}

	return true, &lock, nil
}

/* cleanupExpiredLocks periodically cleans up expired locks */
func (lm *LockManager) cleanupExpiredLocks(ctx context.Context) {
	for range lm.cleanup.C {
		query := `SELECT neuronip.cleanup_expired_locks()`
		lm.pool.Exec(ctx, query)
	}
}

/* Close stops the lock manager */
func (lm *LockManager) Close() {
	if lm.cleanup != nil {
		lm.cleanup.Stop()
	}
}

/* Lock represents a distributed lock */
type Lock struct {
	ID         uuid.UUID              `json:"id"`
	LockKey    string                 `json:"lock_key"`
	LockHolder string                 `json:"lock_holder"`
	ExpiresAt  time.Time              `json:"expires_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
