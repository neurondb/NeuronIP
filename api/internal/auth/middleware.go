package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const apiKeyContextKey contextKey = "api_key"
const userIDContextKey contextKey = "user_id"

/* Middleware provides API key authentication middleware */
func Middleware(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/api/v1/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			key, err := ExtractAPIKey(authHeader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			apiKey, err := ValidateAPIKey(r.Context(), queries, key)
			if err != nil {
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
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
	if len(key) < 8 {
		return nil, fmt.Errorf("API key too short")
	}

	prefix := key[:8]
	apiKey, err := queries.GetAPIKeyByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("API key not found: %w", err)
	}

	// Hash the provided key and compare
	hasher := sha256.New()
	hasher.Write([]byte(key))
	keyHash := hex.EncodeToString(hasher.Sum(nil))

	if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(keyHash)); err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	// Update last used timestamp
	queries.UpdateAPIKeyLastUsed(ctx, apiKey.ID)

	return apiKey, nil
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
