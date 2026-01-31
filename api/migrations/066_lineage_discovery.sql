-- Migration: 066_lineage_discovery.sql
-- Description: Lineage discovery rules and discovered lineage for automatic lineage from audit/API/ETL.

-- Lineage discovery rules
CREATE TABLE IF NOT EXISTS neuronip.lineage_discovery_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    source_type TEXT NOT NULL CHECK (source_type IN ('query_log', 'sql_parser', 'api_call', 'etl_job')),
    pattern JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.lineage_discovery_rules IS 'Rules for automatic lineage discovery';

CREATE INDEX IF NOT EXISTS idx_lineage_discovery_rules_enabled ON neuronip.lineage_discovery_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_lineage_discovery_rules_source_type ON neuronip.lineage_discovery_rules(source_type);

-- Discovered lineage (candidates to be verified and promoted to lineage_edges)
CREATE TABLE IF NOT EXISTS neuronip.discovered_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES neuronip.lineage_discovery_rules(id) ON DELETE CASCADE,
    source_node_id UUID NOT NULL REFERENCES neuronip.lineage_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES neuronip.lineage_nodes(id) ON DELETE CASCADE,
    edge_type TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.5 CHECK (confidence >= 0 AND confidence <= 1),
    evidence JSONB DEFAULT '{}',
    verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.discovered_lineage IS 'Auto-discovered lineage candidates';

CREATE INDEX IF NOT EXISTS idx_discovered_lineage_rule ON neuronip.discovered_lineage(rule_id);
CREATE INDEX IF NOT EXISTS idx_discovered_lineage_verified ON neuronip.discovered_lineage(verified);
CREATE INDEX IF NOT EXISTS idx_discovered_lineage_source ON neuronip.discovered_lineage(source_node_id);
CREATE INDEX IF NOT EXISTS idx_discovered_lineage_target ON neuronip.discovered_lineage(target_node_id);
