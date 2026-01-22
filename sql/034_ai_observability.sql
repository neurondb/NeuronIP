-- Migration: AI Observability Enhancements
-- Description: Adds tables for retrieval metrics, hallucination detection, and agent execution logs

-- Retrieval metrics table
CREATE TABLE IF NOT EXISTS neuronip.retrieval_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID,
    agent_run_id UUID,
    retrieval_type TEXT NOT NULL, -- semantic, keyword, hybrid
    documents_retrieved INT NOT NULL,
    documents_used INT NOT NULL,
    hit_rate FLOAT NOT NULL, -- documents_used / documents_retrieved
    evidence_coverage FLOAT NOT NULL, -- percentage of query covered by evidence
    avg_similarity FLOAT,
    retrieval_latency_ms BIGINT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.retrieval_metrics IS 'Retrieval metrics for RAG and semantic search';

CREATE INDEX IF NOT EXISTS idx_retrieval_metrics_query ON neuronip.retrieval_metrics(query_id);
CREATE INDEX IF NOT EXISTS idx_retrieval_metrics_agent ON neuronip.retrieval_metrics(agent_run_id);
CREATE INDEX IF NOT EXISTS idx_retrieval_metrics_created ON neuronip.retrieval_metrics(created_at DESC);

-- Hallucination signals table
CREATE TABLE IF NOT EXISTS neuronip.hallucination_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID,
    agent_run_id UUID,
    response_id UUID,
    confidence_score FLOAT NOT NULL,
    risk_level TEXT NOT NULL CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    citation_accuracy FLOAT NOT NULL, -- 0-1, how accurate citations are
    evidence_strength FLOAT NOT NULL, -- 0-1, how strong supporting evidence is
    flags TEXT[] DEFAULT '{}', -- low_confidence, missing_citations, weak_evidence, etc.
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.hallucination_signals IS 'Hallucination risk signals for AI responses';

CREATE INDEX IF NOT EXISTS idx_hallucination_signals_query ON neuronip.hallucination_signals(query_id);
CREATE INDEX IF NOT EXISTS idx_hallucination_signals_agent ON neuronip.hallucination_signals(agent_run_id);
CREATE INDEX IF NOT EXISTS idx_hallucination_signals_risk ON neuronip.hallucination_signals(risk_level);
CREATE INDEX IF NOT EXISTS idx_hallucination_signals_created ON neuronip.hallucination_signals(created_at DESC);

-- Agent execution logs table
CREATE TABLE IF NOT EXISTS neuronip.agent_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    agent_run_id UUID NOT NULL,
    step_id TEXT,
    step_type TEXT NOT NULL, -- tool_call, reasoning, decision, etc.
    tool_name TEXT,
    input_data JSONB,
    output_data JSONB,
    decision TEXT,
    latency_ms BIGINT NOT NULL,
    tokens_used BIGINT,
    cost FLOAT,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_execution_logs IS 'Detailed agent execution logs with tool usage and decisions';

CREATE INDEX IF NOT EXISTS idx_agent_execution_logs_agent ON neuronip.agent_execution_logs(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_execution_logs_run ON neuronip.agent_execution_logs(agent_run_id);
CREATE INDEX IF NOT EXISTS idx_agent_execution_logs_timestamp ON neuronip.agent_execution_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_agent_execution_logs_step ON neuronip.agent_execution_logs(step_type);

-- Enhance cost_tracking to support per-query and per-agent-run costs
ALTER TABLE neuronip.cost_tracking 
ADD COLUMN IF NOT EXISTS tokens_used BIGINT;

CREATE INDEX IF NOT EXISTS idx_cost_tracking_resource ON neuronip.cost_tracking(resource_type, resource_id);
