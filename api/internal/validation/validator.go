package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/* Validator validates input data */
type Validator struct {
	rules map[string][]ValidationRule
}

/* ValidationRule defines a validation rule */
type ValidationRule interface {
	Validate(value interface{}) error
}

/* NewValidator creates a new validator */
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

/* AddRule adds a validation rule for a field */
func (v *Validator) AddRule(field string, rule ValidationRule) {
	v.rules[field] = append(v.rules[field], rule)
}

/* Validate validates a struct or map */
func (v *Validator) Validate(data interface{}) map[string][]string {
	errors := make(map[string][]string)

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		v.validateStruct(val, errors)
	} else if val.Kind() == reflect.Map {
		v.validateMap(val, errors)
	}

	return errors
}

/* validateStruct validates a struct */
func (v *Validator) validateStruct(val reflect.Value, errors map[string][]string) {
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Get field name from tag or use struct field name
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		// Validate field
		if rules, ok := v.rules[fieldName]; ok {
			for _, rule := range rules {
				if err := rule.Validate(fieldValue.Interface()); err != nil {
					if errors[fieldName] == nil {
						errors[fieldName] = make([]string, 0)
					}
					errors[fieldName] = append(errors[fieldName], err.Error())
				}
			}
		}
	}
}

/* validateMap validates a map */
func (v *Validator) validateMap(val reflect.Value, errors map[string][]string) {
	for _, key := range val.MapKeys() {
		fieldName := key.String()
		fieldValue := val.MapIndex(key)

		if rules, ok := v.rules[fieldName]; ok {
			for _, rule := range rules {
				if err := rule.Validate(fieldValue.Interface()); err != nil {
					if errors[fieldName] == nil {
						errors[fieldName] = make([]string, 0)
					}
					errors[fieldName] = append(errors[fieldName], err.Error())
				}
			}
		}
	}
}

/* RequiredRule validates that a value is not empty */
type RequiredRule struct{}

/* Validate implements ValidationRule */
func (r *RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String && val.String() == "" {
		return fmt.Errorf("field is required")
	}

	if val.Kind() == reflect.Slice && val.Len() == 0 {
		return fmt.Errorf("field is required")
	}

	return nil
}

/* MinLengthRule validates minimum length */
type MinLengthRule struct {
	MinLength int
}

/* Validate implements ValidationRule */
func (r *MinLengthRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		if len(val.String()) < r.MinLength {
			return fmt.Errorf("field must be at least %d characters", r.MinLength)
		}
	}
	return nil
}

/* MaxLengthRule validates maximum length */
type MaxLengthRule struct {
	MaxLength int
}

/* Validate implements ValidationRule */
func (r *MaxLengthRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		if len(val.String()) > r.MaxLength {
			return fmt.Errorf("field must be at most %d characters", r.MaxLength)
		}
	}
	return nil
}

/* PatternRule validates against a regex pattern */
type PatternRule struct {
	Pattern *regexp.Regexp
	Message string
}

/* Validate implements ValidationRule */
func (r *PatternRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		if !r.Pattern.MatchString(val.String()) {
			if r.Message != "" {
				return fmt.Errorf(r.Message)
			}
			return fmt.Errorf("field does not match required pattern")
		}
	}
	return nil
}

/* EmailRule validates email format */
type EmailRule struct{}

/* Validate implements ValidationRule */
func (r *EmailRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailPattern.MatchString(val.String()) {
			return fmt.Errorf("field must be a valid email address")
		}
	}
	return nil
}

/* URLRule validates URL format */
type URLRule struct{}

/* Validate implements ValidationRule */
func (r *URLRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.String {
		urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlPattern.MatchString(val.String()) {
			return fmt.Errorf("field must be a valid URL")
		}
	}
	return nil
}
