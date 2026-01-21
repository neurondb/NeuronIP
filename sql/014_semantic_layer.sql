-- Migration: Semantic Layer Enhancement
-- Description: Adds metrics, dimensions, definitions, owners, lineage

-- Business metrics: Metric definitions with owners and documentation
CREATE TABLE IF NOT EXISTS neuronip.business_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    metric_type TEXT NOT NULL CHECK (metric_type IN ('sum', 'avg', 'count', 'min', 'max', 'custom')),
    formula TEXT, -- SQL formula or expression
    unit TEXT, -- e.g., 'USD', 'count', 'percentage'
    owner_id TEXT, -- User ID of metric owner
    tags TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.business_metrics IS 'Business metrics definitions';

CREATE INDEX IF NOT EXISTS idx_business_metrics_name ON neuronip.business_metrics(name);
CREATE INDEX IF NOT EXISTS idx_business_metrics_owner ON neuronip.business_metrics(owner_id);
CREATE INDEX IF NOT EXISTS idx_business_metrics_tags ON neuronip.business_metrics USING gin(tags);

-- Dimensions: Dimension definitions for metrics
CREATE TABLE IF NOT EXISTS neuronip.dimensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    dimension_type TEXT NOT NULL CHECK (dimension_type IN ('categorical', 'temporal', 'geographic', 'numeric')),
    hierarchy JSONB, -- Dimension hierarchy
    owner_id TEXT,
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.dimensions IS 'Dimension definitions';

CREATE INDEX IF NOT EXISTS idx_dimensions_name ON neuronip.dimensions(name);
CREATE INDEX IF NOT EXISTS idx_dimensions_owner ON neuronip.dimensions(owner_id);

-- Metric dimensions: Many-to-many relationship between metrics and dimensions
CREATE TABLE IF NOT EXISTS neuronip.metric_dimensions (
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    dimension_id UUID NOT NULL REFERENCES neuronip.dimensions(id) ON DELETE CASCADE,
    is_required BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (metric_id, dimension_id)
);
COMMENT ON TABLE neuronip.metric_dimensions IS 'Metric-dimension relationships';

-- Metric definitions: Versioned metric definitions
CREATE TABLE IF NOT EXISTS neuronip.metric_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    definition_data JSONB NOT NULL, -- Full metric definition
    changelog TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(metric_id, version)
);
COMMENT ON TABLE neuronip.metric_definitions IS 'Versioned metric definitions';

CREATE INDEX IF NOT EXISTS idx_metric_definitions_metric ON neuronip.metric_definitions(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_definitions_created ON neuronip.metric_definitions(created_at DESC);

-- Metric lineage: Track metric dependencies and lineage
CREATE TABLE IF NOT EXISTS neuronip.metric_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    depends_on_metric_id UUID REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    depends_on_dataset_id UUID, -- Reference to dataset
    dependency_type TEXT NOT NULL CHECK (dependency_type IN ('metric', 'dataset', 'table', 'column')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metric_lineage IS 'Metric dependency and lineage tracking';

CREATE INDEX IF NOT EXISTS idx_metric_lineage_metric ON neuronip.metric_lineage(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_lineage_depends ON neuronip.metric_lineage(depends_on_metric_id);

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_business_metrics_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_business_metrics_updated_at ON neuronip.business_metrics;
CREATE TRIGGER trigger_update_business_metrics_updated_at
    BEFORE UPDATE ON neuronip.business_metrics
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_business_metrics_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_dimensions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_dimensions_updated_at ON neuronip.dimensions;
CREATE TRIGGER trigger_update_dimensions_updated_at
    BEFORE UPDATE ON neuronip.dimensions
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_dimensions_updated_at();
