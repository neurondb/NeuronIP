-- Migration: Ingestion Enhancements
-- Description: Adds retry tracking, backpressure monitoring, and dead letter queue

-- Enhance ingestion_jobs table
ALTER TABLE neuronip.ingestion_jobs ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE neuronip.ingestion_jobs ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 3;
ALTER TABLE neuronip.ingestion_jobs ADD COLUMN IF NOT EXISTS retry_config JSONB DEFAULT '{}';
ALTER TABLE neuronip.ingestion_jobs ADD COLUMN IF NOT EXISTS backpressure_detected BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE neuronip.ingestion_jobs ADD COLUMN IF NOT EXISTS throttled_at TIMESTAMPTZ;

COMMENT ON COLUMN neuronip.ingestion_jobs.retry_count IS 'Number of retry attempts';
COMMENT ON COLUMN neuronip.ingestion_jobs.max_retries IS 'Maximum retry attempts';
COMMENT ON COLUMN neuronip.ingestion_jobs.retry_config IS 'Retry configuration (backoff, delays, etc.)';
COMMENT ON COLUMN neuronip.ingestion_jobs.backpressure_detected IS 'Whether backpressure was detected';
COMMENT ON COLUMN neuronip.ingestion_jobs.throttled_at IS 'When the job was throttled due to backpressure';

-- Dead letter queue: Failed jobs that exceeded retry limits
CREATE TABLE IF NOT EXISTS neuronip.dlq_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES neuronip.ingestion_jobs(id) ON DELETE CASCADE,
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id) ON DELETE CASCADE,
    error_type TEXT NOT NULL,
    error_message TEXT NOT NULL,
    failed_data JSONB,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT
);
COMMENT ON TABLE neuronip.dlq_entries IS 'Dead letter queue for failed ingestion jobs';

CREATE INDEX IF NOT EXISTS idx_dlq_entries_job ON neuronip.dlq_entries(job_id);
CREATE INDEX IF NOT EXISTS idx_dlq_entries_data_source ON neuronip.dlq_entries(data_source_id);
CREATE INDEX IF NOT EXISTS idx_dlq_entries_error_type ON neuronip.dlq_entries(error_type);
CREATE INDEX IF NOT EXISTS idx_dlq_entries_created ON neuronip.dlq_entries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dlq_entries_resolved ON neuronip.dlq_entries(resolved_at) WHERE resolved_at IS NULL;

-- Retry history: Track retry attempts
CREATE TABLE IF NOT EXISTS neuronip.ingestion_retries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES neuronip.ingestion_jobs(id) ON DELETE CASCADE,
    attempt_number INTEGER NOT NULL,
    error_message TEXT,
    retried_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.ingestion_retries IS 'Retry attempt history';

CREATE INDEX IF NOT EXISTS idx_ingestion_retries_job ON neuronip.ingestion_retries(job_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_retries_attempt ON neuronip.ingestion_retries(attempt_number);

-- Backpressure metrics: Track backpressure events
CREATE TABLE IF NOT EXISTS neuronip.backpressure_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID REFERENCES neuronip.data_sources(id) ON DELETE SET NULL,
    current_jobs INTEGER NOT NULL,
    queue_size INTEGER NOT NULL,
    max_concurrent_jobs INTEGER NOT NULL,
    max_queue_size INTEGER NOT NULL,
    throttled_jobs INTEGER NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.backpressure_metrics IS 'Backpressure monitoring metrics';

CREATE INDEX IF NOT EXISTS idx_backpressure_metrics_recorded ON neuronip.backpressure_metrics(recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_backpressure_metrics_data_source ON neuronip.backpressure_metrics(data_source_id);

-- Function to record retry attempt
CREATE OR REPLACE FUNCTION neuronip.record_retry_attempt(
    p_job_id UUID,
    p_attempt_number INTEGER,
    p_error_message TEXT,
    p_next_retry_at TIMESTAMPTZ
) RETURNS UUID AS $$
DECLARE
    v_retry_id UUID;
BEGIN
    INSERT INTO neuronip.ingestion_retries (job_id, attempt_number, error_message, next_retry_at)
    VALUES (p_job_id, p_attempt_number, p_error_message, p_next_retry_at)
    RETURNING id INTO v_retry_id;
    
    UPDATE neuronip.ingestion_jobs
    SET retry_count = p_attempt_number, updated_at = NOW()
    WHERE id = p_job_id;
    
    RETURN v_retry_id;
END;
$$ LANGUAGE plpgsql;

-- Function to move job to DLQ
CREATE OR REPLACE FUNCTION neuronip.move_to_dlq(
    p_job_id UUID,
    p_error_type TEXT,
    p_error_message TEXT,
    p_failed_data JSONB,
    p_retry_count INTEGER
) RETURNS UUID AS $$
DECLARE
    v_dlq_id UUID;
    v_data_source_id UUID;
BEGIN
    -- Get data source ID from job
    SELECT data_source_id INTO v_data_source_id
    FROM neuronip.ingestion_jobs
    WHERE id = p_job_id;
    
    -- Insert into DLQ
    INSERT INTO neuronip.dlq_entries (
        job_id, data_source_id, error_type, error_message,
        failed_data, retry_count
    ) VALUES (
        p_job_id, v_data_source_id, p_error_type, p_error_message,
        p_failed_data, p_retry_count
    )
    RETURNING id INTO v_dlq_id;
    
    -- Update job status
    UPDATE neuronip.ingestion_jobs
    SET status = 'failed', error_message = p_error_message, updated_at = NOW()
    WHERE id = p_job_id;
    
    RETURN v_dlq_id;
END;
$$ LANGUAGE plpgsql;
