-- Migration: Data Masking
-- Description: Adds data masking policies

-- Masking Policies: Data masking policies
CREATE TABLE IF NOT EXISTS neuronip.masking_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    masking_type TEXT NOT NULL CHECK (masking_type IN ('tokenization', 'encryption', 'format_preserving', 'partial', 'full')),
    algorithm TEXT,
    masking_rule TEXT,
    user_roles JSONB DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, column_name)
);
COMMENT ON TABLE neuronip.masking_policies IS 'Data masking policies';

CREATE INDEX IF NOT EXISTS idx_masking_policies_connector ON neuronip.masking_policies(connector_id);
CREATE INDEX IF NOT EXISTS idx_masking_policies_table ON neuronip.masking_policies(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_masking_policies_enabled ON neuronip.masking_policies(enabled) WHERE enabled = true;
