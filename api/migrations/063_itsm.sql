-- Migration: ITSM
-- Description: Incidents, changes, problems, runbooks, automation policies

-- Incidents
CREATE TABLE IF NOT EXISTS neuronip.itsm_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    priority TEXT NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    status TEXT NOT NULL CHECK (status IN ('new', 'assigned', 'in_progress', 'pending', 'resolved', 'closed')) DEFAULT 'new',
    assignee_id TEXT,
    requester_id TEXT NOT NULL,
    runbook_id UUID,
    workflow_execution_id UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.itsm_incidents IS 'ITSM incidents';

CREATE INDEX IF NOT EXISTS idx_itsm_incidents_status ON neuronip.itsm_incidents(status);
CREATE INDEX IF NOT EXISTS idx_itsm_incidents_assignee ON neuronip.itsm_incidents(assignee_id);
CREATE INDEX IF NOT EXISTS idx_itsm_incidents_created ON neuronip.itsm_incidents(created_at DESC);

-- Changes
CREATE TABLE IF NOT EXISTS neuronip.itsm_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT,
    change_type TEXT NOT NULL CHECK (change_type IN ('standard', 'normal', 'emergency')) DEFAULT 'normal',
    status TEXT NOT NULL CHECK (status IN ('draft', 'pending_approval', 'approved', 'scheduled', 'in_progress', 'completed', 'cancelled')) DEFAULT 'draft',
    requester_id TEXT NOT NULL,
    approver_id TEXT,
    approved_at TIMESTAMPTZ,
    scheduled_start TIMESTAMPTZ,
    scheduled_end TIMESTAMPTZ,
    runbook_id UUID,
    workflow_execution_id UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.itsm_changes IS 'ITSM changes';

CREATE INDEX IF NOT EXISTS idx_itsm_changes_status ON neuronip.itsm_changes(status);
CREATE INDEX IF NOT EXISTS idx_itsm_changes_scheduled ON neuronip.itsm_changes(scheduled_start);

-- Runbooks (workflow-linked playbooks)
CREATE TABLE IF NOT EXISTS neuronip.itsm_runbooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    workflow_id UUID NOT NULL,
    trigger_conditions JSONB DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.itsm_runbooks IS 'ITSM runbooks (workflow-linked)';

CREATE INDEX IF NOT EXISTS idx_itsm_runbooks_workflow ON neuronip.itsm_runbooks(workflow_id);

-- Automation policies (event -> action)
CREATE TABLE IF NOT EXISTS neuronip.itsm_automation_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    event_source TEXT NOT NULL,
    event_filter JSONB DEFAULT '{}',
    action_type TEXT NOT NULL CHECK (action_type IN ('runbook', 'workflow', 'webhook', 'slack', 'teams', 'create_incident', 'create_change')) DEFAULT 'workflow',
    action_config JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.itsm_automation_policies IS 'Event ingestion -> automation actions';

CREATE INDEX IF NOT EXISTS idx_itsm_automation_policies_enabled ON neuronip.itsm_automation_policies(enabled) WHERE enabled = true;
