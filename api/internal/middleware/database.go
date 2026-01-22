package middleware

import (
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* DatabaseMiddleware injects the correct database pool into request context based on session */
func DatabaseMiddleware(multiPool *db.MultiPool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get database name from context (set by session middleware)
			databaseName := db.GetDatabaseNameFromContext(r.Context())

			// Get the appropriate pool
			pool, err := multiPool.GetPool(databaseName)
			if err != nil {
				// Fallback to default pool if database not found
				pool, _ = multiPool.GetPool("neuronip")
			}

			// Add pool to context
			ctx := db.WithDatabase(r.Context(), pool)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
