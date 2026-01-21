-- Migration: API Keys Enhancements
-- Description: Adds scopes, rotation, expiry, per-key rate limits, usage analytics

-- Enhance api_keys table with new fields
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS scopes TEXT[];
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS rotation_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS rotation_interval_days INTEGER;
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS last_rotated_at TIMESTAMPTZ;
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS next_rotation_at TIMESTAMPTZ;
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ;
ALTER TABLE neuronip.api_keys ADD COLUMN IF NOT EXISTS revoked_reason TEXT;

COMMENT ON COLUMN neuronip.api_keys.scopes IS 'Fine-grained permissions for this API key';
COMMENT ON COLUMN neuronip.api_keys.metadata IS 'Additional metadata for the API key';
COMMENT ON COLUMN neuronip.api_keys.tags IS 'Tags for organizing API keys';
COMMENT ON COLUMN neuronip.api_keys.rotation_enabled IS 'Whether automatic rotation is enabled';
COMMENT ON COLUMN neuronip.api_keys.rotation_interval_days IS 'Days between rotations';
COMMENT ON COLUMN neuronip.api_keys.last_rotated_at IS 'When the key was last rotated';
COMMENT ON COLUMN neuronip.api_keys.next_rotation_at IS 'When the key should be rotated next';
COMMENT ON COLUMN neuronip.api_keys.revoked_at IS 'When the key was revoked';
COMMENT ON COLUMN neuronip.api_keys.revoked_reason IS 'Reason for revocation';

-- API key usage analytics: Track API key usage
CREATE TABLE IF NOT EXISTS neuronip.api_key_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES neuronip.api_keys(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.api_key_usage IS 'API key usage analytics';

CREATE INDEX IF NOT EXISTS idx_api_key_usage_key ON neuronip.api_key_usage(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_created ON neuronip.api_key_usage(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_endpoint ON neuronip.api_key_usage(endpoint);

-- API key rotation history: Track key rotations
CREATE TABLE IF NOT EXISTS neuronip.api_key_rotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES neuronip.api_keys(id) ON DELETE CASCADE,
    old_key_prefix TEXT NOT NULL,
    new_key_prefix TEXT NOT NULL,
    rotated_by UUID REFERENCES neuronip.users(id),
    rotation_type TEXT NOT NULL CHECK (rotation_type IN ('manual', 'automatic', 'scheduled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.api_key_rotations IS 'API key rotation history';

CREATE INDEX IF NOT EXISTS idx_api_key_rotations_key ON neuronip.api_key_rotations(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_rotations_created ON neuronip.api_key_rotations(created_at DESC);

-- API key rate limit tracking: Per-key rate limit enforcement
CREATE TABLE IF NOT EXISTS neuronip.api_key_rate_limits (
    api_key_id UUID PRIMARY KEY REFERENCES neuronip.api_keys(id) ON DELETE CASCADE,
    requests_per_minute INTEGER NOT NULL DEFAULT 60,
    requests_per_hour INTEGER NOT NULL DEFAULT 1000,
    requests_per_day INTEGER NOT NULL DEFAULT 10000,
    burst_limit INTEGER NOT NULL DEFAULT 10,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.api_key_rate_limits IS 'Per-key rate limit configuration';

-- API key rate limit counters: Track current rate limit state
CREATE TABLE IF NOT EXISTS neuronip.api_key_rate_counters (
    api_key_id UUID NOT NULL REFERENCES neuronip.api_keys(id) ON DELETE CASCADE,
    window_start TIMESTAMPTZ NOT NULL,
    window_type TEXT NOT NULL CHECK (window_type IN ('minute', 'hour', 'day')),
    request_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (api_key_id, window_start, window_type)
);
COMMENT ON TABLE neuronip.api_key_rate_counters IS 'Rate limit counters for sliding windows';

CREATE INDEX IF NOT EXISTS idx_rate_counters_key_window ON neuronip.api_key_rate_counters(api_key_id, window_start);

-- Function to check and update rate limits
CREATE OR REPLACE FUNCTION neuronip.check_api_key_rate_limit(
    p_api_key_id UUID,
    p_window_type TEXT
) RETURNS BOOLEAN AS $$
DECLARE
    v_limit INTEGER;
    v_count INTEGER;
    v_window_start TIMESTAMPTZ;
BEGIN
    -- Get rate limit for the key
    SELECT 
        CASE p_window_type
            WHEN 'minute' THEN requests_per_minute
            WHEN 'hour' THEN requests_per_hour
            WHEN 'day' THEN requests_per_day
        END
    INTO v_limit
    FROM neuronip.api_key_rate_limits
    WHERE api_key_id = p_api_key_id;

    -- Default limit if not set
    IF v_limit IS NULL THEN
        v_limit := CASE p_window_type
            WHEN 'minute' THEN 60
            WHEN 'hour' THEN 1000
            WHEN 'day' THEN 10000
        END;
    END IF;

    -- Calculate window start
    v_window_start := CASE p_window_type
        WHEN 'minute' THEN date_trunc('minute', NOW())
        WHEN 'hour' THEN date_trunc('hour', NOW())
        WHEN 'day' THEN date_trunc('day', NOW())
    END;

    -- Get or create counter
    INSERT INTO neuronip.api_key_rate_counters (api_key_id, window_start, window_type, request_count)
    VALUES (p_api_key_id, v_window_start, p_window_type, 1)
    ON CONFLICT (api_key_id, window_start, window_type)
    DO UPDATE SET request_count = api_key_rate_counters.request_count + 1
    RETURNING request_count INTO v_count;

    -- Check if limit exceeded
    RETURN v_count <= v_limit;
END;
$$ LANGUAGE plpgsql;

-- Cleanup old rate limit counters (run periodically)
CREATE OR REPLACE FUNCTION neuronip.cleanup_old_rate_counters()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM neuronip.api_key_rate_counters
    WHERE window_start < NOW() - INTERVAL '2 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- View for API key usage summary
CREATE OR REPLACE VIEW neuronip.api_key_usage_summary AS
SELECT
    ak.id,
    ak.key_prefix,
    ak.name,
    ak.user_id,
    COUNT(aku.id) as total_requests,
    COUNT(CASE WHEN aku.status_code >= 200 AND aku.status_code < 300 THEN 1 END) as successful_requests,
    COUNT(CASE WHEN aku.status_code >= 400 THEN 1 END) as failed_requests,
    AVG(aku.response_time_ms) as avg_response_time_ms,
    MAX(aku.created_at) as last_used_at,
    MIN(aku.created_at) as first_used_at
FROM neuronip.api_keys ak
LEFT JOIN neuronip.api_key_usage aku ON ak.id = aku.api_key_id
GROUP BY ak.id, ak.key_prefix, ak.name, ak.user_id;

COMMENT ON VIEW neuronip.api_key_usage_summary IS 'Summary of API key usage statistics';
