-- Migration: Query Result Caching
-- Description: Adds TTL-based caching with invalidation rules

-- Query cache: Cache query results with TTL
CREATE TABLE IF NOT EXISTS neuronip.query_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key TEXT NOT NULL UNIQUE,
    query_hash TEXT NOT NULL, -- Hash of query (without params)
    query_text TEXT NOT NULL,
    result_data JSONB NOT NULL,
    ttl_seconds INTEGER NOT NULL DEFAULT 3600, -- Default 1 hour
    expires_at TIMESTAMPTZ NOT NULL,
    hit_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.query_cache IS 'Cached query results with TTL';

CREATE INDEX IF NOT EXISTS idx_query_cache_key ON neuronip.query_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_query_cache_hash ON neuronip.query_cache(query_hash);
CREATE INDEX IF NOT EXISTS idx_query_cache_expires ON neuronip.query_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_query_cache_accessed ON neuronip.query_cache(last_accessed_at DESC);

-- Cache invalidation rules: Rules for cache invalidation
CREATE TABLE IF NOT EXISTS neuronip.cache_invalidation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name TEXT NOT NULL UNIQUE,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('query_hash', 'schema', 'table', 'time', 'custom')),
    rule_value TEXT,
    rule_config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.cache_invalidation_rules IS 'Cache invalidation rules';

CREATE INDEX IF NOT EXISTS idx_cache_invalidation_rules_type ON neuronip.cache_invalidation_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_cache_invalidation_rules_enabled ON neuronip.cache_invalidation_rules(enabled) WHERE enabled = true;

-- Cache warming: Track cache warming operations
CREATE TABLE IF NOT EXISTS neuronip.cache_warming (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_text TEXT NOT NULL,
    cache_key TEXT NOT NULL,
    warmed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT
);
COMMENT ON TABLE neuronip.cache_warming IS 'Cache warming operations';

CREATE INDEX IF NOT EXISTS idx_cache_warming_warmed ON neuronip.cache_warming(warmed_at DESC);

-- Cache metrics: Track cache performance
CREATE TABLE IF NOT EXISTS neuronip.cache_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key TEXT,
    hit_count INTEGER NOT NULL DEFAULT 0,
    miss_count INTEGER NOT NULL DEFAULT 0,
    total_requests INTEGER NOT NULL DEFAULT 0,
    avg_response_time_ms INTEGER,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.cache_metrics IS 'Cache performance metrics';

CREATE INDEX IF NOT EXISTS idx_cache_metrics_key ON neuronip.cache_metrics(cache_key);
CREATE INDEX IF NOT EXISTS idx_cache_metrics_recorded ON neuronip.cache_metrics(recorded_at DESC);

-- Function to clean expired cache entries
CREATE OR REPLACE FUNCTION neuronip.cleanup_expired_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM neuronip.query_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_cache_invalidation_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_cache_invalidation_rules_updated_at ON neuronip.cache_invalidation_rules;
CREATE TRIGGER trigger_update_cache_invalidation_rules_updated_at
    BEFORE UPDATE ON neuronip.cache_invalidation_rules
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_cache_invalidation_rules_updated_at();
