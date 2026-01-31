-- Migration: Data Quality Checks Schema
-- Description: Adds tables for data quality rules and validation results

-- Quality rules: Data quality validation rules
CREATE TABLE IF NOT EXISTS neuronip.quality_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('not_null', 'unique', 'range', 'regex', 'format', 'custom')),
    table_name TEXT NOT NULL,
    column_name TEXT,
    rule_config JSONB DEFAULT '{}',
    severity TEXT NOT NULL CHECK (severity IN ('error', 'warning', 'info')) DEFAULT 'error',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.quality_rules IS 'Data quality validation rules';

CREATE INDEX IF NOT EXISTS idx_quality_rules_table ON neuronip.quality_rules(table_name);
CREATE INDEX IF NOT EXISTS idx_quality_rules_enabled ON neuronip.quality_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_quality_rules_type ON neuronip.quality_rules(rule_type);

-- Quality validation results: Results of quality checks
CREATE TABLE IF NOT EXISTS neuronip.quality_validation_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES neuronip.quality_rules(id) ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    passed BOOLEAN NOT NULL,
    message TEXT,
    severity TEXT NOT NULL CHECK (severity IN ('error', 'warning', 'info')),
    details JSONB DEFAULT '{}',
    data_sample JSONB,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.quality_validation_results IS 'Data quality validation results';

CREATE INDEX IF NOT EXISTS idx_quality_results_rule ON neuronip.quality_validation_results(rule_id);
CREATE INDEX IF NOT EXISTS idx_quality_results_table ON neuronip.quality_validation_results(table_name);
CREATE INDEX IF NOT EXISTS idx_quality_results_passed ON neuronip.quality_validation_results(passed);
CREATE INDEX IF NOT EXISTS idx_quality_results_checked_at ON neuronip.quality_validation_results(checked_at DESC);

-- Quality metrics: Aggregated quality metrics
CREATE TABLE IF NOT EXISTS neuronip.quality_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    rule_id UUID REFERENCES neuronip.quality_rules(id) ON DELETE SET NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.quality_metrics IS 'Aggregated data quality metrics';

CREATE INDEX IF NOT EXISTS idx_quality_metrics_table ON neuronip.quality_metrics(table_name);
CREATE INDEX IF NOT EXISTS idx_quality_metrics_period ON neuronip.quality_metrics(period_start, period_end);
