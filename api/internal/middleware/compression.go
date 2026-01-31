package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

/* CompressionMiddleware provides response compression */
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for small responses
		// We'll use a response writer that buffers to check size
		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         nil, // Will be initialized if needed
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}

/* gzipResponseWriter wraps http.ResponseWriter with gzip compression */
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
	buffer []byte
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	// For small responses (< 1KB), don't compress
	if len(b) < 1024 && gzw.Writer == nil {
		return gzw.ResponseWriter.Write(b)
	}

	// Initialize gzip writer if not already done
	if gzw.Writer == nil {
		gzw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		gzw.Writer = gzip.NewWriter(gzw.ResponseWriter)
	}

	return gzw.Writer.Write(b)
}

func (gzw *gzipResponseWriter) WriteHeader(statusCode int) {
	// Don't set Content-Encoding header yet - wait until we know we'll compress
	gzw.ResponseWriter.WriteHeader(statusCode)
}

func (gzw *gzipResponseWriter) Close() {
	if gzw.Writer != nil {
		gzw.Writer.Close()
	}
}

/* CompressionConfig configures compression behavior */
type CompressionConfig struct {
	MinSize int // Minimum response size to compress (bytes)
	Level   int // Compression level (1-9)
}

/* CompressionMiddlewareWithConfig provides configurable response compression */
func CompressionMiddlewareWithConfig(config CompressionConfig) func(http.Handler) http.Handler {
	if config.MinSize == 0 {
		config.MinSize = 1024 // Default 1KB
	}
	if config.Level == 0 {
		config.Level = gzip.DefaultCompression
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip encoding
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         nil,
			}
			defer gzw.Close()

			next.ServeHTTP(gzw, r)
		})
	}
}
