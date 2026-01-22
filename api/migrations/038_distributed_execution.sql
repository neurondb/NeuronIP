-- Migration: Distributed Execution and Scale Story
-- Description: Adds read replicas, job queue, resource quotas, and sharding

-- Read Replicas: Database read replica configuration
CREATE TABLE IF NOT EXISTS neuronip.read_replicas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    region TEXT NOT NULL,
    connection_string TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'degraded')) DEFAULT 'active',
    lag_ms INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 100,
    max_connections INTEGER DEFAULT 10,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_health_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.read_replicas IS 'Read replica configuration';

CREATE INDEX IF NOT EXISTS idx_read_replicas_region ON neuronip.read_replicas(region);
CREATE INDEX IF NOT EXISTS idx_read_replicas_status ON neuronip.read_replicas(status);
CREATE INDEX IF NOT EXISTS idx_read_replicas_enabled ON neuronip.read_replicas(enabled) WHERE enabled = true;

-- Job Queue: Job queue for long-running agents and queries
CREATE TABLE IF NOT EXISTS neuronip.job_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type TEXT NOT NULL CHECK (job_type IN ('query', 'agent', 'workflow', 'ingestion', 'export')),
    job_payload JSONB NOT NULL,
    priority INTEGER DEFAULT 100,
    status TEXT NOT NULL CHECK (status IN ('pending', 'queued', 'running', 'completed', 'failed', 'cancelled')) DEFAULT 'pending',
    resource_requirements JSONB DEFAULT '{}',
    worker_id TEXT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.job_queue IS 'Job queue for async execution';

CREATE INDEX IF NOT EXISTS idx_job_queue_status ON neuronip.job_queue(status);
CREATE INDEX IF NOT EXISTS idx_job_queue_priority ON neuronip.job_queue(priority DESC, scheduled_at);
CREATE INDEX IF NOT EXISTS idx_job_queue_job_type ON neuronip.job_queue(job_type);
CREATE INDEX IF NOT EXISTS idx_job_queue_scheduled ON neuronip.job_queue(scheduled_at) WHERE status IN ('pending', 'queued');

-- Resource Quotas: Per-user and per-workspace resource limits
CREATE TABLE IF NOT EXISTS neuronip.resource_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id TEXT,
    resource_type TEXT NOT NULL CHECK (resource_type IN ('queries', 'agents', 'storage', 'compute', 'api_calls')),
    max_limit BIGINT NOT NULL,
    current_usage BIGINT DEFAULT 0,
    period TEXT NOT NULL CHECK (period IN ('hour', 'day', 'month')) DEFAULT 'day',
    reset_at TIMESTAMPTZ NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, user_id, resource_type, period)
);
COMMENT ON TABLE neuronip.resource_quotas IS 'Resource quota limits per user and workspace';

CREATE INDEX IF NOT EXISTS idx_resource_quotas_workspace ON neuronip.resource_quotas(workspace_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_user ON neuronip.resource_quotas(user_id);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_type ON neuronip.resource_quotas(resource_type);
CREATE INDEX IF NOT EXISTS idx_resource_quotas_reset ON neuronip.resource_quotas(reset_at) WHERE enabled = true;

-- Query Shards: Sharding strategy for large tables
CREATE TABLE IF NOT EXISTS neuronip.query_shards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    schema_name TEXT NOT NULL,
    shard_key TEXT NOT NULL,
    shard_count INTEGER NOT NULL DEFAULT 1,
    shard_strategy TEXT NOT NULL CHECK (shard_strategy IN ('hash', 'range', 'list')) DEFAULT 'hash',
    shard_config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(schema_name, table_name)
);
COMMENT ON TABLE neuronip.query_shards IS 'Sharding configuration for large tables';

CREATE INDEX IF NOT EXISTS idx_query_shards_table ON neuronip.query_shards(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_query_shards_enabled ON neuronip.query_shards(enabled) WHERE enabled = true;

-- Async Query Execution: Track async query execution
CREATE TABLE IF NOT EXISTS neuronip.async_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_text TEXT NOT NULL,
    query_params JSONB DEFAULT '{}',
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')) DEFAULT 'pending',
    result_location TEXT,
    error_message TEXT,
    rows_returned BIGINT,
    execution_time_ms INTEGER,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.async_queries IS 'Async query execution tracking';

CREATE INDEX IF NOT EXISTS idx_async_queries_status ON neuronip.async_queries(status);
CREATE INDEX IF NOT EXISTS idx_async_queries_created_by ON neuronip.async_queries(created_by);
CREATE INDEX IF NOT EXISTS idx_async_queries_expires ON neuronip.async_queries(expires_at) WHERE expires_at IS NOT NULL;

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_read_replicas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_read_replicas_updated_at ON neuronip.read_replicas;
CREATE TRIGGER trigger_update_read_replicas_updated_at
    BEFORE UPDATE ON neuronip.read_replicas
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_read_replicas_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_job_queue_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_job_queue_updated_at ON neuronip.job_queue;
CREATE TRIGGER trigger_update_job_queue_updated_at
    BEFORE UPDATE ON neuronip.job_queue
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_job_queue_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_resource_quotas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_resource_quotas_updated_at ON neuronip.resource_quotas;
CREATE TRIGGER trigger_update_resource_quotas_updated_at
    BEFORE UPDATE ON neuronip.resource_quotas
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_resource_quotas_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_query_shards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_query_shards_updated_at ON neuronip.query_shards;
CREATE TRIGGER trigger_update_query_shards_updated_at
    BEFORE UPDATE ON neuronip.query_shards
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_query_shards_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_async_queries_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_async_queries_updated_at ON neuronip.async_queries;
CREATE TRIGGER trigger_update_async_queries_updated_at
    BEFORE UPDATE ON neuronip.async_queries
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_async_queries_updated_at();
