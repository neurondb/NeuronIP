-- Collaboration features: Comments, Reactions, Mentions
-- Migration: 035_collaboration.sql

-- Comments table for inline comments on any resource
CREATE TABLE IF NOT EXISTS neuronip.comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL, -- e.g., 'query', 'document', 'workflow'
    resource_id UUID NOT NULL,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id UUID REFERENCES neuronip.comments(id) ON DELETE CASCADE, -- For threaded replies
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Composite index for efficient lookups
    CONSTRAINT comments_resource_check CHECK (resource_type IN ('query', 'document', 'workflow', 'agent', 'model', 'collection'))
);

CREATE INDEX IF NOT EXISTS idx_comments_resource ON neuronip.comments(resource_type, resource_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_user ON neuronip.comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_parent ON neuronip.comments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_created ON neuronip.comments(created_at DESC) WHERE deleted_at IS NULL;

COMMENT ON TABLE neuronip.comments IS 'Inline comments on resources for collaboration';

-- Reactions table for emoji reactions on comments
CREATE TABLE IF NOT EXISTS neuronip.reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES neuronip.comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    emoji TEXT NOT NULL, -- e.g., '👍', '❤️', '🎉', '💡'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- One reaction per user per comment per emoji
    UNIQUE(comment_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_reactions_comment ON neuronip.reactions(comment_id);
CREATE INDEX IF NOT EXISTS idx_reactions_user ON neuronip.reactions(user_id);

COMMENT ON TABLE neuronip.reactions IS 'Emoji reactions on comments';

-- Mentions table for @mentions in comments
CREATE TABLE IF NOT EXISTS neuronip.mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES neuronip.comments(id) ON DELETE CASCADE,
    mentioned_user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- One mention record per user per comment
    UNIQUE(comment_id, mentioned_user_id)
);

CREATE INDEX IF NOT EXISTS idx_mentions_comment ON neuronip.mentions(comment_id);
CREATE INDEX IF NOT EXISTS idx_mentions_user ON neuronip.mentions(mentioned_user_id);
CREATE INDEX IF NOT EXISTS idx_mentions_created ON neuronip.mentions(created_at DESC);

COMMENT ON TABLE neuronip.mentions IS 'User mentions in comments for notifications';

-- Presence tracking for live collaboration
CREATE TABLE IF NOT EXISTS neuronip.presence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id, resource_type, resource_id)
);

CREATE INDEX IF NOT EXISTS idx_presence_resource ON neuronip.presence(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_presence_user ON neuronip.presence(user_id);
CREATE INDEX IF NOT EXISTS idx_presence_last_seen ON neuronip.presence(last_seen DESC);

-- Clean up old presence records (older than 5 minutes)
CREATE OR REPLACE FUNCTION neuronip.cleanup_presence()
RETURNS void AS $$
BEGIN
    DELETE FROM neuronip.presence
    WHERE last_seen < NOW() - INTERVAL '5 minutes';
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE neuronip.presence IS 'Real-time presence tracking for live collaboration';

-- Sharing links table
CREATE TABLE IF NOT EXISTS neuronip.share_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    created_by UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    permission TEXT NOT NULL DEFAULT 'view' CHECK (permission IN ('view', 'edit', 'full')),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT share_links_resource_check CHECK (resource_type IN ('query', 'document', 'workflow', 'agent', 'model', 'collection'))
);

CREATE INDEX IF NOT EXISTS idx_share_links_resource ON neuronip.share_links(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_share_links_created_by ON neuronip.share_links(created_by);
CREATE INDEX IF NOT EXISTS idx_share_links_expires ON neuronip.share_links(expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE neuronip.share_links IS 'Shareable links with permissions';

-- Function to update comment updated_at
CREATE OR REPLACE FUNCTION neuronip.update_comment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_comment_updated_at
    BEFORE UPDATE ON neuronip.comments
    FOR EACH ROW
    EXECUTE FUNCTION neuronip.update_comment_updated_at();
