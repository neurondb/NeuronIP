package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

/* SavepointManager manages savepoints for nested transactions */
type SavepointManager struct {
	tx      pgx.Tx
	savepoints []string
	mu      sync.Mutex
}

/* NewSavepointManager creates a new savepoint manager */
func NewSavepointManager(tx pgx.Tx) *SavepointManager {
	return &SavepointManager{
		tx:         tx,
		savepoints: make([]string, 0),
	}
}

/* CreateSavepoint creates a new savepoint */
func (sm *SavepointManager) CreateSavepoint(ctx context.Context, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	savepointName := fmt.Sprintf("sp_%s", name)
	_, err := sm.tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", savepointName))
	if err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", savepointName, err)
	}

	sm.savepoints = append(sm.savepoints, savepointName)
	return nil
}

/* RollbackToSavepoint rolls back to a specific savepoint */
func (sm *SavepointManager) RollbackToSavepoint(ctx context.Context, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	savepointName := fmt.Sprintf("sp_%s", name)
	_, err := sm.tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName))
	if err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", savepointName, err)
	}

	// Remove savepoints after this one
	index := -1
	for i, sp := range sm.savepoints {
		if sp == savepointName {
			index = i
			break
		}
	}
	if index >= 0 {
		sm.savepoints = sm.savepoints[:index+1]
	}

	return nil
}

/* ReleaseSavepoint releases a savepoint */
func (sm *SavepointManager) ReleaseSavepoint(ctx context.Context, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	savepointName := fmt.Sprintf("sp_%s", name)
	_, err := sm.tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName))
	if err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", savepointName, err)
	}

	// Remove savepoint from list
	index := -1
	for i, sp := range sm.savepoints {
		if sp == savepointName {
			index = i
			break
		}
	}
	if index >= 0 {
		sm.savepoints = append(sm.savepoints[:index], sm.savepoints[index+1:]...)
	}

	return nil
}

/* NestedTransactionManager manages nested transactions using savepoints */
type NestedTransactionManager struct {
	tx      pgx.Tx
	manager *SavepointManager
	level   int
	mu      sync.Mutex
}

/* NewNestedTransactionManager creates a new nested transaction manager */
func NewNestedTransactionManager(tx pgx.Tx) *NestedTransactionManager {
	return &NestedTransactionManager{
		tx:      tx,
		manager: NewSavepointManager(tx),
		level:   0,
	}
}

/* Begin starts a nested transaction */
func (ntm *NestedTransactionManager) Begin(ctx context.Context) (string, error) {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()

	ntm.level++
	savepointName := fmt.Sprintf("nested_%d", ntm.level)
	err := ntm.manager.CreateSavepoint(ctx, savepointName)
	if err != nil {
		ntm.level--
		return "", err
	}

	return savepointName, nil
}

/* Rollback rolls back the current nested transaction */
func (ntm *NestedTransactionManager) Rollback(ctx context.Context, savepointName string) error {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()

	if ntm.level <= 0 {
		return fmt.Errorf("no nested transaction to rollback")
	}

	err := ntm.manager.RollbackToSavepoint(ctx, savepointName)
	if err != nil {
		return err
	}

	ntm.level--
	return nil
}

/* Commit commits the current nested transaction (releases savepoint) */
func (ntm *NestedTransactionManager) Commit(ctx context.Context, savepointName string) error {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()

	if ntm.level <= 0 {
		return fmt.Errorf("no nested transaction to commit")
	}

	err := ntm.manager.ReleaseSavepoint(ctx, savepointName)
	if err != nil {
		return err
	}

	ntm.level--
	return nil
}

/* GetLevel returns the current nesting level */
func (ntm *NestedTransactionManager) GetLevel() int {
	ntm.mu.Lock()
	defer ntm.mu.Unlock()
	return ntm.level
}

/* WithNestedTransaction executes a function within a nested transaction */
func WithNestedTransaction(ctx context.Context, tx pgx.Tx, fn func(ctx context.Context, ntm *NestedTransactionManager) error) error {
	ntm := NewNestedTransactionManager(tx)
	savepointName, err := ntm.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin nested transaction: %w", err)
	}

	err = fn(ctx, ntm)
	if err != nil {
		rollbackErr := ntm.Rollback(ctx, savepointName)
		if rollbackErr != nil {
			return fmt.Errorf("nested transaction error: %w, rollback error: %v", err, rollbackErr)
		}
		return err
	}

	err = ntm.Commit(ctx, savepointName)
	if err != nil {
		return fmt.Errorf("failed to commit nested transaction: %w", err)
	}

	return nil
}
