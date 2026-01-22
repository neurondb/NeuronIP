-- Migration: Collaboration Features
-- Description: Adds shared dashboards, comments, saved questions, and team workspaces

-- Shared dashboards table
CREATE TABLE IF NOT EXISTS neuronip.shared_dashboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    dashboard_config JSONB NOT NULL,
    created_by TEXT NOT NULL,
    workspace_id UUID,
    is_public BOOLEAN NOT NULL DEFAULT false,
    shared_with TEXT[], -- Array of user IDs or workspace IDs
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.shared_dashboards IS 'Shared dashboards for collaboration';

CREATE INDEX IF NOT EXISTS idx_shared_dashboards_workspace ON neuronip.shared_dashboards(workspace_id);
CREATE INDEX IF NOT EXISTS idx_shared_dashboards_created_by ON neuronip.shared_dashboards(created_by);
CREATE INDEX IF NOT EXISTS idx_shared_dashboards_public ON neuronip.shared_dashboards(is_public);

-- Dashboard comments table
CREATE TABLE IF NOT EXISTS neuronip.dashboard_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dashboard_id UUID NOT NULL REFERENCES neuronip.shared_dashboards(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    comment_text TEXT NOT NULL,
    parent_comment_id UUID REFERENCES neuronip.dashboard_comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.dashboard_comments IS 'Comments on shared dashboards';

CREATE INDEX IF NOT EXISTS idx_dashboard_comments_dashboard ON neuronip.dashboard_comments(dashboard_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_comments_user ON neuronip.dashboard_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_comments_parent ON neuronip.dashboard_comments(parent_comment_id);

-- Answer cards table (shared query results)
CREATE TABLE IF NOT EXISTS neuronip.answer_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    query_text TEXT NOT NULL,
    query_result JSONB NOT NULL,
    explanation TEXT,
    created_by TEXT NOT NULL,
    workspace_id UUID,
    is_public BOOLEAN NOT NULL DEFAULT false,
    shared_with TEXT[],
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.answer_cards IS 'Shared query results and answer cards';

CREATE INDEX IF NOT EXISTS idx_answer_cards_workspace ON neuronip.answer_cards(workspace_id);
CREATE INDEX IF NOT EXISTS idx_answer_cards_created_by ON neuronip.answer_cards(created_by);
CREATE INDEX IF NOT EXISTS idx_answer_cards_public ON neuronip.answer_cards(is_public);

-- Saved questions table
CREATE TABLE IF NOT EXISTS neuronip.saved_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_text TEXT NOT NULL,
    answer_text TEXT,
    explanation TEXT,
    query_used TEXT,
    created_by TEXT NOT NULL,
    workspace_id UUID,
    is_shared BOOLEAN NOT NULL DEFAULT false,
    shared_with TEXT[],
    tags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.saved_questions IS 'Saved questions and explanations';

CREATE INDEX IF NOT EXISTS idx_saved_questions_workspace ON neuronip.saved_questions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_saved_questions_created_by ON neuronip.saved_questions(created_by);
CREATE INDEX IF NOT EXISTS idx_saved_questions_shared ON neuronip.saved_questions(is_shared);

-- Team workspaces table (extends existing workspaces)
-- Note: This assumes workspace_id references an existing workspaces table
-- If workspaces table doesn't exist, create it separately

-- Annotations table (for annotating specific parts of dashboards/cards)
CREATE TABLE IF NOT EXISTS neuronip.annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('dashboard', 'answer_card', 'saved_question')),
    resource_id UUID NOT NULL,
    user_id TEXT NOT NULL,
    annotation_text TEXT NOT NULL,
    position JSONB, -- Position/coordinates for visual annotations
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.annotations IS 'Annotations on dashboards, cards, and questions';

CREATE INDEX IF NOT EXISTS idx_annotations_resource ON neuronip.annotations(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_annotations_user ON neuronip.annotations(user_id);

-- Update triggers
CREATE OR REPLACE FUNCTION neuronip.update_shared_dashboards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_shared_dashboards_updated_at ON neuronip.shared_dashboards;
CREATE TRIGGER trigger_update_shared_dashboards_updated_at
    BEFORE UPDATE ON neuronip.shared_dashboards
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_shared_dashboards_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_dashboard_comments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_dashboard_comments_updated_at ON neuronip.dashboard_comments;
CREATE TRIGGER trigger_update_dashboard_comments_updated_at
    BEFORE UPDATE ON neuronip.dashboard_comments
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_dashboard_comments_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_answer_cards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_answer_cards_updated_at ON neuronip.answer_cards;
CREATE TRIGGER trigger_update_answer_cards_updated_at
    BEFORE UPDATE ON neuronip.answer_cards
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_answer_cards_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_saved_questions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_saved_questions_updated_at ON neuronip.saved_questions;
CREATE TRIGGER trigger_update_saved_questions_updated_at
    BEFORE UPDATE ON neuronip.saved_questions
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_saved_questions_updated_at();

CREATE OR REPLACE FUNCTION neuronip.update_annotations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_annotations_updated_at ON neuronip.annotations;
CREATE TRIGGER trigger_update_annotations_updated_at
    BEFORE UPDATE ON neuronip.annotations
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_annotations_updated_at();
