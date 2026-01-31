# Workload & Data Products

<div align="center">

**Workload management and data product catalog**

[← Notion UI](notion-ui-blocks-databases.md) • [Decision Dashboards & ITSM →](decision-dashboards-and-itsm.md)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [API Reference](#api-reference)
- [Related Documentation](#related-documentation)

---

## Overview

NeuronIP supports workload management (prioritization, quotas, distributed execution) and data products. Data products are curated datasets or assets exposed for consumption; workload management ensures fair resource usage and prioritization across tenants and jobs.

---

## Key Capabilities

- **Workload management** – Priority queues, quota enforcement, distributed execution
- **Data products** – Define and expose data products for discovery and consumption
- **Quotas** – Set, list, and check quotas (see [API Endpoints: Quotas](../api/endpoints.md#-quotas))
- **Execution** – Priority-based execution and workload scheduling

---

## API Reference

- **Quotas:** `POST /api/v1/quotas/set`, `GET /api/v1/quotas/list`, `POST /api/v1/quotas/check`
- Data product and workload endpoints are documented in [API Endpoints](../api/endpoints.md) where exposed (e.g. warehouse and catalog areas).

---

## Related Documentation

- [API Endpoints: Quotas](../api/endpoints.md#-quotas)
- [Warehouse Q&A](warehouse-qa.md)
- [Data Catalog](../api/endpoints.md#-catalog)
- [FEATURE_MAP: Workload / Data Products](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
