package config

import (
	"fmt"
)

/* Validate validates the configuration */
func (c *Config) Validate() error {
	// Validate database configuration
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate server configuration
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Validate auth configuration
	if c.Auth.EnableAPIKeys && c.Auth.JWTSecret == "" {
		// JWT secret is optional if using API keys only, but warn
		// This is acceptable for API key only auth
	}

	// Validate logging configuration
	if c.Logging.Level != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
		}
		if !validLevels[c.Logging.Level] {
			return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", c.Logging.Level)
		}
	}

	if c.Logging.Format != "" {
		validFormats := map[string]bool{
			"json": true,
			"text": true,
		}
		if !validFormats[c.Logging.Format] {
			return fmt.Errorf("invalid log format: %s (valid: json, text)", c.Logging.Format)
		}
	}

	return nil
}
