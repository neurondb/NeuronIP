package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* QueryTimeoutMiddleware adds query timeout context to requests */
func QueryTimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	manager := db.NewQueryTimeoutManager(db.DefaultQueryTimeoutConfig())
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create context with query timeout
			queryCtx, cancel := manager.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Add timeout to context
			ctx := context.WithValue(queryCtx, "query_timeout", timeout)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

/* GetQueryTimeoutFromContext retrieves query timeout from context */
func GetQueryTimeoutFromContext(ctx context.Context) time.Duration {
	if timeout, ok := ctx.Value("query_timeout").(time.Duration); ok {
		return timeout
	}
	return 30 * time.Second // Default timeout
}
