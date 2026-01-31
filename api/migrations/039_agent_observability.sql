-- Migration: Agent Observability Schema
-- Description: Adds tables for agent execution traces, evidence coverage, and hallucination risk

-- Agent traces: Execution traces for agents
CREATE TABLE IF NOT EXISTS neuronip.agent_traces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    session_id TEXT,
    task TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed', 'cancelled')) DEFAULT 'running',
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    duration_ms BIGINT,
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.agent_traces IS 'Agent execution traces';

CREATE INDEX IF NOT EXISTS idx_agent_traces_agent ON neuronip.agent_traces(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_traces_session ON neuronip.agent_traces(session_id);
CREATE INDEX IF NOT EXISTS idx_agent_traces_status ON neuronip.agent_traces(status);
CREATE INDEX IF NOT EXISTS idx_agent_traces_start_time ON neuronip.agent_traces(start_time DESC);

-- Agent trace steps: Individual steps in an execution trace
CREATE TABLE IF NOT EXISTS neuronip.agent_trace_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL REFERENCES neuronip.agent_traces(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    step_type TEXT NOT NULL CHECK (step_type IN ('tool_call', 'reasoning', 'response', 'error')),
    description TEXT NOT NULL,
    input_data JSONB DEFAULT '{}',
    output_data JSONB DEFAULT '{}',
    tool_name TEXT,
    duration BIGINT NOT NULL DEFAULT 0,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_message TEXT
);
COMMENT ON TABLE neuronip.agent_trace_steps IS 'Individual steps in agent execution traces';

CREATE INDEX IF NOT EXISTS idx_trace_steps_trace ON neuronip.agent_trace_steps(trace_id);
CREATE INDEX IF NOT EXISTS idx_trace_steps_number ON neuronip.agent_trace_steps(trace_id, step_number);

-- Agent evidence coverage: Evidence coverage for agent responses
CREATE TABLE IF NOT EXISTS neuronip.agent_evidence_coverage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL REFERENCES neuronip.agent_traces(id) ON DELETE CASCADE,
    response_id UUID,
    total_claims INTEGER NOT NULL DEFAULT 0,
    supported_claims INTEGER NOT NULL DEFAULT 0,
    unsupported_claims INTEGER NOT NULL DEFAULT 0,
    coverage_score NUMERIC NOT NULL CHECK (coverage_score >= 0 AND coverage_score <= 1),
    evidence_sources JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_evidence_coverage IS 'Evidence coverage metrics for agent responses';

CREATE INDEX IF NOT EXISTS idx_evidence_coverage_trace ON neuronip.agent_evidence_coverage(trace_id);
CREATE INDEX IF NOT EXISTS idx_evidence_coverage_score ON neuronip.agent_evidence_coverage(coverage_score DESC);

-- Agent hallucination risks: Hallucination risk assessments
CREATE TABLE IF NOT EXISTS neuronip.agent_hallucination_risks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id UUID NOT NULL REFERENCES neuronip.agent_traces(id) ON DELETE CASCADE,
    response_id UUID,
    risk_score NUMERIC NOT NULL CHECK (risk_score >= 0 AND risk_score <= 1),
    risk_factors JSONB DEFAULT '[]',
    confidence NUMERIC NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    recommendation TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_hallucination_risks IS 'Hallucination risk assessments for agent responses';

CREATE INDEX IF NOT EXISTS idx_hallucination_risks_trace ON neuronip.agent_hallucination_risks(trace_id);
CREATE INDEX IF NOT EXISTS idx_hallucination_risks_score ON neuronip.agent_hallucination_risks(risk_score DESC);
