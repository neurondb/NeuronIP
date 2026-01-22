package security

import (
	"fmt"
	"net/http"
)

/* SecurityHeaderConfig holds security header configuration */
type SecurityHeaderConfig struct {
	FrameOptions           string
	ContentTypeOptions     string
	XSSProtection          string
	HSTSMaxAge             int
	HSTSIncludeSubdomains  bool
	HSTSPreload            bool
	ReferrerPolicy         string
	PermissionsPolicy      string
	ContentSecurityPolicy  string
	CrossOriginEmbedderPolicy string
	CrossOriginOpenerPolicy string
	CrossOriginResourcePolicy string
}

/* DefaultSecurityHeaderConfig returns default security header configuration */
func DefaultSecurityHeaderConfig() *SecurityHeaderConfig {
	return &SecurityHeaderConfig{
		FrameOptions:           "DENY",
		ContentTypeOptions:     "nosniff",
		XSSProtection:          "1; mode=block",
		HSTSMaxAge:             31536000, // 1 year
		HSTSIncludeSubdomains:  true,
		HSTSPreload:            false,
		ReferrerPolicy:         "strict-origin-when-cross-origin",
		PermissionsPolicy:      "geolocation=(), microphone=(), camera=()",
		ContentSecurityPolicy:  "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy: "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}
}

/* ApplySecurityHeaders applies security headers to a response */
func ApplySecurityHeaders(w http.ResponseWriter, r *http.Request, config *SecurityHeaderConfig) {
	if config == nil {
		config = DefaultSecurityHeaderConfig()
	}

	// X-Frame-Options
	if config.FrameOptions != "" {
		w.Header().Set("X-Frame-Options", config.FrameOptions)
	}

	// X-Content-Type-Options
	if config.ContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
	}

	// X-XSS-Protection
	if config.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", config.XSSProtection)
	}

	// Strict-Transport-Security (HSTS) - only if using HTTPS
	if r.TLS != nil && config.HSTSMaxAge > 0 {
		hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
		if config.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if config.HSTSPreload {
			hstsValue += "; preload"
		}
		w.Header().Set("Strict-Transport-Security", hstsValue)
	}

	// Referrer-Policy
	if config.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
	}

	// Permissions-Policy
	if config.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
	}

	// Content-Security-Policy
	if config.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
	}

	// Cross-Origin-Embedder-Policy
	if config.CrossOriginEmbedderPolicy != "" {
		w.Header().Set("Cross-Origin-Embedder-Policy", config.CrossOriginEmbedderPolicy)
	}

	// Cross-Origin-Opener-Policy
	if config.CrossOriginOpenerPolicy != "" {
		w.Header().Set("Cross-Origin-Opener-Policy", config.CrossOriginOpenerPolicy)
	}

	// Cross-Origin-Resource-Policy
	if config.CrossOriginResourcePolicy != "" {
		w.Header().Set("Cross-Origin-Resource-Policy", config.CrossOriginResourcePolicy)
	}
}
