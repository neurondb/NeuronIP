-- Performance optimization indexes
-- This migration adds indexes to improve query performance

-- Composite indexes for common query patterns

-- Warehouse queries: user_id + status + created_at (common filter pattern)
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_user_status_created 
    ON neuronip.warehouse_queries(user_id, status, created_at DESC)
    WHERE executed_at IS NOT NULL;

-- Warehouse queries: schema_id + status (for schema-based queries)
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_schema_status 
    ON neuronip.warehouse_queries(schema_id, status)
    WHERE schema_id IS NOT NULL;

-- Query results: query_id + created_at (for result retrieval)
CREATE INDEX IF NOT EXISTS idx_query_results_query_created 
    ON neuronip.query_results(query_id, created_at DESC);

-- Knowledge documents: collection_id + content_type (common filter)
CREATE INDEX IF NOT EXISTS idx_knowledge_documents_collection_type 
    ON neuronip.knowledge_documents(collection_id, content_type)
    WHERE collection_id IS NOT NULL;

-- Knowledge embeddings: document_id + chunk_index (for chunk retrieval)
CREATE INDEX IF NOT EXISTS idx_knowledge_embeddings_doc_chunk 
    ON neuronip.knowledge_embeddings(document_id, chunk_index);

-- Support tickets: customer_id + status + created_at (common dashboard query)
CREATE INDEX IF NOT EXISTS idx_support_tickets_customer_status_created 
    ON neuronip.support_tickets(customer_id, status, created_at DESC);

-- Support conversations: ticket_id + created_at (for conversation history)
CREATE INDEX IF NOT EXISTS idx_support_conversations_ticket_created 
    ON neuronip.support_conversations(ticket_id, created_at DESC);

-- Workflow executions: workflow_id + status + created_at
CREATE INDEX IF NOT EXISTS idx_workflow_executions_workflow_status_created 
    ON neuronip.workflow_executions(workflow_id, status, created_at DESC);

-- Agent performance: agent_id + timestamp (for performance queries)
CREATE INDEX IF NOT EXISTS idx_agent_performance_agent_timestamp 
    ON neuronip.agent_performance(agent_id, timestamp DESC);

-- Usage metrics: user_id + metric_type + timestamp (for user analytics)
CREATE INDEX IF NOT EXISTS idx_usage_metrics_user_type_timestamp 
    ON neuronip.usage_metrics(user_id, metric_type, timestamp DESC)
    WHERE user_id IS NOT NULL;

-- Billing records: user_id + period_start (for billing queries)
CREATE INDEX IF NOT EXISTS idx_billing_records_user_period 
    ON neuronip.billing_records(user_id, period_start DESC)
    WHERE user_id IS NOT NULL;

-- Catalog datasets: owner + tags (for owner-based and tag-based queries)
CREATE INDEX IF NOT EXISTS idx_catalog_datasets_owner_created 
    ON neuronip.catalog_datasets(owner, created_at DESC)
    WHERE owner IS NOT NULL;

-- Metrics: kpi_type + reusable (for metric catalog queries)
CREATE INDEX IF NOT EXISTS idx_metrics_type_reusable 
    ON neuronip.metrics(kpi_type, reusable)
    WHERE reusable = true;

-- Data sources: source_type + sync_status (for connector management)
CREATE INDEX IF NOT EXISTS idx_data_sources_type_status 
    ON neuronip.data_sources(source_type, sync_status);

-- Partial indexes for filtered queries

-- Active workflows only
CREATE INDEX IF NOT EXISTS idx_workflows_active 
    ON neuronip.workflows(created_at DESC)
    WHERE status = 'active';

-- Recent query performance (last 24 hours)
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_recent_performance 
    ON neuronip.warehouse_queries(execution_time_ms, executed_at DESC)
    WHERE executed_at > NOW() - INTERVAL '24 hours' 
    AND execution_time_ms IS NOT NULL;

-- Failed queries for debugging
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_failed 
    ON neuronip.warehouse_queries(created_at DESC, error_message)
    WHERE status = 'failed';

-- Expired sessions cleanup
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires 
    ON neuronip.user_sessions(expires_at)
    WHERE expires_at < NOW();

-- Indexes for foreign keys that are frequently queried

-- API keys by user
CREATE INDEX IF NOT EXISTS idx_api_keys_user_created 
    ON neuronip.api_keys(user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

-- Metric dimensions by metric
CREATE INDEX IF NOT EXISTS idx_metric_dimensions_metric_created 
    ON neuronip.metric_dimensions(metric_id, created_at DESC);

-- Version history by version
CREATE INDEX IF NOT EXISTS idx_version_history_version_created 
    ON neuronip.version_history(version_id, created_at DESC);

-- Lineage edges for traversal
CREATE INDEX IF NOT EXISTS idx_lineage_edges_source_target 
    ON neuronip.lineage_edges(source_node_id, target_node_id);

-- Composite index for lineage queries
CREATE INDEX IF NOT EXISTS idx_lineage_nodes_type_resource 
    ON neuronip.lineage_nodes(node_type, resource_id)
    WHERE resource_id IS NOT NULL;

-- Performance monitoring indexes

-- Slow queries (execution time > 1 second)
CREATE INDEX IF NOT EXISTS idx_warehouse_queries_slow 
    ON neuronip.warehouse_queries(execution_time_ms DESC, executed_at DESC)
    WHERE execution_time_ms > 1000;

-- System logs by type and level (common filter)
CREATE INDEX IF NOT EXISTS idx_system_logs_type_level_timestamp 
    ON neuronip.system_logs(log_type, level, timestamp DESC);

-- Audit events by entity (for audit trail queries)
CREATE INDEX IF NOT EXISTS idx_audit_events_entity_created 
    ON neuronip.audit_events(entity_type, entity_id, created_at DESC)
    WHERE entity_id IS NOT NULL;

-- Anomaly detections by status and type
CREATE INDEX IF NOT EXISTS idx_anomaly_detections_status_type_created 
    ON neuronip.anomaly_detections(status, detection_type, created_at DESC);

-- Compliance matches by policy and status
CREATE INDEX IF NOT EXISTS idx_compliance_matches_policy_status_created 
    ON neuronip.compliance_matches(policy_id, status, created_at DESC);

-- Analyze tables after index creation for query planner
ANALYZE neuronip.warehouse_queries;
ANALYZE neuronip.query_results;
ANALYZE neuronip.knowledge_documents;
ANALYZE neuronip.knowledge_embeddings;
ANALYZE neuronip.support_tickets;
ANALYZE neuronip.workflow_executions;
ANALYZE neuronip.usage_metrics;
ANALYZE neuronip.billing_records;
ANALYZE neuronip.catalog_datasets;
ANALYZE neuronip.metrics;
ANALYZE neuronip.data_sources;
ANALYZE neuronip.system_logs;
ANALYZE neuronip.audit_events;
