-- Migration: Model Versioning Schema
-- Description: Adds tables for model versioning, A/B testing, and performance tracking

-- Model versions: Version history for models
CREATE TABLE IF NOT EXISTS neuronip.model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    model_path TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    performance JSONB DEFAULT '{}',
    changelog TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, version)
);
COMMENT ON TABLE neuronip.model_versions IS 'Model version history';

CREATE INDEX IF NOT EXISTS idx_model_versions_model ON neuronip.model_versions(model_id);
CREATE INDEX IF NOT EXISTS idx_model_versions_created_at ON neuronip.model_versions(created_at DESC);

-- Model experiments: A/B testing experiments
CREATE TABLE IF NOT EXISTS neuronip.model_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    model_a_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    model_b_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    traffic_split JSONB DEFAULT '{"a": 50, "b": 50}',
    status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'completed', 'cancelled')) DEFAULT 'draft',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    results JSONB DEFAULT '{}',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.model_experiments IS 'A/B testing experiments for models';

CREATE INDEX IF NOT EXISTS idx_model_experiments_status ON neuronip.model_experiments(status);
CREATE INDEX IF NOT EXISTS idx_model_experiments_created_at ON neuronip.model_experiments(created_at DESC);

-- Model metrics: Time-series performance metrics
CREATE TABLE IF NOT EXISTS neuronip.model_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.model_metrics IS 'Time-series performance metrics for models';

CREATE INDEX IF NOT EXISTS idx_model_metrics_model ON neuronip.model_metrics(model_id);
CREATE INDEX IF NOT EXISTS idx_model_metrics_timestamp ON neuronip.model_metrics(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_model_metrics_name ON neuronip.model_metrics(metric_name);

-- Extend model_registry with versioning fields
ALTER TABLE neuronip.model_registry ADD COLUMN IF NOT EXISTS current_version TEXT DEFAULT '1.0.0';
ALTER TABLE neuronip.model_registry ADD COLUMN IF NOT EXISTS version_count INTEGER DEFAULT 1;
ALTER TABLE neuronip.model_registry ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
