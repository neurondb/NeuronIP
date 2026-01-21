-- Migration: SSO (SAML, OAuth, OIDC) Support
-- Description: Adds tables for SSO authentication providers and configurations

-- SSO Providers: Supported identity providers
CREATE TABLE IF NOT EXISTS neuronip.sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    provider_type TEXT NOT NULL CHECK (provider_type IN ('saml', 'oauth2', 'oidc')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    configuration JSONB NOT NULL DEFAULT '{}',
    metadata_url TEXT,
    entity_id TEXT,
    sso_url TEXT,
    slo_url TEXT,
    certificate TEXT,
    client_id TEXT,
    client_secret TEXT,
    scopes TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.sso_providers IS 'SSO identity provider configurations';

CREATE INDEX IF NOT EXISTS idx_sso_providers_type ON neuronip.sso_providers(provider_type);
CREATE INDEX IF NOT EXISTS idx_sso_providers_enabled ON neuronip.sso_providers(enabled);

-- SSO User Mappings: Maps SSO users to NeuronIP users
CREATE TABLE IF NOT EXISTS neuronip.sso_user_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES neuronip.sso_providers(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    external_id TEXT NOT NULL,
    email TEXT,
    attributes JSONB DEFAULT '{}',
    first_login_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, external_id)
);
COMMENT ON TABLE neuronip.sso_user_mappings IS 'Maps SSO external users to NeuronIP users';

CREATE INDEX IF NOT EXISTS idx_sso_user_mappings_provider ON neuronip.sso_user_mappings(provider_id);
CREATE INDEX IF NOT EXISTS idx_sso_user_mappings_user ON neuronip.sso_user_mappings(user_id);
CREATE INDEX IF NOT EXISTS idx_sso_user_mappings_external ON neuronip.sso_user_mappings(external_id);

-- SSO Sessions: Active SSO sessions
CREATE TABLE IF NOT EXISTS neuronip.sso_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES neuronip.sso_providers(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    id_token TEXT,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.sso_sessions IS 'Active SSO authentication sessions';

CREATE INDEX IF NOT EXISTS idx_sso_sessions_provider ON neuronip.sso_sessions(provider_id);
CREATE INDEX IF NOT EXISTS idx_sso_sessions_user ON neuronip.sso_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sso_sessions_token ON neuronip.sso_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sso_sessions_expires ON neuronip.sso_sessions(expires_at);

-- SSO Audit Log: SSO authentication events
CREATE TABLE IF NOT EXISTS neuronip.sso_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID REFERENCES neuronip.sso_providers(id) ON DELETE SET NULL,
    user_id TEXT,
    external_id TEXT,
    event_type TEXT NOT NULL CHECK (event_type IN ('login', 'logout', 'token_refresh', 'error', 'mapping_created')),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.sso_audit_log IS 'SSO authentication audit log';

CREATE INDEX IF NOT EXISTS idx_sso_audit_provider ON neuronip.sso_audit_log(provider_id);
CREATE INDEX IF NOT EXISTS idx_sso_audit_user ON neuronip.sso_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_sso_audit_event ON neuronip.sso_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_sso_audit_created ON neuronip.sso_audit_log(created_at);
