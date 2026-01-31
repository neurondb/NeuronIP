-- Migration: 053_connector_registry.sql
-- Description: Enhanced connector registry with versioning and discovery

-- Connector registry for available connectors
CREATE TABLE IF NOT EXISTS neuronip.connector_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_type VARCHAR(100) NOT NULL UNIQUE,
    connector_name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    capabilities TEXT[], -- Array of capabilities
    config_schema JSONB, -- JSON schema for connector configuration
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'deprecated', 'experimental'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_connector_registry_type ON neuronip.connector_registry(connector_type);
CREATE INDEX idx_connector_registry_status ON neuronip.connector_registry(status);

-- Connector versions for versioning support
CREATE TABLE IF NOT EXISTS neuronip.connector_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_type VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    changelog TEXT,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(connector_type, version)
);

CREATE INDEX idx_connector_versions_type ON neuronip.connector_versions(connector_type);
CREATE INDEX idx_connector_versions_current ON neuronip.connector_versions(connector_type, is_current) WHERE is_current = true;

COMMENT ON TABLE neuronip.connector_registry IS 'Registry of available data source connectors';
COMMENT ON TABLE neuronip.connector_versions IS 'Connector version history';
