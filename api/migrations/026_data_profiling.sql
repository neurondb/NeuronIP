-- Migration: Data Profiling
-- Description: Adds tables for data profiling statistics and analysis

-- Data Profiles: Profiling results for tables and columns
CREATE TABLE IF NOT EXISTS neuronip.data_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID NOT NULL REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT,
    profile_type TEXT NOT NULL CHECK (profile_type IN ('table', 'column')),
    statistics JSONB NOT NULL DEFAULT '{}',
    data_type TEXT,
    null_count BIGINT DEFAULT 0,
    non_null_count BIGINT DEFAULT 0,
    distinct_count BIGINT,
    min_value TEXT,
    max_value TEXT,
    avg_value NUMERIC,
    median_value NUMERIC,
    patterns JSONB DEFAULT '{}',
    sample_values TEXT[],
    profiled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, column_name)
);
COMMENT ON TABLE neuronip.data_profiles IS 'Data profiling statistics for tables and columns';

CREATE INDEX IF NOT EXISTS idx_data_profiles_connector ON neuronip.data_profiles(connector_id);
CREATE INDEX IF NOT EXISTS idx_data_profiles_table ON neuronip.data_profiles(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_data_profiles_type ON neuronip.data_profiles(profile_type);
CREATE INDEX IF NOT EXISTS idx_data_profiles_profiled ON neuronip.data_profiles(profiled_at);

-- Data Patterns: Detected patterns in data
CREATE TABLE IF NOT EXISTS neuronip.data_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES neuronip.data_profiles(id) ON DELETE CASCADE,
    pattern_type TEXT NOT NULL CHECK (pattern_type IN ('email', 'phone', 'ssn', 'credit_card', 'url', 'ip_address', 'date', 'custom')),
    pattern_regex TEXT,
    match_count BIGINT DEFAULT 0,
    match_percentage NUMERIC,
    confidence NUMERIC,
    examples TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_patterns IS 'Detected patterns in data';

CREATE INDEX IF NOT EXISTS idx_data_patterns_profile ON neuronip.data_patterns(profile_id);
CREATE INDEX IF NOT EXISTS idx_data_patterns_type ON neuronip.data_patterns(pattern_type);
