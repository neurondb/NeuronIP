package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/neurondb/NeuronIP/api/internal/tracing"
)

/* Tracing is a middleware that adds distributed tracing */
func Tracing(tracer *tracing.TracerService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start trace for this request
			ctx, trace := tracer.StartSpan(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			
			// Add request tags
			tracer.AddTag(trace, "http.method", r.Method)
			tracer.AddTag(trace, "http.url", r.URL.String())
			tracer.AddTag(trace, "http.user_agent", r.UserAgent())
			
			// Wrap response writer to capture status code
			wrapped := &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			start := time.Now()
			next.ServeHTTP(wrapped, r.WithContext(ctx))
			duration := time.Since(start)
			
			// Add response tags
			tracer.AddTag(trace, "http.status_code", wrapped.statusCode)
			tracer.AddTag(trace, "http.duration_ms", duration.Milliseconds())
			
			// End trace
			tracer.EndSpan(ctx, trace)
		})
	}
}

/* tracingResponseWriter wraps http.ResponseWriter to capture status code */
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *tracingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
