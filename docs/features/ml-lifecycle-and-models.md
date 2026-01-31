# ML Lifecycle & Models

<div align="center">

**ML lifecycle, model registry, model governance, and prompts**

[← Streaming & CDC](streaming-and-cdc.md) • [Notion UI (Blocks & Databases) →](notion-ui-blocks-databases.md)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [API Reference](#api-reference)
- [UI Routes](#ui-routes)
- [Related Documentation](#related-documentation)

---

## Overview

NeuronIP provides an ML lifecycle and model governance layer: experiment tracking, model registry, model versioning, approval workflows, and prompt management. Models and prompts can be registered, versioned, approved, and rolled back for safe deployment.

---

## Key Capabilities

- **Model registry** – Register and version ML models
- **Model governance** – List models, get versions, approve, rollback
- **Prompt management** – Version and govern prompts (approve, rollback)
- **ML lifecycle** – Experiments, pipelines, serving (see backend packages)
- **Quality scoring** – Model quality metrics and monitoring

---

## API Reference

- **Models (governance):** `GET/POST /api/v1/models`, `GET /api/v1/models/{id}`, versions, approve, rollback
- **Prompts:** `GET /api/v1/prompts`, `GET /api/v1/prompts/{id}`, versions, approve, rollback
- **Inference:** Model infer endpoints as documented in [API Endpoints](../api/endpoints.md#-models--prompts-governance)

See [API Endpoints](../api/endpoints.md) for full request/response details.

---

## UI Routes

- **Models** – Dashboard route `/models` (list-detail). See [UX Map](../UX_MAP.md).

---

## Related Documentation

- [API Endpoints: Models & Prompts](../api/endpoints.md#-models--prompts-governance)
- [Architecture: Backend](../architecture/backend.md)
- [FEATURE_MAP: Models / ML](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
