-- Migration: Backup and Restore System
-- Description: Adds backup tracking and management

-- Backups: Backup records
CREATE TABLE IF NOT EXISTS neuronip.backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL CHECK (type IN ('full', 'incremental', 'config')),
    status TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'failed')),
    file_path TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.backups IS 'Backup records for disaster recovery';

CREATE INDEX IF NOT EXISTS idx_backups_status ON neuronip.backups(status);
CREATE INDEX IF NOT EXISTS idx_backups_type ON neuronip.backups(type);
CREATE INDEX IF NOT EXISTS idx_backups_started_at ON neuronip.backups(started_at DESC);
