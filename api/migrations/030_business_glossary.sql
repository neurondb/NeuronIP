-- Migration: Business Glossary and Data Dictionary
-- Description: Adds tables for business glossary and data dictionary

-- Glossary Terms: Business glossary terms
CREATE TABLE IF NOT EXISTS neuronip.glossary_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    definition TEXT NOT NULL,
    category TEXT,
    tags TEXT[] DEFAULT '{}',
    related_terms UUID[] DEFAULT '{}',
    owned_by TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.glossary_terms IS 'Business glossary terms';

CREATE INDEX IF NOT EXISTS idx_glossary_terms_name ON neuronip.glossary_terms(name);
CREATE INDEX IF NOT EXISTS idx_glossary_terms_category ON neuronip.glossary_terms(category);
CREATE INDEX IF NOT EXISTS idx_glossary_terms_tags ON neuronip.glossary_terms USING GIN(tags);

-- Data Dictionary: Data dictionary entries with business context
CREATE TABLE IF NOT EXISTS neuronip.data_dictionary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT,
    business_name TEXT,
    description TEXT,
    data_type TEXT,
    business_rules TEXT[] DEFAULT '{}',
    examples TEXT[] DEFAULT '{}',
    term_id UUID REFERENCES neuronip.glossary_terms(id) ON DELETE SET NULL,
    owned_by TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.data_dictionary IS 'Data dictionary entries with business context';

CREATE INDEX IF NOT EXISTS idx_data_dictionary_connector ON neuronip.data_dictionary(connector_id);
CREATE INDEX IF NOT EXISTS idx_data_dictionary_table ON neuronip.data_dictionary(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_data_dictionary_column ON neuronip.data_dictionary(column_name);
CREATE INDEX IF NOT EXISTS idx_data_dictionary_term ON neuronip.data_dictionary(term_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_data_dictionary_unique ON neuronip.data_dictionary(connector_id, schema_name, table_name, column_name);
