# 💾 Database Schema Reference

<div align="center">

**Complete database schema documentation**

[← Environment Variables](environment-variables.md) • [Glossary →](glossary.md)

</div>

---

## 📋 Table of Contents

- [Schema Overview](#schema-overview)
- [Migration Index](#migration-index)
- [Core Tables](#core-tables)
- [Relationships](#relationships)

---

## 📊 Schema Overview

See [Database Architecture](../architecture/database.md) for detailed schema documentation. Full schema definitions live in the migration files under `api/migrations/`.

---

## 📑 Migration Index

Migrations are applied in filename order. Below is a one-line purpose for each migration file.

| Migration | Purpose |
|-----------|---------|
| 000_baseline.sql | Baseline schema |
| 001_users_schema.sql | Users and accounts |
| 002_ingestion_schema.sql | Ingestion jobs and data sources |
| 003_metadata_schema.sql | Metadata and catalog |
| 004_governance_schema.sql | Governance and policies |
| 005_model_versioning.sql | Model versioning |
| 006_observability_schema.sql | Observability and metrics |
| 007_auth_enhancements.sql | Auth enhancements |
| 008_rbac_enhancements.sql | RBAC enhancements |
| 009_tenancy_schema.sql | Tenancy and workspaces |
| 010_api_keys_enhancements.sql | API key management |
| 010_workflow_enhancements.sql | Workflow enhancements |
| 011_ingestion_enhancements.sql | Ingestion enhancements |
| 012_pipeline_versioning.sql | Pipeline versioning |
| 013_saved_searches.sql | Saved searches |
| 014_semantic_layer.sql | Semantic layer |
| 015_query_governance.sql | Query governance |
| 016_query_cache.sql | Query cache |
| 017_agent_hub.sql | Agent hub |
| 018_agent_evaluation.sql | Agent evaluation |
| 019_hitl.sql | Human-in-the-loop |
| 020_agent_audit.sql | Agent audit |
| 021_observability_metrics.sql | Observability metrics |
| 022_sso_schema.sql | SSO schema |
| 023_connector_framework.sql | Connector framework |
| 024_data_quality_engine.sql | Data quality engine |
| 025_quick_wins.sql | Quick wins (comments, ownership, etc.) |
| 026_data_profiling.sql | Data profiling |
| 027_column_lineage.sql | Column-level lineage |
| 027_pii_detection.sql | PII detection |
| 028_automated_classification.sql | Automated classification |
| 028_data_freshness.sql | Data freshness |
| 029_schema_evolution.sql | Schema evolution |
| 030_business_glossary.sql | Business glossary |
| 031_advanced_rbac.sql | Advanced RBAC |
| 032_multi_region.sql | Multi-region |
| 032_session_management.sql | Session management |
| 033_backup_system.sql | Backup system |
| 034_privacy_compliance.sql | Privacy compliance |
| 035_collaboration.sql | Collaboration |
| 035_data_masking.sql | Data masking |
| 035_glossary_linking.sql | Glossary linking |
| 035_kafka_ingestion.sql | Kafka ingestion |
| 036_data_quality_checks.sql | Data quality checks |
| 036_semantic_layer_enhancements.sql | Semantic layer enhancements |
| 037_distributed_execution.sql | Distributed execution |
| 037_ingestion_connectors.sql | Ingestion connectors |
| 038_distributed_execution.sql | Distributed execution (continued) |
| 038_model_quality.sql | Model quality |
| 039_agent_observability.sql | Agent observability |
| 039_model_governance.sql | Model governance |
| 040_clustering_schema.sql | Clustering schema |
| 040_collaboration_enhancements.sql | Collaboration enhancements |
| 041_approval_workflows.sql | Approval workflows |
| 041_enterprise_ingestion.sql | Enterprise ingestion |
| 042_integration_ecosystem.sql | Integration ecosystem |
| 042_streaming_pipelines.sql | Streaming pipelines |
| 043_advanced_workflows.sql | Advanced workflows |
| 044_ml_lifecycle.sql | ML lifecycle |
| 045_model_registry.sql | Model registry |
| 046_enterprise_observability.sql | Enterprise observability |
| 047_usage_analytics.sql | Usage analytics |
| 048_integration_ecosystem.sql | Integration ecosystem (continued) |
| 049_knowledge_graph_enhancements.sql | Knowledge graph enhancements |
| 050_partner_ecosystem.sql | Partner ecosystem |
| 051_clustering.sql | Clustering |
| 051_sales_enablement.sql | Sales enablement |
| 052_observability_enhancements.sql | Observability enhancements |
| 053_connector_registry.sql | Connector registry |
| 054_streaming_schema.sql | Streaming schema |
| 055_ml_lifecycle.sql | ML lifecycle (continued) |
| 056_model_governance.sql | Model governance |
| 057_budgets.sql | Budgets |
| 058_notion_ui_blocks.sql | Notion UI blocks |
| 059_notion_ui_databases.sql | Notion UI databases |
| 060_workload_and_data_products.sql | Workload and data products |
| 061_notebooks.sql | Notebooks |
| 062_decision_dashboards.sql | Decision dashboards |
| 063_itsm.sql | ITSM |
| 064_notion_templates.sql | Notion templates |
| 065_streaming_engine_and_events.sql | Streaming engine and events |
| 066_lineage_discovery.sql | Lineage discovery |
| 067_cdc_events.sql | CDC events |

---

## Core Tables

See [Database Architecture](../architecture/database.md) for core table definitions and relationships.

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
