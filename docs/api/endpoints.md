# 📡 API Endpoints Reference

<div align="center">

**Complete API Endpoint Documentation**

[← Overview](overview.md) • [Authentication →](authentication.md)

</div>

---

## 📋 Table of Contents

- [Health Check](#health-check)
- [Semantic Search](#semantic-search)
- [Warehouse Q&A](#warehouse-qa)
- [Support System](#support-system)
- [Compliance](#compliance)
- [Workflows](#workflows)
- [Analytics](#analytics)
- [Knowledge Graph](#knowledge-graph)
- [Data Sources](#data-sources)
- [Metrics](#metrics)
- [Agents](#agents)
- [Observability](#observability)
- [Lineage](#lineage)
- [Audit](#audit)
- [Billing](#billing)
- [Versioning](#versioning)
- [Catalog](#catalog)
- [Ingestion](#-ingestion)
- [Data Quality](#-data-quality)
- [Data Profiling](#-data-profiling)
- [Data Classification](#-data-classification)
- [Column Lineage](#-column-lineage)
- [SSO (Single Sign-On)](#-sso-single-sign-on)
- [Comments](#-comments)
- [Ownership](#-ownership)
- [Connectors](#-connectors)
- [Backup & Recovery](#-backup--recovery)
- [Multi-Region](#-multi-region)
- [Privacy & Compliance](#-privacy--compliance)
- [Auth](#-auth)
- [API Keys](#-api-keys)
- [Blocks](#-blocks)
- [Databases (Notion UI)](#-databases-notion-ui)
- [Quotas](#-quotas)
- [Governance RLS](#-governance-rls)
- [Webhooks](#-webhooks)
- [Alerts](#-alerts)
- [Models & Prompts (Governance)](#-models--prompts-governance)
- [Integrations](#-integrations)
- [Policies](#-policies)

---

## ❤️ Health Check

### GET `/health`

Check API health status.

**No authentication required**

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "database": "connected",
  "neurondb": "connected"
}
```

---

## 🔍 Semantic Search

### POST `/api/v1/semantic/search`

Perform semantic search across knowledge base.

**Request:**
```json
{
  "query": "What is NeuronIP?",
  "collection_id": "uuid-optional",
  "limit": 10,
  "threshold": 0.7
}
```

**Response:**
```json
{
  "results": [
    {
      "id": "uuid",
      "title": "Document Title",
      "content": "Document content...",
      "similarity": 0.95,
      "metadata": {}
    }
  ],
  "count": 10
}
```

### POST `/api/v1/semantic/rag`

Retrieval-Augmented Generation pipeline.

**Request:**
```json
{
  "query": "Explain NeuronIP",
  "collection_id": "uuid-optional",
  "limit": 5,
  "max_context": 2000
}
```

**Response:**
```json
{
  "query": "Explain NeuronIP",
  "context": "Retrieved context...",
  "answer": "Generated answer...",
  "sources": [
    {
      "id": "uuid",
      "title": "Source document",
      "similarity": 0.92
    }
  ]
}
```

### POST `/api/v1/semantic/documents`

Create a new document.

**Request:**
```json
{
  "document": {
    "title": "Document Title",
    "content": "Document content...",
    "content_type": "document",
    "collection_id": "uuid-optional",
    "source": "source-name",
    "source_url": "https://example.com",
    "metadata": {}
  },
  "chunking_config": {
    "chunk_size": 500,
    "chunk_overlap": 50
  }
}
```

**Response:**
```json
{
  "id": "uuid",
  "title": "Document Title",
  "content": "Document content...",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### PUT `/api/v1/semantic/documents/{id}`

Update an existing document.

### GET `/api/v1/semantic/collections/{id}`

Get collection details.

---

## 💬 Warehouse Q&A

### POST `/api/v1/warehouse/query`

Execute a natural language query on the warehouse.

**Request:**
```json
{
  "query": "What are the top 10 customers by revenue?",
  "schema_id": "uuid-optional"
}
```

**Response:**
```json
{
  "id": "uuid",
  "natural_language_query": "What are the top 10 customers by revenue?",
  "generated_sql": "SELECT ...",
  "results": [...],
  "explanation": "This query finds...",
  "status": "completed",
  "executed_at": "2024-01-01T00:00:00Z"
}
```

### GET `/api/v1/warehouse/queries/{id}`

Get query details and results.

### GET `/api/v1/warehouse/schemas`

List all warehouse schemas.

### POST `/api/v1/warehouse/schemas`

Create a new schema.

**Request:**
```json
{
  "name": "Sales Schema",
  "description": "Sales data schema",
  "schema_definition": {
    "tables": [...]
  }
}
```

### GET `/api/v1/warehouse/schemas/{id}`

Get schema details.

---

## 🎫 Support System

### POST `/api/v1/support/tickets`

Create a support ticket.

**Request:**
```json
{
  "customer_id": "customer-123",
  "customer_email": "customer@example.com",
  "subject": "Issue with feature X",
  "priority": "high",
  "metadata": {}
}
```

**Response:**
```json
{
  "id": "uuid",
  "ticket_number": "TKT-2024-001",
  "status": "open",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### GET `/api/v1/support/tickets`

List support tickets.

**Query Parameters:**
- `status` - Filter by status
- `priority` - Filter by priority
- `customer_id` - Filter by customer

### GET `/api/v1/support/tickets/{id}`

Get ticket details.

### POST `/api/v1/support/tickets/{id}/conversations`

Add a conversation to a ticket.

**Request:**
```json
{
  "message": "Customer message",
  "sender": "customer",
  "metadata": {}
}
```

### GET `/api/v1/support/tickets/{id}/conversations`

Get ticket conversations.

### GET `/api/v1/support/tickets/{id}/similar-cases`

Find similar past cases.

---

## 🛡️ Compliance

### POST `/api/v1/compliance/check`

Check compliance against policies.

**Request:**
```json
{
  "data": {...},
  "policy_ids": ["uuid1", "uuid2"]
}
```

**Response:**
```json
{
  "compliant": true,
  "violations": [],
  "checks": [
    {
      "policy_id": "uuid",
      "policy_name": "Policy Name",
      "compliant": true
    }
  ]
}
```

### GET `/api/v1/compliance/anomalies`

Get anomaly detections.

**Query Parameters:**
- `start_date` - Start date filter
- `end_date` - End date filter
- `severity` - Filter by severity

---

## ⚙️ Workflows

### POST `/api/v1/workflows/{id}/execute`

Execute a workflow.

**Request:**
```json
{
  "input": {
    "param1": "value1",
    "param2": "value2"
  }
}
```

**Response:**
```json
{
  "execution_id": "uuid",
  "workflow_id": "uuid",
  "status": "completed",
  "output": {...},
  "started_at": "2024-01-01T00:00:00Z",
  "completed_at": "2024-01-01T00:01:00Z"
}
```

### GET `/api/v1/workflows/{id}`

Get workflow details.

---

## 📊 Analytics

### GET `/api/v1/analytics/search`

Get search analytics.

**Query Parameters:**
- `start_date` - Start date
- `end_date` - End date
- `collection_id` - Filter by collection

**Response:**
```json
{
  "total_searches": 1000,
  "unique_queries": 500,
  "avg_response_time_ms": 150,
  "top_queries": [...],
  "trends": [...]
}
```

### GET `/api/v1/analytics/warehouse`

Get warehouse analytics.

### GET `/api/v1/analytics/workflows`

Get workflow analytics.

### GET `/api/v1/analytics/compliance`

Get compliance analytics.

### GET `/api/v1/analytics/retrieval-quality`

Get retrieval quality metrics.

---

## 🧠 Knowledge Graph

### POST `/api/v1/knowledge-graph/entities/extract`

Extract entities from text.

**Request:**
```json
{
  "text": "Text to extract entities from",
  "entity_types": ["person", "organization"]
}
```

### GET `/api/v1/knowledge-graph/entities/{id}`

Get entity details.

### GET `/api/v1/knowledge-graph/entities/{id}/links`

Get entity relationships.

### POST `/api/v1/knowledge-graph/entities/search`

Search entities.

### POST `/api/v1/knowledge-graph/entities/link`

Link two entities.

### POST `/api/v1/knowledge-graph/traverse`

Traverse the knowledge graph.

---

## 📦 Data Sources

### GET `/api/v1/data-sources`

List data sources.

### POST `/api/v1/data-sources`

Create a data source.

### GET `/api/v1/data-sources/{id}`

Get data source details.

### PUT `/api/v1/data-sources/{id}`

Update data source.

### DELETE `/api/v1/data-sources/{id}`

Delete data source.

### POST `/api/v1/data-sources/{id}/sync`

Trigger data source sync.

### GET `/api/v1/data-sources/{id}/status`

Get sync status.

---

## 📈 Metrics

### GET `/api/v1/metrics`

List business metrics.

### POST `/api/v1/metrics`

Create a metric.

### GET `/api/v1/metrics/{id}`

Get metric details.

### PUT `/api/v1/metrics/{id}`

Update metric.

### DELETE `/api/v1/metrics/{id}`

Delete metric.

### POST `/api/v1/metrics/search`

Search metrics.

---

## 🤖 Agents

### GET `/api/v1/agents`

List agents.

### POST `/api/v1/agents`

Create an agent.

### GET `/api/v1/agents/{id}`

Get agent details.

### PUT `/api/v1/agents/{id}`

Update agent.

### DELETE `/api/v1/agents/{id}`

Delete agent.

### GET `/api/v1/agents/{id}/performance`

Get agent performance metrics.

### POST `/api/v1/agents/{id}/deploy`

Deploy an agent.

---

## 👁️ Observability

### GET `/api/v1/observability/queries/performance`

Get query performance metrics.

### GET `/api/v1/observability/logs`

Get system logs.

### GET `/api/v1/observability/metrics`

Get system metrics.

### GET `/api/v1/observability/agent-logs`

Get agent logs.

### GET `/api/v1/observability/workflow-logs`

Get workflow logs.

---

## 🔗 Lineage

### GET `/api/v1/lineage/{resource_type}/{resource_id}`

Get resource lineage.

### POST `/api/v1/lineage/track`

Track a transformation.

### GET `/api/v1/lineage/impact/{resource_id}`

Get impact analysis.

### GET `/api/v1/lineage/graph`

Get full lineage graph.

---

## 📋 Audit

### GET `/api/v1/audit/events`

Get audit events.

### GET `/api/v1/audit/activity`

Get activity timeline.

### GET `/api/v1/audit/compliance-trail`

Get compliance trail.

### POST `/api/v1/audit/search`

Search audit events.

---

## 💳 Billing

### GET `/api/v1/billing/usage`

Get usage metrics.

### GET `/api/v1/billing/metrics`

Get detailed billing metrics.

### GET `/api/v1/billing/dashboard`

Get billing dashboard data.

### POST `/api/v1/billing/track`

Track usage for billing.

---

## 📚 Versioning

### GET `/api/v1/versions/{resource_type}/{resource_id}`

List versions of a resource.

### POST `/api/v1/versions/create`

Create a version.

### GET `/api/v1/versions/{id}`

Get version details.

### POST `/api/v1/versions/{id}/rollback`

Rollback to a version.

### GET `/api/v1/versions/{id}/history`

Get version history.

---

## 📖 Catalog

### GET `/api/v1/catalog/datasets`

List datasets.

### GET `/api/v1/catalog/datasets/{id}`

Get dataset details.

### GET `/api/v1/catalog/search`

Search catalog.

### GET `/api/v1/catalog/owners`

List dataset owners.

### POST `/api/v1/catalog/discover`

Discover datasets.

---

## 📥 Ingestion

### POST `/api/v1/ingestion/jobs`

Create a new ingestion job.

**Request:**
```json
{
  "connector_type": "snowflake",
  "connector_config": {
    "account": "your-account",
    "warehouse": "COMPUTE_WH",
    "database": "SNOWFLAKE_SAMPLE_DATA"
  },
  "schedule": {
    "frequency": "daily",
    "time": "02:00"
  }
}
```

### GET `/api/v1/ingestion/jobs`

List all ingestion jobs.

### GET `/api/v1/ingestion/jobs/{id}`

Get ingestion job details.

### POST `/api/v1/ingestion/jobs/{id}/execute`

Execute an ingestion job manually.

---

## 🔍 Data Quality

### POST `/api/v1/data-quality/rules`

Create a data quality rule.

**Request:**
```json
{
  "name": "Column Not Null Check",
  "type": "not_null",
  "target": {
    "connector_id": "uuid",
    "schema_name": "public",
    "table_name": "users",
    "column_name": "email"
  },
  "threshold": 0.95
}
```

### GET `/api/v1/data-quality/rules/{id}`

Get data quality rule details.

### POST `/api/v1/data-quality/rules/{id}/execute`

Execute a data quality rule.

### GET `/api/v1/data-quality/dashboard`

Get data quality dashboard metrics.

### GET `/api/v1/data-quality/trends`

Get data quality trends over time.

---

## 📊 Data Profiling

### POST `/api/v1/profiling/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}`

Profile a table.

**Response:**
```json
{
  "table_name": "users",
  "row_count": 1000000,
  "column_count": 15,
  "columns": [
    {
      "name": "id",
      "type": "integer",
      "null_count": 0,
      "distinct_count": 1000000,
      "min": 1,
      "max": 1000000
    }
  ],
  "profiled_at": "2024-01-01T00:00:00Z"
}
```

### POST `/api/v1/profiling/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}/columns/{column_name}`

Profile a specific column.

---

## 🏷️ Data Classification

### POST `/api/v1/classification/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}/columns/{column_name}`

Classify a column (e.g., PII detection).

**Response:**
```json
{
  "column_name": "email",
  "classification": "pii",
  "sub_classification": "email_address",
  "confidence": 0.98,
  "classified_at": "2024-01-01T00:00:00Z"
}
```

### POST `/api/v1/classification/connectors/{id}/classify`

Classify all columns in a connector.

### POST `/api/v1/classification/rules`

Create a custom classification rule.

---

## 🔗 Column Lineage

### GET `/api/v1/lineage/columns/{connector_id}/{schema_name}/{table_name}/{column_name}`

Get column-level lineage.

**Response:**
```json
{
  "column": {
    "connector_id": "uuid",
    "schema_name": "public",
    "table_name": "orders",
    "column_name": "total_amount"
  },
  "upstream": [
    {
      "connector_id": "uuid",
      "schema_name": "public",
      "table_name": "order_items",
      "column_name": "price"
    }
  ],
  "downstream": []
}
```

### POST `/api/v1/lineage/columns/track`

Track column lineage transformation.

### POST `/api/v1/lineage/columns/nodes`

Create a column lineage node.

---

## 🔐 SSO (Single Sign-On)

### POST `/api/v1/sso/providers`

Create an SSO provider.

**Request:**
```json
{
  "name": "Okta",
  "type": "saml",
  "config": {
    "sso_url": "https://your-org.okta.com/sso/saml",
    "certificate": "..."
  }
}
```

### GET `/api/v1/sso/providers`

List SSO providers.

### GET `/api/v1/sso/providers/{id}`

Get SSO provider details.

### GET `/api/v1/sso/providers/{id}/initiate`

Initiate SSO login.

### GET `/api/v1/sso/callback`

Handle SSO callback.

### POST `/api/v1/sso/validate`

Validate SSO session.

---

## 💬 Comments

### POST `/api/v1/comments`

Create a comment on a resource.

**Request:**
```json
{
  "resource_type": "dataset",
  "resource_id": "uuid",
  "content": "This dataset needs better documentation"
}
```

### GET `/api/v1/comments/{id}`

Get comment details.

### GET `/api/v1/comments/{resource_type}/{resource_id}`

List comments for a resource.

### POST `/api/v1/comments/{id}/resolve`

Resolve a comment.

### DELETE `/api/v1/comments/{id}`

Delete a comment.

---

## 👤 Ownership

### POST `/api/v1/ownership`

Assign ownership to a resource.

**Request:**
```json
{
  "resource_type": "dataset",
  "resource_id": "uuid",
  "owner_id": "user-uuid",
  "owner_type": "user"
}
```

### GET `/api/v1/ownership/{resource_type}/{resource_id}`

Get ownership information.

### GET `/api/v1/ownership/by-owner`

List resources owned by a user.

### DELETE `/api/v1/ownership/{resource_type}/{resource_id}`

Remove ownership.

---

## 🔌 Connectors

### POST `/api/v1/connectors`

Create a connector configuration.

### GET `/api/v1/connectors`

List all connectors.

### GET `/api/v1/connectors/{id}`

Get connector details.

### POST `/api/v1/connectors/{id}/sync`

Trigger a connector sync.

---

## 🔄 Backup & Recovery

### POST `/api/v1/backups`

Create a backup.

**Response:**
```json
{
  "id": "uuid",
  "status": "completed",
  "backup_type": "full",
  "size_bytes": 1073741824,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### GET `/api/v1/backups`

List all backups.

### POST `/api/v1/backups/{id}/restore`

Restore from a backup.

---

## 🌐 Multi-Region

### POST `/api/v1/regions`

Create a new region.

**Request:**
```json
{
  "name": "us-west-2",
  "display_name": "US West (Oregon)",
  "endpoint": "https://us-west-2.api.neurondb.ai"
}
```

### GET `/api/v1/regions`

List all regions.

### GET `/api/v1/regions/{id}`

Get region details.

### GET `/api/v1/regions/{id}/health`

Check region health.

### POST `/api/v1/regions/{id}/failover`

Failover to a region.

---

## 🔒 Privacy & Compliance

### DSAR (Data Subject Access Request)

### POST `/api/v1/dsar/requests`

Create a DSAR request.

**Request:**
```json
{
  "subject_id": "user-123",
  "request_type": "access",
  "email": "user@example.com"
}
```

### GET `/api/v1/dsar/requests`

List DSAR requests.

### GET `/api/v1/dsar/requests/{id}`

Get DSAR request details.

### POST `/api/v1/dsar/requests/{id}/complete`

Mark DSAR request as complete.

### PIA (Privacy Impact Assessment)

### POST `/api/v1/pia/requests`

Create a PIA request.

### GET `/api/v1/pia/requests/{id}`

Get PIA request details.

### POST `/api/v1/pia/requests/{id}/submit`

Submit PIA for review.

### POST `/api/v1/pia/requests/{id}/review`

Review PIA request.

### Consent Management

### POST `/api/v1/consent`

Record user consent.

**Request:**
```json
{
  "subject_id": "user-123",
  "purpose": "marketing",
  "consent_given": true
}
```

### POST `/api/v1/consent/withdraw`

Withdraw consent.

### GET `/api/v1/consent/{subject_id}`

Check consent status.

### GET `/api/v1/consent/subject/{subject_id}`

Get all consents for a subject.

### Data Masking

### POST `/api/v1/masking/policies`

Create a masking policy.

**Request:**
```json
{
  "name": "Email Masking",
  "target": {
    "connector_id": "uuid",
    "schema_name": "public",
    "table_name": "users",
    "column_name": "email"
  },
  "masking_type": "partial",
  "config": {
    "reveal_first": 2,
    "reveal_domain": true
  }
}
```

### GET `/api/v1/masking/policies`

List masking policies.

### POST `/api/v1/masking/apply`

Apply masking to data.

---

## 🔐 Auth

Authentication and session management. **Authentication required** (except login/register).

### POST `/api/v1/auth/login`

Authenticate with credentials and receive a token.

### POST `/api/v1/auth/register`

Register a new user.

### GET `/api/v1/auth/me`

Get current authenticated user.

### POST `/api/v1/auth/logout`

Invalidate current session.

### POST `/api/v1/auth/refresh`

Refresh access token.

### GET `/api/v1/auth/oidc/{provider}/initiate`

Initiate OIDC login flow for a provider.

### GET `/api/v1/auth/oidc/{provider}/callback`

OIDC callback after provider authentication.

### POST `/api/v1/auth/2fa/generate`

Generate two-factor authentication secret.

### GET `/api/v1/auth/sessions`

List active sessions for the current user.

---

## 🔑 API Keys

Create and manage API keys for programmatic access. **Authentication required.**

### POST `/api/v1/api-keys`

Create a new API key.

### GET `/api/v1/api-keys/{id}`

Get API key metadata (key value is only returned on create).

### POST `/api/v1/api-keys/{id}/rotate`

Rotate an existing API key.

### GET `/api/v1/api-keys/{id}/usage`

Get usage statistics for an API key.

### POST `/api/v1/api-keys/{id}/revoke`

Revoke an API key.

---

## 📦 Blocks

Notion-like content blocks. **Authentication required.**

### POST `/api/v1/blocks`

Create a block.

### GET `/api/v1/blocks/{id}`

Get a block by ID.

### PUT `/api/v1/blocks/{id}`

Update a block.

### DELETE `/api/v1/blocks/{id}`

Delete a block.

### POST `/api/v1/blocks/reorder`

Reorder blocks (e.g. within a page).

---

## 🗄️ Databases (Notion UI)

Notion-style databases (tables with views). **Authentication required.**

### POST `/api/v1/databases`

Create a database.

### GET `/api/v1/databases`

List databases.

### GET `/api/v1/databases/{id}`

Get a database by ID.

### PUT `/api/v1/databases/{id}`

Update a database.

### DELETE `/api/v1/databases/{id}`

Delete a database.

### GET `/api/v1/databases/{id}/rows`

Get rows in a database.

### POST `/api/v1/databases/{id}/rows`

Create a row.

### GET `/api/v1/databases/view-preferences`

Get view preferences for databases.

---

## 📊 Quotas

Tenancy and usage quotas. **Authentication required.**

### POST `/api/v1/quotas/set`

Set quota for a resource or tenant.

### GET `/api/v1/quotas/list`

List quotas.

### POST `/api/v1/quotas/check`

Check quota usage/availability.

---

## 🛡️ Governance RLS

Row-level security policies. **Authentication required.**

### GET `/api/v1/governance/rls/policies`

List RLS policies.

### POST `/api/v1/governance/rls/policies`

Create an RLS policy.

---

## 🔔 Webhooks

Event-driven webhooks. **Authentication required.**

### POST `/api/v1/webhooks`

Create a webhook.

### GET `/api/v1/webhooks`

List webhooks.

### GET `/api/v1/webhooks/{id}`

Get a webhook.

### PUT `/api/v1/webhooks/{id}`

Update a webhook.

### DELETE `/api/v1/webhooks/{id}`

Delete a webhook.

### POST `/api/v1/webhooks/trigger`

Trigger a webhook (e.g. for testing).

---

## 🚨 Alerts

Alerts and alert rules. **Authentication required.**

### POST `/api/v1/alerts/check`

Run alert checks.

### GET `/api/v1/alerts`

List alerts.

### POST `/api/v1/alerts/{id}/resolve`

Resolve an alert.

### POST `/api/v1/alerts/rules`

Create an alert rule.

### GET `/api/v1/alerts/rules`

List alert rules.

### PUT `/api/v1/alerts/rules/{id}`

Update an alert rule.

### DELETE `/api/v1/alerts/rules/{id}`

Delete an alert rule.

---

## 🤖 Models & Prompts (Governance)

Model and prompt registry with governance (approve, rollback). **Authentication required.**

### GET `/api/v1/models`

List registered models (governance).

### GET `/api/v1/models/{id}`

Get model metadata and versions.

### GET `/api/v1/models/{id}/versions`

List model versions.

### POST `/api/v1/models/{id}/approve`

Approve a model version.

### POST `/api/v1/models/{id}/rollback`

Rollback to a previous version.

### GET `/api/v1/prompts`

List prompts.

### GET `/api/v1/prompts/{id}`

Get prompt and versions.

### GET `/api/v1/prompts/{id}/versions`

List prompt versions.

### POST `/api/v1/prompts/{id}/approve`

Approve a prompt version.

### POST `/api/v1/prompts/{id}/rollback`

Rollback a prompt.

---

## 🔌 Integrations

External integrations (Slack, Teams, BI, helpdesk). **Authentication required.**

### POST `/api/v1/integrations`

Create an integration.

### GET `/api/v1/integrations`

List integrations.

### GET `/api/v1/integrations/{id}`

Get an integration.

### PUT `/api/v1/integrations/{id}`

Update an integration.

### DELETE `/api/v1/integrations/{id}`

Delete an integration.

### POST `/api/v1/integrations/{id}/test`

Test an integration connection.

### GET `/api/v1/integrations/{id}/health`

Check integration health.

### POST `/api/v1/integrations/helpdesk/sync`

Sync helpdesk data.

### POST `/api/v1/integrations/slack/command`

Handle Slack command.

### POST `/api/v1/integrations/teams/message`

Send/handle Teams message.

### POST `/api/v1/integrations/bi/export`

Export to BI tool.

---

## 📜 Policies

Governance policies (create, evaluate). **Authentication required.**

### POST `/api/v1/policies`

Create a policy.

### GET `/api/v1/policies/{id}`

Get a policy.

### POST `/api/v1/policies/{id}/evaluate`

Evaluate a policy against context.

---

## 📚 Related Documentation

- [API Overview](overview.md) - API introduction
- [Authentication](authentication.md) - Auth details
- [Rate Limiting](rate-limiting.md) - Limits

---

<div align="center">

[← Back to API Docs](README.md)

</div>
