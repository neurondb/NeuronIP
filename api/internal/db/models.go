package db

import (
	"time"

	"github.com/google/uuid"
)

/* KnowledgeCollection represents a knowledge collection */
type KnowledgeCollection struct {
	ID          uuid.UUID              `db:"id" json:"id"`
	Name        string                 `db:"name" json:"name"`
	Description *string                `db:"description" json:"description,omitempty"`
	CreatedBy   *string                `db:"created_by" json:"created_by,omitempty"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at" json:"updated_at"`
	Metadata    map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
}

/* KnowledgeDocument represents a knowledge document */
type KnowledgeDocument struct {
	ID          uuid.UUID              `db:"id" json:"id"`
	CollectionID *uuid.UUID            `db:"collection_id" json:"collection_id,omitempty"`
	Title       string                 `db:"title" json:"title"`
	Content     string                 `db:"content" json:"content"`
	ContentType string                 `db:"content_type" json:"content_type"`
	Source      *string                `db:"source" json:"source,omitempty"`
	SourceURL   *string                `db:"source_url" json:"source_url,omitempty"`
	Metadata    map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at" json:"updated_at"`
}

/* WarehouseQuery represents a warehouse query */
type WarehouseQuery struct {
	ID                  uuid.UUID              `db:"id" json:"id"`
	UserID              *string                `db:"user_id" json:"user_id,omitempty"`
	NaturalLanguageQuery string                 `db:"natural_language_query" json:"natural_language_query"`
	GeneratedSQL        *string                `db:"generated_sql" json:"generated_sql,omitempty"`
	SchemaID            *uuid.UUID             `db:"schema_id" json:"schema_id,omitempty"`
	Status              string                 `db:"status" json:"status"`
	ErrorMessage        *string                `db:"error_message" json:"error_message,omitempty"`
	CreatedAt           time.Time              `db:"created_at" json:"created_at"`
	ExecutedAt          *time.Time             `db:"executed_at" json:"executed_at,omitempty"`
}

/* SupportTicket represents a support ticket */
type SupportTicket struct {
	ID            uuid.UUID              `db:"id" json:"id"`
	TicketNumber  string                  `db:"ticket_number" json:"ticket_number"`
	CustomerID    string                  `db:"customer_id" json:"customer_id"`
	CustomerEmail *string                 `db:"customer_email" json:"customer_email,omitempty"`
	Subject       string                  `db:"subject" json:"subject"`
	Status        string                  `db:"status" json:"status"`
	Priority      string                  `db:"priority" json:"priority"`
	AssignedAgentID *uuid.UUID            `db:"assigned_agent_id" json:"assigned_agent_id,omitempty"`
	Metadata      map[string]interface{}  `db:"metadata" json:"metadata,omitempty"`
	CreatedAt     time.Time               `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time               `db:"updated_at" json:"updated_at"`
	ResolvedAt    *time.Time              `db:"resolved_at" json:"resolved_at,omitempty"`
}

/* CompliancePolicy represents a compliance policy */
type CompliancePolicy struct {
	ID          uuid.UUID              `db:"id" json:"id"`
	PolicyName  string                 `db:"policy_name" json:"policy_name"`
	PolicyType  string                  `db:"policy_type" json:"policy_type"`
	Description *string                 `db:"description" json:"description,omitempty"`
	PolicyText  string                  `db:"policy_text" json:"policy_text"`
	Rules       []interface{}           `db:"rules" json:"rules,omitempty"`
	Enabled     bool                    `db:"enabled" json:"enabled"`
	CreatedAt   time.Time               `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time               `db:"updated_at" json:"updated_at"`
}

/* Workflow represents a workflow definition */
type Workflow struct {
	ID               uuid.UUID              `db:"id" json:"id"`
	Name             string                 `db:"name" json:"name"`
	Description      *string                `db:"description" json:"description,omitempty"`
	WorkflowDefinition map[string]interface{} `db:"workflow_definition" json:"workflow_definition"`
	AgentID          *uuid.UUID            `db:"agent_id" json:"agent_id,omitempty"`
	Enabled          bool                  `db:"enabled" json:"enabled"`
	CreatedBy        *string               `db:"created_by" json:"created_by,omitempty"`
	CreatedAt        time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time             `db:"updated_at" json:"updated_at"`
}

/* APIKey represents an API key */
type APIKey struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	KeyHash    string     `db:"key_hash" json:"-"`
	KeyPrefix  string     `db:"key_prefix" json:"key_prefix"`
	UserID     *string    `db:"user_id" json:"user_id,omitempty"`
	Name       *string    `db:"name" json:"name,omitempty"`
	RateLimit  int        `db:"rate_limit" json:"rate_limit"`
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
}

/* User represents a user account */
type User struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	Email           string                 `db:"email" json:"email"`
	EmailVerified   bool                   `db:"email_verified" json:"email_verified"`
	PasswordHash    *string                `db:"password_hash" json:"-"`
	Name            *string                `db:"name" json:"name,omitempty"`
	AvatarURL       *string                `db:"avatar_url" json:"avatar_url,omitempty"`
	Role            string                 `db:"role" json:"role"`
	TwoFactorEnabled bool                  `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret *string                `db:"two_factor_secret" json:"-"`
	Preferences     map[string]interface{} `db:"preferences" json:"preferences,omitempty"`
	LastLoginAt     *time.Time             `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

/* UserProfile represents extended user profile information */
type UserProfile struct {
	UserID    uuid.UUID              `db:"user_id" json:"user_id"`
	Bio       *string                `db:"bio" json:"bio,omitempty"`
	Company   *string                `db:"company" json:"company,omitempty"`
	JobTitle  *string                `db:"job_title" json:"job_title,omitempty"`
	Location  *string                `db:"location" json:"location,omitempty"`
	Website   *string                `db:"website" json:"website,omitempty"`
	Metadata  map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt time.Time              `db:"updated_at" json:"updated_at"`
}

/* UserSession represents a user session */
type UserSession struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	SessionToken string     `db:"session_token" json:"-"`
	RefreshToken string     `db:"refresh_token" json:"-"`
	IPAddress    *string    `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent    *string    `db:"user_agent" json:"user_agent,omitempty"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

/* OAuthProvider represents an OAuth provider link */
type OAuthProvider struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	Provider     string     `db:"provider" json:"provider"`
	ProviderUserID string   `db:"provider_user_id" json:"provider_user_id"`
	AccessToken  *string    `db:"access_token" json:"-"`
	RefreshToken *string    `db:"refresh_token" json:"-"`
	ExpiresAt    *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

/* UserNotification represents a user notification */
type UserNotification struct {
	ID        uuid.UUID              `db:"id" json:"id"`
	UserID    uuid.UUID              `db:"user_id" json:"user_id"`
	Type      string                 `db:"type" json:"type"`
	Title     string                 `db:"title" json:"title"`
	Message   string                 `db:"message" json:"message"`
	Read      bool                   `db:"read" json:"read"`
	Metadata  map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
}

/* EmailVerification represents an email verification token */
type EmailVerification struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

/* PasswordReset represents a password reset token */
type PasswordReset struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

/* UserActivityLog represents a user activity log entry */
type UserActivityLog struct {
	ID           uuid.UUID              `db:"id" json:"id"`
	UserID       uuid.UUID              `db:"user_id" json:"user_id"`
	ActivityType string                 `db:"activity_type" json:"activity_type"`
	ResourceType *string                `db:"resource_type" json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID             `db:"resource_id" json:"resource_id,omitempty"`
	Metadata     map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	IPAddress    *string                `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt    time.Time              `db:"created_at" json:"created_at"`
}
