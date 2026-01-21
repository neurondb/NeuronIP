-- Migration: Agent Evaluation
-- Description: Adds golden sets, regression tests, scoring, drift tracking

-- Golden sets: Test cases for agent evaluation
CREATE TABLE IF NOT EXISTS neuronip.golden_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    agent_id UUID REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    test_cases JSONB NOT NULL, -- Array of test cases
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.golden_sets IS 'Golden sets for agent evaluation';

CREATE INDEX IF NOT EXISTS idx_golden_sets_agent ON neuronip.golden_sets(agent_id);
CREATE INDEX IF NOT EXISTS idx_golden_sets_name ON neuronip.golden_sets(name);

-- Regression tests: Regression test suite
CREATE TABLE IF NOT EXISTS neuronip.regression_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    test_config JSONB NOT NULL,
    baseline_score NUMERIC(5,2), -- Baseline performance score
    current_score NUMERIC(5,2), -- Current performance score
    status TEXT NOT NULL CHECK (status IN ('passing', 'failing', 'warning')) DEFAULT 'passing',
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.regression_tests IS 'Regression test suite for agents';

CREATE INDEX IF NOT EXISTS idx_regression_tests_agent ON neuronip.regression_tests(agent_id);
CREATE INDEX IF NOT EXISTS idx_regression_tests_status ON neuronip.regression_tests(status);

-- Agent evaluations: Evaluation runs and results
CREATE TABLE IF NOT EXISTS neuronip.agent_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    golden_set_id UUID REFERENCES neuronip.golden_sets(id) ON DELETE SET NULL,
    evaluation_type TEXT NOT NULL CHECK (evaluation_type IN ('golden_set', 'regression', 'custom')),
    metrics JSONB NOT NULL, -- Evaluation metrics (accuracy, latency, cost, etc.)
    score NUMERIC(5,2), -- Overall score
    status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed')) DEFAULT 'running',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.agent_evaluations IS 'Agent evaluation runs and results';

CREATE INDEX IF NOT EXISTS idx_agent_evaluations_agent ON neuronip.agent_evaluations(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_evaluations_status ON neuronip.agent_evaluations(status);
CREATE INDEX IF NOT EXISTS idx_agent_evaluations_started ON neuronip.agent_evaluations(started_at DESC);

-- Drift tracking: Track performance drift over time
CREATE TABLE IF NOT EXISTS neuronip.agent_drift (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    metric_name TEXT NOT NULL,
    baseline_value NUMERIC(10,4),
    current_value NUMERIC(10,4),
    drift_percentage NUMERIC(5,2), -- Percentage change
    drift_type TEXT CHECK (drift_type IN ('improvement', 'degradation', 'stable')),
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_drift IS 'Agent performance drift tracking';

CREATE INDEX IF NOT EXISTS idx_agent_drift_agent ON neuronip.agent_drift(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_drift_metric ON neuronip.agent_drift(metric_name);
CREATE INDEX IF NOT EXISTS idx_agent_drift_detected ON neuronip.agent_drift(detected_at DESC);

-- Agent A/B tests: A/B testing for agents
CREATE TABLE IF NOT EXISTS neuronip.agent_ab_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    agent_a_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    agent_b_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    traffic_split JSONB DEFAULT '{"a": 50, "b": 50}',
    status TEXT NOT NULL CHECK (status IN ('draft', 'running', 'completed', 'cancelled')) DEFAULT 'draft',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    results JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_ab_tests IS 'A/B testing experiments for agents';

CREATE INDEX IF NOT EXISTS idx_agent_ab_tests_status ON neuronip.agent_ab_tests(status);
CREATE INDEX IF NOT EXISTS idx_agent_ab_tests_created ON neuronip.agent_ab_tests(created_at DESC);

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_golden_sets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_golden_sets_updated_at ON neuronip.golden_sets;
CREATE TRIGGER trigger_update_golden_sets_updated_at
    BEFORE UPDATE ON neuronip.golden_sets
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_golden_sets_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_regression_tests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_regression_tests_updated_at ON neuronip.regression_tests;
CREATE TRIGGER trigger_update_regression_tests_updated_at
    BEFORE UPDATE ON neuronip.regression_tests
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_regression_tests_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_agent_ab_tests_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_agent_ab_tests_updated_at ON neuronip.agent_ab_tests;
CREATE TRIGGER trigger_update_agent_ab_tests_updated_at
    BEFORE UPDATE ON neuronip.agent_ab_tests
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_agent_ab_tests_updated_at();
