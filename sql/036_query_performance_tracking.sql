-- Query performance tracking table
-- This table stores query performance metrics for monitoring

CREATE TABLE IF NOT EXISTS neuronip.query_performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_name VARCHAR(255) NOT NULL,
    execution_time_ms BIGINT NOT NULL,
    has_error BOOLEAN NOT NULL DEFAULT false,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for query performance analysis
CREATE INDEX IF NOT EXISTS idx_query_performance_metrics_name_time 
    ON neuronip.query_performance_metrics(query_name, execution_time_ms DESC);

CREATE INDEX IF NOT EXISTS idx_query_performance_metrics_executed 
    ON neuronip.query_performance_metrics(executed_at DESC);

-- Partial index for slow queries only
CREATE INDEX IF NOT EXISTS idx_query_performance_metrics_slow 
    ON neuronip.query_performance_metrics(execution_time_ms DESC, executed_at DESC)
    WHERE execution_time_ms > 1000;

-- Cleanup old metrics (older than 30 days)
CREATE OR REPLACE FUNCTION neuronip.cleanup_old_query_metrics()
RETURNS void AS $$
BEGIN
    DELETE FROM neuronip.query_performance_metrics
    WHERE executed_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
