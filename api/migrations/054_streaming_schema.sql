-- Migration: 054_streaming_schema.sql
-- Description: Enhanced streaming and CDC schema

-- CDC connectors table
CREATE TABLE IF NOT EXISTS neuronip.cdc_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_name VARCHAR(255) NOT NULL UNIQUE,
    connector_type VARCHAR(50) NOT NULL, -- 'debezium', 'kafka', 'native'
    source_type VARCHAR(50) NOT NULL, -- 'postgresql', 'mysql', 'mongodb', etc.
    config JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive', -- 'active', 'inactive', 'stopped', 'error'
    last_event_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cdc_connectors_type ON neuronip.cdc_connectors(connector_type);
CREATE INDEX idx_cdc_connectors_status ON neuronip.cdc_connectors(status);

COMMENT ON TABLE neuronip.cdc_connectors IS 'CDC connector configurations';
