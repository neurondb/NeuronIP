package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
	"github.com/neurondb/NeuronIP/api/internal/logging"
)

/* Recovery is a middleware that recovers from panics */
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger := logging.DefaultLogger
				if logger != nil {
					logger = logger.WithContext(r.Context())
				}
				logger.Error("Panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path,
				)

				// Write error response
				apiErr := errors.InternalServer("An unexpected error occurred")
				handlers.WriteErrorResponse(w, apiErr)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
