package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/neurondb/NeuronIP/api/internal/errors"
)

/* ErrorResponse represents an error response */
type ErrorResponse struct {
	Error *errors.APIError `json:"error"`
}

/* WriteError writes an error response to the HTTP response writer */
func WriteError(w http.ResponseWriter, err error) {
	apiErr := errors.AsAPIError(err)
	if apiErr == nil {
		// Wrap unknown errors as internal server errors
		apiErr = errors.InternalServer("An unexpected error occurred")
		if err != nil {
			apiErr = errors.Wrap(err, errors.ErrCodeInternalServer, "An unexpected error occurred")
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
	statusCode := apiErr.HTTPStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: apiErr}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, apiErr.Message, statusCode)
	}
}
