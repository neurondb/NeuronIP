-- Migration: 041_enterprise_ingestion.sql
-- Description: Enterprise-grade ingestion enhancements

-- Ingestion schedules table
CREATE TABLE IF NOT EXISTS neuronip.ingestion_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id),
    schedule_type VARCHAR(50) NOT NULL, -- 'cron', 'interval', 'event_driven'
    schedule_config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingestion_schedules_data_source ON neuronip.ingestion_schedules(data_source_id);
CREATE INDEX idx_ingestion_schedules_enabled ON neuronip.ingestion_schedules(enabled) WHERE enabled = true;
CREATE INDEX idx_ingestion_schedules_next_run ON neuronip.ingestion_schedules(next_run_at) WHERE enabled = true;

-- Ingestion metrics table
CREATE TABLE IF NOT EXISTS neuronip.ingestion_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES neuronip.ingestion_jobs(id),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metric_unit VARCHAR(50),
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingestion_metrics_job ON neuronip.ingestion_metrics(job_id);
CREATE INDEX idx_ingestion_metrics_data_source ON neuronip.ingestion_metrics(data_source_id);
CREATE INDEX idx_ingestion_metrics_name ON neuronip.ingestion_metrics(metric_name);
CREATE INDEX idx_ingestion_metrics_timestamp ON neuronip.ingestion_metrics(timestamp);

-- Schema evolution tracking
CREATE TABLE IF NOT EXISTS neuronip.schema_evolution_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id),
    table_name VARCHAR(255) NOT NULL,
    change_type VARCHAR(50) NOT NULL, -- 'column_added', 'column_removed', 'column_modified', 'type_changed'
    old_schema JSONB,
    new_schema JSONB,
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    handled_at TIMESTAMP WITH TIME ZONE,
    action_taken VARCHAR(100), -- 'auto_apply', 'manual_review', 'rejected'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schema_evolution_data_source ON neuronip.schema_evolution_log(data_source_id);
CREATE INDEX idx_schema_evolution_table ON neuronip.schema_evolution_log(table_name);
CREATE INDEX idx_schema_evolution_detected ON neuronip.schema_evolution_log(detected_at);

-- Data validation rules
CREATE TABLE IF NOT EXISTS neuronip.ingestion_validation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    data_source_id UUID NOT NULL REFERENCES neuronip.data_sources(id),
    table_name VARCHAR(255),
    column_name VARCHAR(255),
    rule_type VARCHAR(50) NOT NULL, -- 'not_null', 'unique', 'range', 'regex', 'custom'
    rule_config JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'error', -- 'error', 'warning', 'info'
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_validation_rules_data_source ON neuronip.ingestion_validation_rules(data_source_id);
CREATE INDEX idx_validation_rules_table ON neuronip.ingestion_validation_rules(table_name);
CREATE INDEX idx_validation_rules_enabled ON neuronip.ingestion_validation_rules(enabled) WHERE enabled = true;

-- Validation results
CREATE TABLE IF NOT EXISTS neuronip.ingestion_validation_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES neuronip.ingestion_jobs(id),
    rule_id UUID NOT NULL REFERENCES neuronip.ingestion_validation_rules(id),
    validation_status VARCHAR(50) NOT NULL, -- 'passed', 'failed', 'warning'
    error_message TEXT,
    affected_rows INTEGER,
    sample_data JSONB,
    validated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_validation_results_job ON neuronip.ingestion_validation_results(job_id);
CREATE INDEX idx_validation_results_rule ON neuronip.ingestion_validation_results(rule_id);
CREATE INDEX idx_validation_results_status ON neuronip.ingestion_validation_results(validation_status);

-- Bulk ingestion batches
CREATE TABLE IF NOT EXISTS neuronip.ingestion_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES neuronip.ingestion_jobs(id),
    batch_number INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    rows_in_batch INTEGER DEFAULT 0,
    rows_processed INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingestion_batches_job ON neuronip.ingestion_batches(job_id);
CREATE INDEX idx_ingestion_batches_status ON neuronip.ingestion_batches(status);

-- Parallel ingestion workers
CREATE TABLE IF NOT EXISTS neuronip.ingestion_workers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES neuronip.ingestion_jobs(id),
    worker_id VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'idle', -- 'idle', 'processing', 'completed', 'failed'
    rows_processed INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ingestion_workers_job ON neuronip.ingestion_workers(job_id);
CREATE INDEX idx_ingestion_workers_status ON neuronip.ingestion_workers(status);

COMMENT ON TABLE neuronip.ingestion_schedules IS 'Scheduled ingestion jobs';
COMMENT ON TABLE neuronip.ingestion_metrics IS 'Ingestion performance metrics';
COMMENT ON TABLE neuronip.schema_evolution_log IS 'Schema change tracking for ingestion';
COMMENT ON TABLE neuronip.ingestion_validation_rules IS 'Data validation rules for ingestion';
COMMENT ON TABLE neuronip.ingestion_validation_results IS 'Validation results for ingestion jobs';
COMMENT ON TABLE neuronip.ingestion_batches IS 'Bulk ingestion batch tracking';
COMMENT ON TABLE neuronip.ingestion_workers IS 'Parallel ingestion worker tracking';
