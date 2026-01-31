# Notion UI: Blocks & Databases

<div align="center">

**Notion-like blocks and databases for content and structured data**

[← ML Lifecycle & Models](ml-lifecycle-and-models.md) • [Workload & Data Products →](workload-and-data-products.md)

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

NeuronIP includes a Notion-style UI for flexible content (blocks) and structured databases. Blocks support rich content and reordering; databases support rows, properties, and view preferences. Templates can be used to bootstrap pages and databases.

---

## Key Capabilities

- **Blocks** – Create, read, update, delete, and reorder content blocks
- **Databases** – Notion-style databases with rows and view preferences
- **Templates** – Reusable block and database templates
- **Notion handler** – Backend support for block and database operations

---

## API Reference

- **Blocks:** `POST/GET/PUT/DELETE /api/v1/blocks`, `POST /api/v1/blocks/reorder`
- **Databases:** `POST/GET/PUT/DELETE /api/v1/databases`, `GET/POST /api/v1/databases/{id}/rows`, `GET /api/v1/databases/view-preferences`

See [API Endpoints: Blocks & Databases](../api/endpoints.md#-blocks) and [Databases (Notion UI)](../api/endpoints.md#-databases-notion-ui).

---

## UI Routes

- **Notion UI** – Dashboard route `/notion-ui` (builder archetype: blocks, databases). See [UX Map](../UX_MAP.md).

---

## Related Documentation

- [API Endpoints: Blocks](../api/endpoints.md#-blocks)
- [API Endpoints: Databases (Notion UI)](../api/endpoints.md#-databases-notion-ui)
- [UX Map: notion-ui](../UX_MAP.md)
- [FEATURE_MAP: Blocks UI / Databases UI](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
