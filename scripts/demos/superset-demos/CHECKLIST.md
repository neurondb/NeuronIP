# NeuronIP Demo – Acceptance Checklist

Use this checklist to validate end-to-end flows: elastic analytics, data products, notebooks, decision dashboards, ITSM, knowledge workspace, and AI assistant.

## 1. Warehouse & elastic analytics

- [ ] **POST /api/v1/warehouse/query** – Natural language query returns 200, body contains `query_id`, `generated_sql`, `results`, `explanation`.
- [ ] **GET /api/v1/warehouse/queries/{id}** – Returns query details and results.
- [ ] **GET /api/v1/warehouse/schemas** – List schemas 200.
- [ ] **POST /api/v1/warehouse/governance/validate** – Validate query returns 200.
- [ ] **GET /api/v1/warehouse/workload/queues** – List workload queues 200.
- [ ] **POST /api/v1/warehouse/workload/queues** – Create queue (name, max_concurrency, query_timeout_seconds) returns 201.
- [ ] **GET /api/v1/warehouse/workload/queues/{name}** – Get queue by name 200.
- [ ] **POST /api/v1/data-products** – Create data product (name, owner_id, version, visibility) returns 201.
- [ ] **GET /api/v1/data-products** – List data products 200.
- [ ] **POST /api/v1/data-products/{id}/share** – Share with consumer_workspace_id or consumer_user_id returns 201.
- [ ] **GET /api/v1/data-products/{id}/consumers** – List consumers 200.
- [ ] **GET /api/v1/warehouse/cache/stats** – Cache stats 200.
- [ ] **POST /api/v1/warehouse/cache/invalidate** – Invalidate cache 200.

## 2. Notebooks

- [ ] **POST /api/v1/notebooks** – Create notebook (name, owner_id, default_language) returns 201.
- [ ] **GET /api/v1/notebooks** – List notebooks 200.
- [ ] **GET /api/v1/notebooks/{id}** – Get notebook 200.
- [ ] **POST /api/v1/notebooks/{id}/cells** – Add cell (position, cell_type: sql|python|markdown, content) returns 201.
- [ ] **GET /api/v1/notebooks/{id}/cells** – List cells 200.
- [ ] **POST /api/v1/notebooks/{id}/runs** – Create run (triggered_by, optional workflow_execution_id) returns 201.
- [ ] **GET /api/v1/notebooks/{id}/runs** – List runs 200.

## 3. Decision dashboards & governance

- [ ] **POST /api/v1/decision-dashboards** – Create dashboard (name, owner_id, layout, metric_ids, visibility) returns 201.
- [ ] **GET /api/v1/decision-dashboards** – List dashboards 200.
- [ ] **GET /api/v1/decision-dashboards/{id}** – Get dashboard 200.
- [ ] **POST /api/v1/decision-dashboards/{id}/runs** – Record run (triggered_by, snapshot, status) returns 201.
- [ ] **GET /api/v1/lineage/graph** – Full lineage graph 200.
- [ ] **GET /api/v1/governance/rls/policies** – RLS policies 200.

## 4. ITSM (incidents, changes, runbooks)

- [ ] **POST /api/v1/itsm/incidents** – Create incident (title, description, priority, requester_id) returns 201.
- [ ] **GET /api/v1/itsm/incidents** – List incidents 200.
- [ ] **POST /api/v1/itsm/changes** – Create change (title, description, change_type, requester_id) returns 201.
- [ ] **GET /api/v1/itsm/changes** – List changes 200.
- [ ] **POST /api/v1/itsm/runbooks** – Create runbook (name, workflow_id) returns 201.
- [ ] **GET /api/v1/itsm/runbooks** – List runbooks 200.

## 5. Knowledge workspace & AI assistant

- [ ] **GET /api/v1/notion-ui/templates/pages** – List page templates 200.
- [ ] **GET /api/v1/notion-ui/templates/databases** – List database templates 200.
- [ ] **POST /api/v1/ai/assistant** – RAG query returns 200 with answer, context, sources.
- [ ] **POST /api/v1/rag/query** – RAG query 200 with answer and sources.

## SLO (measurable)

- Warehouse query p95 latency &lt; 30s (or configured timeout).
- Workload queue create/list &lt; 500ms.
- Data product create &lt; 500ms.
- Notebook create &lt; 500ms; add cell &lt; 500ms; create run &lt; 500ms.
- Decision dashboard create &lt; 500ms.
- Incident/create/runbook &lt; 500ms.
- List templates &lt; 500ms; AI assistant (RAG) p95 &lt; 10s (model-dependent).
