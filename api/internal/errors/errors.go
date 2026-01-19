package errors

import (
	"errors"
	"fmt"
	"net/http"
)

/* ErrorCode represents an error code */
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"

	// Server errors (5xx)
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout         ErrorCode = "TIMEOUT"
)

/* APIError represents a structured API error */
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Err     error     `json:"-"`
}

/* Error implements the error interface */
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

/* Unwrap returns the underlying error */
func (e *APIError) Unwrap() error {
	return e.Err
}

/* HTTPStatus returns the HTTP status code for the error */
func (e *APIError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeBadRequest, ErrCodeValidationFailed:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

/* New creates a new APIError */
func New(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

/* Wrap wraps an error with an APIError */
func Wrap(err error, code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

/* WithDetails adds details to the error */
func (e *APIError) WithDetails(details interface{}) *APIError {
	e.Details = details
	return e
}

/* Predefined error constructors */

/* BadRequest creates a bad request error */
func BadRequest(message string) *APIError {
	return New(ErrCodeBadRequest, message)
}

/* Unauthorized creates an unauthorized error */
func Unauthorized(message string) *APIError {
	return New(ErrCodeUnauthorized, message)
}

/* Forbidden creates a forbidden error */
func Forbidden(message string) *APIError {
	return New(ErrCodeForbidden, message)
}

/* NotFound creates a not found error */
func NotFound(resource string) *APIError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

/* Conflict creates a conflict error */
func Conflict(message string) *APIError {
	return New(ErrCodeConflict, message)
}

/* ValidationFailed creates a validation error */
func ValidationFailed(message string, details interface{}) *APIError {
	return New(ErrCodeValidationFailed, message).WithDetails(details)
}

/* TooManyRequests creates a rate limit error */
func TooManyRequests(message string) *APIError {
	return New(ErrCodeTooManyRequests, message)
}

/* InternalServer creates an internal server error */
func InternalServer(message string) *APIError {
	return New(ErrCodeInternalServer, message)
}

/* ServiceUnavailable creates a service unavailable error */
func ServiceUnavailable(message string) *APIError {
	return New(ErrCodeServiceUnavailable, message)
}

/* Timeout creates a timeout error */
func Timeout(message string) *APIError {
	return New(ErrCodeTimeout, message)
}

/* IsAPIError checks if an error is an APIError */
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

/* AsAPIError converts an error to APIError, or returns nil */
func AsAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
