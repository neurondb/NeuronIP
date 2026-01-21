-- Migration: Data Quality Rules Engine
-- Description: Adds tables for data quality rules, checks, and scoring

-- Data Quality Rules: Quality rules definitions
CREATE TABLE IF NOT EXISTS neuronip.data_quality_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('completeness', 'accuracy', 'consistency', 'validity', 'uniqueness', 'timeliness', 'custom')),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT,
    table_name TEXT,
    column_name TEXT,
    rule_expression TEXT NOT NULL,
    threshold NUMERIC,
    enabled BOOLEAN NOT NULL DEFAULT true,
    schedule_cron TEXT,
    metadata JSONB DEFAULT '{}',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_quality_rules IS 'Data quality rules definitions';

CREATE INDEX IF NOT EXISTS idx_quality_rules_type ON neuronip.data_quality_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_quality_rules_connector ON neuronip.data_quality_rules(connector_id);
CREATE INDEX IF NOT EXISTS idx_quality_rules_table ON neuronip.data_quality_rules(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_quality_rules_enabled ON neuronip.data_quality_rules(enabled);

-- Data Quality Checks: Execution results of quality rules
CREATE TABLE IF NOT EXISTS neuronip.data_quality_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES neuronip.data_quality_rules(id) ON DELETE CASCADE,
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT,
    table_name TEXT,
    column_name TEXT,
    status TEXT NOT NULL CHECK (status IN ('pass', 'fail', 'warning', 'error')),
    score NUMERIC NOT NULL CHECK (score >= 0 AND score <= 100),
    passed_count BIGINT DEFAULT 0,
    failed_count BIGINT DEFAULT 0,
    total_count BIGINT DEFAULT 0,
    error_message TEXT,
    execution_time_ms INTEGER,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.data_quality_checks IS 'Data quality check execution results';

CREATE INDEX IF NOT EXISTS idx_quality_checks_rule ON neuronip.data_quality_checks(rule_id);
CREATE INDEX IF NOT EXISTS idx_quality_checks_connector ON neuronip.data_quality_checks(connector_id);
CREATE INDEX IF NOT EXISTS idx_quality_checks_table ON neuronip.data_quality_checks(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_quality_checks_status ON neuronip.data_quality_checks(status);
CREATE INDEX IF NOT EXISTS idx_quality_checks_executed ON neuronip.data_quality_checks(executed_at);

-- Data Quality Scores: Aggregated quality scores
CREATE TABLE IF NOT EXISTS neuronip.data_quality_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT,
    table_name TEXT,
    column_name TEXT,
    score NUMERIC NOT NULL CHECK (score >= 0 AND score <= 100),
    score_breakdown JSONB DEFAULT '{}',
    rule_count INTEGER DEFAULT 0,
    passed_rules INTEGER DEFAULT 0,
    failed_rules INTEGER DEFAULT 0,
    last_calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_quality_scores IS 'Aggregated data quality scores';

CREATE UNIQUE INDEX IF NOT EXISTS idx_quality_scores_unique ON neuronip.data_quality_scores(connector_id, schema_name, table_name, column_name);
CREATE INDEX IF NOT EXISTS idx_quality_scores_connector ON neuronip.data_quality_scores(connector_id);
CREATE INDEX IF NOT EXISTS idx_quality_scores_table ON neuronip.data_quality_scores(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_quality_scores_calculated ON neuronip.data_quality_scores(calculated_at);

-- Data Quality Violations: Individual violations from quality checks
CREATE TABLE IF NOT EXISTS neuronip.data_quality_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_id UUID NOT NULL REFERENCES neuronip.data_quality_checks(id) ON DELETE CASCADE,
    rule_id UUID NOT NULL REFERENCES neuronip.data_quality_rules(id) ON DELETE CASCADE,
    row_identifier TEXT,
    column_value TEXT,
    violation_type TEXT NOT NULL,
    violation_message TEXT NOT NULL,
    severity TEXT CHECK (severity IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.data_quality_violations IS 'Individual data quality violations';

CREATE INDEX IF NOT EXISTS idx_quality_violations_check ON neuronip.data_quality_violations(check_id);
CREATE INDEX IF NOT EXISTS idx_quality_violations_rule ON neuronip.data_quality_violations(rule_id);
CREATE INDEX IF NOT EXISTS idx_quality_violations_severity ON neuronip.data_quality_violations(severity);
CREATE INDEX IF NOT EXISTS idx_quality_violations_created ON neuronip.data_quality_violations(created_at);
