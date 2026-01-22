-- Migration: Resource Quotas
-- Description: Adds resource quota management for workspaces and users

-- Resource quotas table
CREATE TABLE IF NOT EXISTS neuronip.resource_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id TEXT,
    resource_type TEXT NOT NULL, -- query, agent, storage, etc.
    limit_value BIGINT NOT NULL,
    current_usage BIGINT NOT NULL DEFAULT 0,
    period TEXT NOT NULL CHECK (period IN ('daily', 'weekly', 'monthly')),
    reset_at TIMESTAMPTZ NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(COALESCE(workspace_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(user_id, ''), resource_type, period)
);
COMMENT ON TABLE neuronip.resource_quotas IS 'Resource quotas for workspaces and users';

CREATE INDEX IF NOT EXISTS idx_resource_quotas_workspace ON neuronip.resource_quotas(workspace_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_user ON neuronip.resource_quotas(user_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_type ON neuronip.resource_quotas(resource_type);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_reset ON neuronip.resource_quotas(reset_at);

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_resource_quotas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_resource_quotas_updated_at ON neuronip.resource_quotas;
CREATE TRIGGER trigger_update_resource_quotas_updated_at
    BEFORE UPDATE ON neuronip.resource_quotas
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_resource_quotas_updated_at();
