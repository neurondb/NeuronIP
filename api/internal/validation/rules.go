package validation

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

/* SQLInjectionRule validates against SQL injection patterns */
type SQLInjectionRule struct{}

/* Validate implements ValidationRule */
func (r *SQLInjectionRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.String {
		return nil
	}

	str := val.String()
	
	// Common SQL injection patterns
	sqlPatterns := []string{
		`(?i)(\bUNION\b.*\bSELECT\b)`,
		`(?i)(\bSELECT\b.*\bFROM\b)`,
		`(?i)(\bINSERT\b.*\bINTO\b)`,
		`(?i)(\bUPDATE\b.*\bSET\b)`,
		`(?i)(\bDELETE\b.*\bFROM\b)`,
		`(?i)(\bDROP\b.*\bTABLE\b)`,
		`(?i)(\bEXEC\b|\bEXECUTE\b)`,
		`['";]`,
		`(--|#|/\*|\*/)`,
		`(\bOR\b.*=.*)`,
		`(\bAND\b.*=.*)`,
	}

	for _, pattern := range sqlPatterns {
		matched, err := regexp.MatchString(pattern, str)
		if err != nil {
			continue
		}
		if matched {
			return fmt.Errorf("potential SQL injection detected")
		}
	}

	return nil
}

/* XSSRule validates against XSS patterns */
type XSSRule struct{}

/* Validate implements ValidationRule */
func (r *XSSRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.String {
		return nil
	}

	str := val.String()
	
	// Common XSS patterns
	xssPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`javascript:`,
		`on\w+\s*=`,
		`<iframe[^>]*>`,
		`<img[^>]*src\s*=\s*["']?javascript:`,
		`<svg[^>]*onload\s*=`,
		`expression\s*\(`,
		`vbscript:`,
	}

	for _, pattern := range xssPatterns {
		matched, err := regexp.MatchString(`(?i)`+pattern, str)
		if err != nil {
			continue
		}
		if matched {
			return fmt.Errorf("potential XSS detected")
		}
	}

	return nil
}

/* SanitizeString sanitizes a string to prevent SQL injection and XSS */
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except newlines and tabs
	input = regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]`).ReplaceAllString(input, "")
	
	return input
}

/* SanitizeSQL sanitizes input for SQL queries */
func SanitizeSQL(input string) string {
	// Remove SQL comment patterns
	input = regexp.MustCompile(`(--|#|/\*|\*/)`).ReplaceAllString(input, "")
	
	// Remove semicolons (statement terminators)
	input = strings.ReplaceAll(input, ";", "")
	
	// Remove quotes (basic protection)
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, `"`, `""`)
	
	return input
}

/* EscapeHTML escapes HTML special characters */
func EscapeHTML(input string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		`"`:  "&quot;",
		"'":  "&#39;",
		"/":  "&#47;",
	}

	result := input
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

/* ValidateRequestSize validates request size */
func ValidateRequestSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("request size %d exceeds maximum allowed size %d", size, maxSize)
	}
	return nil
}

/* RequestSizeLimitMiddleware creates middleware to limit request size */
func RequestSizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Limit request body reader
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}
