-- Migration: Blocks and Databases
-- Description: Schema for block-based editor and database views

-- Blocks table for block-based editor
CREATE TABLE IF NOT EXISTS neuronip.blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    content JSONB NOT NULL DEFAULT '{}',
    order_index INTEGER NOT NULL DEFAULT 0,
    parent_id UUID REFERENCES neuronip.blocks(id) ON DELETE CASCADE,
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT blocks_page_id_fkey FOREIGN KEY (page_id) REFERENCES neuronip.pages(id) ON DELETE CASCADE
);

-- Pages table for organizing blocks
CREATE TABLE IF NOT EXISTS neuronip.pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- Databases table
CREATE TABLE IF NOT EXISTS neuronip.databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE CASCADE,
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- Database columns definition
CREATE TABLE IF NOT EXISTS neuronip.database_columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- text, number, date, select, multiSelect, person, file, checkbox
    options JSONB DEFAULT '[]', -- For select/multiSelect options
    order_index INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT database_columns_database_id_fkey FOREIGN KEY (database_id) REFERENCES neuronip.databases(id) ON DELETE CASCADE
);

-- Database rows
CREATE TABLE IF NOT EXISTS neuronip.database_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}', -- Flexible column data
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT database_rows_database_id_fkey FOREIGN KEY (database_id) REFERENCES neuronip.databases(id) ON DELETE CASCADE
);

-- Block comments
CREATE TABLE IF NOT EXISTS neuronip.block_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    block_id UUID NOT NULL REFERENCES neuronip.blocks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT block_comments_block_id_fkey FOREIGN KEY (block_id) REFERENCES neuronip.blocks(id) ON DELETE CASCADE
);

-- View preferences for databases
CREATE TABLE IF NOT EXISTS neuronip.database_view_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    view_type VARCHAR(50) NOT NULL DEFAULT 'table', -- table, kanban, calendar, gallery, list
    filters JSONB DEFAULT '[]',
    sort JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT database_view_preferences_database_id_user_id_key UNIQUE (database_id, user_id),
    CONSTRAINT database_view_preferences_database_id_fkey FOREIGN KEY (database_id) REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    CONSTRAINT database_view_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES neuronip.users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_blocks_page_id ON neuronip.blocks(page_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_blocks_parent_id ON neuronip.blocks(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_blocks_order ON neuronip.blocks(page_id, order_index) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pages_workspace_id ON neuronip.pages(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_database_rows_database_id ON neuronip.database_rows(database_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_database_columns_database_id ON neuronip.database_columns(database_id);
CREATE INDEX IF NOT EXISTS idx_block_comments_block_id ON neuronip.block_comments(block_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_database_view_preferences_database_user ON neuronip.database_view_preferences(database_id, user_id);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION neuronip.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_blocks_updated_at BEFORE UPDATE ON neuronip.blocks
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

CREATE TRIGGER update_pages_updated_at BEFORE UPDATE ON neuronip.pages
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

CREATE TRIGGER update_databases_updated_at BEFORE UPDATE ON neuronip.databases
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

CREATE TRIGGER update_database_rows_updated_at BEFORE UPDATE ON neuronip.database_rows
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

CREATE TRIGGER update_block_comments_updated_at BEFORE UPDATE ON neuronip.block_comments
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

CREATE TRIGGER update_database_view_preferences_updated_at BEFORE UPDATE ON neuronip.database_view_preferences
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at_column();

-- Add workspace_id to pages if workspaces table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'neuronip' AND table_name = 'workspaces') THEN
        -- Workspaces table exists, foreign key already added
        NULL;
    ELSE
        -- Create a simple workspaces reference (or make workspace_id nullable)
        ALTER TABLE neuronip.pages ALTER COLUMN workspace_id DROP NOT NULL;
        ALTER TABLE neuronip.databases ALTER COLUMN workspace_id DROP NOT NULL;
    END IF;
END $$;
