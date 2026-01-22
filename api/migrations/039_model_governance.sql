-- Migration: Model and Prompt Governance
-- Description: Adds model registry, prompt versioning, and approval workflows

-- Model Registry: Model versioning and approval
CREATE TABLE IF NOT EXISTS neuronip.model_registry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name TEXT NOT NULL,
    version TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('openai', 'anthropic', 'cohere', 'huggingface', 'custom')),
    model_id TEXT NOT NULL, -- e.g., 'gpt-4', 'claude-3', etc.
    status TEXT NOT NULL CHECK (status IN ('draft', 'pending_approval', 'approved', 'deprecated')) DEFAULT 'draft',
    approved_by TEXT,
    approved_at TIMESTAMPTZ,
    config JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(model_name, version)
);
COMMENT ON TABLE neuronip.model_registry IS 'Model registry with versioning';

CREATE INDEX IF NOT EXISTS idx_model_registry_name ON neuronip.model_registry(model_name);
CREATE INDEX IF NOT EXISTS idx_model_registry_status ON neuronip.model_registry(status);
CREATE INDEX IF NOT EXISTS idx_model_registry_provider ON neuronip.model_registry(provider);

-- Prompt Templates: Prompt template versioning
CREATE TABLE IF NOT EXISTS neuronip.prompt_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    template_text TEXT NOT NULL,
    variables TEXT[] DEFAULT '{}',
    description TEXT,
    status TEXT NOT NULL CHECK (status IN ('draft', 'pending_approval', 'approved', 'deprecated')) DEFAULT 'draft',
    approved_by TEXT,
    approved_at TIMESTAMPTZ,
    parent_template_id UUID REFERENCES neuronip.prompt_templates(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, version)
);
COMMENT ON TABLE neuronip.prompt_templates IS 'Prompt template versioning';

CREATE INDEX IF NOT EXISTS idx_prompt_templates_name ON neuronip.prompt_templates(name);
CREATE INDEX IF NOT EXISTS idx_prompt_templates_status ON neuronip.prompt_templates(status);
CREATE INDEX IF NOT EXISTS idx_prompt_templates_parent ON neuronip.prompt_templates(parent_template_id);

-- Prompt Approvals: Approval workflow for prompts
CREATE TABLE IF NOT EXISTS neuronip.prompt_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id UUID NOT NULL REFERENCES neuronip.prompt_templates(id) ON DELETE CASCADE,
    approver_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'changes_requested')) DEFAULT 'pending',
    comments TEXT,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.prompt_approvals IS 'Prompt approval workflow';

CREATE INDEX IF NOT EXISTS idx_prompt_approvals_prompt ON neuronip.prompt_approvals(prompt_id);
CREATE INDEX IF NOT EXISTS idx_prompt_approvals_status ON neuronip.prompt_approvals(status);

-- Workspace Model Selection: Per-workspace model selection
CREATE TABLE IF NOT EXISTS neuronip.workspace_model_selection (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    is_default BOOLEAN NOT NULL DEFAULT false,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, model_id)
);
COMMENT ON TABLE neuronip.workspace_model_selection IS 'Per-workspace model selection';

CREATE INDEX IF NOT EXISTS idx_workspace_model_selection_workspace ON neuronip.workspace_model_selection(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_model_selection_model ON neuronip.workspace_model_selection(model_id);
CREATE INDEX IF NOT EXISTS idx_workspace_model_selection_default ON neuronip.workspace_model_selection(workspace_id, is_default) WHERE is_default = true;

-- Model Approvals: Approval workflow for models
CREATE TABLE IF NOT EXISTS neuronip.model_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES neuronip.model_registry(id) ON DELETE CASCADE,
    approver_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'changes_requested')) DEFAULT 'pending',
    comments TEXT,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.model_approvals IS 'Model approval workflow';

CREATE INDEX IF NOT EXISTS idx_model_approvals_model ON neuronip.model_approvals(model_id);
CREATE INDEX IF NOT EXISTS idx_model_approvals_status ON neuronip.model_approvals(status);

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_model_registry_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_model_registry_updated_at ON neuronip.model_registry;
CREATE TRIGGER trigger_update_model_registry_updated_at
    BEFORE UPDATE ON neuronip.model_registry
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_model_registry_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_prompt_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_prompt_templates_updated_at ON neuronip.prompt_templates;
CREATE TRIGGER trigger_update_prompt_templates_updated_at
    BEFORE UPDATE ON neuronip.prompt_templates
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_prompt_templates_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_prompt_approvals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_prompt_approvals_updated_at ON neuronip.prompt_approvals;
CREATE TRIGGER trigger_update_prompt_approvals_updated_at
    BEFORE UPDATE ON neuronip.prompt_approvals
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_prompt_approvals_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_model_approvals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_model_approvals_updated_at ON neuronip.model_approvals;
CREATE TRIGGER trigger_update_model_approvals_updated_at
    BEFORE UPDATE ON neuronip.model_approvals
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_model_approvals_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_workspace_model_selection_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_workspace_model_selection_updated_at ON neuronip.workspace_model_selection;
CREATE TRIGGER trigger_update_workspace_model_selection_updated_at
    BEFORE UPDATE ON neuronip.workspace_model_selection
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_workspace_model_selection_updated_at();
