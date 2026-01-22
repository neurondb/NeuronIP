-- Migration: Semantic Layer Enhancements
-- Description: Adds metric approval workflows, time grains, default filters, and glossary linkage

-- Metric Approvals: Approval workflow for metrics
CREATE TABLE IF NOT EXISTS neuronip.metric_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    approver_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'changes_requested')) DEFAULT 'pending',
    comments TEXT,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metric_approvals IS 'Metric approval workflow';

CREATE INDEX IF NOT EXISTS idx_metric_approvals_metric ON neuronip.metric_approvals(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_approvals_status ON neuronip.metric_approvals(status);

-- Time Grains: Time grain definitions for metrics
CREATE TABLE IF NOT EXISTS neuronip.time_grains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    duration TEXT NOT NULL, -- e.g., '1 hour', '1 day', '1 week', '1 month'
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.time_grains IS 'Time grain definitions';

-- Insert default time grains
INSERT INTO neuronip.time_grains (name, display_name, duration, is_default) VALUES
    ('hour', 'Hourly', '1 hour', false),
    ('day', 'Daily', '1 day', true),
    ('week', 'Weekly', '1 week', false),
    ('month', 'Monthly', '1 month', false)
ON CONFLICT (name) DO NOTHING;

-- Metric Filters: Default filters for metrics
CREATE TABLE IF NOT EXISTS neuronip.metric_filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    filter_expression TEXT NOT NULL,
    filter_name TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metric_filters IS 'Default filters for metrics';

CREATE INDEX IF NOT EXISTS idx_metric_filters_metric ON neuronip.metric_filters(metric_id);

-- Metric Glossary Links: Link metrics to glossary terms
CREATE TABLE IF NOT EXISTS neuronip.metric_glossary_links (
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    glossary_term_id UUID NOT NULL REFERENCES neuronip.glossary_terms(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (metric_id, glossary_term_id)
);
COMMENT ON TABLE neuronip.metric_glossary_links IS 'Links between metrics and glossary terms';

CREATE INDEX IF NOT EXISTS idx_metric_glossary_links_metric ON neuronip.metric_glossary_links(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_glossary_links_term ON neuronip.metric_glossary_links(glossary_term_id);

-- Enhance business_metrics table with ownership and version tracking
ALTER TABLE neuronip.business_metrics 
    ADD COLUMN IF NOT EXISTS owner_id TEXT,
    ADD COLUMN IF NOT EXISTS version TEXT NOT NULL DEFAULT '1.0.0',
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL CHECK (status IN ('draft', 'pending_approval', 'approved', 'deprecated')) DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS approval_required BOOLEAN NOT NULL DEFAULT true;

CREATE INDEX IF NOT EXISTS idx_business_metrics_status ON neuronip.business_metrics(status);

-- Metric Time Grains: Many-to-many relationship between metrics and time grains
CREATE TABLE IF NOT EXISTS neuronip.metric_time_grains (
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    time_grain_id UUID NOT NULL REFERENCES neuronip.time_grains(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (metric_id, time_grain_id)
);
COMMENT ON TABLE neuronip.metric_time_grains IS 'Metric-time grain relationships';

-- Update triggers for updated_at
CREATE OR REPLACE FUNCTION neuronip.update_metric_approvals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_metric_approvals_updated_at ON neuronip.metric_approvals;
CREATE TRIGGER trigger_update_metric_approvals_updated_at
    BEFORE UPDATE ON neuronip.metric_approvals
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_metric_approvals_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_time_grains_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_time_grains_updated_at ON neuronip.time_grains;
CREATE TRIGGER trigger_update_time_grains_updated_at
    BEFORE UPDATE ON neuronip.time_grains
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_time_grains_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_metric_filters_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_metric_filters_updated_at ON neuronip.metric_filters;
CREATE TRIGGER trigger_update_metric_filters_updated_at
    BEFORE UPDATE ON neuronip.metric_filters
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_metric_filters_updated_at();
