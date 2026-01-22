package shutdown

import (
	"context"
	"fmt"
	"sync"
	"time"
)

/* ShutdownHook is a function that performs cleanup during shutdown */
type ShutdownHook func(ctx context.Context) error

/* Manager manages graceful shutdown of services */
type Manager struct {
	hooks       []ShutdownHook
	hooksMu     sync.Mutex
	timeout     time.Duration
	shutdownCh  chan struct{}
	shutdownMu  sync.Mutex
	shuttingDown bool
}

/* NewManager creates a new shutdown manager */
func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		hooks:      make([]ShutdownHook, 0),
		timeout:    timeout,
		shutdownCh: make(chan struct{}),
	}
}

/* RegisterHook registers a shutdown hook */
func (m *Manager) RegisterHook(hook ShutdownHook) {
	m.hooksMu.Lock()
	defer m.hooksMu.Unlock()
	m.hooks = append(m.hooks, hook)
}

/* Shutdown performs graceful shutdown of all registered hooks */
func (m *Manager) Shutdown(ctx context.Context) error {
	m.shutdownMu.Lock()
	if m.shuttingDown {
		m.shutdownMu.Unlock()
		return fmt.Errorf("shutdown already in progress")
	}
	m.shuttingDown = true
	close(m.shutdownCh)
	m.shutdownMu.Unlock()

	// Create context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Execute all hooks
	m.hooksMu.Lock()
	hooks := make([]ShutdownHook, len(m.hooks))
	copy(hooks, m.hooks)
	m.hooksMu.Unlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(hooks))

	for _, hook := range hooks {
		wg.Add(1)
		go func(h ShutdownHook) {
			defer wg.Done()
			if err := h(shutdownCtx); err != nil {
				errors <- fmt.Errorf("hook error: %w", err)
			}
		}(hook)
	}

	// Wait for all hooks to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All hooks completed
	case <-shutdownCtx.Done():
		// Timeout reached
		return fmt.Errorf("shutdown timeout after %v", m.timeout)
	}

	// Collect errors
	close(errors)
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

/* IsShuttingDown returns true if shutdown is in progress */
func (m *Manager) IsShuttingDown() bool {
	m.shutdownMu.Lock()
	defer m.shutdownMu.Unlock()
	return m.shuttingDown
}

/* ShutdownChannel returns a channel that is closed when shutdown begins */
func (m *Manager) ShutdownChannel() <-chan struct{} {
	return m.shutdownCh
}

/* ConnectionDrainer manages draining of connections */
type ConnectionDrainer struct {
	drainTimeout time.Duration
	drained      bool
	mu           sync.Mutex
}

/* NewConnectionDrainer creates a new connection drainer */
func NewConnectionDrainer(drainTimeout time.Duration) *ConnectionDrainer {
	return &ConnectionDrainer{
		drainTimeout: drainTimeout,
		drained:      false,
	}
}

/* Drain waits for connections to drain */
func (cd *ConnectionDrainer) Drain(ctx context.Context, activeConnections func() int) error {
	cd.mu.Lock()
	if cd.drained {
		cd.mu.Unlock()
		return nil
	}
	cd.mu.Unlock()

	drainCtx, cancel := context.WithTimeout(ctx, cd.drainTimeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		active := activeConnections()
		if active == 0 {
			cd.mu.Lock()
			cd.drained = true
			cd.mu.Unlock()
			return nil
		}

		select {
		case <-drainCtx.Done():
			// Timeout reached, connections still active
			return fmt.Errorf("drain timeout: %d connections still active", active)
		case <-ticker.C:
			// Continue checking
		}
	}
}

/* IsDrained returns true if connections have been drained */
func (cd *ConnectionDrainer) IsDrained() bool {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	return cd.drained
}
