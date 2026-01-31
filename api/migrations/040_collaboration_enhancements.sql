-- Migration: Collaboration Enhancements Schema
-- Description: Adds tables for inline annotations, discussion threads, and decision history

-- Annotations: Inline annotations on resources
CREATE TABLE IF NOT EXISTS neuronip.annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('query_result', 'dashboard', 'document', 'metric', 'table', 'column')),
    resource_id UUID NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('cell', 'row', 'column', 'section', 'element')),
    target_path TEXT NOT NULL,
    annotation_text TEXT NOT NULL,
    author_id TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.annotations IS 'Inline annotations on resources';

CREATE INDEX IF NOT EXISTS idx_annotations_resource ON neuronip.annotations(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_annotations_author ON neuronip.annotations(author_id);
CREATE INDEX IF NOT EXISTS idx_annotations_created_at ON neuronip.annotations(created_at DESC);

-- Discussion threads: Discussion threads on resources
CREATE TABLE IF NOT EXISTS neuronip.discussion_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('dashboard', 'query', 'metric', 'document', 'table')),
    resource_id UUID NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('open', 'resolved', 'archived')) DEFAULT 'open',
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.discussion_threads IS 'Discussion threads on resources';

CREATE INDEX IF NOT EXISTS idx_threads_resource ON neuronip.discussion_threads(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_threads_status ON neuronip.discussion_threads(status);
CREATE INDEX IF NOT EXISTS idx_threads_updated_at ON neuronip.discussion_threads(updated_at DESC);

-- Discussion posts: Posts in discussion threads
CREATE TABLE IF NOT EXISTS neuronip.discussion_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES neuronip.discussion_threads(id) ON DELETE CASCADE,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_post_id UUID REFERENCES neuronip.discussion_posts(id) ON DELETE SET NULL,
    attachments TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.discussion_posts IS 'Posts in discussion threads';

CREATE INDEX IF NOT EXISTS idx_posts_thread ON neuronip.discussion_posts(thread_id);
CREATE INDEX IF NOT EXISTS idx_posts_parent ON neuronip.discussion_posts(parent_post_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON neuronip.discussion_posts(created_at ASC);

-- Decision history: Team decision history
CREATE TABLE IF NOT EXISTS neuronip.decision_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL,
    resource_id UUID NOT NULL,
    decision_type TEXT NOT NULL CHECK (decision_type IN ('approval', 'rejection', 'change', 'creation', 'deletion')),
    decision TEXT NOT NULL,
    reasoning TEXT,
    made_by TEXT NOT NULL,
    participants TEXT[] DEFAULT '{}',
    context JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.decision_history IS 'Team decision history';

CREATE INDEX IF NOT EXISTS idx_decisions_resource ON neuronip.decision_history(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_decisions_type ON neuronip.decision_history(decision_type);
CREATE INDEX IF NOT EXISTS idx_decisions_made_by ON neuronip.decision_history(made_by);
CREATE INDEX IF NOT EXISTS idx_decisions_created_at ON neuronip.decision_history(created_at DESC);
