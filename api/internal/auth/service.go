package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"golang.org/x/crypto/bcrypt"
)

/* AuthService provides authentication services */
type AuthService struct {
	queries   *db.Queries
	jwtSecret string
}

/* NewAuthService creates a new authentication service */
func NewAuthService(queries *db.Queries, jwtSecret string) *AuthService {
	return &AuthService{
		queries:   queries,
		jwtSecret: jwtSecret,
	}
}

/* RegisterRequest represents a user registration request */
type RegisterRequest struct {
	Email    string
	Password string
	Name     string
}

/* LoginRequest represents a user login request */
type LoginRequest struct {
	Email    string
	Password string
	TOTPCode string // Optional 2FA code
}

/* AuthResponse represents an authentication response */
type AuthResponse struct {
	User         *db.User
	SessionToken string
	RefreshToken string
	ExpiresAt    time.Time
}

/* RegisterUser registers a new user with email and password */
func (s *AuthService) RegisterUser(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	_, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	passwordHashStr := string(passwordHash)

	// Create user
	namePtr := &req.Name
	if req.Name == "" {
		namePtr = nil
	}
	user, err := s.queries.CreateUser(ctx, req.Email, &passwordHashStr, namePtr, "analyst")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default user profile
	profile := &db.UserProfile{
		UserID:   user.ID,
		Metadata: make(map[string]interface{}),
	}
	if err := s.queries.CreateUserProfile(ctx, profile); err != nil {
		// Log error but don't fail registration
	}

	// Create session
	sessionToken, refreshToken, expiresAt, err := s.generateSessionTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session tokens: %w", err)
	}

	session := &db.UserSession{
		UserID:       user.ID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	if err := s.queries.CreateUserSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		User:         user,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

/* LoginUser authenticates a user with email and password */
func (s *AuthService) LoginUser(ctx context.Context, req LoginRequest, ipAddress, userAgent string) (*AuthResponse, error) {
	// Get user by email
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if user.PasswordHash == nil {
		return nil, fmt.Errorf("account does not have a password set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check 2FA if enabled
	if user.TwoFactorEnabled {
		if req.TOTPCode == "" {
			return nil, fmt.Errorf("2FA code required")
		}
		// 2FA verification will be handled by twofactor service
		// For now, we'll skip it here and let the handler handle it
	}

	// Update last login
	if err := s.queries.UpdateUserLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
	}

	// Create session
	sessionToken, refreshToken, expiresAt, err := s.generateSessionTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session tokens: %w", err)
	}

	ipPtr := &ipAddress
	if ipAddress == "" {
		ipPtr = nil
	}
	uaPtr := &userAgent
	if userAgent == "" {
		uaPtr = nil
	}

	session := &db.UserSession{
		UserID:       user.ID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		IPAddress:    ipPtr,
		UserAgent:    uaPtr,
		ExpiresAt:    expiresAt,
	}
	if err := s.queries.CreateUserSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		User:         user,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

/* LogoutUser invalidates a user session */
func (s *AuthService) LogoutUser(ctx context.Context, sessionID uuid.UUID) error {
	return s.queries.DeleteUserSession(ctx, sessionID)
}

/* RefreshToken refreshes a session token using a refresh token */
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Get session by refresh token
	session, err := s.queries.GetUserSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user
	user, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Delete old session
	if err := s.queries.DeleteUserSession(ctx, session.ID); err != nil {
		// Log error but continue
	}

	// Create new session
	newSessionToken, newRefreshToken, expiresAt, err := s.generateSessionTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session tokens: %w", err)
	}

	newSession := &db.UserSession{
		UserID:       user.ID,
		SessionToken: newSessionToken,
		RefreshToken: newRefreshToken,
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		ExpiresAt:    expiresAt,
	}
	if err := s.queries.CreateUserSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		User:         user,
		SessionToken: newSessionToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

/* ValidateSession validates a session token and returns the user */
func (s *AuthService) ValidateSession(ctx context.Context, sessionToken string) (*db.User, error) {
	session, err := s.queries.GetUserSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	user, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

/* generateSessionTokens generates a new session token and refresh token */
func (s *AuthService) generateSessionTokens() (sessionToken, refreshToken string, expiresAt time.Time, err error) {
	// Generate secure random tokens
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate session token: %w", err)
	}
	sessionToken = base64.URLEncoding.EncodeToString(sessionBytes)

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshToken = hex.EncodeToString(refreshBytes)

	// Session expires in 24 hours, refresh token expires in 7 days
	expiresAt = time.Now().Add(24 * time.Hour)

	return sessionToken, refreshToken, expiresAt, nil
}
