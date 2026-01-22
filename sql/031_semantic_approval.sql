-- Migration: Semantic Layer Approval Workflow
-- Description: Adds approval workflow, multi-owner support, and enhanced metric governance

-- Add approval_status to business_metrics
ALTER TABLE neuronip.business_metrics 
ADD COLUMN IF NOT EXISTS approval_status TEXT DEFAULT 'draft' 
CHECK (approval_status IN ('draft', 'pending_approval', 'approved', 'rejected'));

-- Add status column if it doesn't exist (for backward compatibility)
ALTER TABLE neuronip.business_metrics 
ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'draft';

-- Update existing metrics to have approved status if they have owners
UPDATE neuronip.business_metrics 
SET approval_status = 'approved', status = 'approved' 
WHERE owner_id IS NOT NULL AND approval_status = 'draft';

-- Metric approvals table for approval workflow
CREATE TABLE IF NOT EXISTS neuronip.metric_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    approver_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'changes_requested')),
    comments TEXT,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metric_approvals IS 'Metric approval workflow records';

CREATE INDEX IF NOT EXISTS idx_metric_approvals_metric ON neuronip.metric_approvals(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_approvals_approver ON neuronip.metric_approvals(approver_id);
CREATE INDEX IF NOT EXISTS idx_metric_approvals_status ON neuronip.metric_approvals(status);

-- Metric owners table for multi-owner support
CREATE TABLE IF NOT EXISTS neuronip.metric_owners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.business_metrics(id) ON DELETE CASCADE,
    owner_id TEXT NOT NULL,
    owner_type TEXT NOT NULL DEFAULT 'primary' CHECK (owner_type IN ('primary', 'secondary', 'steward')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(metric_id, owner_id)
);
COMMENT ON TABLE neuronip.metric_owners IS 'Multi-owner support for metrics';

CREATE INDEX IF NOT EXISTS idx_metric_owners_metric ON neuronip.metric_owners(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_owners_owner ON neuronip.metric_owners(owner_id);

-- Migration: Migrate existing owner_id to metric_owners table
INSERT INTO neuronip.metric_owners (metric_id, owner_id, owner_type, created_at, updated_at)
SELECT id, owner_id, 'primary', created_at, updated_at
FROM neuronip.business_metrics
WHERE owner_id IS NOT NULL
ON CONFLICT (metric_id, owner_id) DO NOTHING;

-- Approval queue view for easy querying
CREATE OR REPLACE VIEW neuronip.metric_approval_queue AS
SELECT 
    ma.id as approval_id,
    ma.metric_id,
    bm.name as metric_name,
    bm.display_name as metric_display_name,
    ma.approver_id,
    ma.status as approval_status,
    ma.comments,
    ma.created_at as requested_at,
    ma.updated_at,
    bm.approval_status as metric_status,
    bm.owner_id as primary_owner_id
FROM neuronip.metric_approvals ma
JOIN neuronip.business_metrics bm ON ma.metric_id = bm.id
WHERE ma.status = 'pending'
ORDER BY ma.created_at ASC;

COMMENT ON VIEW neuronip.metric_approval_queue IS 'View of pending metric approvals';

-- Update triggers for metric_approvals
CREATE OR REPLACE FUNCTION neuronip.update_metric_approvals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_metric_approvals_updated_at ON neuronip.metric_approvals;
CREATE TRIGGER trigger_update_metric_approvals_updated_at
    BEFORE UPDATE ON neuronip.metric_approvals
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_metric_approvals_updated_at();

-- Update triggers for metric_owners
CREATE OR REPLACE FUNCTION neuronip.update_metric_owners_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_metric_owners_updated_at ON neuronip.metric_owners;
CREATE TRIGGER trigger_update_metric_owners_updated_at
    BEFORE UPDATE ON neuronip.metric_owners
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_metric_owners_updated_at();
