-- Migration: 049_knowledge_graph_enhancements.sql
-- Description: Knowledge graph maturity enhancements

-- Graph relationships (enhanced)
CREATE TABLE IF NOT EXISTS neuronip.kg_relationships_enhanced (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity_id UUID NOT NULL,
    target_entity_id UUID NOT NULL,
    relationship_type VARCHAR(100) NOT NULL,
    relationship_strength DOUBLE PRECISION DEFAULT 1.0,
    confidence_score DOUBLE PRECISION DEFAULT 1.0,
    temporal_start TIMESTAMP WITH TIME ZONE,
    temporal_end TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kg_relationships_source ON neuronip.kg_relationships_enhanced(source_entity_id);
CREATE INDEX idx_kg_relationships_target ON neuronip.kg_relationships_enhanced(target_entity_id);
CREATE INDEX idx_kg_relationships_type ON neuronip.kg_relationships_enhanced(relationship_type);
CREATE INDEX idx_kg_relationships_temporal ON neuronip.kg_relationships_enhanced(temporal_start, temporal_end);

-- Graph analytics results
CREATE TABLE IF NOT EXISTS neuronip.kg_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    analysis_type VARCHAR(50) NOT NULL, -- 'pagerank', 'community_detection', 'centrality', 'path_analysis'
    entity_id UUID,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    metadata JSONB,
    computed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kg_analytics_type ON neuronip.kg_analytics(analysis_type);
CREATE INDEX idx_kg_analytics_entity ON neuronip.kg_analytics(entity_id);
CREATE INDEX idx_kg_analytics_computed ON neuronip.kg_analytics(computed_at);

-- Graph inference rules
CREATE TABLE IF NOT EXISTS neuronip.kg_inference_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name VARCHAR(255) NOT NULL,
    rule_type VARCHAR(50) NOT NULL, -- 'transitive', 'symmetric', 'inverse', 'custom'
    rule_definition JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kg_inference_rules_enabled ON neuronip.kg_inference_rules(enabled) WHERE enabled = true;

-- Inferred relationships
CREATE TABLE IF NOT EXISTS neuronip.kg_inferred_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity_id UUID NOT NULL,
    target_entity_id UUID NOT NULL,
    relationship_type VARCHAR(100) NOT NULL,
    inference_rule_id UUID NOT NULL REFERENCES neuronip.kg_inference_rules(id),
    confidence_score DOUBLE PRECISION NOT NULL,
    inference_path JSONB, -- Path of relationships that led to inference
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kg_inferred_source ON neuronip.kg_inferred_relationships(source_entity_id);
CREATE INDEX idx_kg_inferred_target ON neuronip.kg_inferred_relationships(target_entity_id);
CREATE INDEX idx_kg_inferred_rule ON neuronip.kg_inferred_relationships(inference_rule_id);

COMMENT ON TABLE neuronip.kg_relationships_enhanced IS 'Enhanced knowledge graph relationships with temporal support';
COMMENT ON TABLE neuronip.kg_analytics IS 'Knowledge graph analytics results';
COMMENT ON TABLE neuronip.kg_inference_rules IS 'Rules for relationship inference';
COMMENT ON TABLE neuronip.kg_inferred_relationships IS 'Inferred relationships from rules';
