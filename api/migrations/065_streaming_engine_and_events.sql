-- Migration: 065_streaming_engine_and_events.sql
-- Description: Streaming engine columns and Postgres-native events table

-- Ensure streaming_pipelines has columns expected by StreamingEngine (sources, transformations, sinks as JSONB)
ALTER TABLE neuronip.streaming_pipelines ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE neuronip.streaming_pipelines ADD COLUMN IF NOT EXISTS sources JSONB;
ALTER TABLE neuronip.streaming_pipelines ADD COLUMN IF NOT EXISTS transformations JSONB;
ALTER TABLE neuronip.streaming_pipelines ADD COLUMN IF NOT EXISTS sinks JSONB;
ALTER TABLE neuronip.streaming_pipelines ADD COLUMN IF NOT EXISTS config JSONB;

-- Allow nullable legacy columns so engine can insert with only new columns
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'neuronip' AND table_name = 'streaming_pipelines' AND column_name = 'source_type') THEN
    ALTER TABLE neuronip.streaming_pipelines ALTER COLUMN source_type DROP NOT NULL;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'neuronip' AND table_name = 'streaming_pipelines' AND column_name = 'source_config') THEN
    ALTER TABLE neuronip.streaming_pipelines ALTER COLUMN source_config DROP NOT NULL;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'neuronip' AND table_name = 'streaming_pipelines' AND column_name = 'destination_type') THEN
    ALTER TABLE neuronip.streaming_pipelines ALTER COLUMN destination_type DROP NOT NULL;
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'neuronip' AND table_name = 'streaming_pipelines' AND column_name = 'destination_config') THEN
    ALTER TABLE neuronip.streaming_pipelines ALTER COLUMN destination_config DROP NOT NULL;
  END IF;
EXCEPTION WHEN OTHERS THEN
  NULL; -- Ignore if columns do not exist
END $$;

-- Postgres-native append-only events table for streaming source (no Kafka required)
CREATE TABLE IF NOT EXISTS neuronip.stream_events (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    source_table TEXT,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stream_events_created_at ON neuronip.stream_events(created_at);
CREATE INDEX IF NOT EXISTS idx_stream_events_event_type ON neuronip.stream_events(event_type);
CREATE INDEX IF NOT EXISTS idx_stream_events_id ON neuronip.stream_events(id);

COMMENT ON TABLE neuronip.stream_events IS 'Append-only events for Postgres-native streaming pipelines';

-- Sink results table for pipeline output (when sink type is database)
CREATE TABLE IF NOT EXISTS neuronip.stream_sink_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL,
    event_id BIGINT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stream_sink_results_pipeline ON neuronip.stream_sink_results(pipeline_id);
COMMENT ON TABLE neuronip.stream_sink_results IS 'Pipeline output sink for Postgres-native streaming';

-- CDC polling state (for Postgres CDC without replication connection)
CREATE TABLE IF NOT EXISTS neuronip.cdc_polling_state (
    connector_key TEXT PRIMARY KEY,
    last_id BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.cdc_polling_state IS 'Checkpoint for polling-based Postgres CDC (no replication slot required)';
