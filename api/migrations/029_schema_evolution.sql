-- Migration: Schema Evolution Tracking
-- Description: Adds tables for schema evolution tracking

-- Schema Changes: Schema evolution change tracking
CREATE TABLE IF NOT EXISTS neuronip.schema_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    change_type TEXT NOT NULL CHECK (change_type IN ('table_added', 'table_dropped', 'column_added', 'column_dropped', 'column_modified', 'index_added', 'index_dropped', 'constraint_added', 'constraint_dropped')),
    column_name TEXT,
    old_schema JSONB DEFAULT '{}',
    new_schema JSONB DEFAULT '{}',
    change_summary TEXT,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.schema_changes IS 'Schema evolution change tracking';

CREATE INDEX IF NOT EXISTS idx_schema_changes_connector ON neuronip.schema_changes(connector_id);
CREATE INDEX IF NOT EXISTS idx_schema_changes_table ON neuronip.schema_changes(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_schema_changes_type ON neuronip.schema_changes(change_type);
CREATE INDEX IF NOT EXISTS idx_schema_changes_detected ON neuronip.schema_changes(detected_at DESC);
