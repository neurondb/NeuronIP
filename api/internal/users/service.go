package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

/* Service provides user management functionality */
type Service struct {
	queries *db.Queries
}

/* NewService creates a new user service */
func NewService(queries *db.Queries) *Service {
	return &Service{
		queries: queries,
	}
}

/* GetUser retrieves a user by ID */
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*db.User, error) {
	return s.queries.GetUserByID(ctx, userID)
}

/* UpdateUser updates user information */
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, name *string, avatarURL *string, emailVerified *bool) error {
	return s.queries.UpdateUser(ctx, userID, name, avatarURL, emailVerified)
}

/* GetUserProfile retrieves a user profile */
func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*db.UserProfile, error) {
	profile, err := s.queries.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		// Create empty profile
		profile = &db.UserProfile{
			UserID:   userID,
			Metadata: make(map[string]interface{}),
		}
		if err := s.queries.CreateUserProfile(ctx, profile); err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	}
	return profile, nil
}

/* UpdateUserProfile updates user profile information */
func (s *Service) UpdateUserProfile(ctx context.Context, userID uuid.UUID, profile *db.UserProfile) error {
	profile.UserID = userID
	return s.queries.CreateUserProfile(ctx, profile)
}

/* ChangePassword changes a user's password */
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.PasswordHash == nil {
		return fmt.Errorf("password not set for this account")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("incorrect password")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	passwordHashStr := string(passwordHash)
	return s.queries.UpdateUserPassword(ctx, userID, passwordHashStr)
}

/* GetUserPreferences retrieves user preferences */
func (s *Service) GetUserPreferences(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Preferences == nil {
		return make(map[string]interface{}), nil
	}
	return user.Preferences, nil
}

/* UpdateUserPreferences updates user preferences */
func (s *Service) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, preferences map[string]interface{}) error {
	query := `UPDATE neuronip.users SET preferences = $1 WHERE id = $2`
	_, err := s.queries.DB.Exec(ctx, query, preferences, userID)
	if err != nil {
		return fmt.Errorf("failed to update preferences: %w", err)
	}
	return nil
}
