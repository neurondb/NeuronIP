-- Migration: Kafka Ingestion Schema
-- Description: Adds tables for Kafka streaming ingestion

-- Kafka ingestion configs: Configuration for Kafka ingestion streams
CREATE TABLE IF NOT EXISTS neuronip.kafka_ingestion_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id) ON DELETE CASCADE,
    topic_name TEXT NOT NULL,
    consumer_group TEXT NOT NULL,
    brokers TEXT[] NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(data_source_id, topic_name)
);
COMMENT ON TABLE neuronip.kafka_ingestion_configs IS 'Kafka ingestion stream configurations';

CREATE INDEX IF NOT EXISTS idx_kafka_configs_data_source ON neuronip.kafka_ingestion_configs(data_source_id);
CREATE INDEX IF NOT EXISTS idx_kafka_configs_enabled ON neuronip.kafka_ingestion_configs(enabled);

-- Kafka ingestion checkpoints: Track Kafka consumer offsets
CREATE TABLE IF NOT EXISTS neuronip.kafka_ingestion_checkpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_id UUID NOT NULL REFERENCES neuronip.kafka_ingestion_configs(id) ON DELETE CASCADE,
    topic_name TEXT NOT NULL,
    partition INTEGER NOT NULL,
    offset BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(config_id, topic_name, partition)
);
COMMENT ON TABLE neuronip.kafka_ingestion_checkpoints IS 'Kafka consumer offset checkpoints';

CREATE INDEX IF NOT EXISTS idx_kafka_checkpoints_config ON neuronip.kafka_ingestion_checkpoints(config_id);
