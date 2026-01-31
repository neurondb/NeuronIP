package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"github.com/neurondb/NeuronIP/api/internal/cache"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const apiKeyContextKey contextKey = "api_key"
const userIDContextKey contextKey = "user_id"

/* Middleware provides API key authentication middleware */
func Middleware(queries *db.Queries, cacheService cache.CacheInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/api/v1/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeErrorResponse(w, errors.Unauthorized("Missing authorization header"))
				return
			}

			key, err := ExtractAPIKey(authHeader)
			if err != nil {
				writeErrorResponse(w, errors.Unauthorized(err.Error()))
				return
			}

			apiKey, err := ValidateAPIKeyWithCache(r.Context(), queries, cacheService, key)
			if err != nil {
				writeErrorResponse(w, errors.Unauthorized("Invalid API key"))
				return
			}

			ctx := SetAPIKey(r.Context(), apiKey)
			if apiKey.UserID != nil {
				ctx = SetUserID(ctx, *apiKey.UserID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

/* ExtractAPIKey extracts the API key from the Authorization header */
func ExtractAPIKey(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return parts[1], nil
}

/* ValidateAPIKey validates an API key and returns the API key record */
func ValidateAPIKey(ctx context.Context, queries *db.Queries, key string) (*db.APIKey, error) {
	return ValidateAPIKeyWithCache(ctx, queries, nil, key)
}

/* ValidateAPIKeyWithCache validates an API key with caching support */
func ValidateAPIKeyWithCache(ctx context.Context, queries *db.Queries, cacheService cache.CacheInterface, key string) (*db.APIKey, error) {
	if len(key) < 8 {
		return nil, fmt.Errorf("API key too short")
	}

	// Generate cache key
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))
	cacheKey := "api_key:" + keyHash

	// Try to get from cache first
	if cacheService != nil {
		if cached, found := cacheService.Get(ctx, cacheKey); found {
			if apiKey, ok := cached.(*db.APIKey); ok {
				// Update last used timestamp asynchronously
				go func() {
					queries.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
				}()
				return apiKey, nil
			}
		}
	}

	prefix := key[:8]
	apiKey, err := queries.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Hash the provided key and compare
	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(keyHash)); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Cache the validated API key (10 minute TTL)
	if cacheService != nil {
		cacheService.Set(ctx, cacheKey, apiKey, 10*time.Minute)
	}

	// Update last used timestamp
	queries.UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey, nil
}

/* InvalidateAPIKeyCache invalidates cached API key */
func InvalidateAPIKeyCache(cacheService cache.CacheInterface, key string) error {
	if cacheService == nil {
		return nil
	}

	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))
	cacheKey := "api_key:" + keyHash

	return cacheService.Delete(context.Background(), cacheKey)
}

/* SetAPIKey sets the API key in the context */
func SetAPIKey(ctx context.Context, key *db.APIKey) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, key)
}

/* GetAPIKeyFromContext gets the API key from context */
func GetAPIKeyFromContext(ctx context.Context) (*db.APIKey, bool) {
	key, ok := ctx.Value(apiKeyContextKey).(*db.APIKey)
	return key, ok
}

/* SetUserID sets the user ID in the context */
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

/* GetUserIDFromContext gets the user ID from context */
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

/* writeErrorResponse writes an error response without importing handlers to avoid cycle */
func writeErrorResponse(w http.ResponseWriter, apiErr *errors.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPStatus())

	response := map[string]interface{}{
		"error": apiErr,
	}
	json.NewEncoder(w).Encode(response)
}
