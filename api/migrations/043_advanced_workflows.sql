-- Migration: 043_advanced_workflows.sql
-- Description: Advanced workflow capabilities (extends existing workflow tables)

-- Workflow state snapshots for recovery
CREATE TABLE IF NOT EXISTS neuronip.workflow_state_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id),
    step_id VARCHAR(255) NOT NULL,
    state_data JSONB NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL DEFAULT 'checkpoint', -- 'checkpoint', 'error', 'manual'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_state_snapshots_execution ON neuronip.workflow_state_snapshots(execution_id);
CREATE INDEX idx_workflow_state_snapshots_step ON neuronip.workflow_state_snapshots(step_id);

-- Workflow compensation actions (for error recovery)
CREATE TABLE IF NOT EXISTS neuronip.workflow_compensations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id),
    step_id VARCHAR(255) NOT NULL,
    compensation_action JSONB NOT NULL, -- Action to undo/compensate for step
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_compensations_workflow ON neuronip.workflow_compensations(workflow_id);
CREATE INDEX idx_workflow_compensations_step ON neuronip.workflow_compensations(step_id);

-- Sub-workflow references
CREATE TABLE IF NOT EXISTS neuronip.workflow_sub_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id),
    child_workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id),
    step_id VARCHAR(255) NOT NULL,
    input_mapping JSONB, -- How to map parent workflow data to child workflow input
    output_mapping JSONB, -- How to map child workflow output back to parent
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_sub_workflows_parent ON neuronip.workflow_sub_workflows(parent_workflow_id);
CREATE INDEX idx_workflow_sub_workflows_child ON neuronip.workflow_sub_workflows(child_workflow_id);

COMMENT ON TABLE neuronip.workflow_state_snapshots IS 'Workflow execution state snapshots for recovery';
COMMENT ON TABLE neuronip.workflow_compensations IS 'Compensation actions for workflow error recovery';
COMMENT ON TABLE neuronip.workflow_sub_workflows IS 'Sub-workflow composition';
