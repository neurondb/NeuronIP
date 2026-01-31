-- Migration: Notebooks
-- Description: Notebook entities with cells (SQL/Python/Markdown), backed by workflow executions

-- Notebooks
CREATE TABLE IF NOT EXISTS neuronip.notebooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    workspace_id UUID,
    owner_id TEXT NOT NULL,
    default_language TEXT NOT NULL DEFAULT 'sql' CHECK (default_language IN ('sql', 'python', 'markdown')),
    workflow_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.notebooks IS 'Notebooks, linked to workflows for execution';

CREATE INDEX IF NOT EXISTS idx_notebooks_workspace ON neuronip.notebooks(workspace_id);
CREATE INDEX IF NOT EXISTS idx_notebooks_owner ON neuronip.notebooks(owner_id);
CREATE INDEX IF NOT EXISTS idx_notebooks_workflow ON neuronip.notebooks(workflow_id);

-- Notebook cells (SQL, Python, Markdown)
CREATE TABLE IF NOT EXISTS neuronip.notebook_cells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notebook_id UUID NOT NULL REFERENCES neuronip.notebooks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    cell_type TEXT NOT NULL CHECK (cell_type IN ('sql', 'python', 'markdown')),
    content TEXT NOT NULL DEFAULT '',
    output TEXT,
    run_id UUID,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.notebook_cells IS 'Notebook cells (SQL/Python/Markdown)';

CREATE INDEX IF NOT EXISTS idx_notebook_cells_notebook ON neuronip.notebook_cells(notebook_id);
CREATE INDEX IF NOT EXISTS idx_notebook_cells_position ON neuronip.notebook_cells(notebook_id, position);

-- Notebook runs (execution linked to workflow run)
CREATE TABLE IF NOT EXISTS neuronip.notebook_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notebook_id UUID NOT NULL REFERENCES neuronip.notebooks(id) ON DELETE CASCADE,
    workflow_execution_id UUID,
    triggered_by TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.notebook_runs IS 'Notebook execution runs (backed by workflow)';

CREATE INDEX IF NOT EXISTS idx_notebook_runs_notebook ON neuronip.notebook_runs(notebook_id);
CREATE INDEX IF NOT EXISTS idx_notebook_runs_status ON neuronip.notebook_runs(status);
