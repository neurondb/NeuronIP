-- Migration: Column-Level Lineage
-- Description: Extends lineage system to support column-level tracking

-- Column Lineage Nodes: Column-level lineage nodes
CREATE TABLE IF NOT EXISTS neuronip.column_lineage_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connector_id UUID REFERENCES neuronip.data_source_connectors(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    node_type TEXT NOT NULL CHECK (node_type IN ('source', 'derived', 'aggregated', 'transformed')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connector_id, schema_name, table_name, column_name)
);
COMMENT ON TABLE neuronip.column_lineage_nodes IS 'Column-level lineage nodes';

CREATE INDEX IF NOT EXISTS idx_column_lineage_connector ON neuronip.column_lineage_nodes(connector_id);
CREATE INDEX IF NOT EXISTS idx_column_lineage_table ON neuronip.column_lineage_nodes(schema_name, table_name);
CREATE INDEX IF NOT EXISTS idx_column_lineage_column ON neuronip.column_lineage_nodes(column_name);

-- Column Lineage Edges: Column-level lineage relationships
CREATE TABLE IF NOT EXISTS neuronip.column_lineage_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_node_id UUID NOT NULL REFERENCES neuronip.column_lineage_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES neuronip.column_lineage_nodes(id) ON DELETE CASCADE,
    edge_type TEXT NOT NULL CHECK (edge_type IN ('reads', 'transforms', 'writes', 'depends_on', 'aggregates')),
    transformation JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_node_id, target_node_id, edge_type)
);
COMMENT ON TABLE neuronip.column_lineage_edges IS 'Column-level lineage edges';

CREATE INDEX IF NOT EXISTS idx_column_lineage_edges_source ON neuronip.column_lineage_edges(source_node_id);
CREATE INDEX IF NOT EXISTS idx_column_lineage_edges_target ON neuronip.column_lineage_edges(target_node_id);
CREATE INDEX IF NOT EXISTS idx_column_lineage_edges_type ON neuronip.column_lineage_edges(edge_type);
