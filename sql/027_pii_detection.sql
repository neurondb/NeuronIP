-- Migration: PII Detection
-- Description: Adds tables for automated PII detection

-- PII Detections: Automated PII detection results
CREATE TABLE IF NOT EXISTS neuronip.pii_detections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    pii_types JSONB NOT NULL DEFAULT '[]',
    classification TEXT NOT NULL CHECK (classification IN ('sensitive', 'public', 'internal')),
    risk_level TEXT NOT NULL CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.pii_detections IS 'Automated PII detection results';

CREATE INDEX IF NOT EXISTS idx_pii_detections_connector ON neuronip.pii_detections(connector_id);
CREATE INDEX IF NOT EXISTS idx_pii_detections_table ON neuronip.pii_detections(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_pii_detections_column ON neuronip.pii_detections(column_name);
CREATE INDEX IF NOT EXISTS idx_pii_detections_risk ON neuronip.pii_detections(risk_level);
CREATE INDEX IF NOT EXISTS idx_pii_detections_detected ON neuronip.pii_detections(detected_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pii_detections_unique ON neuronip.pii_detections(connector_id, schema_name, table_name, column_name);
