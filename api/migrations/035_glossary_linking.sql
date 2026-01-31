-- Migration: Glossary Linking Schema
-- Description: Adds tables for linking glossary terms to columns, metrics, and documents

-- Glossary links: Links between glossary terms and entities
CREATE TABLE IF NOT EXISTS neuronip.glossary_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    glossary_id UUID NOT NULL REFERENCES neuronip.glossary(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL CHECK (entity_type IN ('column', 'metric', 'document', 'table')),
    entity_id UUID NOT NULL,
    entity_name TEXT NOT NULL,
    relationship TEXT NOT NULL CHECK (relationship IN ('defines', 'related_to', 'synonym_of', 'derived_from')),
    confidence FLOAT DEFAULT 1.0 CHECK (confidence >= 0 AND confidence <= 1),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(glossary_id, entity_type, entity_id)
);
COMMENT ON TABLE neuronip.glossary_links IS 'Links between glossary terms and entities (columns, metrics, documents, tables)';

CREATE INDEX IF NOT EXISTS idx_glossary_links_glossary ON neuronip.glossary_links(glossary_id);
CREATE INDEX IF NOT EXISTS idx_glossary_links_entity ON neuronip.glossary_links(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_glossary_links_confidence ON neuronip.glossary_links(confidence DESC);
