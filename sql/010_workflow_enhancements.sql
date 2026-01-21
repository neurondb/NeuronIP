-- Workflow enhancements migration
-- Adds tables for workflow versioning, scheduling, logging, and metrics

-- Workflow versions: Track workflow version history
CREATE TABLE IF NOT EXISTS neuronip.workflow_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    parent_workflow_id UUID REFERENCES neuronip.workflows(id) ON DELETE SET NULL,
    changes JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workflow_id, version)
);
COMMENT ON TABLE neuronip.workflow_versions IS 'Workflow version history';

-- Workflow schedules: Scheduled workflow executions
CREATE TABLE IF NOT EXISTS neuronip.workflow_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id) ON DELETE CASCADE,
    schedule_config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workflow_id)
);
COMMENT ON TABLE neuronip.workflow_schedules IS 'Scheduled workflow executions';

-- Workflow step executions: Track individual step execution
CREATE TABLE IF NOT EXISTS neuronip.workflow_step_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id) ON DELETE CASCADE,
    step_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped')),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workflow_step_executions IS 'Individual workflow step execution tracking';

-- Workflow step results: Store step execution results
CREATE TABLE IF NOT EXISTS neuronip.workflow_step_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id) ON DELETE CASCADE,
    step_id TEXT NOT NULL,
    result_data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(execution_id, step_id)
);
COMMENT ON TABLE neuronip.workflow_step_results IS 'Workflow step execution results';

-- Workflow logs: Execution logs
CREATE TABLE IF NOT EXISTS neuronip.workflow_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id) ON DELETE CASCADE,
    step_id TEXT,
    level TEXT NOT NULL DEFAULT 'info' CHECK (level IN ('debug', 'info', 'warn', 'error')),
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workflow_logs IS 'Workflow execution logs';

-- Workflow metrics: Performance metrics
CREATE TABLE IF NOT EXISTS neuronip.workflow_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id) ON DELETE CASCADE,
    step_id TEXT,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    metric_unit TEXT,
    description TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workflow_metrics IS 'Workflow performance metrics';

-- Add execution_time_ms to workflow_executions if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'neuronip' 
        AND table_name = 'workflow_executions' 
        AND column_name = 'execution_time_ms'
    ) THEN
        ALTER TABLE neuronip.workflow_executions 
        ADD COLUMN execution_time_ms BIGINT;
    END IF;
END $$;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_workflow_versions_workflow_id 
    ON neuronip.workflow_versions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_schedules_workflow_id 
    ON neuronip.workflow_schedules(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_schedules_next_run 
    ON neuronip.workflow_schedules(next_run_at) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_workflow_step_executions_execution_id 
    ON neuronip.workflow_step_executions(execution_id);
CREATE INDEX IF NOT EXISTS idx_workflow_step_results_execution_id 
    ON neuronip.workflow_step_results(execution_id);
CREATE INDEX IF NOT EXISTS idx_workflow_logs_execution_id 
    ON neuronip.workflow_logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_workflow_logs_timestamp 
    ON neuronip.workflow_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_workflow_metrics_execution_id 
    ON neuronip.workflow_metrics(execution_id);
