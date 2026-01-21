-- Migration: Data Freshness Monitoring
-- Description: Adds tables for data freshness monitoring

-- Freshness Monitors: Data freshness monitoring configuration
CREATE TABLE IF NOT EXISTS neuronip.freshness_monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    timestamp_column TEXT NOT NULL,
    expected_interval_minutes INTEGER NOT NULL,
    alert_threshold INTEGER NOT NULL DEFAULT 60,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_check_at TIMESTAMPTZ,
    last_update_at TIMESTAMPTZ,
    status TEXT NOT NULL CHECK (status IN ('fresh', 'stale', 'critical')) DEFAULT 'fresh',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.freshness_monitors IS 'Data freshness monitoring configuration';

CREATE INDEX IF NOT EXISTS idx_freshness_monitors_connector ON neuronip.freshness_monitors(connector_id);
CREATE INDEX IF NOT EXISTS idx_freshness_monitors_table ON neuronip.freshness_monitors(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_freshness_monitors_status ON neuronip.freshness_monitors(status);
CREATE INDEX IF NOT EXISTS idx_freshness_monitors_enabled ON neuronip.freshness_monitors(enabled);
CREATE UNIQUE INDEX IF NOT EXISTS idx_freshness_monitors_unique ON neuronip.freshness_monitors(connector_id, schema_name, table_name, timestamp_column);
