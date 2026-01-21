-- Migration: Automated Data Classification
-- Description: Adds tables for automated data classification (PII, PHI, PCI detection)

-- Data Classifications: Classification results for columns
CREATE TABLE IF NOT EXISTS neuronip.data_classifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID NOT NULL REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    classification_type TEXT NOT NULL CHECK (classification_type IN ('pii', 'phi', 'pci', 'sensitive', 'public', 'internal', 'confidential')),
    confidence NUMERIC NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    detection_method TEXT NOT NULL CHECK (detection_method IN ('pattern', 'ml', 'manual', 'rule')),
    detected_patterns TEXT[],
    metadata JSONB DEFAULT '{}',
    classified_by TEXT,
    classified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, column_name)
);
COMMENT ON TABLE neuronip.data_classifications IS 'Automated data classification results';

CREATE INDEX IF NOT EXISTS idx_classifications_connector ON neuronip.data_classifications(connector_id);
CREATE INDEX IF NOT EXISTS idx_classifications_table ON neuronip.data_classifications(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_classifications_type ON neuronip.data_classifications(classification_type);
CREATE INDEX IF NOT EXISTS idx_classifications_confidence ON neuronip.data_classifications(confidence);

-- Classification Rules: Rules for automated classification
CREATE TABLE IF NOT EXISTS neuronip.classification_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    classification_type TEXT NOT NULL CHECK (classification_type IN ('pii', 'phi', 'pci', 'sensitive', 'public', 'internal', 'confidential')),
    rule_type TEXT NOT NULL CHECK (rule_type IN ('pattern', 'keyword', 'column_name', 'ml_model')),
    rule_expression TEXT NOT NULL,
    confidence_threshold NUMERIC DEFAULT 0.7,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.classification_rules IS 'Rules for automated data classification';

CREATE INDEX IF NOT EXISTS idx_classification_rules_type ON neuronip.classification_rules(classification_type);
CREATE INDEX IF NOT EXISTS idx_classification_rules_enabled ON neuronip.classification_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_classification_rules_priority ON neuronip.classification_rules(priority);
