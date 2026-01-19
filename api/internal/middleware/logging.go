package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/logging"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
)

/* responseWriter wraps http.ResponseWriter to capture status code and response size */
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)
	return size, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrHijacked
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

/* HTTPLogging is a middleware that logs HTTP requests and responses */
func HTTPLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get route pattern
		var routePattern string
		if route := mux.CurrentRoute(r); route != nil {
			if pathTemplate, err := route.GetPathTemplate(); err == nil {
				routePattern = pathTemplate
			}
		}
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		// Get request size
		var requestSize int64
		if r.ContentLength > 0 {
			requestSize = r.ContentLength
		} else if r.Body != nil {
			// Try to read body to get size (this consumes the body)
			body, err := io.ReadAll(r.Body)
			if err == nil {
				requestSize = int64(len(body))
				r.Body = io.NopCloser(bytes.NewReader(body))
			}
		}

		// Wrap response writer
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request
		logger := logging.DefaultLogger
		if logger != nil {
			logger = logger.WithContext(r.Context())
			
			logger.Info("HTTP request",
				"method", r.Method,
				"path", routePattern,
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_size", requestSize,
				"response_size", rw.responseSize,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		}

		// Record metrics
		metrics.RecordHTTPRequest(
			r.Method,
			routePattern,
			rw.statusCode,
			duration,
			requestSize,
			rw.responseSize,
		)
	})
}
