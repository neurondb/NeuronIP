package errors

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
)

/* ErrorType represents the type of error */
type ErrorType string

const (
	// ErrorTypeTransient represents errors that may succeed on retry
	ErrorTypeTransient ErrorType = "transient"
	// ErrorTypePermanent represents errors that won't succeed on retry
	ErrorTypePermanent ErrorType = "permanent"
	// ErrorTypeUnknown represents unknown error types
	ErrorTypeUnknown ErrorType = "unknown"
)

/* ErrorContext holds contextual information about an error */
type ErrorContext struct {
	RequestID string
	UserID    string
	Operation string
	Stack     []string
	Type      ErrorType
}

/* WithContext adds context information to an APIError */
func (e *APIError) WithContext(ctx ErrorContext) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	
	detailsMap, ok := e.Details.(map[string]interface{})
	if !ok {
		detailsMap = make(map[string]interface{})
		if e.Details != nil {
			detailsMap["original_details"] = e.Details
		}
		e.Details = detailsMap
	}
	
	if ctx.RequestID != "" {
		detailsMap["request_id"] = ctx.RequestID
	}
	if ctx.UserID != "" {
		detailsMap["user_id"] = ctx.UserID
	}
	if ctx.Operation != "" {
		detailsMap["operation"] = ctx.Operation
	}
	if len(ctx.Stack) > 0 {
		detailsMap["stack"] = ctx.Stack
	}
	if ctx.Type != "" {
		detailsMap["error_type"] = ctx.Type
	}
	
	return e
}

/* WithRequestID adds a request ID to the error */
func (e *APIError) WithRequestID(requestID string) *APIError {
	return e.WithContext(ErrorContext{RequestID: requestID})
}

/* WithUserID adds a user ID to the error */
func (e *APIError) WithUserID(userID string) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	detailsMap, ok := e.Details.(map[string]interface{})
	if !ok {
		detailsMap = make(map[string]interface{})
		if e.Details != nil {
			detailsMap["original_details"] = e.Details
		}
		e.Details = detailsMap
	}
	detailsMap["user_id"] = userID
	return e
}

/* WithOperation adds an operation name to the error */
func (e *APIError) WithOperation(operation string) *APIError {
	return e.WithContext(ErrorContext{Operation: operation})
}

/* WithStack adds a stack trace to the error (for development) */
func (e *APIError) WithStack() *APIError {
	stack := captureStack(10) // Capture 10 frames
	return e.WithContext(ErrorContext{Stack: stack})
}

/* WithType marks the error as transient or permanent */
func (e *APIError) WithType(errType ErrorType) *APIError {
	return e.WithContext(ErrorContext{Type: errType})
}

/* IsTransient checks if the error is transient */
func (e *APIError) IsTransient() bool {
	if e.Details == nil {
		return false
	}
	detailsMap, ok := e.Details.(map[string]interface{})
	if !ok {
		return false
	}
	errType, ok := detailsMap["error_type"].(ErrorType)
	if !ok {
		// Try string conversion
		if errTypeStr, ok := detailsMap["error_type"].(string); ok {
			errType = ErrorType(errTypeStr)
		} else {
			return false
		}
	}
	return errType == ErrorTypeTransient
}

/* IsPermanent checks if the error is permanent */
func (e *APIError) IsPermanent() bool {
	if e.Details == nil {
		return true // Default to permanent if not specified
	}
	detailsMap, ok := e.Details.(map[string]interface{})
	if !ok {
		return true
	}
	errType, ok := detailsMap["error_type"].(ErrorType)
	if !ok {
		// Try string conversion
		if errTypeStr, ok := detailsMap["error_type"].(string); ok {
			errType = ErrorType(errTypeStr)
		} else {
			return true // Default to permanent
		}
	}
	return errType == ErrorTypePermanent
}

/* captureStack captures the current stack trace */
func captureStack(skip int) []string {
	stack := make([]string, 0, 10)
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		funcName := fn.Name()
		// Clean up function name
		parts := strings.Split(funcName, ".")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, funcName))
	}
	return stack
}

/* ClassifyError classifies an error as transient or permanent based on error code */
func ClassifyError(err *APIError) ErrorType {
	switch err.Code {
	case ErrCodeTimeout, ErrCodeServiceUnavailable:
		return ErrorTypeTransient
	case ErrCodeBadRequest, ErrCodeUnauthorized, ErrCodeForbidden, 
		 ErrCodeNotFound, ErrCodeConflict, ErrCodeValidationFailed:
		return ErrorTypePermanent
	case ErrCodeInternalServer:
		// Internal server errors could be transient (database connection) or permanent (logic error)
		// Default to transient to allow retries
		return ErrorTypeTransient
	default:
		return ErrorTypeUnknown
	}
}

/* WrapWithContext wraps an error with context information */
func WrapWithContext(err error, code ErrorCode, message string, ctx context.Context) *APIError {
	apiErr := Wrap(err, code, message)
	
	// Extract context information
	errorCtx := ErrorContext{}
	
	// Try to get request ID from context
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		errorCtx.RequestID = requestID
	}
	
	// Try to get user ID from context
	if userID := getUserIDFromContext(ctx); userID != "" {
		errorCtx.UserID = userID
	}
	
	// Classify error type
	errorCtx.Type = ClassifyError(apiErr)
	
	// Add stack trace in development mode
	if isDevelopment() {
		errorCtx.Stack = captureStack(3)
	}
	
	return apiErr.WithContext(errorCtx)
}

/* Helper functions to extract context values */
func getRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// Try common context key patterns
	type requestIDKey struct{}
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	// Try logging package's request ID key
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// Try common context key patterns
	type userIDKey struct{}
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	if id, ok := ctx.Value("user_id").(string); ok {
		return id
	}
	return ""
}

func isDevelopment() bool {
	// Check environment variable or other indicators
	env := os.Getenv("ENVIRONMENT")
	return env == "development" || env == "dev" || env == "local"
}
