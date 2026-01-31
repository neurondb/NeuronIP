# NeuronIP Feature Map

Single source of truth: **Module → UI pages → API routes → service packages → migrations**.  
Derived from `api/cmd/server/main.go`, `frontend/app/(dashboard)/`, and `api/migrations/`.

---

## Module Index

| Module | UI Page(s) | API Prefix | Service Package(s) | Key Migrations |
|--------|------------|------------|--------------------|----------------|
| **AI / RAG** | semantic | `/ai/*`, `/rag/*` | ai, rag | (uses semantic/warehouse) |
| **Semantic / Pipelines** | semantic | `/semantic/pipelines*` | semantic | 012_pipeline_versioning, 014_semantic_layer |
| **Warehouse** | warehouse | `/warehouse/*` | warehouse | 013_saved_searches, 014_semantic_layer, 015_query_governance, 016_query_cache |
| **Workflows** | workflows, workflows/builder | `/workflows/*` | workflows | 010_workflow_enhancements, 043_advanced_workflows |
| **Compliance** | compliance | `/compliance/*`, `/dsar/*`, `/pia/*`, `/consent/*`, `/masking/*` | compliance, masking | 004_governance_schema, 034_privacy_compliance, 035_data_masking |
| **Analytics** | analytics | `/analytics/*` | analytics | 006_observability_schema, 047_usage_analytics |
| **Models / ML** | models | `/models/*`, `/prompts/*` | models, governance | 005_model_versioning, 038_model_quality, 039_model_governance, 044_ml_lifecycle, 045_model_registry, 055_ml_lifecycle, 056_model_governance |
| **Integrations** | integrations | `/integrations/*` | integrations, slack, teams, bi | 042_integration_ecosystem, 048_integration_ecosystem |
| **Alerts** | alerts | `/alerts/*` | alerts | (uses compliance/anomaly) |
| **Support** | support | `/support/*` | support | (uses db/semantic) |
| **Knowledge Graph** | knowledge-graph | `/knowledge-graph/*` | knowledgegraph | 030_business_glossary, 049_knowledge_graph_enhancements |
| **Data Sources** | data-sources | `/data-sources/*` | datasources | 002_ingestion_schema, 011_ingestion_enhancements |
| **Metrics (Semantic Layer)** | metrics | `/metrics/*`, `/catalog/metrics*` | metrics, catalog | 014_semantic_layer, 036_semantic_layer_enhancements |
| **Agents** | agents | `/agents/*` | agents, agent | 017_agent_hub, 018_agent_evaluation, 019_hitl, 020_agent_audit, 039_agent_observability |
| **Observability** | observability | `/observability/*` | observability | 006_observability_schema, 021_observability_metrics, 046_enterprise_observability, 052_observability_enhancements |
| **Lineage** | lineage | `/lineage/*` | lineage | 003_metadata_schema, 027_column_lineage, 032_lineage_discovery |
| **Audit** | audit | `/audit/*` | audit | (audit tables in governance/observability) |
| **Billing** | billing | `/billing/*` | billing | 047_usage_analytics, 057_budgets |
| **Versioning** | versioning | `/versions/*` | versioning | 005_model_versioning, 012_pipeline_versioning |
| **Catalog** | catalog | `/catalog/*` | catalog | 003_metadata_schema, 030_business_glossary |
| **Ingestion** | data-sources | `/ingestion/*` | ingestion | 002_ingestion_schema, 011_ingestion_enhancements, 037_ingestion_connectors, 041_enterprise_ingestion |
| **Policy** | — | `/policies/*` | policy | 004_governance_schema |
| **Webhooks** | integrations | `/webhooks/*` | webhooks, integrations | 025_quick_wins, 042_integration_ecosystem |
| **Auth** | login, settings | `/auth/*`, `/api-keys/*` | auth, session | 001_users_schema, 007_auth_enhancements, 008_rbac_enhancements, 022_sso_schema, 032_session_management |
| **RBAC** | users, api-keys, settings | `/rbac/*` | auth | 008_rbac_enhancements, 031_advanced_rbac |
| **SSO** | settings | `/sso/*` | auth | 022_sso_schema |
| **Comments** | (collaboration) | `/comments/*` | comments | 025_quick_wins, 035_collaboration |
| **Ownership** | catalog | `/ownership/*` | ownership | 025_quick_wins |
| **Connectors** | data-sources | `/connectors/*` | connectors | 023_connector_framework, 053_connector_registry |
| **Data Quality** | quality | `/data-quality/*` | dataquality | 024_data_quality_engine, 036_data_quality_checks |
| **Profiling** | quality, catalog | `/profiling/*` | profiling | 026_data_profiling |
| **Classification** | compliance, catalog | `/classification/*` | classification | 027_pii_detection, 028_automated_classification |
| **Column Lineage** | lineage | `/lineage/columns/*` | lineage | 027_column_lineage |
| **Regions** | — | `/regions/*` | tenancy | 032_multi_region |
| **Backup** | — | `/backups/*` | backup | 033_backup_system |
| **Blocks UI** | notion-ui | `/blocks/*` | blocks | 058_notion_ui_blocks |
| **Databases UI** | notion-ui | `/databases/*` | databases | 059_notion_ui_databases |
| **Collaboration** | (dashboards) | `/collaboration/*` | collaboration, handlers | 035_collaboration, 040_collaboration_enhancements |
| **Governance (RLS)** | compliance | `/governance/rls/*` | auth, handlers | 031_advanced_rbac |
| **Quotas** | billing, settings | `/quotas/*` | tenancy, handlers | 009_tenancy_schema, 037_distributed_execution |
| **BI Export** | warehouse, integrations | `/integrations/bi/export` | integrations/bi | — |

---

## API Routes by Prefix (from main.go)

All routes are under `POST/GET/etc. /api/v1/<path>`.

- **ai**: embedding, workflow, register-tools  
- **rag**: query, query/stream, status  
- **semantic**: pipelines (CRUD, versions, replay, activate)  
- **warehouse**: query, query/rewrite, query/semantics, queries/{id}, queries/history, optimize, schemas, saved-searches, governance/validate, governance/sanitize, cache, cache/invalidate, cache/stats  
- **workflows**: CRUD, execute, versions, schedule, schedules, monitoring, executions/{id}/status|recover|logs|metrics|decisions  
- **compliance**: check, anomalies, policies, report  
- **analytics**: search, warehouse, workflows, compliance, retrieval-quality  
- **models**: RegisterModel, GetModel, InferModel; (governance) ListModels, GetModel, versions, approve, rollback  
- **prompts**: ListPrompts, GetPrompt, versions, approve, rollback  
- **integrations**: CRUD, test, health, helpdesk/sync, slack/command, teams/message, bi/export  
- **alerts**: check, list, resolve, rules CRUD  
- **support**: tickets CRUD, conversations, similar-cases  
- **knowledge-graph**: entities/extract|{id}|search|link, entity-types, glossary, glossary/{id}, glossary/search, traverse, query  
- **data-sources**: CRUD, sync, status  
- **metrics**: CRUD, search, discover, calculate, lineage, approve; approvals/queue, {id}/approvals, {id}/owners  
- **agents**: CRUD, performance, deploy  
- **observability**: queries/performance, logs, logs/stream, metrics, realtime, benchmark, cost/breakdown, agent-logs, workflow-logs; metrics/latency|error-rate|token-usage|embedding-cost; agents/{id}/logs, retrieval/*, hallucination/*, queries/{id}/cost, agents/runs/{id}/cost  
- **lineage**: {resource_type}/{resource_id}, track, impact/{id}, graph; columns/*  
- **audit**: events, activity, compliance-trail, search  
- **billing**: usage, metrics, dashboard, track  
- **versions**: {resource_type}/{resource_id}, create, {id}, rollback, history  
- **catalog**: datasets, datasets/{id}, search, owners, discover; metrics (via businessMetricsHandler)  
- **ingestion**: jobs CRUD, execute, data-sources/{id}/status, failures, jobs/{id}/retry  
- **policies**: create, {id}, evaluate  
- **webhooks**: CRUD, trigger  
- **auth**: login, register, me, logout, refresh; oidc/{provider}/initiate|callback; scim/*; 2fa/generate; sessions  
- **rbac**: workspaces, permissions/check  
- **api-keys**: create, {id}/rotate|usage|revoke  
- **sso**: providers CRUD, initiate, callback, validate  
- **comments**: create, {id}, {resource_type}/{resource_id}, resolve, delete  
- **ownership**: assign, get, by-owner, remove  
- **connectors**: CRUD, sync  
- **data-quality**: rules, rules/{id}, rules/{id}/execute, dashboard, trends  
- **profiling**: connectors/.../tables/..., columns/...  
- **classification**: connectors/.../columns/..., connectors/{id}/classify, rules  
- **regions**: CRUD, health, failover  
- **backups**: create, list, restore  
- **dsar**: requests CRUD, complete  
- **pia**: requests, submit, review  
- **consent**: record, withdraw, {subject_id}, subject/{id}  
- **masking**: policies, apply  
- **blocks**: CRUD, reorder  
- **databases**: CRUD, rows, view-preferences  
- **governance/rls**: policies GET/POST  
- **quotas**: set, list, check  

---

## UI Pages (frontend/app/(dashboard)/)

| Route | Path | Primary Module(s) |
|-------|------|-------------------|
| Dashboard | / | page.tsx |
| Agents | /agents | agents |
| Alerts | /alerts | alerts |
| Analytics | /analytics | analytics |
| API Keys | /api-keys | auth |
| Audit | /audit | audit |
| Billing | /billing | billing |
| Catalog | /catalog | catalog, metrics |
| Compliance | /compliance | compliance |
| Data Sources | /data-sources | data-sources, ingestion, connectors |
| Features | /features | marketing |
| Integrations | /integrations | integrations |
| Knowledge Graph | /knowledge-graph | knowledgegraph |
| Lineage | /lineage | lineage |
| Metrics | /metrics | metrics (semantic layer) |
| Models | /models | models, governance |
| Blocks / Databases | /notion-ui | blocks, databases |
| Observability | /observability | observability |
| Quality | /quality | dataquality, profiling |
| Semantic | /semantic | semantic, pipelines, RAG |
| Settings | /settings | auth, RBAC, SSO |
| Support | /support | support |
| Users | /users | auth, RBAC |
| Versioning | /versioning | versioning |
| Warehouse | /warehouse | warehouse |
| Why NeuronIP | /why-neuronip | marketing |
| Workflows | /workflows | workflows |
| Workflow Builder | /workflows/builder | workflows |

---

## Migrations by Theme

- **Auth / RBAC / SSO**: 001, 007, 008, 022, 031, 032_session_management  
- **Ingestion / Connectors**: 002, 011, 023, 037_ingestion_connectors, 041_enterprise_ingestion, 053  
- **Metadata / Catalog / Semantic**: 003, 014, 030, 036_semantic_layer_enhancements  
- **Governance / Compliance**: 004, 015, 034, 035_data_masking  
- **Models / ML**: 005, 038_model_quality, 039_model_governance, 044, 045, 055, 056  
- **Observability**: 006, 021, 046, 052  
- **Tenancy / Multi-region**: 009, 032_multi_region  
- **Workflows / Pipelines**: 010_workflow_enhancements, 012, 043, 042_streaming_pipelines, 054  
- **Saved Searches / Cache**: 013, 016  
- **Agents**: 017, 018, 019, 020, 039_agent_observability  
- **Data Quality / Profiling / Classification**: 024, 026, 027_pii_detection, 028_automated_classification, 036_data_quality_checks  
- **Quick Wins (comments, ownership, webhooks)**: 025  
- **Lineage**: 027_column_lineage, 032_lineage_discovery  
- **Freshness / Schema Evolution**: 028_data_freshness, 029  
- **Collaboration**: 035_collaboration, 040  
- **Backup / Regions**: 033  
- **Integration / Ecosystem**: 042_integration_ecosystem, 048  
- **Usage / Billing**: 047, 057  
- **Knowledge Graph**: 049  
- **Clustering / Distributed**: 040_clustering_schema, 037_distributed_execution, 038_distributed_execution, 051  
- **Notion UI**: 058, 059  
- **Approval Workflows**: 041_approval_workflows  
- **Kafka / Streaming**: 035_kafka_ingestion, 054  
- **Glossary**: 035_glossary_linking  

---

## How to Use This Map

1. **Proof a feature**: Pick module → find UI page → confirm API routes in main.go → trace handler to service in `api/internal/<package>` → find tables in `api/migrations/*.sql`.  
2. **Add a feature**: Add migration if needed → implement or extend service → add handler and route in main.go → add or update UI under `frontend/app/(dashboard)/`.  
3. **Full-stack demo**: Use this map to list which modules and routes the single demo script exercises (warehouse, workload, data products, notebooks, decision dashboards, ITSM, templates, AI assistant).
