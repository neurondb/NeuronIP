-- Migration: Integration Ecosystem Schema
-- Description: Adds tables for CRM automation, ITSM triggers, and BI exports

-- CRM automation hooks: CRM automation hooks
CREATE TABLE IF NOT EXISTS neuronip.crm_automation_hooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crm_type TEXT NOT NULL CHECK (crm_type IN ('salesforce', 'hubspot', 'dynamics')),
    event_type TEXT NOT NULL,
    trigger_config JSONB DEFAULT '{}',
    action_config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.crm_automation_hooks IS 'CRM automation hooks';

CREATE INDEX IF NOT EXISTS idx_crm_hooks_type ON neuronip.crm_automation_hooks(crm_type, event_type);
CREATE INDEX IF NOT EXISTS idx_crm_hooks_enabled ON neuronip.crm_automation_hooks(enabled);

-- ITSM triggers: ITSM trigger configurations
CREATE TABLE IF NOT EXISTS neuronip.itsm_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itsm_type TEXT NOT NULL CHECK (itsm_type IN ('servicenow', 'jira', 'zendesk')),
    trigger_type TEXT NOT NULL,
    config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.itsm_triggers IS 'ITSM trigger configurations';

CREATE INDEX IF NOT EXISTS idx_itsm_triggers_type ON neuronip.itsm_triggers(itsm_type, trigger_type);

-- BI export configs: BI export configurations
CREATE TABLE IF NOT EXISTS neuronip.bi_export_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bi_type TEXT NOT NULL CHECK (bi_type IN ('tableau', 'powerbi', 'looker', 'qlik')),
    query_id UUID NOT NULL REFERENCES neuronip.warehouse_queries(id) ON DELETE CASCADE,
    export_format TEXT NOT NULL CHECK (export_format IN ('csv', 'json', 'xlsx', 'parquet')),
    config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.bi_export_configs IS 'BI export configurations';

CREATE INDEX IF NOT EXISTS idx_bi_exports_type ON neuronip.bi_export_configs(bi_type);
CREATE INDEX IF NOT EXISTS idx_bi_exports_query ON neuronip.bi_export_configs(query_id);

-- UI RLS policies: UI-driven RLS policies
CREATE TABLE IF NOT EXISTS neuronip.ui_rls_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    policy_name TEXT NOT NULL,
    policy_type TEXT NOT NULL CHECK (policy_type IN ('select', 'insert', 'update', 'delete', 'all')),
    condition TEXT NOT NULL,
    condition_builder JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(schema_name, table_name, policy_name)
);
COMMENT ON TABLE neuronip.ui_rls_policies IS 'UI-driven row-level security policies';

CREATE INDEX IF NOT EXISTS idx_ui_rls_table ON neuronip.ui_rls_policies(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_ui_rls_enabled ON neuronip.ui_rls_policies(enabled);

-- Data residency rules: Data residency enforcement rules
CREATE TABLE IF NOT EXISTS neuronip.data_residency_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    required_region TEXT NOT NULL,
    enforcement_level TEXT NOT NULL CHECK (enforcement_level IN ('strict', 'warning', 'audit')) DEFAULT 'strict',
    metadata JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_residency_rules IS 'Data residency enforcement rules';

CREATE INDEX IF NOT EXISTS idx_residency_rules_table ON neuronip.data_residency_rules(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_residency_rules_region ON neuronip.data_residency_rules(required_region);

-- Agent audit trail: Decision audit trail for agents
CREATE TABLE IF NOT EXISTS neuronip.agent_audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL REFERENCES neuronip.agent_traces(id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL,
    decision_type TEXT NOT NULL CHECK (decision_type IN ('tool_selection', 'response_generation', 'reasoning_step')),
    decision JSONB NOT NULL,
    reasoning TEXT,
    alternatives JSONB DEFAULT '[]',
    context JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id TEXT
);
COMMENT ON TABLE neuronip.agent_audit_trail IS 'Decision audit trail for agents';

CREATE INDEX IF NOT EXISTS idx_audit_trail_trace ON neuronip.agent_audit_trail(trace_id);
CREATE INDEX IF NOT EXISTS idx_audit_trail_agent ON neuronip.agent_audit_trail(agent_id);
CREATE INDEX IF NOT EXISTS idx_audit_trail_timestamp ON neuronip.agent_audit_trail(timestamp DESC);
