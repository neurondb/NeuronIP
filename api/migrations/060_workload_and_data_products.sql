-- Migration: Elastic analytics (workload queues, cache policies, data products)
-- Description: Workload queues, concurrency controls, cache policies, data products/shares

-- Workload queues: isolation and priority per workspace/tenant
CREATE TABLE IF NOT EXISTS neuronip.workload_queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 100,
    max_concurrency INTEGER NOT NULL DEFAULT 10,
    query_timeout_seconds INTEGER NOT NULL DEFAULT 300,
    query_budget_per_period BIGINT,
    period TEXT CHECK (period IN ('hour', 'day', 'month')),
    workspace_id UUID,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workload_queues IS 'Workload isolation queues';

CREATE INDEX IF NOT EXISTS idx_workload_queues_workspace ON neuronip.workload_queues(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workload_queues_enabled ON neuronip.workload_queues(enabled) WHERE enabled = true;

-- Active query slots: current concurrency per queue (best-effort, can use in-memory instead)
CREATE TABLE IF NOT EXISTS neuronip.workload_queue_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL REFERENCES neuronip.workload_queues(id) ON DELETE CASCADE,
    query_id UUID NOT NULL,
    user_id TEXT,
    workspace_id UUID,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
COMMENT ON TABLE neuronip.workload_queue_slots IS 'Active concurrency slots per workload queue';

CREATE INDEX IF NOT EXISTS idx_workload_queue_slots_queue ON neuronip.workload_queue_slots(queue_id);
CREATE INDEX IF NOT EXISTS idx_workload_queue_slots_expires ON neuronip.workload_queue_slots(expires_at);

-- Cache policies: per-schema or per-pattern TTL and rules (extends 016_query_cache)
CREATE TABLE IF NOT EXISTS neuronip.cache_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    scope TEXT NOT NULL CHECK (scope IN ('global', 'schema', 'table', 'query_pattern')),
    scope_value TEXT,
    ttl_seconds INTEGER NOT NULL DEFAULT 3600,
    max_result_rows INTEGER,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.cache_policies IS 'Cache TTL and rules per scope';

CREATE INDEX IF NOT EXISTS idx_cache_policies_scope ON neuronip.cache_policies(scope, scope_value);
CREATE INDEX IF NOT EXISTS idx_cache_policies_enabled ON neuronip.cache_policies(enabled) WHERE enabled = true;

-- Data products (shares): publish datasets/metrics with versioning and consumption
CREATE TABLE IF NOT EXISTS neuronip.data_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    owner_id TEXT NOT NULL,
    workspace_id UUID,
    version TEXT NOT NULL DEFAULT '1.0.0',
    schema_ids JSONB DEFAULT '[]',
    metric_ids JSONB DEFAULT '[]',
    dataset_ids JSONB DEFAULT '[]',
    sla_freshness_hours INTEGER,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'workspace', 'public')) DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, version)
);
COMMENT ON TABLE neuronip.data_products IS 'Data products (sharing / consumption layer)';

CREATE INDEX IF NOT EXISTS idx_data_products_owner ON neuronip.data_products(owner_id);
CREATE INDEX IF NOT EXISTS idx_data_products_workspace ON neuronip.data_products(workspace_id);
CREATE INDEX IF NOT EXISTS idx_data_products_visibility ON neuronip.data_products(visibility);

-- Data product consumers (who can access a share)
CREATE TABLE IF NOT EXISTS neuronip.data_product_consumers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_product_id UUID NOT NULL REFERENCES neuronip.data_products(id) ON DELETE CASCADE,
    consumer_workspace_id UUID,
    consumer_user_id TEXT,
    permissions TEXT NOT NULL CHECK (permissions IN ('read', 'read_write')) DEFAULT 'read',
    granted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by TEXT,
    expires_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.data_product_consumers IS 'Consumers of data products (share grants)';

CREATE UNIQUE INDEX IF NOT EXISTS idx_data_product_consumers_unique
    ON neuronip.data_product_consumers (data_product_id, COALESCE(consumer_workspace_id::text, ''), COALESCE(consumer_user_id, ''));

CREATE INDEX IF NOT EXISTS idx_data_product_consumers_product ON neuronip.data_product_consumers(data_product_id);
CREATE INDEX IF NOT EXISTS idx_data_product_consumers_consumer ON neuronip.data_product_consumers(consumer_workspace_id, consumer_user_id);
