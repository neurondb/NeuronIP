-- Migration: 067_cdc_events.sql
-- Description: CDC change events table for Debezium/Kafka-consumed or DB-backed events

CREATE TABLE IF NOT EXISTS neuronip.cdc_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_name VARCHAR(255) NOT NULL,
    table_name TEXT NOT NULL,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('insert', 'update', 'delete')),
    lsn TEXT,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cdc_events_connector ON neuronip.cdc_events(connector_name);
CREATE INDEX IF NOT EXISTS idx_cdc_events_created ON neuronip.cdc_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cdc_events_table ON neuronip.cdc_events(table_name);

COMMENT ON TABLE neuronip.cdc_events IS 'CDC change events (populated by Kafka consumer or DB triggers)';
