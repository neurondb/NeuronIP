package config

import (
	"fmt"
	"reflect"
	"strings"
)

/* SchemaField defines a configuration field schema */
type SchemaField struct {
	Name        string
	Type        string
	Required    bool
	Default     interface{}
	Description string
	Validation  func(interface{}) error
}

/* ConfigSchema defines the configuration schema */
type ConfigSchema struct {
	Fields map[string]*SchemaField
}

/* NewConfigSchema creates a new configuration schema */
func NewConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Fields: make(map[string]*SchemaField),
	}
}

/* AddField adds a field to the schema */
func (cs *ConfigSchema) AddField(field *SchemaField) {
	cs.Fields[field.Name] = field
}

/* Validate validates configuration against schema */
func (cs *ConfigSchema) Validate(cfg *Config) []error {
	var errors []error

	val := reflect.ValueOf(cfg).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Get field name from tag or use struct field name
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = strings.ToLower(field.Name)
		}

		schemaField, exists := cs.Fields[fieldName]
		if !exists {
			continue
		}

		// Check required
		if schemaField.Required {
			if isZeroValue(fieldValue) {
				errors = append(errors, fmt.Errorf("field %s is required", fieldName))
				continue
			}
		}

		// Check type
		if schemaField.Type != "" {
			if !isTypeMatch(fieldValue, schemaField.Type) {
				errors = append(errors, fmt.Errorf("field %s has invalid type, expected %s", fieldName, schemaField.Type))
				continue
			}
		}

		// Run custom validation
		if schemaField.Validation != nil {
			if err := schemaField.Validation(fieldValue.Interface()); err != nil {
				errors = append(errors, fmt.Errorf("field %s validation failed: %w", fieldName, err))
			}
		}
	}

	return errors
}

/* isZeroValue checks if a value is zero */
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

/* isTypeMatch checks if a value matches a type */
func isTypeMatch(v reflect.Value, expectedType string) bool {
	actualType := v.Type().String()
	return strings.Contains(actualType, expectedType)
}

/* DefaultConfigSchema returns the default configuration schema */
func DefaultConfigSchema() *ConfigSchema {
	schema := NewConfigSchema()

	// Database fields
	schema.AddField(&SchemaField{
		Name:     "database.host",
		Type:     "string",
		Required: true,
		Description: "Database host",
	})

	schema.AddField(&SchemaField{
		Name:     "database.port",
		Type:     "string",
		Required: true,
		Description: "Database port",
	})

	schema.AddField(&SchemaField{
		Name:     "database.user",
		Type:     "string",
		Required: true,
		Description: "Database user",
	})

	schema.AddField(&SchemaField{
		Name:     "database.password",
		Type:     "string",
		Required: true,
		Description: "Database password",
	})

	schema.AddField(&SchemaField{
		Name:     "database.name",
		Type:     "string",
		Required: true,
		Description: "Database name",
	})

	// Server fields
	schema.AddField(&SchemaField{
		Name:     "server.port",
		Type:     "string",
		Required: true,
		Description: "Server port",
	})

	return schema
}
