-- Migration: Agent Audit Trail
-- Description: Logs every agent action and tool call

-- Agent actions: Log all agent actions
CREATE TABLE IF NOT EXISTS neuronip.agent_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL, -- e.g., 'query', 'tool_call', 'response', 'decision'
    action_data JSONB NOT NULL, -- Action details
    input_data JSONB, -- Input to the action
    output_data JSONB, -- Output from the action
    reasoning TEXT, -- Decision reasoning
    user_id TEXT, -- User who triggered the action
    request_id TEXT, -- Request ID for correlation
    execution_time_ms INTEGER,
    token_usage INTEGER, -- Tokens used
    cost NUMERIC(10,4), -- Cost in USD
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_actions IS 'Complete audit trail of agent actions';

CREATE INDEX IF NOT EXISTS idx_agent_actions_agent ON neuronip.agent_actions(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_actions_type ON neuronip.agent_actions(action_type);
CREATE INDEX IF NOT EXISTS idx_agent_actions_user ON neuronip.agent_actions(user_id);
CREATE INDEX IF NOT EXISTS idx_agent_actions_request ON neuronip.agent_actions(request_id);
CREATE INDEX IF NOT EXISTS idx_agent_actions_created ON neuronip.agent_actions(created_at DESC);

-- Tool calls: Detailed tool call tracking
CREATE TABLE IF NOT EXISTS neuronip.tool_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_action_id UUID NOT NULL REFERENCES neuronip.agent_actions(id) ON DELETE CASCADE,
    tool_id UUID REFERENCES neuronip.tools_registry(id) ON DELETE SET NULL,
    tool_name TEXT NOT NULL,
    tool_input JSONB NOT NULL,
    tool_output JSONB,
    execution_time_ms INTEGER,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.tool_calls IS 'Detailed tool call tracking';

CREATE INDEX IF NOT EXISTS idx_tool_calls_action ON neuronip.tool_calls(agent_action_id);
CREATE INDEX IF NOT EXISTS idx_tool_calls_tool ON neuronip.tool_calls(tool_id);
CREATE INDEX IF NOT EXISTS idx_tool_calls_success ON neuronip.tool_calls(success);
CREATE INDEX IF NOT EXISTS idx_tool_calls_created ON neuronip.tool_calls(created_at DESC);

-- Agent performance: Performance metrics per action
CREATE TABLE IF NOT EXISTS neuronip.agent_performance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    action_id UUID REFERENCES neuronip.agent_actions(id) ON DELETE SET NULL,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC(10,4),
    metric_unit TEXT, -- e.g., 'ms', 'tokens', 'USD'
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_performance IS 'Performance metrics per agent action';

CREATE INDEX IF NOT EXISTS idx_agent_performance_agent ON neuronip.agent_performance(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_performance_metric ON neuronip.agent_performance(metric_name);
CREATE INDEX IF NOT EXISTS idx_agent_performance_recorded ON neuronip.agent_performance(recorded_at DESC);
