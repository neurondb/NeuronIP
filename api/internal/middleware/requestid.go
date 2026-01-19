package middleware

import (
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/logging"
)

/* RequestID is a middleware that adds a request ID to each request */
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = logging.GenerateRequestID()
		}

		// Add request ID to context
		ctx := logging.SetRequestID(r.Context(), requestID)

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
