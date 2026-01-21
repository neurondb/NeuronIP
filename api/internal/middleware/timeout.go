package middleware

import (
	"context"
	"net/http"
	"time"
)

/* TimeoutConfig holds timeout configuration for different route types */
type TimeoutConfig struct {
	Default   time.Duration
	Query     time.Duration
	Workflow  time.Duration
	Ingestion time.Duration
}

/* DefaultTimeoutConfig returns default timeout configuration */
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Default:   30 * time.Second,
		Query:     5 * time.Minute,
		Workflow:  1 * time.Hour,
		Ingestion: 10 * time.Minute,
	}
}

/* Timeout creates a middleware that enforces request timeouts */
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a response writer that detects context cancellation
			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Request completed normally
			case <-ctx.Done():
				// Timeout occurred
				if ctx.Err() == context.DeadlineExceeded {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusRequestTimeout)
					w.Write([]byte(`{"error":{"code":"TIMEOUT","message":"Request timeout exceeded"}}`))
				}
			}
		})
	}
}

/* TimeoutByRoute creates a middleware that applies different timeouts based on route */
func TimeoutByRoute(config TimeoutConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var timeout time.Duration

			// Determine timeout based on route path
			path := r.URL.Path
			switch {
			case contains(path, "/warehouse/query") || contains(path, "/semantic/search") || contains(path, "/semantic/rag"):
				timeout = config.Query
			case contains(path, "/workflows/") && contains(path, "/execute"):
				timeout = config.Workflow
			case contains(path, "/ingestion/"):
				timeout = config.Ingestion
			default:
				timeout = config.Default
			}

			// Apply timeout middleware
			Timeout(timeout)(next).ServeHTTP(w, r)
		})
	}
}

/* contains checks if a string contains a substring */
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

/* containsSubstring is a simple substring check */
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
