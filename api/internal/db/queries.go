package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* Queries provides database query methods */
type Queries struct {
	DB *pgxpool.Pool
}

/* NewQueries creates a new Queries instance */
func NewQueries(db *pgxpool.Pool) *Queries {
	return &Queries{DB: db}
}

/* GetKnowledgeCollectionByID retrieves a knowledge collection by ID */
func (q *Queries) GetKnowledgeCollectionByID(ctx context.Context, id uuid.UUID) (*KnowledgeCollection, error) {
	var collection KnowledgeCollection
	query := `SELECT id, name, description, created_by, created_at, updated_at, metadata 
	          FROM neuronip.knowledge_collections WHERE id = $1`
	
	err := q.DB.QueryRow(ctx, query, id).Scan(
		&collection.ID, &collection.Name, &collection.Description,
		&collection.CreatedBy, &collection.CreatedAt, &collection.UpdatedAt, &collection.Metadata,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("knowledge collection not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge collection: %w", err)
	}
	return &collection, nil
}

/* ListKnowledgeCollections retrieves all knowledge collections */
func (q *Queries) ListKnowledgeCollections(ctx context.Context) ([]KnowledgeCollection, error) {
	query := `SELECT id, name, description, created_by, created_at, updated_at, metadata 
	          FROM neuronip.knowledge_collections ORDER BY created_at DESC`
	
	rows, err := q.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list knowledge collections: %w", err)
	}
	defer rows.Close()

	var collections []KnowledgeCollection
	for rows.Next() {
		var collection KnowledgeCollection
		err := rows.Scan(
			&collection.ID, &collection.Name, &collection.Description,
			&collection.CreatedBy, &collection.CreatedAt, &collection.UpdatedAt, &collection.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan knowledge collection: %w", err)
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

/* GetKnowledgeDocumentByID retrieves a knowledge document by ID */
func (q *Queries) GetKnowledgeDocumentByID(ctx context.Context, id uuid.UUID) (*KnowledgeDocument, error) {
	var doc KnowledgeDocument
	query := `SELECT id, collection_id, title, content, content_type, source, source_url, 
	          metadata, created_at, updated_at 
	          FROM neuronip.knowledge_documents WHERE id = $1`
	
	err := q.DB.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.CollectionID, &doc.Title, &doc.Content, &doc.ContentType,
		&doc.Source, &doc.SourceURL, &doc.Metadata, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("knowledge document not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge document: %w", err)
	}
	return &doc, nil
}

/* GetSupportTicketByID retrieves a support ticket by ID */
func (q *Queries) GetSupportTicketByID(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	var ticket SupportTicket
	query := `SELECT id, ticket_number, customer_id, customer_email, subject, status, 
	          priority, assigned_agent_id, metadata, created_at, updated_at, resolved_at 
	          FROM neuronip.support_tickets WHERE id = $1`
	
	err := q.DB.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.TicketNumber, &ticket.CustomerID, &ticket.CustomerEmail,
		&ticket.Subject, &ticket.Status, &ticket.Priority, &ticket.AssignedAgentID,
		&ticket.Metadata, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("support ticket not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get support ticket: %w", err)
	}
	return &ticket, nil
}

/* GetAPIKeyByPrefix retrieves an API key by prefix */
func (q *Queries) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*APIKey, error) {
	var key APIKey
	query := `SELECT id, key_hash, key_prefix, user_id, name, rate_limit, last_used_at, expires_at, created_at 
	          FROM neuronip.api_keys WHERE key_prefix = $1`
	
	err := q.DB.QueryRow(ctx, query, prefix).Scan(
		&key.ID, &key.KeyHash, &key.KeyPrefix, &key.UserID, &key.Name,
		&key.RateLimit, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	return &key, nil
}

/* UpdateAPIKeyLastUsed updates the last used timestamp for an API key */
func (q *Queries) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE neuronip.api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := q.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}
	return nil
}

/* GetUserByID retrieves a user by ID */
func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT id, email, email_verified, password_hash, name, avatar_url, role, 
	          two_factor_enabled, two_factor_secret, preferences, last_login_at, created_at, updated_at 
	          FROM neuronip.users WHERE id = $1`
	
	err := q.DB.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.PasswordHash,
		&user.Name, &user.AvatarURL, &user.Role, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.Preferences, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

/* GetUserByEmail retrieves a user by email */
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `SELECT id, email, email_verified, password_hash, name, avatar_url, role, 
	          two_factor_enabled, two_factor_secret, preferences, last_login_at, created_at, updated_at 
	          FROM neuronip.users WHERE email = $1`
	
	err := q.DB.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.PasswordHash,
		&user.Name, &user.AvatarURL, &user.Role, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.Preferences, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

/* CreateUser creates a new user */
func (q *Queries) CreateUser(ctx context.Context, email string, passwordHash *string, name *string, role string) (*User, error) {
	var user User
	query := `INSERT INTO neuronip.users (email, password_hash, name, role) 
	          VALUES ($1, $2, $3, $4) 
	          RETURNING id, email, email_verified, password_hash, name, avatar_url, role, 
	          two_factor_enabled, two_factor_secret, preferences, last_login_at, created_at, updated_at`
	
	err := q.DB.QueryRow(ctx, query, email, passwordHash, name, role).Scan(
		&user.ID, &user.Email, &user.EmailVerified, &user.PasswordHash,
		&user.Name, &user.AvatarURL, &user.Role, &user.TwoFactorEnabled,
		&user.TwoFactorSecret, &user.Preferences, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

/* UpdateUser updates a user */
func (q *Queries) UpdateUser(ctx context.Context, id uuid.UUID, name *string, avatarURL *string, emailVerified *bool) error {
	query := `UPDATE neuronip.users SET name = COALESCE($1, name), 
	          avatar_url = COALESCE($2, avatar_url), 
	          email_verified = COALESCE($3, email_verified) 
	          WHERE id = $4`
	_, err := q.DB.Exec(ctx, query, name, avatarURL, emailVerified, id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

/* UpdateUserPassword updates a user's password */
func (q *Queries) UpdateUserPassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `UPDATE neuronip.users SET password_hash = $1 WHERE id = $2`
	_, err := q.DB.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

/* UpdateUserLastLogin updates the last login timestamp */
func (q *Queries) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE neuronip.users SET last_login_at = NOW() WHERE id = $1`
	_, err := q.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

/* GetUserProfile retrieves a user profile */
func (q *Queries) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	var profile UserProfile
	query := `SELECT user_id, bio, company, job_title, location, website, metadata, created_at, updated_at 
	          FROM neuronip.user_profiles WHERE user_id = $1`
	
	err := q.DB.QueryRow(ctx, query, userID).Scan(
		&profile.UserID, &profile.Bio, &profile.Company, &profile.JobTitle,
		&profile.Location, &profile.Website, &profile.Metadata,
		&profile.CreatedAt, &profile.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // Profile doesn't exist yet, return nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return &profile, nil
}

/* CreateUserProfile creates a user profile */
func (q *Queries) CreateUserProfile(ctx context.Context, profile *UserProfile) error {
	query := `INSERT INTO neuronip.user_profiles (user_id, bio, company, job_title, location, website, metadata) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) 
	          ON CONFLICT (user_id) DO UPDATE SET 
	          bio = EXCLUDED.bio, company = EXCLUDED.company, job_title = EXCLUDED.job_title, 
	          location = EXCLUDED.location, website = EXCLUDED.website, metadata = EXCLUDED.metadata`
	
	_, err := q.DB.Exec(ctx, query, profile.UserID, profile.Bio, profile.Company,
		profile.JobTitle, profile.Location, profile.Website, profile.Metadata)
	if err != nil {
		return fmt.Errorf("failed to create user profile: %w", err)
	}
	return nil
}

/* GetUserSessionByToken retrieves a session by token */
func (q *Queries) GetUserSessionByToken(ctx context.Context, token string) (*UserSession, error) {
	var session UserSession
	query := `SELECT id, user_id, session_token, refresh_token, ip_address, user_agent, expires_at, created_at 
	          FROM neuronip.user_sessions WHERE session_token = $1 AND expires_at > NOW()`
	
	err := q.DB.QueryRow(ctx, query, token).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

/* GetUserSessionByRefreshToken retrieves a session by refresh token */
func (q *Queries) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*UserSession, error) {
	var session UserSession
	query := `SELECT id, user_id, session_token, refresh_token, ip_address, user_agent, expires_at, created_at 
	          FROM neuronip.user_sessions WHERE refresh_token = $1 AND expires_at > NOW()`
	
	err := q.DB.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

/* CreateUserSession creates a new user session */
func (q *Queries) CreateUserSession(ctx context.Context, session *UserSession) error {
	query := `INSERT INTO neuronip.user_sessions (user_id, session_token, refresh_token, ip_address, user_agent, expires_at) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	
	err := q.DB.QueryRow(ctx, query, session.UserID, session.SessionToken, session.RefreshToken,
		session.IPAddress, session.UserAgent, session.ExpiresAt).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

/* DeleteUserSession deletes a session */
func (q *Queries) DeleteUserSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM neuronip.user_sessions WHERE id = $1`
	_, err := q.DB.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

/* DeleteUserSessionsByUserID deletes all sessions for a user */
func (q *Queries) DeleteUserSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM neuronip.user_sessions WHERE user_id = $1`
	_, err := q.DB.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

/* ListUserSessions lists all active sessions for a user */
func (q *Queries) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]UserSession, error) {
	query := `SELECT id, user_id, session_token, refresh_token, ip_address, user_agent, expires_at, created_at 
	          FROM neuronip.user_sessions WHERE user_id = $1 AND expires_at > NOW() 
	          ORDER BY created_at DESC`
	
	rows, err := q.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
			&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
