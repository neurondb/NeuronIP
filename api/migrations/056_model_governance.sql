-- Migration: 056_model_governance.sql
-- Description: Model governance, compliance, and monitoring

-- Model approval workflows
CREATE TABLE IF NOT EXISTS neuronip.model_approval_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_name VARCHAR(255) NOT NULL,
    description TEXT,
    steps JSONB NOT NULL, -- Array of workflow steps
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_approval_workflows_enabled ON neuronip.model_approval_workflows(enabled) WHERE enabled = true;

-- Model compliance checks
CREATE TABLE IF NOT EXISTS neuronip.model_compliance_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'passed', 'failed')),
    results JSONB NOT NULL, -- Compliance check results
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_compliance_checks_model ON neuronip.model_compliance_checks(model_id);
CREATE INDEX idx_model_compliance_checks_status ON neuronip.model_compliance_checks(status);

-- Model predictions for monitoring
CREATE TABLE IF NOT EXISTS neuronip.model_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    features JSONB NOT NULL,
    prediction JSONB NOT NULL,
    actual_value JSONB,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_predictions_model ON neuronip.model_predictions(model_id);
CREATE INDEX idx_model_predictions_timestamp ON neuronip.model_predictions(timestamp DESC);

-- Model drift detections
CREATE TABLE IF NOT EXISTS neuronip.model_drift_detections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    drift_score DOUBLE PRECISION NOT NULL,
    baseline_mean DOUBLE PRECISION,
    current_mean DOUBLE PRECISION,
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_drift_detections_model ON neuronip.model_drift_detections(model_id);
CREATE INDEX idx_model_drift_detections_detected ON neuronip.model_drift_detections(detected_at DESC);

-- Model training data (for baseline comparison)
CREATE TABLE IF NOT EXISTS neuronip.model_training_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    features JSONB NOT NULL,
    target JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_training_data_model ON neuronip.model_training_data(model_id);

COMMENT ON TABLE neuronip.model_approval_workflows IS 'Model approval workflow definitions';
COMMENT ON TABLE neuronip.model_compliance_checks IS 'Model compliance check results';
COMMENT ON TABLE neuronip.model_predictions IS 'Model predictions for monitoring';
COMMENT ON TABLE neuronip.model_drift_detections IS 'Data drift detection results';
COMMENT ON TABLE neuronip.model_training_data IS 'Model training data for baseline comparison';
