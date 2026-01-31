-- Migration: Distributed Execution Schema
-- Description: Adds tables for distributed job execution and priority queues

-- Distributed jobs: Jobs for distributed execution
CREATE TABLE IF NOT EXISTS neuronip.distributed_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    tenant_id TEXT,
    resource_quota JSONB DEFAULT '{}',
    job_data JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')) DEFAULT 'pending',
    assigned_node TEXT,
    result_data JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.distributed_jobs IS 'Distributed job execution queue';

CREATE INDEX IF NOT EXISTS idx_distributed_jobs_status ON neuronip.distributed_jobs(status);
CREATE INDEX IF NOT EXISTS idx_distributed_jobs_priority ON neuronip.distributed_jobs(priority DESC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_distributed_jobs_tenant ON neuronip.distributed_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_distributed_jobs_node ON neuronip.distributed_jobs(assigned_node);

-- Execution nodes: Registered execution nodes
CREATE TABLE IF NOT EXISTS neuronip.execution_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id TEXT NOT NULL UNIQUE,
    node_type TEXT NOT NULL CHECK (node_type IN ('worker', 'coordinator', 'scheduler')),
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive', 'failed')) DEFAULT 'active',
    capacity JSONB DEFAULT '{}',
    current_load INTEGER DEFAULT 0,
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.execution_nodes IS 'Registered execution nodes';

CREATE INDEX IF NOT EXISTS idx_execution_nodes_status ON neuronip.execution_nodes(status);
CREATE INDEX IF NOT EXISTS idx_execution_nodes_heartbeat ON neuronip.execution_nodes(last_heartbeat);

-- Tenant resource quotas: per-tenant quotas (isolation service). Distinct from workspace/user resource_quotas.
CREATE TABLE IF NOT EXISTS neuronip.tenant_resource_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    resource_type TEXT NOT NULL CHECK (resource_type IN ('cpu', 'memory', 'storage', 'concurrent_jobs', 'api_calls')),
    quota_limit NUMERIC NOT NULL,
    current_usage NUMERIC DEFAULT 0,
    period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, resource_type, period_start)
);
COMMENT ON TABLE neuronip.tenant_resource_quotas IS 'Resource quotas per tenant (isolation)';

CREATE INDEX IF NOT EXISTS idx_tenant_resource_quotas_tenant ON neuronip.tenant_resource_quotas(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_resource_quotas_type ON neuronip.tenant_resource_quotas(resource_type);
