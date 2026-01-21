-- Migration: Connector Framework
-- Description: Adds tables for data source connectors and catalog management

-- Data Source Connectors: Registered data source connectors
CREATE TABLE IF NOT EXISTS neuronip.data_source_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    connector_type TEXT NOT NULL CHECK (connector_type IN ('postgresql', 'mysql', 'sqlserver', 'oracle', 'snowflake', 'bigquery', 'redshift', 'databricks', 'mongodb', 'elasticsearch', 'custom')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    configuration JSONB NOT NULL DEFAULT '{}',
    connection_string TEXT,
    credentials JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    last_sync_at TIMESTAMPTZ,
    sync_status TEXT CHECK (sync_status IN ('idle', 'syncing', 'success', 'error')) DEFAULT 'idle',
    sync_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_source_connectors IS 'Data source connector configurations';

CREATE INDEX IF NOT EXISTS idx_connectors_type ON neuronip.data_source_connectors(connector_type);
CREATE INDEX IF NOT EXISTS idx_connectors_enabled ON neuronip.data_source_connectors(enabled);
CREATE INDEX IF NOT EXISTS idx_connectors_sync_status ON neuronip.data_source_connectors(sync_status);

-- Catalog Tables: Discovered tables from connectors
CREATE TABLE IF NOT EXISTS neuronip.catalog_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID NOT NULL REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    table_type TEXT NOT NULL CHECK (table_type IN ('table', 'view', 'materialized_view', 'external_table')),
    description TEXT,
    row_count BIGINT,
    size_bytes BIGINT,
    owner TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    last_analyzed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name)
);
COMMENT ON TABLE neuronip.catalog_tables IS 'Discovered tables from data source connectors';

CREATE INDEX IF NOT EXISTS idx_catalog_tables_connector ON neuronip.catalog_tables(connector_id);
CREATE INDEX IF NOT EXISTS idx_catalog_tables_schema ON neuronip.catalog_tables(schema_name);
CREATE INDEX IF NOT EXISTS idx_catalog_tables_name ON neuronip.catalog_tables(table_name);

-- Catalog Columns: Discovered columns from connectors
CREATE TABLE IF NOT EXISTS neuronip.catalog_columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES neuronip.catalog_tables(id) ON DELETE CASCADE,
    column_name TEXT NOT NULL,
    column_type TEXT NOT NULL,
    ordinal_position INTEGER NOT NULL,
    is_nullable BOOLEAN NOT NULL DEFAULT true,
    is_primary_key BOOLEAN NOT NULL DEFAULT false,
    is_foreign_key BOOLEAN NOT NULL DEFAULT false,
    default_value TEXT,
    description TEXT,
    metadata JSONB DEFAULT '{}',
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(table_id, column_name)
);
COMMENT ON TABLE neuronip.catalog_columns IS 'Discovered columns from data source connectors';

CREATE INDEX IF NOT EXISTS idx_catalog_columns_table ON neuronip.catalog_columns(table_id);
CREATE INDEX IF NOT EXISTS idx_catalog_columns_name ON neuronip.catalog_columns(column_name);

-- Connector Sync History: History of connector synchronization
CREATE TABLE IF NOT EXISTS neuronip.connector_sync_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID NOT NULL REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    sync_type TEXT NOT NULL CHECK (sync_type IN ('full', 'incremental', 'schema_only')),
    status TEXT NOT NULL CHECK (status IN ('running', 'success', 'error', 'cancelled')),
    tables_discovered INTEGER DEFAULT 0,
    columns_discovered INTEGER DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.connector_sync_history IS 'History of connector synchronization operations';

CREATE INDEX IF NOT EXISTS idx_sync_history_connector ON neuronip.connector_sync_history(connector_id);
CREATE INDEX IF NOT EXISTS idx_sync_history_status ON neuronip.connector_sync_history(status);
CREATE INDEX IF NOT EXISTS idx_sync_history_started ON neuronip.connector_sync_history(started_at);
