-- Migration: Saved Searches
-- Description: Adds saved hybrid searches with SQL filters + vector queries

-- Saved searches: Reusable hybrid searches
CREATE TABLE IF NOT EXISTS neuronip.saved_searches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    query TEXT NOT NULL, -- Natural language query
    semantic_query TEXT, -- Semantic similarity query
    schema_id UUID REFERENCES neuronip.warehouse_schemas(id) ON DELETE SET NULL,
    sql_filters JSONB DEFAULT '{}', -- SQL filter conditions
    semantic_table TEXT, -- Table to search semantically
    semantic_column TEXT, -- Column with embeddings
    limit_count INTEGER NOT NULL DEFAULT 10,
    threshold NUMERIC(3,2) NOT NULL DEFAULT 0.5,
    is_public BOOLEAN NOT NULL DEFAULT false,
    owner_id TEXT,
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.saved_searches IS 'Saved hybrid searches (SQL + vector)';

CREATE INDEX IF NOT EXISTS idx_saved_searches_owner ON neuronip.saved_searches(owner_id);
CREATE INDEX IF NOT EXISTS idx_saved_searches_public ON neuronip.saved_searches(is_public) WHERE is_public = true;
CREATE INDEX IF NOT EXISTS idx_saved_searches_tags ON neuronip.saved_searches USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_saved_searches_created ON neuronip.saved_searches(created_at DESC);

-- Search templates: Pre-defined search templates
CREATE TABLE IF NOT EXISTS neuronip.search_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    template_config JSONB NOT NULL, -- Template configuration
    category TEXT,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.search_templates IS 'Pre-defined search templates';

CREATE INDEX IF NOT EXISTS idx_search_templates_category ON neuronip.search_templates(category);
CREATE INDEX IF NOT EXISTS idx_search_templates_public ON neuronip.search_templates(is_public);

-- Search sharing: Share saved searches with specific users or teams
CREATE TABLE IF NOT EXISTS neuronip.search_sharing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    search_id UUID NOT NULL REFERENCES neuronip.saved_searches(id) ON DELETE CASCADE,
    shared_with_user_id TEXT,
    shared_with_team_id TEXT,
    permission TEXT NOT NULL CHECK (permission IN ('read', 'execute', 'edit')) DEFAULT 'read',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(search_id, COALESCE(shared_with_user_id, ''), COALESCE(shared_with_team_id, ''))
);
COMMENT ON TABLE neuronip.search_sharing IS 'Saved search sharing and permissions';

CREATE INDEX IF NOT EXISTS idx_search_sharing_search ON neuronip.search_templates(id);
CREATE INDEX IF NOT EXISTS idx_search_sharing_user ON neuronip.search_sharing(shared_with_user_id);
CREATE INDEX IF NOT EXISTS idx_search_sharing_team ON neuronip.search_sharing(shared_with_team_id);

-- Update trigger
CREATE OR REPLACE FUNCTION neuronip.update_saved_searches_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_saved_searches_updated_at ON neuronip.saved_searches;
CREATE TRIGGER trigger_update_saved_searches_updated_at
    BEFORE UPDATE ON neuronip.saved_searches
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_saved_searches_updated_at();
