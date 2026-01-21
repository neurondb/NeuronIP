-- Migration: Metadata & Semantic Layer Schema
-- Description: Adds tables for metric catalog, lineage, and semantic definitions

-- Metrics catalog: Central metric definitions
CREATE TABLE IF NOT EXISTS neuronip.metric_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    sql_expression TEXT NOT NULL,
    metric_type TEXT NOT NULL CHECK (metric_type IN ('count', 'sum', 'avg', 'min', 'max', 'custom')),
    unit TEXT,
    category TEXT,
    tags TEXT[] DEFAULT '{}',
    owner_id TEXT,
    version TEXT NOT NULL DEFAULT '1.0.0',
    status TEXT NOT NULL CHECK (status IN ('draft', 'approved', 'deprecated')) DEFAULT 'draft',
    approval_workflow JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    approved_by TEXT
);
COMMENT ON TABLE neuronip.metric_catalog IS 'Central metric catalog with SQL expressions';

CREATE INDEX IF NOT EXISTS idx_metric_catalog_name ON neuronip.metric_catalog(name);
CREATE INDEX IF NOT EXISTS idx_metric_catalog_status ON neuronip.metric_catalog(status);
CREATE INDEX IF NOT EXISTS idx_metric_catalog_category ON neuronip.metric_catalog(category);

-- Metric lineage: Track metric dependencies
CREATE TABLE IF NOT EXISTS neuronip.metric_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.metric_catalog(id) ON DELETE CASCADE,
    depends_on_metric_id UUID REFERENCES neuronip.metric_catalog(id) ON DELETE CASCADE,
    depends_on_table TEXT,
    depends_on_column TEXT,
    relationship_type TEXT NOT NULL CHECK (relationship_type IN ('uses', 'derived_from', 'aggregates', 'filters')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metric_lineage IS 'Metric dependencies and lineage';

CREATE INDEX IF NOT EXISTS idx_metric_lineage_metric ON neuronip.metric_lineage(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_lineage_depends_on ON neuronip.metric_lineage(depends_on_metric_id);

-- Semantic definitions: Business logic definitions
CREATE TABLE IF NOT EXISTS neuronip.semantic_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    term TEXT NOT NULL,
    definition TEXT NOT NULL,
    category TEXT,
    sql_expression TEXT,
    ai_model_mapping JSONB,
    synonyms TEXT[] DEFAULT '{}',
    related_terms TEXT[] DEFAULT '{}',
    examples JSONB DEFAULT '[]',
    owner_id TEXT,
    version TEXT NOT NULL DEFAULT '1.0.0',
    status TEXT NOT NULL CHECK (status IN ('draft', 'approved', 'deprecated')) DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(term, version)
);
COMMENT ON TABLE neuronip.semantic_definitions IS 'Business logic and semantic definitions';

CREATE INDEX IF NOT EXISTS idx_semantic_definitions_term ON neuronip.semantic_definitions(term);
CREATE INDEX IF NOT EXISTS idx_semantic_definitions_category ON neuronip.semantic_definitions(category);

-- Extend existing glossary table if needed
ALTER TABLE neuronip.glossary ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE neuronip.glossary ADD COLUMN IF NOT EXISTS related_entity_id UUID REFERENCES neuronip.entities(id) ON DELETE SET NULL;
