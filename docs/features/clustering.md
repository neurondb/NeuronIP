# Clustering & Distributed Execution

<div align="center">

**Multi-node clustering, Raft consensus, sharding, and task distribution**

[← Agent Workflows](agent-workflows.md) • [Streaming & CDC →](streaming-and-cdc.md)

</div>

---

## Table of Contents

- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [API Reference](#api-reference)
- [UI & Deployment](#ui--deployment)
- [Related Documentation](#related-documentation)

---

## Overview

NeuronIP supports multi-node clustering for high availability, horizontal scaling, and distributed task execution. The cluster layer provides membership, coordination (e.g. Raft), sharding, autoscaling, load balancing, and a task queue for distributing work across nodes.

---

## Key Capabilities

- **Membership** – Node discovery and cluster membership
- **Coordination** – Raft-based consensus for leader election and coordination
- **Sharding** – Data/work sharding across nodes
- **Task queue** – Distributed task execution and workload distribution
- **Autoscaling** – Scale nodes based on load
- **Load balancing** – Distribute requests across healthy nodes
- **Region support** – Multi-region deployment (see [Multi-Region](../deployment/multi-region.md))

---

## API Reference

Cluster APIs are used internally for node coordination. For public API surface (e.g. quotas, workload), see [API Endpoints](../api/endpoints.md). Cluster configuration and health are typically exposed via deployment and health endpoints.

---

## UI & Deployment

Clustering is primarily an operational concern. Dashboard routes for observability and workload may show cluster-backed data. See [UX Map](../UX_MAP.md) for dashboard routes. For deployment: [Kubernetes](../deployment/kubernetes.md), [Multi-Region](../deployment/multi-region.md).

---

## Related Documentation

- [Deployment: Multi-Region](../deployment/multi-region.md)
- [Deployment: Kubernetes](../deployment/kubernetes.md)
- [Architecture: Backend](../architecture/backend.md)
- [FEATURE_MAP: Clustering / Distributed](../FEATURE_MAP.md)

---

<div align="center">[← Back to Features](../README.md)</div>
