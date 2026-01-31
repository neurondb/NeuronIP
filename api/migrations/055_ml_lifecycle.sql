-- Migration: 055_ml_lifecycle.sql
-- Description: ML lifecycle with training pipelines, experiment tracking, and model serving

-- ML training pipelines
CREATE TABLE IF NOT EXISTS neuronip.ml_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_name VARCHAR(255) NOT NULL,
    description TEXT,
    steps JSONB NOT NULL, -- Array of pipeline steps
    config JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'active', 'archived'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_pipelines_status ON neuronip.ml_pipelines(status);

-- ML pipeline executions
CREATE TABLE IF NOT EXISTS neuronip.ml_pipeline_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.ml_pipelines(id),
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- 'running', 'completed', 'failed', 'cancelled'
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    results JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_pipeline_executions_pipeline ON neuronip.ml_pipeline_executions(pipeline_id);
CREATE INDEX idx_ml_pipeline_executions_status ON neuronip.ml_pipeline_executions(status);

-- ML experiments
CREATE TABLE IF NOT EXISTS neuronip.ml_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_name VARCHAR(255) NOT NULL,
    description TEXT,
    parameters JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    tags JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- 'running', 'completed', 'failed'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_experiments_status ON neuronip.ml_experiments(status);
CREATE INDEX idx_ml_experiments_created ON neuronip.ml_experiments(created_at DESC);

-- ML experiment metrics (time-series)
CREATE TABLE IF NOT EXISTS neuronip.ml_experiment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES neuronip.ml_experiments(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    step INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_experiment_metrics_experiment ON neuronip.ml_experiment_metrics(experiment_id);
CREATE INDEX idx_ml_experiment_metrics_name ON neuronip.ml_experiment_metrics(metric_name);
CREATE INDEX idx_ml_experiment_metrics_timestamp ON neuronip.ml_experiment_metrics(timestamp);

-- Model deployments
CREATE TABLE IF NOT EXISTS neuronip.model_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id),
    model_version VARCHAR(50) NOT NULL,
    deployment_name VARCHAR(255) NOT NULL,
    config JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'deploying', -- 'deploying', 'active', 'inactive', 'failed'
    endpoint_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_deployments_model ON neuronip.model_deployments(model_id);
CREATE INDEX idx_model_deployments_status ON neuronip.model_deployments(status);

-- Model A/B tests
CREATE TABLE IF NOT EXISTS neuronip.model_ab_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    model_a_id UUID NOT NULL REFERENCES neuronip.model_registry(id),
    model_b_id UUID NOT NULL REFERENCES neuronip.model_registry(id),
    traffic_split JSONB NOT NULL DEFAULT '{"a": 50, "b": 50}',
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'running', 'completed', 'cancelled'
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    results JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_ab_tests_status ON neuronip.model_ab_tests(status);

COMMENT ON TABLE neuronip.ml_pipelines IS 'ML training pipeline definitions';
COMMENT ON TABLE neuronip.ml_pipeline_executions IS 'ML pipeline execution history';
COMMENT ON TABLE neuronip.ml_experiments IS 'ML experiment tracking';
COMMENT ON TABLE neuronip.ml_experiment_metrics IS 'Time-series experiment metrics';
COMMENT ON TABLE neuronip.model_deployments IS 'Model deployment configurations';
COMMENT ON TABLE neuronip.model_ab_tests IS 'Model A/B testing experiments';
