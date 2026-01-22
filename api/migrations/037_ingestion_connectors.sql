-- Migration: Ingestion Connectors Enhancement
-- Description: Adds SaaS connectors, file ingestion, and enhanced failure handling

-- Extend connector_types with SaaS connectors
DO $$ 
BEGIN
    -- Add SaaS connector types if they don't exist
    IF NOT EXISTS (SELECT 1 FROM neuronip.connector_types WHERE type_name = 'zendesk') THEN
        INSERT INTO neuronip.connector_types (type_name, display_name, connector_category, config_schema, created_at)
        VALUES ('zendesk', 'Zendesk', 'saas', '{"subdomain": "string", "email": "string", "api_token": "string"}'::jsonb, NOW());
    END IF;

    IF NOT EXISTS (SELECT 1 FROM neuronip.connector_types WHERE type_name = 'jira') THEN
        INSERT INTO neuronip.connector_types (type_name, display_name, connector_category, config_schema, created_at)
        VALUES ('jira', 'Jira', 'saas', '{"url": "string", "email": "string", "api_token": "string"}'::jsonb, NOW());
    END IF;

    IF NOT EXISTS (SELECT 1 FROM neuronip.connector_types WHERE type_name = 'salesforce') THEN
        INSERT INTO neuronip.connector_types (type_name, display_name, connector_category, config_schema, created_at)
        VALUES ('salesforce', 'Salesforce', 'saas', '{"instance_url": "string", "client_id": "string", "client_secret": "string", "username": "string", "password": "string"}'::jsonb, NOW());
    END IF;

    IF NOT EXISTS (SELECT 1 FROM neuronip.connector_types WHERE type_name = 'hubspot') THEN
        INSERT INTO neuronip.connector_types (type_name, display_name, connector_category, config_schema, created_at)
        VALUES ('hubspot', 'HubSpot', 'saas', '{"api_key": "string", "portal_id": "string"}'::jsonb, NOW());
    END IF;
END $$;

-- File ingestion jobs: Track file uploads and parsing
CREATE TABLE IF NOT EXISTS neuronip.file_ingestion_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_type TEXT NOT NULL CHECK (file_type IN ('csv', 'pdf', 'xlsx', 'docx', 'pptx', 'json', 'parquet')),
    parser_type TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')) DEFAULT 'pending',
    data_source_id UUID REFERENCES neuronip.data_sources(id) ON DELETE SET NULL,
    rows_processed INTEGER DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.file_ingestion_jobs IS 'File ingestion job tracking';

CREATE INDEX IF NOT EXISTS idx_file_ingestion_jobs_status ON neuronip.file_ingestion_jobs(status);
CREATE INDEX IF NOT EXISTS idx_file_ingestion_jobs_data_source ON neuronip.file_ingestion_jobs(data_source_id);
CREATE INDEX IF NOT EXISTS idx_file_ingestion_jobs_created_at ON neuronip.file_ingestion_jobs(created_at DESC);

-- Enhance ingestion_jobs with retry policies and failure details
ALTER TABLE neuronip.ingestion_jobs 
    ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS max_retries INTEGER DEFAULT 3,
    ADD COLUMN IF NOT EXISTS retry_backoff_ms INTEGER DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS last_error TEXT,
    ADD COLUMN IF NOT EXISTS last_error_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS incremental_sync_enabled BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS watermark_column TEXT,
    ADD COLUMN IF NOT EXISTS last_watermark_value TEXT;

CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_retry ON neuronip.ingestion_jobs(retry_count, max_retries, status);

-- Connector sync schedules
CREATE TABLE IF NOT EXISTS neuronip.connector_sync_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID NOT NULL REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schedule_type TEXT NOT NULL CHECK (schedule_type IN ('interval', 'cron', 'manual')) DEFAULT 'manual',
    schedule_config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id)
);
COMMENT ON TABLE neuronip.connector_sync_schedules IS 'Scheduled sync configurations for connectors';

CREATE INDEX IF NOT EXISTS idx_connector_sync_schedules_connector ON neuronip.connector_sync_schedules(connector_id);
CREATE INDEX IF NOT EXISTS idx_connector_sync_schedules_next_run ON neuronip.connector_sync_schedules(next_run_at) WHERE enabled = true;

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_file_ingestion_jobs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_file_ingestion_jobs_updated_at ON neuronip.file_ingestion_jobs;
CREATE TRIGGER trigger_update_file_ingestion_jobs_updated_at
    BEFORE UPDATE ON neuronip.file_ingestion_jobs
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_file_ingestion_jobs_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_connector_sync_schedules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_connector_sync_schedules_updated_at ON neuronip.connector_sync_schedules;
CREATE TRIGGER trigger_update_connector_sync_schedules_updated_at
    BEFORE UPDATE ON neuronip.connector_sync_schedules
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_connector_sync_schedules_updated_at();
