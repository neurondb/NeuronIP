-- Migration: 048_integration_ecosystem.sql
-- Description: Large SaaS integration ecosystem

-- Integration registry
CREATE TABLE IF NOT EXISTS neuronip.integration_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_name VARCHAR(255) NOT NULL UNIQUE,
    integration_type VARCHAR(100) NOT NULL, -- 'saas', 'api', 'webhook', 'database'
    category VARCHAR(100), -- 'crm', 'support', 'analytics', 'storage', etc.
    description TEXT,
    icon_url TEXT,
    documentation_url TEXT,
    auth_type VARCHAR(50) NOT NULL, -- 'oauth2', 'api_key', 'basic', 'custom'
    auth_config_schema JSONB, -- Schema for auth configuration
    capabilities JSONB, -- Available capabilities
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'beta', 'deprecated'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_registry_type ON neuronip.integration_registry(integration_type);
CREATE INDEX idx_integration_registry_category ON neuronip.integration_registry(category);
CREATE INDEX idx_integration_registry_status ON neuronip.integration_registry(status);

-- Integration instances (user-configured integrations)
CREATE TABLE IF NOT EXISTS neuronip.integration_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES neuronip.integration_registry(id),
    workspace_id UUID,
    user_id VARCHAR(255),
    instance_name VARCHAR(255) NOT NULL,
    auth_config JSONB NOT NULL, -- Encrypted auth credentials
    config JSONB, -- Instance-specific configuration
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'inactive', 'error'
    last_sync_at TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(50) DEFAULT 'unknown', -- 'healthy', 'degraded', 'unhealthy'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_instances_integration ON neuronip.integration_instances(integration_id);
CREATE INDEX idx_integration_instances_workspace ON neuronip.integration_instances(workspace_id);
CREATE INDEX idx_integration_instances_status ON neuronip.integration_instances(status);

-- Webhook subscriptions
CREATE TABLE IF NOT EXISTS neuronip.webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_instance_id UUID NOT NULL REFERENCES neuronip.integration_instances(id),
    webhook_url TEXT NOT NULL,
    events JSONB NOT NULL, -- Array of event types to subscribe to
    secret_token VARCHAR(255), -- For webhook verification
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_delivery_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_subscriptions_instance ON neuronip.webhook_subscriptions(integration_instance_id);
CREATE INDEX idx_webhook_subscriptions_status ON neuronip.webhook_subscriptions(status);

-- Integration templates
CREATE TABLE IF NOT EXISTS neuronip.integration_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_name VARCHAR(255) NOT NULL,
    integration_id UUID NOT NULL REFERENCES neuronip.integration_registry(id),
    template_config JSONB NOT NULL, -- Pre-configured setup
    use_cases JSONB, -- Common use cases
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_templates_integration ON neuronip.integration_templates(integration_id);

COMMENT ON TABLE neuronip.integration_registry IS 'Registry of available integrations';
COMMENT ON TABLE neuronip.integration_instances IS 'User-configured integration instances';
COMMENT ON TABLE neuronip.webhook_subscriptions IS 'Webhook subscriptions for integrations';
COMMENT ON TABLE neuronip.integration_templates IS 'Pre-configured integration templates';
