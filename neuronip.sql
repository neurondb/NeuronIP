/*-------------------------------------------------------------------------
 *
 * neuronip.sql
 *    Complete NeuronIP Database Setup Script
 *
 * This script sets up everything needed for NeuronIP:
 * - Database schema (tables, indexes, views, triggers)
 * - Management functions
 * - Pre-populated default data
 *
 * This script is idempotent and can be run multiple times safely.
 *
 * Copyright (c) 2024-2026, neurondb, Inc. <support@neurondb.ai>
 *
 * IDENTIFICATION
 *    NeuronIP/neuronip.sql
 *
 *-------------------------------------------------------------------------
 *
 * PREREQUISITES
 * =============
 *
 * - PostgreSQL 16 or later
 * - NeuronDB extension installed
 * - Database user with CREATE privileges
 *
 * USAGE
 * =====
 *
 * To run this setup script on a database:
 *
 *   psql -d your_database -f neuronip.sql
 *
 * Or from within psql:
 *
 *   \i neuronip.sql
 *
 *-------------------------------------------------------------------------
 */

-- ============================================================================
-- SECTION 1: EXTENSIONS
-- ============================================================================

-- Ensure required extensions are available
CREATE EXTENSION IF NOT EXISTS neurondb;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- SECTION 2: SCHEMA CREATION
-- ============================================================================

-- Schema: neuronip
CREATE SCHEMA IF NOT EXISTS neuronip;

-- ============================================================================
-- SECTION 3: SEMANTIC KNOWLEDGE SEARCH TABLES
-- ============================================================================

-- Knowledge collections: Organized knowledge collections
CREATE TABLE IF NOT EXISTS neuronip.knowledge_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',
    UNIQUE(name)
);
COMMENT ON TABLE neuronip.knowledge_collections IS 'Organized knowledge collections';

-- Knowledge documents: Documents, tickets, policies, logs
CREATE TABLE IF NOT EXISTS neuronip.knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID REFERENCES neuronip.knowledge_collections(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    content_type TEXT NOT NULL CHECK (content_type IN ('document', 'ticket', 'policy', 'log', 'other')),
    source TEXT,
    source_url TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.knowledge_documents IS 'Documents, tickets, policies, logs for semantic search';

-- Knowledge embeddings: Vector embeddings for documents
CREATE TABLE IF NOT EXISTS neuronip.knowledge_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES neuronip.knowledge_documents(id) ON DELETE CASCADE,
    embedding vector(1536),  -- Default embedding dimension, adjustable
    model_name TEXT NOT NULL DEFAULT 'sentence-transformers/all-MiniLM-L6-v2',
    chunk_index INTEGER DEFAULT 0,
    chunk_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, chunk_index)
);
COMMENT ON TABLE neuronip.knowledge_embeddings IS 'Vector embeddings for knowledge documents';

-- Create HNSW index on embeddings
CREATE INDEX IF NOT EXISTS idx_knowledge_embeddings_hnsw 
    ON neuronip.knowledge_embeddings 
    USING hnsw (embedding vector_cosine_ops);

-- Search history: User search queries and results
CREATE TABLE IF NOT EXISTS neuronip.search_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    query_text TEXT NOT NULL,
    query_embedding vector(1536),
    results JSONB DEFAULT '[]',
    result_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.search_history IS 'User search queries and results';

-- ============================================================================
-- SECTION 4: DATA WAREHOUSE Q&A TABLES
-- ============================================================================

-- Warehouse schemas: Data warehouse schema metadata
CREATE TABLE IF NOT EXISTS neuronip.warehouse_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_name TEXT NOT NULL,
    database_name TEXT NOT NULL,
    description TEXT,
    tables JSONB DEFAULT '[]',  -- Array of table metadata
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(schema_name, database_name)
);
COMMENT ON TABLE neuronip.warehouse_schemas IS 'Data warehouse schema metadata';

-- Warehouse queries: Stored natural language queries
CREATE TABLE IF NOT EXISTS neuronip.warehouse_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    natural_language_query TEXT NOT NULL,
    generated_sql TEXT,
    schema_id UUID REFERENCES neuronip.warehouse_schemas(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'executing', 'completed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.warehouse_queries IS 'Stored natural language queries';

-- Query results: Cached query results with SQL + charts
CREATE TABLE IF NOT EXISTS neuronip.query_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID NOT NULL REFERENCES neuronip.warehouse_queries(id) ON DELETE CASCADE,
    result_data JSONB DEFAULT '[]',  -- Query result rows
    chart_config JSONB,  -- Chart configuration
    chart_type TEXT CHECK (chart_type IN ('table', 'bar', 'line', 'pie', 'scatter', 'area')),
    execution_time_ms INTEGER,
    row_count INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.query_results IS 'Cached query results with SQL + charts';

-- Query explanations: AI-generated explanations
CREATE TABLE IF NOT EXISTS neuronip.query_explanations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query_id UUID NOT NULL REFERENCES neuronip.warehouse_queries(id) ON DELETE CASCADE,
    explanation_text TEXT NOT NULL,
    explanation_type TEXT NOT NULL CHECK (explanation_type IN ('sql', 'result', 'insight')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.query_explanations IS 'AI-generated query explanations';

-- ============================================================================
-- SECTION 5: CUSTOMER SUPPORT MEMORY TABLES
-- ============================================================================

-- Support agents: Support agent configurations
CREATE TABLE IF NOT EXISTS neuronip.support_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    agent_id UUID,  -- Reference to NeuronAgent agent
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name)
);
COMMENT ON TABLE neuronip.support_agents IS 'Support agent configurations';

-- Support tickets: Customer support tickets
CREATE TABLE IF NOT EXISTS neuronip.support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number TEXT NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    customer_email TEXT,
    subject TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    assigned_agent_id UUID REFERENCES neuronip.support_agents(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);
COMMENT ON TABLE neuronip.support_tickets IS 'Customer support tickets';

-- Support conversations: Conversation history
CREATE TABLE IF NOT EXISTS neuronip.support_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES neuronip.support_tickets(id) ON DELETE CASCADE,
    message_text TEXT NOT NULL,
    sender_type TEXT NOT NULL CHECK (sender_type IN ('customer', 'agent', 'system')),
    sender_id TEXT,
    embedding vector(1536),  -- For semantic search
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.support_conversations IS 'Support conversation history';

-- Create HNSW index on conversation embeddings
CREATE INDEX IF NOT EXISTS idx_support_conversations_hnsw 
    ON neuronip.support_conversations 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- Support memory: Long-term memory for customers
CREATE TABLE IF NOT EXISTS neuronip.support_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id TEXT NOT NULL,
    memory_type TEXT NOT NULL CHECK (memory_type IN ('preference', 'issue', 'interaction', 'context')),
    memory_content TEXT NOT NULL,
    embedding vector(1536),
    importance_score FLOAT DEFAULT 0.5,
    last_accessed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.support_memory IS 'Long-term memory for customers';

-- Create HNSW index on support memory embeddings
CREATE INDEX IF NOT EXISTS idx_support_memory_hnsw 
    ON neuronip.support_memory 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- ============================================================================
-- SECTION 6: COMPLIANCE & AUDIT ANALYTICS TABLES
-- ============================================================================

-- Compliance policies: Compliance policies and rules
CREATE TABLE IF NOT EXISTS neuronip.compliance_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name TEXT NOT NULL,
    policy_type TEXT NOT NULL CHECK (policy_type IN ('regulation', 'internal', 'standard')),
    description TEXT,
    policy_text TEXT NOT NULL,
    rules JSONB DEFAULT '[]',  -- Structured rules
    embedding vector(1536),  -- For semantic matching
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(policy_name)
);
COMMENT ON TABLE neuronip.compliance_policies IS 'Compliance policies and rules';

-- Create HNSW index on policy embeddings
CREATE INDEX IF NOT EXISTS idx_compliance_policies_hnsw 
    ON neuronip.compliance_policies 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- Audit events: Audit trail events
CREATE TABLE IF NOT EXISTS neuronip.audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    user_id TEXT,
    action TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.audit_events IS 'Audit trail events';

-- Compliance matches: Policy matching results
CREATE TABLE IF NOT EXISTS neuronip.compliance_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES neuronip.compliance_policies(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    match_score FLOAT NOT NULL,
    match_details JSONB DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'resolved', 'false_positive')),
    reviewed_by TEXT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.compliance_matches IS 'Policy matching results';

-- Anomaly detections: Detected anomalies
CREATE TABLE IF NOT EXISTS neuronip.anomaly_detections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    detection_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    anomaly_score FLOAT NOT NULL,
    details JSONB DEFAULT '{}',
    model_name TEXT,
    status TEXT NOT NULL DEFAULT 'detected' CHECK (status IN ('detected', 'investigating', 'resolved', 'false_positive')),
    resolved_by TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.anomaly_detections IS 'Detected anomalies';

-- ============================================================================
-- SECTION 7: AGENT WORKFLOWS TABLES
-- ============================================================================

-- Workflows: Workflow definitions
CREATE TABLE IF NOT EXISTS neuronip.workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    workflow_definition JSONB NOT NULL,  -- DAG structure, steps, conditions
    agent_id UUID,  -- Reference to NeuronAgent agent
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name)
);
COMMENT ON TABLE neuronip.workflows IS 'Workflow definitions';

-- Workflow executions: Execution history
CREATE TABLE IF NOT EXISTS neuronip.workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    input_data JSONB DEFAULT '{}',
    output_data JSONB DEFAULT '{}',
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workflow_executions IS 'Workflow execution history';

-- Workflow memory: Long-term workflow memory
CREATE TABLE IF NOT EXISTS neuronip.workflow_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES neuronip.workflows(id) ON DELETE CASCADE,
    execution_id UUID REFERENCES neuronip.workflow_executions(id) ON DELETE SET NULL,
    memory_key TEXT NOT NULL,
    memory_value JSONB NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workflow_id, memory_key)
);
COMMENT ON TABLE neuronip.workflow_memory IS 'Long-term workflow memory';

-- Create HNSW index on workflow memory embeddings
CREATE INDEX IF NOT EXISTS idx_workflow_memory_hnsw 
    ON neuronip.workflow_memory 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- Workflow decisions: Decision tracking
CREATE TABLE IF NOT EXISTS neuronip.workflow_decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES neuronip.workflow_executions(id) ON DELETE CASCADE,
    decision_point TEXT NOT NULL,
    decision_value TEXT NOT NULL,
    decision_reason TEXT,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.workflow_decisions IS 'Workflow decision tracking';

-- Agent memory: Persistent memory storage for agents
CREATE TABLE IF NOT EXISTS neuronip.agent_memory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    memory_key TEXT NOT NULL,
    memory_value JSONB NOT NULL,
    embedding vector(1536),
    importance_score FLOAT DEFAULT 0.5,
    last_accessed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, memory_key)
);
COMMENT ON TABLE neuronip.agent_memory IS 'Persistent memory storage for agents';

-- Create HNSW index on agent memory embeddings
CREATE INDEX IF NOT EXISTS idx_agent_memory_hnsw 
    ON neuronip.agent_memory 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- ============================================================================
-- SECTION 8: KNOWLEDGE GRAPH TABLES
-- ============================================================================

-- Entity types: Classification of entities in the knowledge graph
CREATE TABLE IF NOT EXISTS neuronip.entity_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type_name TEXT NOT NULL UNIQUE,
    description TEXT,
    parent_type_id UUID REFERENCES neuronip.entity_types(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.entity_types IS 'Entity type classifications for knowledge graph';

-- Entities: Extracted entities from documents and data
CREATE TABLE IF NOT EXISTS neuronip.entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_name TEXT NOT NULL,
    entity_type_id UUID REFERENCES neuronip.entity_types(id) ON DELETE SET NULL,
    entity_value TEXT,
    description TEXT,
    source_document_id UUID REFERENCES neuronip.knowledge_documents(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536),
    confidence_score FLOAT DEFAULT 1.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(entity_name, entity_type_id, source_document_id)
);
COMMENT ON TABLE neuronip.entities IS 'Extracted entities for knowledge graph';

-- Create HNSW index on entity embeddings
CREATE INDEX IF NOT EXISTS idx_entities_hnsw 
    ON neuronip.entities 
    USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

-- Create index on entity type
CREATE INDEX IF NOT EXISTS idx_entities_type ON neuronip.entities(entity_type_id);

-- Entity links: Relationships between entities
CREATE TABLE IF NOT EXISTS neuronip.entity_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_entity_id UUID NOT NULL REFERENCES neuronip.entities(id) ON DELETE CASCADE,
    target_entity_id UUID NOT NULL REFERENCES neuronip.entities(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL,
    relationship_strength FLOAT DEFAULT 1.0,
    description TEXT,
    source_document_id UUID REFERENCES neuronip.knowledge_documents(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_entity_id, target_entity_id, relationship_type)
);
COMMENT ON TABLE neuronip.entity_links IS 'Entity relationships in knowledge graph';

-- Create indexes on entity links
CREATE INDEX IF NOT EXISTS idx_entity_links_source ON neuronip.entity_links(source_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_links_target ON neuronip.entity_links(target_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_links_type ON neuronip.entity_links(relationship_type);

-- Glossary: Semantic layer annotations and business terms
CREATE TABLE IF NOT EXISTS neuronip.glossary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    term TEXT NOT NULL UNIQUE,
    definition TEXT NOT NULL,
    category TEXT,
    related_entity_id UUID REFERENCES neuronip.entities(id) ON DELETE SET NULL,
    synonyms TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.glossary IS 'Business glossary and semantic layer annotations';

-- Create index on glossary category
CREATE INDEX IF NOT EXISTS idx_glossary_category ON neuronip.glossary(category);

-- ============================================================================
-- SECTION 9: INTEGRATION TABLES
-- ============================================================================

-- Integrations: External system integrations
CREATE TABLE IF NOT EXISTS neuronip.integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    integration_type TEXT NOT NULL CHECK (integration_type IN ('api', 'database', 'webhook', 'file', 'other')),
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_sync_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name)
);
COMMENT ON TABLE neuronip.integrations IS 'External system integrations';

-- Data sources: Connected data sources
CREATE TABLE IF NOT EXISTS neuronip.data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('postgresql', 'mysql', 'mongodb', 'api', 'file', 'other')),
    connection_string TEXT,
    config JSONB DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_accessed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name)
);
COMMENT ON TABLE neuronip.data_sources IS 'Connected data sources';

-- API keys: API key management
CREATE TABLE IF NOT EXISTS neuronip.api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL UNIQUE,
    user_id TEXT,
    name TEXT,
    rate_limit INTEGER NOT NULL DEFAULT 100,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.api_keys IS 'API key management';

-- Enhance data_sources table with sync schedules and status
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS sync_schedule TEXT;
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS sync_status TEXT CHECK (sync_status IN ('idle', 'syncing', 'error', 'paused')) DEFAULT 'idle';
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS last_sync_at TIMESTAMPTZ;
ALTER TABLE neuronip.data_sources ADD COLUMN IF NOT EXISTS sync_error TEXT;

-- Metrics / Semantic Layer
CREATE TABLE IF NOT EXISTS neuronip.metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    definition TEXT NOT NULL,
    kpi_type TEXT CHECK (kpi_type IN ('revenue', 'growth', 'efficiency', 'quality', 'other')),
    business_term TEXT,
    reusable BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.metrics IS 'Metrics and KPIs for semantic layer';

CREATE TABLE IF NOT EXISTS neuronip.metric_dimensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id UUID NOT NULL REFERENCES neuronip.metrics(id) ON DELETE CASCADE,
    dimension_name TEXT NOT NULL,
    dimension_type TEXT CHECK (dimension_type IN ('time', 'geographic', 'category', 'custom')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(metric_id, dimension_name)
);
COMMENT ON TABLE neuronip.metric_dimensions IS 'Dimensions for metrics';

-- Agent Hub
CREATE TABLE IF NOT EXISTS neuronip.agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    agent_type TEXT NOT NULL CHECK (agent_type IN ('workflow', 'support', 'analytics', 'automation', 'custom')),
    config JSONB DEFAULT '{}',
    status TEXT NOT NULL CHECK (status IN ('draft', 'active', 'paused', 'error')) DEFAULT 'draft',
    performance_metrics JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agents IS 'Agent management';

CREATE TABLE IF NOT EXISTS neuronip.agent_performance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES neuronip.agents(id) ON DELETE CASCADE,
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.agent_performance IS 'Agent performance metrics';

-- Observability: System logs
CREATE TABLE IF NOT EXISTS neuronip.system_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_type TEXT NOT NULL CHECK (log_type IN ('query', 'agent', 'workflow', 'system', 'error')),
    level TEXT NOT NULL CHECK (level IN ('debug', 'info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL,
    context JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.system_logs IS 'System logs for observability';

-- Data Lineage
CREATE TABLE IF NOT EXISTS neuronip.lineage_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_type TEXT NOT NULL CHECK (node_type IN ('source', 'table', 'view', 'transformation', 'target')),
    node_name TEXT NOT NULL,
    schema_info JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.lineage_nodes IS 'Lineage graph nodes';

CREATE TABLE IF NOT EXISTS neuronip.lineage_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_node_id UUID NOT NULL REFERENCES neuronip.lineage_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES neuronip.lineage_nodes(id) ON DELETE CASCADE,
    edge_type TEXT NOT NULL CHECK (edge_type IN ('reads', 'transforms', 'writes', 'depends_on')),
    transformation JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_node_id, target_node_id, edge_type)
);
COMMENT ON TABLE neuronip.lineage_edges IS 'Lineage graph edges';

-- Billing / Usage
CREATE TABLE IF NOT EXISTS neuronip.usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type TEXT NOT NULL CHECK (metric_type IN ('api_call', 'query', 'embedding', 'seat', 'storage', 'compute')),
    metric_name TEXT NOT NULL,
    count INTEGER NOT NULL DEFAULT 0,
    user_id TEXT,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);
COMMENT ON TABLE neuronip.usage_metrics IS 'Usage metrics tracking';

CREATE TABLE IF NOT EXISTS neuronip.billing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    seats INTEGER NOT NULL DEFAULT 0,
    api_calls INTEGER NOT NULL DEFAULT 0,
    queries INTEGER NOT NULL DEFAULT 0,
    embeddings INTEGER NOT NULL DEFAULT 0,
    cost NUMERIC(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.billing_records IS 'Billing records';

-- Versioning
CREATE TABLE IF NOT EXISTS neuronip.versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('model', 'embedding', 'workflow', 'metric', 'agent')),
    resource_id UUID NOT NULL,
    version_number TEXT NOT NULL,
    version_data JSONB NOT NULL,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_current BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(resource_type, resource_id, version_number)
);
COMMENT ON TABLE neuronip.versions IS 'Version control for resources';

CREATE TABLE IF NOT EXISTS neuronip.version_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id UUID NOT NULL REFERENCES neuronip.versions(id) ON DELETE CASCADE,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'rollback', 'delete')),
    changes JSONB DEFAULT '{}',
    rollback_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.version_history IS 'Version history and changes';

-- Data Catalog
CREATE TABLE IF NOT EXISTS neuronip.catalog_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    schema_info JSONB DEFAULT '{}',
    fields JSONB DEFAULT '[]',
    owner TEXT,
    description TEXT,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE neuronip.catalog_datasets IS 'Data catalog datasets';

CREATE TABLE IF NOT EXISTS neuronip.catalog_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES neuronip.catalog_datasets(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL,
    field_type TEXT NOT NULL,
    description TEXT,
    semantic_tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(dataset_id, field_name)
);
COMMENT ON TABLE neuronip.catalog_fields IS 'Data catalog field metadata';

-- Enhance integrations table for webhooks and triggers
ALTER TABLE neuronip.integrations ADD COLUMN IF NOT EXISTS webhook_url TEXT;
ALTER TABLE neuronip.integrations ADD COLUMN IF NOT EXISTS webhook_secret TEXT;
ALTER TABLE neuronip.integrations ADD COLUMN IF NOT EXISTS triggers JSONB DEFAULT '[]';

-- ============================================================================
-- SECTION 9: INDEXES
-- ============================================================================

-- Knowledge documents indexes
CREATE INDEX IF NOT EXISTS idx_knowledge_documents_collection ON neuronip.knowledge_documents(collection_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_documents_type ON neuronip.knowledge_documents(content_type);
CREATE INDEX IF NOT EXISTS idx_knowledge_documents_created ON neuronip.knowledge_documents(created_at DESC);

-- Search history indexes
CREATE INDEX IF NOT EXISTS idx_search_history_user ON neuronip.search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_search_history_created ON neuronip.search_history(created_at DESC);

-- Warehouse queries indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_user ON neuronip.warehouse_queries(user_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_status ON neuronip.warehouse_queries(status);
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_created ON neuronip.warehouse_queries(created_at DESC);

-- Support tickets indexes
CREATE INDEX IF NOT EXISTS idx_support_tickets_customer ON neuronip.support_tickets(customer_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON neuronip.support_tickets(status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_agent ON neuronip.support_tickets(assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_created ON neuronip.support_tickets(created_at DESC);

-- Support conversations indexes
CREATE INDEX IF NOT EXISTS idx_support_conversations_ticket ON neuronip.support_conversations(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_conversations_created ON neuronip.support_conversations(created_at DESC);

-- Support memory indexes
CREATE INDEX IF NOT EXISTS idx_support_memory_customer ON neuronip.support_memory(customer_id);
CREATE INDEX IF NOT EXISTS idx_support_memory_type ON neuronip.support_memory(memory_type);

-- Audit events indexes
CREATE INDEX IF NOT EXISTS idx_audit_events_type ON neuronip.audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_events_entity ON neuronip.audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_user ON neuronip.audit_events(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_created ON neuronip.audit_events(created_at DESC);

-- Compliance matches indexes
CREATE INDEX IF NOT EXISTS idx_compliance_matches_policy ON neuronip.compliance_matches(policy_id);
CREATE INDEX IF NOT EXISTS idx_compliance_matches_entity ON neuronip.compliance_matches(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_compliance_matches_status ON neuronip.compliance_matches(status);

-- Anomaly detections indexes
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_type ON neuronip.anomaly_detections(detection_type);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_entity ON neuronip.anomaly_detections(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_status ON neuronip.anomaly_detections(status);
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_created ON neuronip.anomaly_detections(created_at DESC);

-- Workflow executions indexes
CREATE INDEX IF NOT EXISTS idx_workflow_executions_workflow ON neuronip.workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_status ON neuronip.workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_created ON neuronip.workflow_executions(created_at DESC);

-- Workflow memory indexes
CREATE INDEX IF NOT EXISTS idx_workflow_memory_workflow ON neuronip.workflow_memory(workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_memory_execution ON neuronip.workflow_memory(execution_id);

-- Workflow decisions indexes
CREATE INDEX IF NOT EXISTS idx_workflow_decisions_execution ON neuronip.workflow_decisions(execution_id);

-- Data sources indexes
CREATE INDEX IF NOT EXISTS idx_data_sources_type ON neuronip.data_sources(source_type);
CREATE INDEX IF NOT EXISTS idx_data_sources_status ON neuronip.data_sources(sync_status);
CREATE INDEX IF NOT EXISTS idx_data_sources_enabled ON neuronip.data_sources(enabled);

-- Metrics indexes
CREATE INDEX IF NOT EXISTS idx_metrics_kpi_type ON neuronip.metrics(kpi_type);
CREATE INDEX IF NOT EXISTS idx_metrics_reusable ON neuronip.metrics(reusable);
CREATE INDEX IF NOT EXISTS idx_metric_dimensions_metric ON neuronip.metric_dimensions(metric_id);

-- Agents indexes
CREATE INDEX IF NOT EXISTS idx_agents_type ON neuronip.agents(agent_type);
CREATE INDEX IF NOT EXISTS idx_agents_status ON neuronip.agents(status);
CREATE INDEX IF NOT EXISTS idx_agent_performance_agent ON neuronip.agent_performance(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_performance_timestamp ON neuronip.agent_performance(timestamp DESC);

-- System logs indexes
CREATE INDEX IF NOT EXISTS idx_system_logs_type ON neuronip.system_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_system_logs_level ON neuronip.system_logs(level);
CREATE INDEX IF NOT EXISTS idx_system_logs_timestamp ON neuronip.system_logs(timestamp DESC);

-- Lineage indexes
CREATE INDEX IF NOT EXISTS idx_lineage_nodes_type ON neuronip.lineage_nodes(node_type);
CREATE INDEX IF NOT EXISTS idx_lineage_edges_source ON neuronip.lineage_edges(source_node_id);
CREATE INDEX IF NOT EXISTS idx_lineage_edges_target ON neuronip.lineage_edges(target_node_id);

-- Usage metrics indexes
CREATE INDEX IF NOT EXISTS idx_usage_metrics_type ON neuronip.usage_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_user ON neuronip.usage_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_timestamp ON neuronip.usage_metrics(timestamp DESC);

-- Billing records indexes
CREATE INDEX IF NOT EXISTS idx_billing_records_user ON neuronip.billing_records(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_records_period ON neuronip.billing_records(period_start, period_end);

-- Versions indexes
CREATE INDEX IF NOT EXISTS idx_versions_resource ON neuronip.versions(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_versions_current ON neuronip.versions(is_current);
CREATE INDEX IF NOT EXISTS idx_version_history_version ON neuronip.version_history(version_id);

-- Catalog indexes
CREATE INDEX IF NOT EXISTS idx_catalog_datasets_owner ON neuronip.catalog_datasets(owner);
CREATE INDEX IF NOT EXISTS idx_catalog_fields_dataset ON neuronip.catalog_fields(dataset_id);
CREATE INDEX IF NOT EXISTS idx_catalog_datasets_tags ON neuronip.catalog_datasets USING GIN(tags);

-- ============================================================================
-- SECTION 10: TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION neuronip.update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers to relevant tables
CREATE TRIGGER update_knowledge_collections_updated_at
    BEFORE UPDATE ON neuronip.knowledge_collections
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_knowledge_documents_updated_at
    BEFORE UPDATE ON neuronip.knowledge_documents
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_warehouse_schemas_updated_at
    BEFORE UPDATE ON neuronip.warehouse_schemas
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_support_agents_updated_at
    BEFORE UPDATE ON neuronip.support_agents
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_support_tickets_updated_at
    BEFORE UPDATE ON neuronip.support_tickets
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_support_memory_updated_at
    BEFORE UPDATE ON neuronip.support_memory
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_compliance_policies_updated_at
    BEFORE UPDATE ON neuronip.compliance_policies
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_workflows_updated_at
    BEFORE UPDATE ON neuronip.workflows
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_workflow_memory_updated_at
    BEFORE UPDATE ON neuronip.workflow_memory
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_integrations_updated_at
    BEFORE UPDATE ON neuronip.integrations
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_data_sources_updated_at
    BEFORE UPDATE ON neuronip.data_sources
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_metrics_updated_at
    BEFORE UPDATE ON neuronip.metrics
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_agents_updated_at
    BEFORE UPDATE ON neuronip.agents
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

CREATE TRIGGER update_catalog_datasets_updated_at
    BEFORE UPDATE ON neuronip.catalog_datasets
    FOR EACH ROW EXECUTE FUNCTION neuronip.update_updated_at();

-- ============================================================================
-- SECTION 11: VIEWS
-- ============================================================================

-- View: Knowledge search summary
CREATE OR REPLACE VIEW neuronip.knowledge_search_summary AS
SELECT 
    kc.id as collection_id,
    kc.name as collection_name,
    COUNT(DISTINCT kd.id) as document_count,
    COUNT(DISTINCT ke.id) as embedding_count,
    MAX(kd.updated_at) as last_updated
FROM neuronip.knowledge_collections kc
LEFT JOIN neuronip.knowledge_documents kd ON kd.collection_id = kc.id
LEFT JOIN neuronip.knowledge_embeddings ke ON ke.document_id = kd.id
GROUP BY kc.id, kc.name;

-- View: Support ticket summary
CREATE OR REPLACE VIEW neuronip.support_ticket_summary AS
SELECT 
    st.status,
    st.priority,
    COUNT(*) as ticket_count,
    COUNT(DISTINCT st.customer_id) as customer_count,
    AVG(EXTRACT(EPOCH FROM (COALESCE(st.resolved_at, NOW()) - st.created_at))/3600) as avg_hours_to_resolve
FROM neuronip.support_tickets st
GROUP BY st.status, st.priority;

-- View: Compliance match summary
CREATE OR REPLACE VIEW neuronip.compliance_match_summary AS
SELECT 
    cp.policy_name,
    cm.status,
    COUNT(*) as match_count,
    AVG(cm.match_score) as avg_match_score
FROM neuronip.compliance_matches cm
JOIN neuronip.compliance_policies cp ON cp.id = cm.policy_id
GROUP BY cp.policy_name, cm.status;

-- View: Workflow execution summary
CREATE OR REPLACE VIEW neuronip.workflow_execution_summary AS
SELECT 
    w.name as workflow_name,
    we.status,
    COUNT(*) as execution_count,
    AVG(EXTRACT(EPOCH FROM (COALESCE(we.completed_at, NOW()) - we.started_at))) as avg_execution_seconds
FROM neuronip.workflow_executions we
JOIN neuronip.workflows w ON w.id = we.workflow_id
WHERE we.started_at IS NOT NULL
GROUP BY w.name, we.status;

-- ============================================================================
-- SECTION 12: GRANTS
-- ============================================================================

-- Grant usage on schema
GRANT USAGE ON SCHEMA neuronip TO PUBLIC;

-- Grant select on all tables (read access)
GRANT SELECT ON ALL TABLES IN SCHEMA neuronip TO PUBLIC;

-- Grant select on all views
GRANT SELECT ON ALL TABLES IN SCHEMA neuronip TO PUBLIC;

-- ============================================================================
-- SECTION 13: INITIAL DATA (Optional)
-- ============================================================================

-- Insert default knowledge collection
INSERT INTO neuronip.knowledge_collections (name, description)
VALUES ('Default Collection', 'Default knowledge collection')
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
