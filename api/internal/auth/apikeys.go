package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

/* APIKeyService provides enhanced API key management */
type APIKeyService struct {
	queries *db.Queries
}

/* NewAPIKeyService creates a new API key service */
func NewAPIKeyService(queries *db.Queries) *APIKeyService {
	return &APIKeyService{
		queries: queries,
	}
}

/* CreateAPIKeyRequest represents API key creation request */
type CreateAPIKeyRequest struct {
	Name              string
	UserID            *uuid.UUID
	Scopes            []string
	RateLimit         int
	ExpiresAt         *time.Time
	RotationEnabled   bool
	RotationIntervalDays int
	Tags              []string
	Metadata          map[string]interface{}
}

/* APIKeyResponse represents API key response */
type APIKeyResponse struct {
	ID                uuid.UUID              `json:"id"`
	Key               string                 `json:"key"` // Only shown once on creation
	KeyPrefix         string                 `json:"key_prefix"`
	Name              *string                `json:"name"`
	Scopes            []string               `json:"scopes"`
	RateLimit         int                    `json:"rate_limit"`
	ExpiresAt         *time.Time             `json:"expires_at"`
	RotationEnabled   bool                   `json:"rotation_enabled"`
	NextRotationAt    *time.Time             `json:"next_rotation_at"`
	Tags              []string               `json:"tags"`
	CreatedAt         time.Time              `json:"created_at"`
}

/* CreateAPIKey creates a new API key with enhanced features */
func (s *APIKeyService) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKeyResponse, error) {
	// Generate API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	key := "sk_live_" + base64.URLEncoding.EncodeToString(keyBytes)
	keyPrefix := key[:12] // First 12 characters for display

	// Hash the key
	keyHash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	// Calculate next rotation date if rotation enabled
	var nextRotationAt *time.Time
	if req.RotationEnabled && req.RotationIntervalDays > 0 {
		next := time.Now().Add(time.Duration(req.RotationIntervalDays) * 24 * time.Hour)
		nextRotationAt = &next
	}

	// Insert API key
	keyID := uuid.New()
	query := `
		INSERT INTO neuronip.api_keys (
			id, key_hash, key_prefix, user_id, name, scopes, rate_limit,
			expires_at, rotation_enabled, rotation_interval_days, next_rotation_at,
			tags, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		RETURNING created_at
	`
	var createdAt time.Time
	err = s.queries.DB.QueryRow(ctx, query,
		keyID, string(keyHash), keyPrefix, req.UserID, req.Name, req.Scopes, req.RateLimit,
		req.ExpiresAt, req.RotationEnabled, req.RotationIntervalDays, nextRotationAt,
		req.Tags, req.Metadata,
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Create rate limit configuration
	rateLimitQuery := `
		INSERT INTO neuronip.api_key_rate_limits (
			api_key_id, requests_per_minute, requests_per_hour, requests_per_day
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (api_key_id) DO UPDATE SET
			requests_per_minute = $2,
			requests_per_hour = $3,
			requests_per_day = $4
	`
	_, err = s.queries.DB.Exec(ctx, rateLimitQuery, keyID, req.RateLimit, req.RateLimit*60, req.RateLimit*1440)
	if err != nil {
		return nil, fmt.Errorf("failed to set rate limits: %w", err)
	}

	namePtr := &req.Name
	return &APIKeyResponse{
		ID:              keyID,
		Key:             key, // Only returned once
		KeyPrefix:       keyPrefix,
		Name:            namePtr,
		Scopes:          req.Scopes,
		RateLimit:       req.RateLimit,
		ExpiresAt:       req.ExpiresAt,
		RotationEnabled: req.RotationEnabled,
		NextRotationAt:  nextRotationAt,
		Tags:            req.Tags,
		CreatedAt:       createdAt,
	}, nil
}

/* RotateAPIKey rotates an API key */
func (s *APIKeyService) RotateAPIKey(ctx context.Context, keyID uuid.UUID, rotationType string) (*APIKeyResponse, error) {
	// Get existing key
	query := `SELECT key_prefix, user_id, name, scopes, rate_limit, rotation_interval_days, tags, metadata
	          FROM neuronip.api_keys WHERE id = $1`
	var oldPrefix string
	var userID *uuid.UUID
	var name *string
	var scopes []string
	var rateLimit int
	var rotationIntervalDays *int
	var tags []string
	var metadata map[string]interface{}

	err := s.queries.DB.QueryRow(ctx, query, keyID).Scan(
		&oldPrefix, &userID, &name, &scopes, &rateLimit, &rotationIntervalDays, &tags, &metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Generate new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	newKey := "sk_live_" + base64.URLEncoding.EncodeToString(keyBytes)
	newPrefix := newKey[:12]

	// Hash the new key
	keyHash, err := bcrypt.GenerateFromPassword([]byte(newKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	// Calculate next rotation
	var nextRotationAt *time.Time
	if rotationIntervalDays != nil && *rotationIntervalDays > 0 {
		next := time.Now().Add(time.Duration(*rotationIntervalDays) * 24 * time.Hour)
		nextRotationAt = &next
	}

	// Update API key
	updateQuery := `
		UPDATE neuronip.api_keys
		SET key_hash = $1, key_prefix = $2, last_rotated_at = NOW(), next_rotation_at = $3
		WHERE id = $4
		RETURNING created_at
	`
	var createdAt time.Time
	err = s.queries.DB.QueryRow(ctx, updateQuery, string(keyHash), newPrefix, nextRotationAt, keyID).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate API key: %w", err)
	}

	// Record rotation
	rotationQuery := `
		INSERT INTO neuronip.api_key_rotations (api_key_id, old_key_prefix, new_key_prefix, rotation_type)
		VALUES ($1, $2, $3, $4)
	`
	_, err = s.queries.DB.Exec(ctx, rotationQuery, keyID, oldPrefix, newPrefix, rotationType)
	if err != nil {
		// Log but don't fail
	}

	return &APIKeyResponse{
		ID:              keyID,
		Key:             newKey, // Only returned once
		KeyPrefix:       newPrefix,
		Name:            name,
		Scopes:          scopes,
		RateLimit:       rateLimit,
		RotationEnabled: rotationIntervalDays != nil,
		NextRotationAt:  nextRotationAt,
		Tags:            tags,
		CreatedAt:       createdAt,
	}, nil
}

/* CheckRateLimit checks if API key is within rate limits */
func (s *APIKeyService) CheckRateLimit(ctx context.Context, keyID uuid.UUID, windowType string) (bool, error) {
	query := `SELECT neuronip.check_api_key_rate_limit($1, $2)`
	var allowed bool
	err := s.queries.DB.QueryRow(ctx, query, keyID, windowType).Scan(&allowed)
	return allowed, err
}

/* RecordUsage records API key usage for analytics */
func (s *APIKeyService) RecordUsage(ctx context.Context, keyID uuid.UUID, endpoint, method string, statusCode, responseTimeMs, requestSize, responseSize int, ipAddress, userAgent string) error {
	query := `
		INSERT INTO neuronip.api_key_usage (
			api_key_id, endpoint, method, status_code, response_time_ms,
			request_size_bytes, response_size_bytes, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.queries.DB.Exec(ctx, query, keyID, endpoint, method, statusCode, responseTimeMs,
		requestSize, responseSize, ipAddress, userAgent)
	return err
}

/* GetUsageAnalytics gets usage analytics for an API key */
func (s *APIKeyService) GetUsageAnalytics(ctx context.Context, keyID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as successful_requests,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) as failed_requests,
			AVG(response_time_ms) as avg_response_time_ms,
			MAX(created_at) as last_used_at
		FROM neuronip.api_key_usage
		WHERE api_key_id = $1 AND created_at BETWEEN $2 AND $3
	`
	var totalRequests, successfulRequests, failedRequests int
	var avgResponseTime *float64
	var lastUsedAt *time.Time

	err := s.queries.DB.QueryRow(ctx, query, keyID, startDate, endDate).Scan(
		&totalRequests, &successfulRequests, &failedRequests, &avgResponseTime, &lastUsedAt,
	)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_requests":      totalRequests,
		"successful_requests": successfulRequests,
		"failed_requests":     failedRequests,
	}
	if avgResponseTime != nil {
		result["avg_response_time_ms"] = *avgResponseTime
	}
	if lastUsedAt != nil {
		result["last_used_at"] = lastUsedAt
	}

	return result, nil
}

/* ValidateScopes validates that API key has required scopes */
func (s *APIKeyService) ValidateScopes(ctx context.Context, keyID uuid.UUID, requiredScopes []string) (bool, error) {
	query := `SELECT scopes FROM neuronip.api_keys WHERE id = $1`
	var scopes []string
	err := s.queries.DB.QueryRow(ctx, query, keyID).Scan(&scopes)
	if err != nil {
		return false, err
	}

	// Check if all required scopes are present
	scopeMap := make(map[string]bool)
	for _, scope := range scopes {
		scopeMap[scope] = true
		// Check wildcard scopes (e.g., "semantic:*" matches "semantic:read")
		if strings.HasSuffix(scope, ":*") {
			prefix := strings.TrimSuffix(scope, ":*")
			for _, reqScope := range requiredScopes {
				if strings.HasPrefix(reqScope, prefix+":") {
					scopeMap[reqScope] = true
				}
			}
		}
	}

	for _, reqScope := range requiredScopes {
		if !scopeMap[reqScope] {
			return false, nil
		}
	}

	return true, nil
}

/* RevokeAPIKey revokes an API key */
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID, reason string) error {
	query := `UPDATE neuronip.api_keys SET revoked_at = NOW(), revoked_reason = $1 WHERE id = $2`
	_, err := s.queries.DB.Exec(ctx, query, reason, keyID)
	return err
}

/* CheckExpiry checks if API key is expired */
func (s *APIKeyService) CheckExpiry(ctx context.Context, keyID uuid.UUID) (bool, error) {
	query := `SELECT expires_at, revoked_at FROM neuronip.api_keys WHERE id = $1`
	var expiresAt *time.Time
	var revokedAt *time.Time
	err := s.queries.DB.QueryRow(ctx, query, keyID).Scan(&expiresAt, &revokedAt)
	if err != nil {
		return false, err
	}

	// Check if revoked
	if revokedAt != nil {
		return true, nil // Expired (revoked)
	}

	// Check if expired
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return true, nil // Expired
	}

	return false, nil // Not expired
}
