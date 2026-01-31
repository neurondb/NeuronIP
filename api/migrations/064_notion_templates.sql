-- Migration: Page and database templates (knowledge workspace)
-- Description: Page and database templates for docs+databases flows

-- Page templates (doc templates)
CREATE TABLE IF NOT EXISTS neuronip.page_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    cover_url TEXT,
    default_blocks JSONB NOT NULL DEFAULT '[]',
    workspace_id UUID,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'workspace', 'public')) DEFAULT 'workspace',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.page_templates IS 'Page templates';

CREATE INDEX IF NOT EXISTS idx_page_templates_workspace ON neuronip.page_templates(workspace_id);
CREATE INDEX IF NOT EXISTS idx_page_templates_visibility ON neuronip.page_templates(visibility);

-- Database templates (database structure templates)
CREATE TABLE IF NOT EXISTS neuronip.database_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    schema_definition JSONB NOT NULL DEFAULT '{}',
    default_views JSONB DEFAULT '[]',
    workspace_id UUID,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'workspace', 'public')) DEFAULT 'workspace',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.database_templates IS 'Database templates';

CREATE INDEX IF NOT EXISTS idx_database_templates_workspace ON neuronip.database_templates(workspace_id);
