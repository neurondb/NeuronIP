-- Migration: Approval Workflows Schema
-- Description: Adds tables for multi-stage approval workflows

-- Approval workflows: Multi-stage approval workflows
CREATE TABLE IF NOT EXISTS neuronip.approval_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('model', 'prompt', 'metric', 'policy', 'workflow')),
    resource_id UUID NOT NULL,
    stages JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'approved', 'rejected', 'cancelled')) DEFAULT 'pending',
    current_stage INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.approval_workflows IS 'Multi-stage approval workflows';

CREATE INDEX IF NOT EXISTS idx_approval_workflows_resource ON neuronip.approval_workflows(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_approval_workflows_status ON neuronip.approval_workflows(status);

-- Approval workflow approvals: Individual approvals in workflow stages
CREATE TABLE IF NOT EXISTS neuronip.approval_workflow_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.approval_workflows(id) ON DELETE CASCADE,
    stage_number INTEGER NOT NULL,
    approver_id TEXT NOT NULL,
    decision TEXT NOT NULL CHECK (decision IN ('approve', 'reject', 'request_changes')),
    comments TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.approval_workflow_approvals IS 'Individual approvals in workflow stages';

CREATE INDEX IF NOT EXISTS idx_workflow_approvals_workflow ON neuronip.approval_workflow_approvals(workflow_id, stage_number);
CREATE INDEX IF NOT EXISTS idx_workflow_approvals_approver ON neuronip.approval_workflow_approvals(approver_id);
