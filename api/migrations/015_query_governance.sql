-- Migration: Query Governance
-- Description: Adds function allow-list, read-only mode, sandbox role

-- Allowed functions: Function allow-list for query governance
CREATE TABLE IF NOT EXISTS neuronip.allowed_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_name TEXT NOT NULL UNIQUE,
    function_category TEXT NOT NULL CHECK (function_category IN ('aggregate', 'window', 'scalar', 'table')),
    description TEXT,
    is_safe BOOLEAN NOT NULL DEFAULT true,
    allowed_for_roles TEXT[], -- Roles that can use this function
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.allowed_functions IS 'Allowed SQL functions for query governance';

CREATE INDEX IF NOT EXISTS idx_allowed_functions_name ON neuronip.allowed_functions(function_name);
CREATE INDEX IF NOT EXISTS idx_allowed_functions_category ON neuronip.allowed_functions(function_category);

-- Query governance rules: Rules for query execution
CREATE TABLE IF NOT EXISTS neuronip.query_governance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name TEXT NOT NULL UNIQUE,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('function_allowlist', 'read_only', 'sandbox', 'cost_limit', 'timeout')),
    rule_config JSONB NOT NULL,
    applies_to_roles TEXT[], -- Roles this rule applies to
    applies_to_users TEXT[], -- Specific users this rule applies to
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.query_governance_rules IS 'Query governance rules';

CREATE INDEX IF NOT EXISTS idx_query_governance_rules_type ON neuronip.query_governance_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_query_governance_rules_enabled ON neuronip.query_governance_rules(enabled) WHERE enabled = true;

-- Query validations: Track query validation results
CREATE TABLE IF NOT EXISTS neuronip.query_validations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID, -- Reference to warehouse_queries
    user_id TEXT,
    query_text TEXT NOT NULL,
    validation_status TEXT NOT NULL CHECK (validation_status IN ('allowed', 'blocked', 'warning')),
    validation_rules JSONB, -- Which rules were checked
    blocked_reason TEXT,
    validated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.query_validations IS 'Query validation tracking';

CREATE INDEX IF NOT EXISTS idx_query_validations_query ON neuronip.query_validations(query_id);
CREATE INDEX IF NOT EXISTS idx_query_validations_user ON neuronip.query_validations(user_id);
CREATE INDEX IF NOT EXISTS idx_query_validations_status ON neuronip.query_validations(validation_status);
CREATE INDEX IF NOT EXISTS idx_query_validations_validated ON neuronip.query_validations(validated_at DESC);

-- Sandbox roles: Restricted query execution roles
CREATE TABLE IF NOT EXISTS neuronip.sandbox_roles (
    role_name TEXT PRIMARY KEY,
    max_query_time_seconds INTEGER NOT NULL DEFAULT 30,
    max_result_rows INTEGER NOT NULL DEFAULT 1000,
    max_query_cost NUMERIC(10,2), -- Estimated cost limit
    allowed_schemas TEXT[], -- Schemas this role can access
    blocked_schemas TEXT[], -- Schemas this role cannot access
    read_only BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.sandbox_roles IS 'Sandbox role configurations for restricted query execution';

-- Query cost estimates: Track query cost estimates
CREATE TABLE IF NOT EXISTS neuronip.query_cost_estimates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID,
    estimated_cost NUMERIC(10,2),
    estimated_rows INTEGER,
    estimated_time_ms INTEGER,
    estimation_method TEXT, -- How the cost was estimated
    estimated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.query_cost_estimates IS 'Query cost estimation tracking';

CREATE INDEX IF NOT EXISTS idx_query_cost_estimates_query ON neuronip.query_cost_estimates(query_id);
CREATE INDEX IF NOT EXISTS idx_query_cost_estimates_estimated ON neuronip.query_cost_estimates(estimated_at DESC);

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_query_governance_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_query_governance_rules_updated_at ON neuronip.query_governance_rules;
CREATE TRIGGER trigger_update_query_governance_rules_updated_at
    BEFORE UPDATE ON neuronip.query_governance_rules
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_query_governance_rules_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_sandbox_roles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_sandbox_roles_updated_at ON neuronip.sandbox_roles;
CREATE TRIGGER trigger_update_sandbox_roles_updated_at
    BEFORE UPDATE ON neuronip.sandbox_roles
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_sandbox_roles_updated_at();
