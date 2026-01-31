-- Migration: 044_ml_lifecycle.sql
-- Description: ML lifecycle and training pipeline management

-- ML experiments table
CREATE TABLE IF NOT EXISTS neuronip.ml_experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'running', -- 'running', 'completed', 'failed', 'cancelled'
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_experiments_status ON neuronip.ml_experiments(status);
CREATE INDEX idx_ml_experiments_created_by ON neuronip.ml_experiments(created_by);

-- ML training runs
CREATE TABLE IF NOT EXISTS neuronip.ml_training_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES neuronip.ml_experiments(id),
    run_name VARCHAR(255) NOT NULL,
    model_type VARCHAR(100) NOT NULL,
    hyperparameters JSONB NOT NULL,
    training_config JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    metrics JSONB,
    artifacts JSONB, -- Model artifact locations
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_training_runs_experiment ON neuronip.ml_training_runs(experiment_id);
CREATE INDEX idx_ml_training_runs_status ON neuronip.ml_training_runs(status);
CREATE INDEX idx_ml_training_runs_model_type ON neuronip.ml_training_runs(model_type);

-- Training pipeline definitions
CREATE TABLE IF NOT EXISTS neuronip.ml_training_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_name VARCHAR(255) NOT NULL,
    description TEXT,
    pipeline_steps JSONB NOT NULL, -- Array of pipeline steps
    schedule_config JSONB, -- Optional scheduling
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_training_pipelines_enabled ON neuronip.ml_training_pipelines(enabled) WHERE enabled = true;

-- Training pipeline executions
CREATE TABLE IF NOT EXISTS neuronip.ml_pipeline_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.ml_training_pipelines(id),
    experiment_id UUID REFERENCES neuronip.ml_experiments(id),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    execution_config JSONB,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_pipeline_executions_pipeline ON neuronip.ml_pipeline_executions(pipeline_id);
CREATE INDEX idx_ml_pipeline_executions_status ON neuronip.ml_pipeline_executions(status);

-- Training metrics
CREATE TABLE IF NOT EXISTS neuronip.ml_training_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    training_run_id UUID NOT NULL REFERENCES neuronip.ml_training_runs(id),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_type VARCHAR(50), -- 'loss', 'accuracy', 'precision', 'recall', 'f1', 'custom'
    epoch INTEGER,
    step INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_training_metrics_run ON neuronip.ml_training_metrics(training_run_id);
CREATE INDEX idx_ml_training_metrics_name ON neuronip.ml_training_metrics(metric_name);
CREATE INDEX idx_ml_training_metrics_timestamp ON neuronip.ml_training_metrics(timestamp);

-- Hyperparameter tuning jobs
CREATE TABLE IF NOT EXISTS neuronip.ml_hyperparameter_tuning (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES neuronip.ml_experiments(id),
    tuning_strategy VARCHAR(50) NOT NULL, -- 'grid_search', 'random_search', 'bayesian', 'evolutionary'
    search_space JSONB NOT NULL,
    objective_metric VARCHAR(100) NOT NULL,
    maximize_objective BOOLEAN NOT NULL DEFAULT true,
    max_trials INTEGER NOT NULL DEFAULT 10,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    best_trial_id UUID,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ml_hyperparameter_tuning_experiment ON neuronip.ml_hyperparameter_tuning(experiment_id);
CREATE INDEX idx_ml_hyperparameter_tuning_status ON neuronip.ml_hyperparameter_tuning(status);

COMMENT ON TABLE neuronip.ml_experiments IS 'ML experiment tracking';
COMMENT ON TABLE neuronip.ml_training_runs IS 'Individual model training runs';
COMMENT ON TABLE neuronip.ml_training_pipelines IS 'Training pipeline definitions';
COMMENT ON TABLE neuronip.ml_pipeline_executions IS 'Training pipeline execution history';
COMMENT ON TABLE neuronip.ml_training_metrics IS 'Training metrics over time';
COMMENT ON TABLE neuronip.ml_hyperparameter_tuning IS 'Hyperparameter tuning jobs';
