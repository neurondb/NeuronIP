package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, queries *db.Queries, email, password string) (uuid.UUID, error) {
	t.Helper()

	ctx := context.Background()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}
	hashStr := string(hashedPassword)

	user, err := queries.CreateUser(ctx, email, &hashStr, nil, "analyst")
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

// CreateTestAPIKey creates a test API key
func CreateTestAPIKey(t *testing.T, queries *db.Queries, userID uuid.UUID, name string) (string, error) {
	t.Helper()

	ctx := context.Background()
	svc := auth.NewAPIKeyService(queries)
	exp := time.Now().Add(24 * time.Hour)
	resp, err := svc.CreateAPIKey(ctx, auth.CreateAPIKeyRequest{
		Name:      name,
		UserID:    &userID,
		Scopes:    []string{},
		RateLimit: 100,
		ExpiresAt: &exp,
	})
	if err != nil {
		return "", err
	}
	return resp.Key, nil
}

// GetTestAuthToken generates a test JWT token
func GetTestAuthToken(t *testing.T, userID uuid.UUID, email string, secret string) (string, error) {
	t.Helper()
	if secret == "" {
		secret = "test-secret"
	}
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// CreateTestSession creates a test session
func CreateTestSession(t *testing.T, queries *db.Queries, userID uuid.UUID) (string, error) {
	t.Helper()

	ctx := context.Background()

	accessBytes := make([]byte, 32)
	if _, err := rand.Read(accessBytes); err != nil {
		return "", err
	}
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return "", err
	}
	accessToken := hex.EncodeToString(accessBytes)
	refreshToken := hex.EncodeToString(refreshBytes)

	session := &db.UserSession{
		ID:           uuid.Nil,
		UserID:       userID,
		SessionToken: accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Time{},
	}
	if err := queries.CreateUserSession(ctx, session); err != nil {
		return "", err
	}
	return accessToken, nil
}

// MockAuthContext returns a mock authentication context for testing
func MockAuthContext(userID uuid.UUID, email string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "user_email", email)
	return ctx
}
