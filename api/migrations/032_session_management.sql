/*-------------------------------------------------------------------------
 *
 * 032_session_management.sql
 *    Session Management Schema
 *
 * This migration creates tables for session-based authentication with
 * refresh token rotation and database selection support.
 *
 *-------------------------------------------------------------------------
 */

-- ============================================================================
-- SESSIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    database_name TEXT NOT NULL DEFAULT 'neuronip',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    user_agent_hash TEXT,
    ip_hash TEXT
);

COMMENT ON TABLE neuronip.sessions IS 'User sessions for cookie-based authentication';
COMMENT ON COLUMN neuronip.sessions.database_name IS 'Selected database (neuronip or neuronai-demo)';
COMMENT ON COLUMN neuronip.sessions.user_agent_hash IS 'SHA256 hash of user agent for privacy';
COMMENT ON COLUMN neuronip.sessions.ip_hash IS 'SHA256 hash of IP address for privacy';

-- ============================================================================
-- REFRESH TOKENS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES neuronip.sessions(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    rotated_from UUID REFERENCES neuronip.refresh_tokens(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.refresh_tokens IS 'Refresh tokens for session rotation';
COMMENT ON COLUMN neuronip.refresh_tokens.token_hash IS 'SHA256 hash of the refresh token';
COMMENT ON COLUMN neuronip.refresh_tokens.rotated_from IS 'Previous token ID when rotated (for reuse detection)';

-- ============================================================================
-- INDEXES
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON neuronip.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_revoked_at ON neuronip.sessions(revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_last_seen_at ON neuronip.sessions(last_seen_at);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON neuronip.refresh_tokens(session_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON neuronip.refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON neuronip.refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked_at ON neuronip.refresh_tokens(revoked_at) WHERE revoked_at IS NULL;

-- ============================================================================
-- FAILED LOGIN ATTEMPTS TABLE (for rate limiting)
-- ============================================================================

CREATE TABLE IF NOT EXISTS neuronip.failed_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier TEXT NOT NULL, -- username or IP address
    ip_address TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE neuronip.failed_login_attempts IS 'Failed login attempts for rate limiting';

CREATE INDEX IF NOT EXISTS idx_failed_login_attempts_identifier ON neuronip.failed_login_attempts(identifier, attempted_at);
CREATE INDEX IF NOT EXISTS idx_failed_login_attempts_attempted_at ON neuronip.failed_login_attempts(attempted_at);

-- Cleanup old failed attempts (older than 1 hour)
CREATE OR REPLACE FUNCTION neuronip.cleanup_old_failed_attempts()
RETURNS void AS $$
BEGIN
    DELETE FROM neuronip.failed_login_attempts 
    WHERE attempted_at < NOW() - INTERVAL '1 hour';
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- UPDATE USERS TABLE (if username column doesn't exist)
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'neuronip' 
        AND table_name = 'users' 
        AND column_name = 'username'
    ) THEN
        ALTER TABLE neuronip.users ADD COLUMN username TEXT UNIQUE;
        CREATE INDEX IF NOT EXISTS idx_users_username ON neuronip.users(username);
    END IF;
END $$;

COMMENT ON COLUMN neuronip.users.username IS 'Username for login (alternative to email)';
