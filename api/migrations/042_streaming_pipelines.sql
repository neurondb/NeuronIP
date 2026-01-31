-- Migration: 042_streaming_pipelines.sql
-- Description: Streaming pipeline and CDC enhancements

-- Streaming pipelines table
CREATE TABLE IF NOT EXISTS neuronip.streaming_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_name VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL, -- 'kafka', 'kinesis', 'pulsar', 'postgres_cdc', 'mysql_cdc'
    source_config JSONB NOT NULL,
    destination_type VARCHAR(50) NOT NULL, -- 'postgres', 'warehouse', 'kafka', etc.
    destination_config JSONB NOT NULL,
    transformation_config JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive', -- 'active', 'inactive', 'paused', 'error'
    enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_streaming_pipelines_status ON neuronip.streaming_pipelines(status);
CREATE INDEX idx_streaming_pipelines_enabled ON neuronip.streaming_pipelines(enabled) WHERE enabled = true;

-- Stream processing windows
CREATE TABLE IF NOT EXISTS neuronip.stream_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.streaming_pipelines(id),
    window_type VARCHAR(50) NOT NULL, -- 'tumbling', 'sliding', 'session'
    window_size_seconds INTEGER NOT NULL,
    slide_interval_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stream_windows_pipeline ON neuronip.stream_windows(pipeline_id);

-- Stream checkpoints (for exactly-once semantics)
CREATE TABLE IF NOT EXISTS neuronip.stream_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.streaming_pipelines(id),
    partition_id VARCHAR(255) NOT NULL,
    offset BIGINT NOT NULL,
    watermark TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(pipeline_id, partition_id)
);

CREATE INDEX idx_stream_checkpoints_pipeline ON neuronip.stream_checkpoints(pipeline_id);
CREATE INDEX idx_stream_checkpoints_partition ON neuronip.stream_checkpoints(partition_id);

-- Stream processing metrics
CREATE TABLE IF NOT EXISTS neuronip.stream_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.streaming_pipelines(id),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_unit VARCHAR(50),
    partition_id VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stream_metrics_pipeline ON neuronip.stream_metrics(pipeline_id);
CREATE INDEX idx_stream_metrics_name ON neuronip.stream_metrics(metric_name);
CREATE INDEX idx_stream_metrics_timestamp ON neuronip.stream_metrics(timestamp);

-- Backfill jobs for streaming
CREATE TABLE IF NOT EXISTS neuronip.stream_backfills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES neuronip.streaming_pipelines(id),
    start_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    end_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    rows_processed BIGINT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stream_backfills_pipeline ON neuronip.stream_backfills(pipeline_id);
CREATE INDEX idx_stream_backfills_status ON neuronip.stream_backfills(status);

COMMENT ON TABLE neuronip.streaming_pipelines IS 'Streaming data pipeline definitions';
COMMENT ON TABLE neuronip.stream_windows IS 'Stream processing window configurations';
COMMENT ON TABLE neuronip.stream_checkpoints IS 'Stream processing checkpoints for exactly-once semantics';
COMMENT ON TABLE neuronip.stream_metrics IS 'Streaming pipeline performance metrics';
COMMENT ON TABLE neuronip.stream_backfills IS 'Historical backfill jobs for streaming pipelines';
