-- Migration: Ingestion Schema
-- Description: Adds tables for data ingestion, connectors, CDC, and ETL/ELT

-- Ingestion jobs: Track data ingestion jobs
CREATE TABLE IF NOT EXISTS neuronip.ingestion_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id) ON DELETE CASCADE,
    job_type TEXT NOT NULL CHECK (job_type IN ('sync', 'cdc', 'etl', 'file_upload')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')) DEFAULT 'pending',
    config JSONB DEFAULT '{}',
    progress JSONB DEFAULT '{}',
    error_message TEXT,
    rows_processed BIGINT DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.ingestion_jobs IS 'Data ingestion job tracking';

CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_data_source ON neuronip.ingestion_jobs(data_source_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_status ON neuronip.ingestion_jobs(status);
CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_created_at ON neuronip.ingestion_jobs(created_at DESC);

-- CDC checkpoints: Track CDC replication state
CREATE TABLE IF NOT EXISTS neuronip.cdc_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id) ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    checkpoint_data JSONB NOT NULL,
    last_lsn TEXT, -- PostgreSQL logical sequence number
    last_timestamp TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(data_source_id, table_name)
);
COMMENT ON TABLE neuronip.cdc_checkpoints IS 'CDC replication checkpoints';

CREATE INDEX IF NOT EXISTS idx_cdc_checkpoints_data_source ON neuronip.cdc_checkpoints(data_source_id);

-- Connector registry: Available connectors and their capabilities
CREATE TABLE IF NOT EXISTS neuronip.connector_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_type TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    version TEXT NOT NULL,
    capabilities TEXT[] DEFAULT '{}',
    config_schema JSONB DEFAULT '{}', -- JSON schema for connector configuration
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.connector_registry IS 'Available connector types and their capabilities';

-- Ingestion logs: Detailed logs for ingestion operations
CREATE TABLE IF NOT EXISTS neuronip.ingestion_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES neuronip.ingestion_jobs(id) ON DELETE CASCADE,
    level TEXT NOT NULL CHECK (level IN ('debug', 'info', 'warning', 'error')),
    message TEXT NOT NULL,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.ingestion_logs IS 'Detailed ingestion operation logs';

CREATE INDEX IF NOT EXISTS idx_ingestion_logs_job_id ON neuronip.ingestion_logs(job_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_logs_created_at ON neuronip.ingestion_logs(created_at DESC);

-- File uploads: Track file uploads for ingestion
CREATE TABLE IF NOT EXISTS neuronip.file_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID REFERENCES neuronip.data_sources(id) ON DELETE SET NULL,
    file_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    file_type TEXT NOT NULL,
    file_path TEXT,
    status TEXT NOT NULL CHECK (status IN ('uploading', 'processing', 'completed', 'failed')) DEFAULT 'uploading',
    rows_imported BIGINT DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    uploaded_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.file_uploads IS 'File upload tracking for ingestion';

CREATE INDEX IF NOT EXISTS idx_file_uploads_data_source ON neuronip.file_uploads(data_source_id);
CREATE INDEX IF NOT EXISTS idx_file_uploads_status ON neuronip.file_uploads(status);

-- ETL transformations: Store ETL transformation definitions
CREATE TABLE IF NOT EXISTS neuronip.etl_transformations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    transformation_type TEXT NOT NULL CHECK (transformation_type IN ('filter', 'map', 'aggregate', 'join', 'custom')),
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.etl_transformations IS 'ETL transformation definitions';

-- Extend data_sources table with connector-specific fields
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS connector_type TEXT;
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS connector_config JSONB DEFAULT '{}';
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS schema_cache JSONB;
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS schema_cached_at TIMESTAMPTZ;

-- Add indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_data_sources_connector_type ON neuronip.data_sources(connector_type);
CREATE INDEX IF NOT EXISTS idx_data_sources_enabled ON neuronip.data_sources(enabled);

-- Webhooks: Webhook registrations
CREATE TABLE IF NOT EXISTS neuronip.webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    secret TEXT,
    headers JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    retry_config JSONB DEFAULT '{"max_retries": 3}',
    last_triggered_at TIMESTAMPTZ,
    trigger_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.webhooks IS 'Webhook registrations';

CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON neuronip.webhooks(enabled);
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON neuronip.webhooks USING GIN(events);
