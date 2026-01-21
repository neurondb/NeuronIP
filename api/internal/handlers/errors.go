package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/logging"
)

/* ErrorResponse represents an error response */
type ErrorResponse struct {
	Error *errors.APIError `json:"error"`
}

/* WriteError writes an error response to the HTTP response writer */
func WriteError(w http.ResponseWriter, err error) {
	WriteErrorWithContext(w, nil, err)
}

/* WriteErrorWithRequest writes an error response with request context */
func WriteErrorWithRequest(w http.ResponseWriter, r *http.Request, err error) {
	WriteErrorWithContext(w, r, err)
}

/* WriteErrorWithContext writes an error response with request context */
func WriteErrorWithContext(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := errors.AsAPIError(err)
	if apiErr == nil {
		// Wrap unknown errors as internal server errors
		apiErr = errors.InternalServer("An unexpected error occurred")
		if err != nil {
			apiErr = errors.Wrap(err, errors.ErrCodeInternalServer, "An unexpected error occurred")
		}
	}

	// Add request ID to error details if available
	if r != nil {
		requestID := logging.GetRequestID(r.Context())
		if requestID != "" {
			if apiErr.Details == nil {
				apiErr.Details = make(map[string]interface{})
			}
			if detailsMap, ok := apiErr.Details.(map[string]interface{}); ok {
				detailsMap["request_id"] = requestID
			} else {
				// If details is not a map, create a new map
				apiErr.Details = map[string]interface{}{
					"request_id": requestID,
					"original_details": apiErr.Details,
				}
			}
			// Ensure request ID is in response header
			w.Header().Set("X-Request-ID", requestID)
		}
	}

	statusCode := apiErr.HTTPStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: apiErr}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback if JSON encoding fails
		http.Error(w, apiErr.Message, statusCode)
	}
}

/* WriteErrorResponse writes an APIError directly to the response */
func WriteErrorResponse(w http.ResponseWriter, apiErr *errors.APIError) {
	WriteErrorResponseWithContext(w, nil, apiErr)
}

/* WriteErrorResponseWithRequest writes an APIError with request context */
func WriteErrorResponseWithRequest(w http.ResponseWriter, r *http.Request, apiErr *errors.APIError) {
	WriteErrorResponseWithContext(w, r, apiErr)
}

/* WriteErrorResponseWithContext writes an APIError with request context */
func WriteErrorResponseWithContext(w http.ResponseWriter, r *http.Request, apiErr *errors.APIError) {
	// Add request ID to error details if available
	if r != nil {
		requestID := logging.GetRequestID(r.Context())
		if requestID != "" {
			if apiErr.Details == nil {
				apiErr.Details = make(map[string]interface{})
			}
			if detailsMap, ok := apiErr.Details.(map[string]interface{}); ok {
				detailsMap["request_id"] = requestID
			} else {
				// If details is not a map, create a new map
				apiErr.Details = map[string]interface{}{
					"request_id": requestID,
					"original_details": apiErr.Details,
				}
			}
			// Ensure request ID is in response header
			w.Header().Set("X-Request-ID", requestID)
		}
	}

	statusCode := apiErr.HTTPStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: apiErr}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, apiErr.Message, statusCode)
	}
}
