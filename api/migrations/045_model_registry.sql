-- Migration: 045_model_registry.sql
-- Description: Enhanced model registry and governance (extends 039_model_governance.sql)

-- Model artifacts storage
CREATE TABLE IF NOT EXISTS neuronip.model_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.models(id),
    artifact_type VARCHAR(50) NOT NULL, -- 'model_file', 'weights', 'config', 'metadata', 'checkpoint'
    storage_type VARCHAR(50) NOT NULL, -- 's3', 'gcs', 'azure_blob', 'local'
    storage_path TEXT NOT NULL,
    storage_config JSONB, -- Storage-specific configuration
    file_size BIGINT,
    checksum VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_artifacts_model ON neuronip.model_artifacts(model_id);
CREATE INDEX idx_model_artifacts_type ON neuronip.model_artifacts(artifact_type);

-- Model versioning (enhanced)
CREATE TABLE IF NOT EXISTS neuronip.model_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.models(id),
    version_number VARCHAR(50) NOT NULL,
    version_type VARCHAR(20) NOT NULL DEFAULT 'patch', -- 'major', 'minor', 'patch'
    parent_version_id UUID REFERENCES neuronip.model_versions(id),
    changelog TEXT,
    artifacts JSONB, -- References to model_artifacts
    metadata JSONB,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, version_number)
);

CREATE INDEX idx_model_versions_model ON neuronip.model_versions(model_id);
CREATE INDEX idx_model_versions_parent ON neuronip.model_versions(parent_version_id);

-- Model deployment tracking
CREATE TABLE IF NOT EXISTS neuronip.model_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_version_id UUID NOT NULL REFERENCES neuronip.model_versions(id),
    deployment_name VARCHAR(255) NOT NULL,
    environment VARCHAR(50) NOT NULL, -- 'development', 'staging', 'production'
    deployment_config JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'deploying', 'active', 'failed', 'retired'
    endpoint_url TEXT,
    replicas INTEGER DEFAULT 1,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    retired_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_deployments_version ON neuronip.model_deployments(model_version_id);
CREATE INDEX idx_model_deployments_status ON neuronip.model_deployments(status);
CREATE INDEX idx_model_deployments_environment ON neuronip.model_deployments(environment);

-- Model performance monitoring
CREATE TABLE IF NOT EXISTS neuronip.model_performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES neuronip.model_deployments(id),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_type VARCHAR(50), -- 'accuracy', 'latency', 'throughput', 'error_rate', 'custom'
    sample_size INTEGER,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_performance_deployment ON neuronip.model_performance_metrics(deployment_id);
CREATE INDEX idx_model_performance_name ON neuronip.model_performance_metrics(metric_name);
CREATE INDEX idx_model_performance_timestamp ON neuronip.model_performance_metrics(timestamp);

-- Model approval workflows
CREATE TABLE IF NOT EXISTS neuronip.model_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_version_id UUID NOT NULL REFERENCES neuronip.model_versions(id),
    approval_type VARCHAR(50) NOT NULL, -- 'deployment', 'promotion', 'retirement'
    requested_by VARCHAR(255) NOT NULL,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    comments TEXT
);

CREATE INDEX idx_model_approvals_version ON neuronip.model_approvals(model_version_id);
CREATE INDEX idx_model_approvals_status ON neuronip.model_approvals(status);

-- Model lineage tracking
CREATE TABLE IF NOT EXISTS neuronip.model_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_version_id UUID NOT NULL REFERENCES neuronip.model_versions(id),
    parent_model_version_id UUID REFERENCES neuronip.model_versions(id),
    relationship_type VARCHAR(50) NOT NULL, -- 'trained_from', 'fine_tuned_from', 'ensemble_of', 'derived_from'
    training_run_id UUID REFERENCES neuronip.ml_training_runs(id),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_lineage_version ON neuronip.model_lineage(model_version_id);
CREATE INDEX idx_model_lineage_parent ON neuronip.model_lineage(parent_model_version_id);

-- Model tags
CREATE TABLE IF NOT EXISTS neuronip.model_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.models(id),
    tag_key VARCHAR(100) NOT NULL,
    tag_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, tag_key)
);

CREATE INDEX idx_model_tags_model ON neuronip.model_tags(model_id);
CREATE INDEX idx_model_tags_key ON neuronip.model_tags(tag_key);

COMMENT ON TABLE neuronip.model_artifacts IS 'Model artifact storage locations';
COMMENT ON TABLE neuronip.model_versions IS 'Model versioning and lineage';
COMMENT ON TABLE neuronip.model_deployments IS 'Model deployment tracking';
COMMENT ON TABLE neuronip.model_performance_metrics IS 'Model performance monitoring';
COMMENT ON TABLE neuronip.model_approvals IS 'Model approval workflows';
COMMENT ON TABLE neuronip.model_lineage IS 'Model lineage and relationships';
COMMENT ON TABLE neuronip.model_tags IS 'Model tags for organization';
