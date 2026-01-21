-- Migration: Multi-Region Deployment
-- Description: Adds region management and region-aware data partitioning

-- Regions: Deployment regions
CREATE TABLE IF NOT EXISTS neuronip.regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    primary_region BOOLEAN NOT NULL DEFAULT false,
    active BOOLEAN NOT NULL DEFAULT true,
    endpoint TEXT,
    database_host TEXT,
    database_port INTEGER,
    replica_of UUID REFERENCES neuronip.regions(id) ON DELETE SET NULL,
    last_sync_at TIMESTAMPTZ,
    health_status TEXT NOT NULL DEFAULT 'healthy' CHECK (health_status IN ('healthy', 'degraded', 'down')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.regions IS 'Deployment regions for multi-region architecture';

CREATE INDEX IF NOT EXISTS idx_regions_code ON neuronip.regions(code);
CREATE INDEX IF NOT EXISTS idx_regions_primary ON neuronip.regions(primary_region) WHERE primary_region = true;
CREATE INDEX IF NOT EXISTS idx_regions_active ON neuronip.regions(active) WHERE active = true;

-- Region-aware partitioning: Add region column to key tables
ALTER TABLE neuronip.users ADD COLUMN IF NOT EXISTS region_code TEXT;
ALTER TABLE neuronip.knowledge_collections ADD COLUMN IF NOT EXISTS region_code TEXT;
ALTER TABLE neuronip.warehouse_schemas ADD COLUMN IF NOT EXISTS region_code TEXT;
ALTER TABLE neuronip.support_tickets ADD COLUMN IF NOT EXISTS region_code TEXT;
ALTER TABLE neuronip.workflows ADD COLUMN IF NOT EXISTS region_code TEXT;

-- Indexes for region-based queries
CREATE INDEX IF NOT EXISTS idx_users_region ON neuronip.users(region_code);
CREATE INDEX IF NOT EXISTS idx_knowledge_collections_region ON neuronip.knowledge_collections(region_code);
CREATE INDEX IF NOT EXISTS idx_warehouse_schemas_region ON neuronip.warehouse_schemas(region_code);
CREATE INDEX IF NOT EXISTS idx_support_tickets_region ON neuronip.support_tickets(region_code);
CREATE INDEX IF NOT EXISTS idx_workflows_region ON neuronip.workflows(region_code);

-- Region sync status: Track replication status between regions
CREATE TABLE IF NOT EXISTS neuronip.region_sync_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_region_id UUID NOT NULL REFERENCES neuronip.regions(id) ON DELETE CASCADE,
    target_region_id UUID NOT NULL REFERENCES neuronip.regions(id) ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    last_synced_at TIMESTAMPTZ,
    sync_status TEXT NOT NULL DEFAULT 'pending' CHECK (sync_status IN ('pending', 'syncing', 'synced', 'failed')),
    sync_error TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_region_id, target_region_id, table_name)
);
COMMENT ON TABLE neuronip.region_sync_status IS 'Region replication sync status';

CREATE INDEX IF NOT EXISTS idx_region_sync_source ON neuronip.region_sync_status(source_region_id);
CREATE INDEX IF NOT EXISTS idx_region_sync_target ON neuronip.region_sync_status(target_region_id);
CREATE INDEX IF NOT EXISTS idx_region_sync_status ON neuronip.region_sync_status(sync_status);
