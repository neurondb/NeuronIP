package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

/* ValidateJSON validates JSON request body against a struct */
func ValidateJSON(v interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := json.NewDecoder(r.Body).Decode(v); err != nil {
				handlers.WriteErrorResponse(w, errors.BadRequest("Invalid JSON request body"))
				return
			}

			if err := validate.Struct(v); err != nil {
				validationErrors := make(map[string]string)
				for _, err := range err.(validator.ValidationErrors) {
					field := err.Field()
					tag := err.Tag()
					validationErrors[field] = getValidationErrorMessage(field, tag, err.Param())
				}
				handlers.WriteErrorResponse(w, errors.ValidationFailed("Validation failed", validationErrors))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

/* getValidationErrorMessage returns a user-friendly validation error message */
func getValidationErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	case "email":
		return field + " must be a valid email address"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	default:
		return field + " is invalid"
	}
}
