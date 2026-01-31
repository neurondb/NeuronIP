-- 059_notion_ui_databases.sql
-- Description: Creates tables for database views and rows.

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS neuronip.databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    workspace_id UUID REFERENCES neuronip.workspaces(id) ON DELETE SET NULL,
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_databases_workspace_id ON neuronip.databases(workspace_id);
CREATE INDEX IF NOT EXISTS idx_databases_created_by ON neuronip.databases(created_by);

CREATE TABLE IF NOT EXISTS neuronip.database_columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- text, number, date, select, multiSelect, person, file, checkbox
    options JSONB NOT NULL DEFAULT '[]', -- For select/multiSelect types
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_database_columns_database_id ON neuronip.database_columns(database_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_database_columns_database_order ON neuronip.database_columns (database_id, order_index);

CREATE TABLE IF NOT EXISTS neuronip.database_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}', -- Flexible row data
    created_by UUID REFERENCES neuronip.users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_database_rows_database_id ON neuronip.database_rows(database_id);
CREATE INDEX IF NOT EXISTS idx_database_rows_created_by ON neuronip.database_rows(created_by);

CREATE TABLE IF NOT EXISTS neuronip.database_view_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES neuronip.databases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES neuronip.users(id) ON DELETE CASCADE,
    view_type TEXT NOT NULL DEFAULT 'table', -- table, kanban, calendar, gallery, list
    filters JSONB NOT NULL DEFAULT '[]',
    sort JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(database_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_database_view_preferences_database_id ON neuronip.database_view_preferences(database_id);
CREATE INDEX IF NOT EXISTS idx_database_view_preferences_user_id ON neuronip.database_view_preferences(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS neuronip.database_view_preferences;
DROP TABLE IF EXISTS neuronip.database_rows;
DROP TABLE IF EXISTS neuronip.database_columns;
DROP TABLE IF EXISTS neuronip.databases;
-- +goose StatementEnd
