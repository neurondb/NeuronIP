-- Migration: Decision dashboards
-- Description: Decision intelligence apps: metrics, evidence, lineage, workflow runs, approvals

-- Decision dashboards
CREATE TABLE IF NOT EXISTS neuronip.decision_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    workspace_id UUID,
    owner_id TEXT NOT NULL,
    layout JSONB NOT NULL DEFAULT '[]',
    metric_ids JSONB DEFAULT '[]',
    evidence_sources JSONB DEFAULT '[]',
    lineage_resource_type TEXT,
    lineage_resource_id UUID,
    workflow_ids JSONB DEFAULT '[]',
    approval_workflow_id UUID,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'workspace', 'public')) DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.decision_dashboards IS 'Decision intelligence dashboards';

CREATE INDEX IF NOT EXISTS idx_decision_dashboards_workspace ON neuronip.decision_dashboards(workspace_id);
CREATE INDEX IF NOT EXISTS idx_decision_dashboards_owner ON neuronip.decision_dashboards(owner_id);

-- Decision dashboard runs (snapshot of metrics + evidence at run time)
CREATE TABLE IF NOT EXISTS neuronip.decision_dashboard_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES neuronip.decision_dashboards(id) ON DELETE CASCADE,
    triggered_by TEXT NOT NULL,
    snapshot JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('running', 'completed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.decision_dashboard_runs IS 'Decision dashboard run snapshots';

CREATE INDEX IF NOT EXISTS idx_decision_dashboard_runs_dashboard ON neuronip.decision_dashboard_runs(dashboard_id);
