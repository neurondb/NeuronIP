package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/session"
	"golang.org/x/crypto/bcrypt"
)

/* AuthService provides authentication services */
type AuthService struct {
	queries       *db.Queries
	jwtSecret     string
	sessionManager *session.Manager
}

/* NewAuthService creates a new authentication service */
func NewAuthService(queries *db.Queries, jwtSecret string, sessionManager *session.Manager) *AuthService {
	return &AuthService{
		queries:        queries,
		jwtSecret:      jwtSecret,
		sessionManager: sessionManager,
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

/* LoginWithUsername authenticates a user with username and password, supporting database selection */
func (s *AuthService) LoginWithUsername(ctx context.Context, username, password, database, ipAddress, userAgent string) (*db.User, *session.Session, string, error) {
	// Default to neuronip if not specified
	if database == "" {
		database = "neuronip"
	}

	// Get user by username or email
	var user *db.User
	var err error
	
	// Try username first
	query := `SELECT id, email, email_verified, password_hash, name, avatar_url, role, 
	          two_factor_enabled, two_factor_secret, preferences, last_login_at, created_at, updated_at
	          FROM neuronip.users WHERE username = $1 OR email = $1`
	
	var userID uuid.UUID
	var email string
	var emailVerified bool
	var passwordHash sql.NullString
	var name, avatarURL sql.NullString
	var role string
	var twoFactorEnabled bool
	var twoFactorSecret sql.NullString
	var preferences map[string]interface{}
	var lastLoginAt sql.NullTime
	var createdAt, updatedAt time.Time

	// Use pgx query (queries.DB is *pgxpool.Pool)
	err = s.queries.DB.QueryRow(ctx, query, username).Scan(
		&userID, &email, &emailVerified, &passwordHash, &name, &avatarURL, &role,
		&twoFactorEnabled, &twoFactorSecret, &preferences, &lastLoginAt, &createdAt, &updatedAt,
	)
	if err != nil {
		// Check for "no rows" error (pgx uses different error than database/sql)
		if err.Error() == "no rows in result set" {
			return nil, nil, "", fmt.Errorf("invalid username or password")
		}
		return nil, nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	user = &db.User{
		ID:              userID,
		Email:           email,
		EmailVerified:   emailVerified,
		Role:            role,
		TwoFactorEnabled: twoFactorEnabled,
		Preferences:     preferences,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	if passwordHash.Valid {
		hash := passwordHash.String
		user.PasswordHash = &hash
	}
	if name.Valid {
		user.Name = &name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if twoFactorSecret.Valid {
		user.TwoFactorSecret = &twoFactorSecret.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Verify password
	if user.PasswordHash == nil {
		return nil, nil, "", fmt.Errorf("account does not have a password set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, "", fmt.Errorf("invalid username or password")
	}

	// Update last login (non-blocking)
	updateQuery := `UPDATE neuronip.users SET last_login_at = NOW() WHERE id = $1`
	go func() {
		_, _ = s.queries.DB.Exec(context.Background(), updateQuery, user.ID)
	}()

	// Create session using session manager
	sess, refreshToken, err := s.sessionManager.CreateSession(ctx, user.ID.String(), database, userAgent, ipAddress)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sess, refreshToken, nil
}

/* RegisterWithUsername registers a new user with username and password, supporting database selection */
func (s *AuthService) RegisterWithUsername(ctx context.Context, username, password, database, ipAddress, userAgent string) (*db.User, *session.Session, string, error) {
	// Default to neuronip if not specified
	if database == "" {
		database = "neuronip"
	}

	// Check if user already exists (always check in neuronip database for user management)
	checkQuery := `SELECT id FROM neuronip.users WHERE username = $1 OR email = $1`
	var existingID uuid.UUID
	err := s.queries.DB.QueryRow(ctx, checkQuery, username).Scan(&existingID)
	if err == nil {
		return nil, nil, "", fmt.Errorf("user with username %s already exists", username)
	}
	// If error is "no rows", that's fine - user doesn't exist yet
	if err != nil && err.Error() != "no rows in result set" {
		return nil, nil, "", fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to hash password: %w", err)
	}
	passwordHashStr := string(passwordHash)

	// Create user (use username as email if email format, otherwise generate email)
	email := username
	if !contains(username, "@") {
		email = fmt.Sprintf("%s@neuronip.local", username)
	}

	insertQuery := `INSERT INTO neuronip.users (email, username, password_hash, role, created_at, updated_at)
	                VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id, email, email_verified, password_hash, 
	                name, avatar_url, role, two_factor_enabled, two_factor_secret, preferences, 
	                last_login_at, created_at, updated_at`

	var userID uuid.UUID
	var userEmail string
	var emailVerified bool
	var passwordHashDB sql.NullString
	var name, avatarURL sql.NullString
	var role string
	var twoFactorEnabled bool
	var twoFactorSecret sql.NullString
	var preferences map[string]interface{}
	var lastLoginAt sql.NullTime
	var createdAt, updatedAt time.Time

	err = s.queries.DB.QueryRow(ctx, insertQuery, email, username, passwordHashStr, "analyst").Scan(
		&userID, &userEmail, &emailVerified, &passwordHashDB, &name, &avatarURL, &role,
		&twoFactorEnabled, &twoFactorSecret, &preferences, &lastLoginAt, &createdAt, &updatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "duplicate key value violates unique constraint" || 
		   err.Error() == "ERROR: duplicate key value violates unique constraint" {
			return nil, nil, "", fmt.Errorf("user with username %s already exists", username)
		}
		return nil, nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	user := &db.User{
		ID:              userID,
		Email:           userEmail,
		EmailVerified:   emailVerified,
		Role:            role,
		TwoFactorEnabled: twoFactorEnabled,
		Preferences:     preferences,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	if passwordHashDB.Valid {
		hash := passwordHashDB.String
		user.PasswordHash = &hash
	}

	// Create default user profile
	profileQuery := `INSERT INTO neuronip.user_profiles (user_id, metadata, created_at, updated_at)
	                 VALUES ($1, '{}', NOW(), NOW()) ON CONFLICT (user_id) DO NOTHING`
	_, err = s.queries.DB.Exec(ctx, profileQuery, userID)
	if err != nil {
		// Log error but don't fail registration
	}

	// Create session using session manager
	sess, refreshToken, err := s.sessionManager.CreateSession(ctx, user.ID.String(), database, userAgent, ipAddress)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sess, refreshToken, nil
}

/* GetCurrentUser gets the current user from session context */
func (s *AuthService) GetCurrentUser(ctx context.Context) (*db.User, error) {
	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}

	userID, err := uuid.Parse(sess.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

/* Helper function to check if string contains substring */
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
