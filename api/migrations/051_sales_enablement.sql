-- Migration: 051_sales_enablement.sql
-- Description: Sales motion and enterprise enablement

-- Demo environments
CREATE TABLE IF NOT EXISTS neuronip.demo_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    demo_name VARCHAR(255) NOT NULL,
    demo_type VARCHAR(50) NOT NULL, -- 'standard', 'custom', 'industry_specific'
    environment_config JSONB NOT NULL, -- Environment setup configuration
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'provisioning', 'active', 'expired', 'deleted'
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_demo_environments_status ON neuronip.demo_environments(status);
CREATE INDEX idx_demo_environments_expires ON neuronip.demo_environments(expires_at) WHERE status = 'active';

-- Trial accounts
CREATE TABLE IF NOT EXISTS neuronip.trial_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_email VARCHAR(255) NOT NULL,
    trial_type VARCHAR(50) NOT NULL DEFAULT 'standard', -- 'standard', 'enterprise', 'custom'
    trial_duration_days INTEGER NOT NULL DEFAULT 14,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'expired', 'converted', 'cancelled'
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    converted_at TIMESTAMP WITH TIME ZONE,
    conversion_revenue DOUBLE PRECISION,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trial_accounts_email ON neuronip.trial_accounts(user_email);
CREATE INDEX idx_trial_accounts_status ON neuronip.trial_accounts(status);
CREATE INDEX idx_trial_accounts_expires ON neuronip.trial_accounts(expires_at) WHERE status = 'active';

-- Enterprise onboarding workflows
CREATE TABLE IF NOT EXISTS neuronip.onboarding_workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_name VARCHAR(255) NOT NULL,
    workflow_type VARCHAR(50) NOT NULL, -- 'trial', 'enterprise', 'partner'
    workflow_steps JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'active', 'archived'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_onboarding_workflows_type ON neuronip.onboarding_workflows(workflow_type);
CREATE INDEX idx_onboarding_workflows_status ON neuronip.onboarding_workflows(status);

-- Onboarding progress tracking
CREATE TABLE IF NOT EXISTS neuronip.onboarding_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID,
    user_id VARCHAR(255),
    workflow_id UUID NOT NULL REFERENCES neuronip.onboarding_workflows(id),
    current_step VARCHAR(255),
    completed_steps JSONB, -- Array of completed step IDs
    progress_percentage INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress', -- 'in_progress', 'completed', 'abandoned'
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_onboarding_progress_workspace ON neuronip.onboarding_progress(workspace_id);
CREATE INDEX idx_onboarding_progress_workflow ON neuronip.onboarding_progress(workflow_id);
CREATE INDEX idx_onboarding_progress_status ON neuronip.onboarding_progress(status);

COMMENT ON TABLE neuronip.demo_environments IS 'Demo environment provisioning';
COMMENT ON TABLE neuronip.trial_accounts IS 'Trial account management';
COMMENT ON TABLE neuronip.onboarding_workflows IS 'Enterprise onboarding workflow definitions';
COMMENT ON TABLE neuronip.onboarding_progress IS 'Onboarding progress tracking';
