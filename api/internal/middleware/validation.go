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
				handlers.WriteErrorResponseWithContext(w, r, errors.BadRequest("Invalid JSON request body"))
				return
			}

			if err := validate.Struct(v); err != nil {
				validationErrors := make(map[string]string)
				for _, err := range err.(validator.ValidationErrors) {
					field := err.Field()
					tag := err.Tag()
					validationErrors[field] = getValidationErrorMessage(field, tag, err.Param())
				}
				handlers.WriteErrorResponseWithContext(w, r, errors.ValidationFailed("Validation failed", validationErrors))
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
		return "This field is required"
	case "min":
		if param == "" {
			return field + " is too short"
		}
		return field + " must be at least " + param + " characters long"
	case "max":
		if param == "" {
			return field + " is too long"
		}
		return field + " must be at most " + param + " characters long"
	case "len":
		return field + " must be exactly " + param + " characters long"
	case "email":
		return "Please enter a valid email address"
	case "url":
		return "Please enter a valid URL"
	case "uuid":
		return "Please enter a valid UUID"
	case "numeric":
		return field + " must be a number"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "gte":
		return field + " must be greater than or equal to " + param
	case "lte":
		return field + " must be less than or equal to " + param
	case "gt":
		return field + " must be greater than " + param
	case "lt":
		return field + " must be less than " + param
	case "oneof":
		return field + " must be one of: " + param
	case "eqfield":
		return field + " must match " + param
	case "nefield":
		return field + " must not match " + param
	default:
		return field + " is invalid (" + tag + ")"
	}
}
