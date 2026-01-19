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

## 📚 Related Documentation

- [API Overview](overview.md) - API introduction
- [Authentication](authentication.md) - Auth details
- [Rate Limiting](rate-limiting.md) - Limits

---

<div align="center">

[← Back to API Docs](README.md)

</div>
