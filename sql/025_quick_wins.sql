-- Migration: Quick Wins (Comments, Ownership, Webhooks)
-- Description: Adds tables for comments, ownership, and webhooks

-- Comments: Comments and annotations on catalog resources
CREATE TABLE IF NOT EXISTS neuronip.comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('table', 'column', 'schema', 'connector', 'document', 'workflow', 'policy')),
    resource_id UUID NOT NULL,
    user_id TEXT NOT NULL,
    comment_text TEXT NOT NULL,
    parent_comment_id UUID REFERENCES neuronip.comments(id) ON DELETE CASCADE,
    is_resolved BOOLEAN NOT NULL DEFAULT false,
    resolved_by TEXT,
    resolved_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.comments IS 'Comments and annotations on catalog resources';

CREATE INDEX IF NOT EXISTS idx_comments_resource ON neuronip.comments(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON neuronip.comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON neuronip.comments(parent_comment_id);
CREATE INDEX IF NOT EXISTS idx_comments_created ON neuronip.comments(created_at);

-- Resource Ownership: Ownership assignment for catalog resources
CREATE TABLE IF NOT EXISTS neuronip.resource_ownership (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('table', 'column', 'schema', 'connector', 'document', 'workflow', 'policy')),
    resource_id UUID NOT NULL,
    owner_id TEXT NOT NULL,
    owner_type TEXT NOT NULL CHECK (owner_type IN ('user', 'team', 'organization')),
    assigned_by TEXT,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    UNIQUE(resource_type, resource_id)
);
COMMENT ON TABLE neuronip.resource_ownership IS 'Ownership assignment for catalog resources';

CREATE INDEX IF NOT EXISTS idx_ownership_resource ON neuronip.resource_ownership(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_ownership_owner ON neuronip.resource_ownership(owner_id, owner_type);

-- Webhooks: Webhook configurations for event notifications
CREATE TABLE IF NOT EXISTS neuronip.webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    secret TEXT,
    headers JSONB DEFAULT '{}',
    retry_count INTEGER DEFAULT 3,
    timeout_seconds INTEGER DEFAULT 30,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.webhooks IS 'Webhook configurations for event notifications';

CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON neuronip.webhooks(enabled);
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON neuronip.webhooks USING GIN(events);

-- Webhook Deliveries: Webhook delivery history
CREATE TABLE IF NOT EXISTS neuronip.webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES neuronip.webhooks(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'success', 'failed', 'retrying')),
    status_code INTEGER,
    response_body TEXT,
    error_message TEXT,
    attempt_number INTEGER DEFAULT 1,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.webhook_deliveries IS 'Webhook delivery history';

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON neuronip.webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON neuronip.webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created ON neuronip.webhook_deliveries(created_at);
