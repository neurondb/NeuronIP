-- Migration: Observability Metrics
-- Description: Adds latency, error rate, token usage, embedding cost metrics

-- Latency metrics: Track endpoint latency
CREATE TABLE IF NOT EXISTS neuronip.latency_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    latency_ms NUMERIC(10,2) NOT NULL,
    request_id TEXT,
    user_id TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.latency_metrics IS 'Endpoint latency tracking';

CREATE INDEX IF NOT EXISTS idx_latency_metrics_endpoint ON neuronip.latency_metrics(endpoint);
CREATE INDEX IF NOT EXISTS idx_latency_metrics_recorded ON neuronip.latency_metrics(recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_latency_metrics_request ON neuronip.latency_metrics(request_id);

-- Error metrics: Track error rates
CREATE TABLE IF NOT EXISTS neuronip.error_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    error_type TEXT,
    error_message TEXT,
    request_id TEXT,
    user_id TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.error_metrics IS 'Error rate tracking';

CREATE INDEX IF NOT EXISTS idx_error_metrics_endpoint ON neuronip.error_metrics(endpoint);
CREATE INDEX IF NOT EXISTS idx_error_metrics_status ON neuronip.error_metrics(status_code);
CREATE INDEX IF NOT EXISTS idx_error_metrics_recorded ON neuronip.error_metrics(recorded_at DESC);

-- Token usage: Track token consumption
CREATE TABLE IF NOT EXISTS neuronip.token_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    endpoint TEXT NOT NULL,
    user_id TEXT,
    request_id TEXT,
    tokens INTEGER NOT NULL,
    cost_usd NUMERIC(10,6) NOT NULL DEFAULT 0,
    model_name TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.token_usage IS 'Token usage tracking';

CREATE INDEX IF NOT EXISTS idx_token_usage_endpoint ON neuronip.token_usage(endpoint);
CREATE INDEX IF NOT EXISTS idx_token_usage_user ON neuronip.token_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_recorded ON neuronip.token_usage(recorded_at DESC);

-- Embedding costs: Track embedding generation costs
CREATE TABLE IF NOT EXISTS neuronip.embedding_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name TEXT NOT NULL,
    tokens INTEGER NOT NULL,
    cost_usd NUMERIC(10,6) NOT NULL,
    request_id TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.embedding_costs IS 'Embedding generation cost tracking';

CREATE INDEX IF NOT EXISTS idx_embedding_costs_model ON neuronip.embedding_costs(model_name);
CREATE INDEX IF NOT EXISTS idx_embedding_costs_recorded ON neuronip.embedding_costs(recorded_at DESC);

-- Request correlation: Track request IDs across UI → API → DB
CREATE TABLE IF NOT EXISTS neuronip.request_correlation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id TEXT NOT NULL UNIQUE,
    ui_request_id TEXT, -- Request ID from frontend
    api_request_id TEXT, -- Request ID in API
    db_query_id TEXT, -- Query ID in database
    user_id TEXT,
    endpoint TEXT,
    method TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.request_correlation IS 'Request ID correlation across UI → API → DB';

CREATE INDEX IF NOT EXISTS idx_request_correlation_request ON neuronip.request_correlation(request_id);
CREATE INDEX IF NOT EXISTS idx_request_correlation_ui ON neuronip.request_correlation(ui_request_id);
CREATE INDEX IF NOT EXISTS idx_request_correlation_api ON neuronip.request_correlation(api_request_id);
CREATE INDEX IF NOT EXISTS idx_request_correlation_db ON neuronip.request_correlation(db_query_id);

-- Cleanup old metrics (run periodically)
CREATE OR REPLACE FUNCTION neuronip.cleanup_old_metrics()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete metrics older than 90 days
    DELETE FROM neuronip.latency_metrics WHERE recorded_at < NOW() - INTERVAL '90 days';
    DELETE FROM neuronip.error_metrics WHERE recorded_at < NOW() - INTERVAL '90 days';
    DELETE FROM neuronip.token_usage WHERE recorded_at < NOW() - INTERVAL '90 days';
    DELETE FROM neuronip.embedding_costs WHERE recorded_at < NOW() - INTERVAL '90 days';
    DELETE FROM neuronip.request_correlation WHERE created_at < NOW() - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
