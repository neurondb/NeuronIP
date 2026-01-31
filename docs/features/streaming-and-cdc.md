# Streaming & CDC

<div align="center">

**Real-time streaming engine and change data capture (CDC)**

[← Clustering](clustering.md) • [ML Lifecycle & Models →](ml-lifecycle-and-models.md)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [API Reference](#api-reference)
- [Related Documentation](#related-documentation)

---

## Overview

NeuronIP supports real-time data ingestion via a streaming engine and change data capture (CDC). CDC captures changes from source databases (e.g. PostgreSQL, MySQL) or event streams (Kafka) and propagates them for downstream processing, lineage, and analytics.

---

## Key Capabilities

- **Streaming engine** – Event ingestion and processing pipeline
- **CDC – PostgreSQL** – Change capture from PostgreSQL (logical replication or polling)
- **CDC – MySQL** – Change capture from MySQL
- **CDC – Kafka** – Consume change events from Kafka (e.g. Debezium)
- **Debezium integration** – Use Debezium-formatted change events
- **Real-time ingestion** – Low-latency pipeline for streaming data

---

## API Reference

Streaming and CDC are typically configured via ingestion and data-source APIs. For endpoints (ingestion jobs, data-sources status, failures, retry), see [API Endpoints](../api/endpoints.md).

---

## Related Documentation

- [API Endpoints: Ingestion](../api/endpoints.md#-ingestion)
- [Data Sources](../api/endpoints.md#-data-sources)
- [Architecture: Data Flow](../architecture/data-flow.md)
- [FEATURE_MAP](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
