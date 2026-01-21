-- Migration: Governance Schema
-- Description: Adds tables for audit logging, policies, and policy enforcement

-- Audit logs: Comprehensive audit trail
CREATE TABLE IF NOT EXISTS neuronip.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    action_type TEXT NOT NULL CHECK (action_type IN ('query', 'agent_execution', 'workflow_execution', 'data_access', 'config_change', 'user_action', 'ai_action')),
    resource_type TEXT,
    resource_id TEXT,
    action TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address TEXT,
    user_agent TEXT,
    status TEXT NOT NULL CHECK (status IN ('success', 'failure', 'error')) DEFAULT 'success',
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.audit_logs IS 'Comprehensive audit trail for all actions';

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON neuronip.audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_type ON neuronip.audit_logs(action_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON neuronip.audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON neuronip.audit_logs(created_at DESC);

-- Policies: Policy definitions
CREATE TABLE IF NOT EXISTS neuronip.policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    policy_type TEXT NOT NULL CHECK (policy_type IN ('data_access', 'query_filter', 'result_filter', 'usage_limit', 'compliance')),
    policy_definition JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 100,
    applies_to TEXT[] DEFAULT '{}', -- Resource types this policy applies to
    conditions JSONB DEFAULT '{}',
    actions JSONB DEFAULT '{}',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.policies IS 'Policy definitions for governance';

CREATE INDEX IF NOT EXISTS idx_policies_type ON neuronip.policies(policy_type);
CREATE INDEX IF NOT EXISTS idx_policies_enabled ON neuronip.policies(enabled);

-- Policy enforcements: Track policy enforcement events
CREATE TABLE IF NOT EXISTS neuronip.policy_enforcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES neuronip.policies(id) ON DELETE CASCADE,
    user_id TEXT,
    resource_type TEXT,
    resource_id TEXT,
    action TEXT NOT NULL,
    enforcement_result TEXT NOT NULL CHECK (enforcement_result IN ('allowed', 'denied', 'filtered', 'modified')),
    enforcement_details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.policy_enforcements IS 'Policy enforcement records';

CREATE INDEX IF NOT EXISTS idx_policy_enforcements_policy ON neuronip.policy_enforcements(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_enforcements_user ON neuronip.policy_enforcements(user_id);
CREATE INDEX IF NOT EXISTS idx_policy_enforcements_created_at ON neuronip.policy_enforcements(created_at DESC);
