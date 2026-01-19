/*-------------------------------------------------------------------------
 *
 * 001_users_schema.sql
 *    User Authentication and Profile Management Schema
 *
 * This migration creates all tables needed for user authentication,
 * profiles, sessions, OAuth, notifications, and activity tracking.
 *
 *-------------------------------------------------------------------------
 */

-- ============================================================================
-- USERS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    password_hash TEXT,  -- Nullable for OAuth-only users
    name TEXT,
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'analyst' CHECK (role IN ('admin', 'analyst', 'support', 'developer')),
    two_factor_enabled BOOLEAN NOT NULL DEFAULT false,
    two_factor_secret TEXT,  -- Encrypted TOTP secret
    preferences JSONB NOT NULL DEFAULT '{}',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.users IS 'Core user accounts';
COMMENT ON COLUMN neuronip.users.password_hash IS 'BCrypt hashed password, nullable for OAuth-only users';
COMMENT ON COLUMN neuronip.users.two_factor_secret IS 'Encrypted TOTP secret for 2FA';

-- ============================================================================
-- USER PROFILES TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.user_profiles (
    user_id UUID PRIMARY KEY REFERENCES neuronip.users(id) ON DELETE CASCADE,
    bio TEXT,
    company TEXT,
    job_title TEXT,
    location TEXT,
    website TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.user_profiles IS 'Extended profile information for users';

-- ============================================================================
-- USER SESSIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    session_token TEXT NOT NULL UNIQUE,
    refresh_token TEXT NOT NULL UNIQUE,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.user_sessions IS 'Active user sessions for token-based authentication';

-- ============================================================================
-- OAUTH PROVIDERS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github', 'microsoft')),
    provider_user_id TEXT NOT NULL,
    access_token TEXT,  -- Encrypted
    refresh_token TEXT,  -- Encrypted
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

COMMENT ON TABLE neuronip.oauth_providers IS 'OAuth provider account links';
COMMENT ON COLUMN neuronip.oauth_providers.access_token IS 'Encrypted OAuth access token';
COMMENT ON COLUMN neuronip.oauth_providers.refresh_token IS 'Encrypted OAuth refresh token';

-- ============================================================================
-- USER NOTIFICATIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.user_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.user_notifications IS 'User notifications';

-- ============================================================================
-- EMAIL VERIFICATIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.email_verifications IS 'Email verification tokens';

-- ============================================================================
-- PASSWORD RESETS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.password_resets IS 'Password reset tokens';

-- ============================================================================
-- USER ACTIVITY LOGS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.user_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    activity_type TEXT NOT NULL,
    resource_type TEXT,
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    ip_address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.user_activity_logs IS 'User activity tracking for analytics';

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Users indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON neuronip.users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON neuronip.users(role);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON neuronip.users(created_at DESC);

-- Sessions indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON neuronip.user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON neuronip.user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON neuronip.user_sessions(expires_at);

-- OAuth providers indexes
CREATE INDEX IF NOT EXISTS idx_oauth_providers_user_id ON neuronip.oauth_providers(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_providers_provider ON neuronip.oauth_providers(provider);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON neuronip.user_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_user_notifications_read ON neuronip.user_notifications(read);
CREATE INDEX IF NOT EXISTS idx_user_notifications_created_at ON neuronip.user_notifications(created_at DESC);

-- Email verifications indexes
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON neuronip.email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON neuronip.email_verifications(user_id);

-- Password resets indexes
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON neuronip.password_resets(token);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON neuronip.password_resets(user_id);

-- Activity logs indexes
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_id ON neuronip.user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_activity_type ON neuronip.user_activity_logs(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at ON neuronip.user_activity_logs(created_at DESC);

-- ============================================================================
-- UPDATE API_KEYS TABLE TO LINK TO USERS
-- ============================================================================

-- Update api_keys.user_id to reference users.id if the table exists
-- This is done conditionally as the table may already exist
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = 'neuronip' AND table_name = 'api_keys'
    ) THEN
        -- Add foreign key if it doesn't exist and user_id column can reference UUID
        -- Note: We'll need to handle this carefully since user_id is currently TEXT
        -- We may need to migrate existing data first
        ALTER TABLE neuronip.api_keys 
        DROP CONSTRAINT IF EXISTS fk_api_keys_user_id;
        
        -- We'll handle the migration of user_id from TEXT to UUID in a separate step
        -- For now, we just ensure the constraint can be added when ready
    END IF;
END $$;

-- ============================================================================
-- UPDATE USER_ROLES TABLE IF IT EXISTS
-- ============================================================================

-- Create user_roles table if it doesn't exist (used by RBAC)
CREATE TABLE IF NOT EXISTS neuronip.user_roles (
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL CHECK (role_name IN ('admin', 'analyst', 'support', 'developer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

COMMENT ON TABLE neuronip.user_roles IS 'User role assignments for RBAC';

CREATE INDEX IF NOT EXISTS idx_user_roles_role_name ON neuronip.user_roles(role_name);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update updated_at timestamp for users
CREATE OR REPLACE FUNCTION neuronip.update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON neuronip.users;
CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON neuronip.users
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_users_updated_at();

-- Update updated_at timestamp for user_profiles
CREATE OR REPLACE FUNCTION neuronip.update_user_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_user_profiles_updated_at ON neuronip.user_profiles;
CREATE TRIGGER trigger_update_user_profiles_updated_at
    BEFORE UPDATE ON neuronip.user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_user_profiles_updated_at();

-- Update updated_at timestamp for oauth_providers
CREATE OR REPLACE FUNCTION neuronip.update_oauth_providers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_oauth_providers_updated_at ON neuronip.oauth_providers;
CREATE TRIGGER trigger_update_oauth_providers_updated_at
    BEFORE UPDATE ON neuronip.oauth_providers
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_oauth_providers_updated_at();
