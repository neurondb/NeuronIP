-- Migration: Advanced RBAC - Column and Row Level Security
-- Description: Adds column-level and row-level security policies

-- Column Security Policies: Column-level access control
CREATE TABLE IF NOT EXISTS neuronip.column_security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    policy_type TEXT NOT NULL CHECK (policy_type IN ('mask', 'hide', 'redact')),
    masking_rule TEXT,
    user_roles JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, column_name)
);
COMMENT ON TABLE neuronip.column_security_policies IS 'Column-level security policies';

CREATE INDEX IF NOT EXISTS idx_column_security_connector ON neuronip.column_security_policies(connector_id);
CREATE INDEX IF NOT EXISTS idx_column_security_table ON neuronip.column_security_policies(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_column_security_enabled ON neuronip.column_security_policies(enabled);

-- Row Security Policies: Row-level access control
CREATE TABLE IF NOT EXISTS neuronip.row_security_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    policy_name TEXT NOT NULL,
    filter_expression TEXT NOT NULL,
    user_roles JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, policy_name)
);
COMMENT ON TABLE neuronip.row_security_policies IS 'Row-level security policies';

CREATE INDEX IF NOT EXISTS idx_row_security_connector ON neuronip.row_security_policies(connector_id);
CREATE INDEX IF NOT EXISTS idx_row_security_table ON neuronip.row_security_policies(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_row_security_enabled ON neuronip.row_security_policies(enabled);
