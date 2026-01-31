# Decision Dashboards & ITSM

<div align="center">

**Decision dashboards and IT service management integration**

[← Workload & Data Products](workload-and-data-products.md) • [Notebooks →](notebooks.md)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [API Reference](#api-reference)
- [Related Documentation](#related-documentation)

---

## Overview

NeuronIP provides decision dashboards for governance and decision tracking, and ITSM (IT Service Management) integration for ticketing, workflows, and helpdesk sync. Decision dashboards surface approval queues and governance metrics; ITSM connects to external ticketing and support systems.

---

## Key Capabilities

- **Decision dashboards** – Governance decision tracking, approval queues, dashboards
- **ITSM integration** – Connect to helpdesk and ticketing systems
- **Approval workflows** – Approval workflow support (see governance)
- **Helpdesk sync** – Sync with external helpdesk (e.g. via integrations API)

---

## API Reference

- **Integrations (helpdesk):** `POST /api/v1/integrations/helpdesk/sync` and related integration endpoints
- Decision dashboard and ITSM-specific routes are documented in [API Endpoints](../api/endpoints.md) and [FEATURE_MAP](../FEATURE_MAP.md)

---

## Related Documentation

- [API Endpoints: Integrations](../api/endpoints.md#-integrations)
- [Compliance](compliance.md)
- [FEATURE_MAP](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
