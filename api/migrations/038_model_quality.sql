-- Migration: Model Quality Schema
-- Description: Adds tables for model output quality scoring

-- Model quality scores: Quality scores for model outputs
CREATE TABLE IF NOT EXISTS neuronip.model_quality_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    model_version TEXT NOT NULL,
    output_id UUID,
    score NUMERIC NOT NULL CHECK (score >= 0 AND score <= 1),
    score_components JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evaluated_by TEXT
);
COMMENT ON TABLE neuronip.model_quality_scores IS 'Model output quality scores';

CREATE INDEX IF NOT EXISTS idx_quality_scores_model ON neuronip.model_quality_scores(model_id);
CREATE INDEX IF NOT EXISTS idx_quality_scores_version ON neuronip.model_quality_scores(model_version);
CREATE INDEX IF NOT EXISTS idx_quality_scores_score ON neuronip.model_quality_scores(score DESC);
CREATE INDEX IF NOT EXISTS idx_quality_scores_evaluated_at ON neuronip.model_quality_scores(evaluated_at DESC);
