-- Migration: Authentication Enhancements
-- Description: Adds tables for OIDC SSO and SCIM 2.0 provisioning

-- OIDC providers: OIDC SSO provider configurations
CREATE TABLE IF NOT EXISTS neuronip.oidc_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name TEXT NOT NULL UNIQUE,
    issuer_url TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret_encrypted TEXT NOT NULL,
    redirect_url TEXT NOT NULL,
    scopes TEXT[] DEFAULT ARRAY['openid', 'profile', 'email'],
    discovery_url TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.oidc_providers IS 'OIDC SSO provider configurations';

CREATE INDEX IF NOT EXISTS idx_oidc_providers_name ON neuronip.oidc_providers(provider_name);
CREATE INDEX IF NOT EXISTS idx_oidc_providers_enabled ON neuronip.oidc_providers(enabled);

-- OIDC sessions: Track OIDC authentication sessions
CREATE TABLE IF NOT EXISTS neuronip.oidc_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    provider_name TEXT NOT NULL,
    state_token TEXT NOT NULL UNIQUE,
    nonce TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.oidc_sessions IS 'OIDC authentication session tracking';

CREATE INDEX IF NOT EXISTS idx_oidc_sessions_user ON neuronip.oidc_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_oidc_sessions_state ON neuronip.oidc_sessions(state_token);
CREATE INDEX IF NOT EXISTS idx_oidc_sessions_expires ON neuronip.oidc_sessions(expires_at);

-- SCIM configuration: SCIM 2.0 provisioning configuration
CREATE TABLE IF NOT EXISTS neuronip.scim_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enabled BOOLEAN NOT NULL DEFAULT false,
    bearer_token_hash TEXT NOT NULL,
    endpoint_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.scim_config IS 'SCIM 2.0 provisioning configuration';

-- SCIM sync log: Track SCIM provisioning operations
CREATE TABLE IF NOT EXISTS neuronip.scim_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation TEXT NOT NULL CHECK (operation IN ('create', 'update', 'delete', 'list')),
    user_id UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    external_id TEXT,
    status TEXT NOT NULL CHECK (status IN ('success', 'error', 'pending')),
    error_message TEXT,
    request_data JSONB,
    response_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.scim_sync_log IS 'SCIM provisioning operation log';

CREATE INDEX IF NOT EXISTS idx_scim_sync_log_user ON neuronip.scim_sync_log(user_id);
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_status ON neuronip.scim_sync_log(status);
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_created ON neuronip.scim_sync_log(created_at DESC);

-- Enhanced user sessions: Add concurrent session limits
ALTER TABLE neuronip.user_sessions ADD COLUMN IF NOT EXISTS device_info TEXT;
ALTER TABLE neuronip.user_sessions ADD COLUMN IF NOT EXISTS location TEXT;
ALTER TABLE neuronip.user_sessions ADD COLUMN IF NOT EXISTS auth_method TEXT CHECK (auth_method IN ('password', 'oidc', 'oauth', 'api_key'));

CREATE INDEX IF NOT EXISTS idx_user_sessions_auth_method ON neuronip.user_sessions(auth_method);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_expires ON neuronip.user_sessions(user_id, expires_at);

-- Session limits: Per-user concurrent session limits
CREATE TABLE IF NOT EXISTS neuronip.session_limits (
    user_id UUID PRIMARY KEY REFERENCES neuronip.users(id) ON DELETE CASCADE,
    max_concurrent_sessions INTEGER NOT NULL DEFAULT 5,
    session_timeout_minutes INTEGER NOT NULL DEFAULT 1440, -- 24 hours
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.session_limits IS 'Per-user session limits and timeouts';

-- Update updated_at trigger for oidc_providers
CREATE OR REPLACE FUNCTION neuronip.update_oidc_providers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_oidc_providers_updated_at ON neuronip.oidc_providers;
CREATE TRIGGER trigger_update_oidc_providers_updated_at
    BEFORE UPDATE ON neuronip.oidc_providers
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_oidc_providers_updated_at();

-- Update updated_at trigger for scim_config
CREATE OR REPLACE FUNCTION neuronip.update_scim_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_scim_config_updated_at ON neuronip.scim_config;
CREATE TRIGGER trigger_update_scim_config_updated_at
    BEFORE UPDATE ON neuronip.scim_config
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_scim_config_updated_at();

-- Cleanup expired OIDC sessions (run periodically)
CREATE OR REPLACE FUNCTION neuronip.cleanup_expired_oidc_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM neuronip.oidc_sessions
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
