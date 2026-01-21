-- Migration: Pipeline Versioning
-- Description: Adds versioned, replayable chunking and embedding pipelines

-- Pipelines: Versioned pipeline configurations
CREATE TABLE IF NOT EXISTS neuronip.pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    chunking_config JSONB NOT NULL,
    embedding_model TEXT NOT NULL,
    embedding_config JSONB DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT false,
    performance_metrics JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by TEXT,
    UNIQUE(id, version)
);
COMMENT ON TABLE neuronip.pipelines IS 'Versioned chunking and embedding pipelines';

CREATE INDEX IF NOT EXISTS idx_pipelines_name ON neuronip.pipelines(name);
CREATE INDEX IF NOT EXISTS idx_pipelines_active ON neuronip.pipelines(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_pipelines_created ON neuronip.pipelines(created_at DESC);

-- Pipeline replays: Track document reprocessing with new pipelines
CREATE TABLE IF NOT EXISTS neuronip.pipeline_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_ids UUID[] NOT NULL,
    old_pipeline_id UUID REFERENCES neuronip.pipelines(id) ON DELETE SET NULL,
    old_version TEXT,
    new_pipeline_id UUID NOT NULL REFERENCES neuronip.pipelines(id) ON DELETE CASCADE,
    new_version TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')) DEFAULT 'pending',
    documents_processed INTEGER DEFAULT 0,
    documents_total INTEGER NOT NULL,
    error_message TEXT,
    replayed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.pipeline_replays IS 'Document reprocessing with new pipeline versions';

CREATE INDEX IF NOT EXISTS idx_pipeline_replays_new_pipeline ON neuronip.pipeline_replays(new_pipeline_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_replays_status ON neuronip.pipeline_replays(status);
CREATE INDEX IF NOT EXISTS idx_pipeline_replays_replayed ON neuronip.pipeline_replays(replayed_at DESC);

-- Pipeline A/B tests: A/B testing for pipeline versions
CREATE TABLE IF NOT EXISTS neuronip.pipeline_ab_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    pipeline_a_id UUID NOT NULL REFERENCES neuronip.pipelines(id) ON DELETE CASCADE,
    pipeline_a_version TEXT NOT NULL,
    pipeline_b_id UUID NOT NULL REFERENCES neuronip.pipelines(id) ON DELETE CASCADE,
    pipeline_b_version TEXT NOT NULL,
    traffic_split JSONB DEFAULT '{"a": 50, "b": 50}',
    status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'completed', 'cancelled')) DEFAULT 'draft',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    results JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.pipeline_ab_tests IS 'A/B testing experiments for pipeline versions';

CREATE INDEX IF NOT EXISTS idx_pipeline_ab_tests_status ON neuronip.pipeline_ab_tests(status);
CREATE INDEX IF NOT EXISTS idx_pipeline_ab_tests_created ON neuronip.pipeline_ab_tests(created_at DESC);

-- Document pipeline tracking: Track which pipeline version processed each document
ALTER TABLE neuronip.knowledge_documents ADD COLUMN IF NOT EXISTS pipeline_id UUID REFERENCES neuronip.pipelines(id) ON DELETE SET NULL;
ALTER TABLE neuronip.knowledge_documents ADD COLUMN IF NOT EXISTS pipeline_version TEXT;
ALTER TABLE neuronip.knowledge_documents ADD COLUMN IF NOT EXISTS processed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_documents_pipeline ON neuronip.knowledge_documents(pipeline_id, pipeline_version);

-- Embedding pipeline tracking: Track pipeline version for embeddings
ALTER TABLE neuronip.knowledge_embeddings ADD COLUMN IF NOT EXISTS pipeline_id UUID REFERENCES neuronip.pipelines(id) ON DELETE SET NULL;
ALTER TABLE neuronip.knowledge_embeddings ADD COLUMN IF NOT EXISTS pipeline_version TEXT;

CREATE INDEX IF NOT EXISTS idx_embeddings_pipeline ON neuronip.knowledge_embeddings(pipeline_id, pipeline_version);

-- Update trigger for pipeline_ab_tests
CREATE OR REPLACE FUNCTION neuronip.update_pipeline_ab_tests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_pipeline_ab_tests_updated_at ON neuronip.pipeline_ab_tests;
CREATE TRIGGER trigger_update_pipeline_ab_tests_updated_at
    BEFORE UPDATE ON neuronip.pipeline_ab_tests
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_pipeline_ab_tests_updated_at();
