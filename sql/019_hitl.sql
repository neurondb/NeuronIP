-- Migration: Human-in-the-Loop
-- Description: Adds approval queue, edit tracking, learning from edits

-- Approval queue: Queue for human approval of agent actions
CREATE TABLE IF NOT EXISTS neuronip.approval_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL, -- e.g., 'response', 'tool_call', 'workflow_step'
    action_data JSONB NOT NULL, -- The action to be approved
    original_response TEXT, -- Original agent response
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'edited')) DEFAULT 'pending',
    requested_by TEXT, -- User who triggered the action
    reviewed_by TEXT, -- User who reviewed/approved
    review_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.approval_queue IS 'Human-in-the-loop approval queue';

CREATE INDEX IF NOT EXISTS idx_approval_queue_agent ON neuronip.approval_queue(agent_id);
CREATE INDEX IF NOT EXISTS idx_approval_queue_status ON neuronip.approval_queue(status) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_approval_queue_created ON neuronip.approval_queue(created_at DESC);

-- Edit tracking: Track edits made to agent responses
CREATE TABLE IF NOT EXISTS neuronip.agent_edits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_id UUID NOT NULL REFERENCES neuronip.approval_queue(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    original_text TEXT NOT NULL,
    edited_text TEXT NOT NULL,
    edit_type TEXT CHECK (edit_type IN ('correction', 'improvement', 'clarification', 'other')),
    edited_by TEXT NOT NULL,
    edit_reason TEXT,
    learned BOOLEAN NOT NULL DEFAULT false, -- Whether the edit was learned
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_edits IS 'Edits made to agent responses';

CREATE INDEX IF NOT EXISTS idx_agent_edits_approval ON neuronip.agent_edits(approval_id);
CREATE INDEX IF NOT EXISTS idx_agent_edits_agent ON neuronip.agent_edits(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_edits_learned ON neuronip.agent_edits(learned) WHERE learned = false;

-- Learning feedback: Feedback loop for learning from edits
CREATE TABLE IF NOT EXISTS neuronip.learning_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    edit_id UUID REFERENCES neuronip.agent_edits(id) ON DELETE SET NULL,
    feedback_type TEXT NOT NULL CHECK (feedback_type IN ('positive', 'negative', 'correction')),
    feedback_data JSONB NOT NULL,
    applied_to_model BOOLEAN NOT NULL DEFAULT false,
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.learning_feedback IS 'Learning feedback from human edits';

CREATE INDEX IF NOT EXISTS idx_learning_feedback_agent ON neuronip.learning_feedback(agent_id);
CREATE INDEX IF NOT EXISTS idx_learning_feedback_applied ON neuronip.learning_feedback(applied_to_model) WHERE applied_to_model = false;

-- Approval workflows: Workflow definitions for approvals
CREATE TABLE IF NOT EXISTS neuronip.approval_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    agent_id UUID REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    workflow_config JSONB NOT NULL, -- Workflow steps and conditions
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.approval_workflows IS 'Approval workflow definitions';

CREATE INDEX IF NOT EXISTS idx_approval_workflows_agent ON neuronip.approval_workflows(agent_id);
CREATE INDEX IF NOT EXISTS idx_approval_workflows_enabled ON neuronip.approval_workflows(enabled) WHERE enabled = true;

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_approval_workflows_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_approval_workflows_updated_at ON neuronip.approval_workflows;
CREATE TRIGGER trigger_update_approval_workflows_updated_at
    BEFORE UPDATE ON neuronip.approval_workflows
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_approval_workflows_updated_at();
