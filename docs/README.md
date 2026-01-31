# 📚 NeuronIP Documentation

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)
![Node Version](https://img.shields.io/badge/node-18+-339933.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**AI-Native Enterprise Intelligence Platform**

[Getting Started](#-getting-started) • [Architecture](#-architecture) • [API Reference](#-api-reference) • [Features](#-features) • [Contributing](#-contributing)

</div>

---

## 📖 Table of Contents

- [Overview](#-overview)
- [Quick Links](#-quick-links)
- [Documentation Structure](#-documentation-structure)
- [Getting Started](#-getting-started)
- [Architecture](#-architecture)
- [API Reference](#-api-reference)
- [Features](#-features)
- [Development](#-development)
- [Deployment](#-deployment)
- [Integrations](#-integrations)
- [Security](#-security)
- [Tutorials](#-tutorials)
- [Troubleshooting](#-troubleshooting)
- [Reference](#-reference)

---

## 🎯 Overview

NeuronIP is a comprehensive enterprise intelligence platform that combines multiple capabilities into a unified AI-native system built on PostgreSQL. The platform provides end-to-end data intelligence, governance, and automation capabilities.

### Core Capabilities

| Feature | Description | Status |
|---------|-------------|--------|
| 🔍 **Semantic Knowledge Search** | Vector-based semantic search across knowledge base with RAG pipeline | ✅ Active |
| 💬 **Data Warehouse Q&A** | Natural language to SQL with query execution and visualization | ✅ Active |
| 🤖 **Customer Support Memory** | AI-powered support with long-term memory and context awareness | ✅ Active |
| 🛡️ **Compliance & Audit Analytics** | Policy matching, anomaly detection, semantic filtering, audit trails | ✅ Active |
| ⚙️ **Agent Workflows** | Multi-step workflow execution with AI agents and NeuronDB integration | ✅ Active |

### Enterprise Features

| Feature Category | Capabilities | Status |
|------------------|--------------|--------|
| **Data Governance** | Column/Row-level security, RBAC, Policy management, Access controls | ✅ Active |
| **Data Lineage** | Table & column-level lineage, Impact analysis, Cross-system tracking | ✅ Active |
| **Data Quality** | Quality metrics, Trend analysis, Automated profiling, Outlier detection | ✅ Active |
| **Data Catalog** | Dataset management, Schema evolution tracking, Search & discovery | ✅ Active |
| **Data Ingestion** | 30+ connectors (Snowflake, BigQuery, MySQL, etc.), ETL integration | ✅ Active |
| **Data Classification** | Automated PII detection, Data classification, Masking | ✅ Active |
| **Observability** | Query performance tracking, Metrics collection, Health monitoring | ✅ Active |
| **Knowledge Graph** | Entity extraction, Relationship mapping, Graph traversal | ✅ Active |
| **Versioning** | Data versioning, Model versioning, Pipeline versioning | ✅ Active |
| **Multi-Tenancy** | Tenant isolation, Multi-region support, Workspace management | ✅ Active |

### Security & Compliance

| Feature | Description | Status |
|---------|-------------|--------|
| **Authentication** | API Keys, JWT, OAuth 2.0, OIDC, SSO, 2FA, SCIM | ✅ Active |
| **Authorization** | RBAC, ABAC, Column/Row-level security, Permission management | ✅ Active |
| **Privacy** | DSAR automation, PIA, Consent management, GDPR compliance | ✅ Active |
| **Audit** | Comprehensive audit logging, Compliance reporting | ✅ Active |
| **Data Protection** | Encryption at rest, Data masking, PII detection | ✅ Active |

### Integrations & Extensions

| Component | Purpose | Status |
|-----------|---------|--------|
| **NeuronDB** | AI-native database extensions for vectors, ML, embeddings | ✅ Active |
| **NeuronAgent** | AI agent framework with sessions and workflows | ✅ Active |
| **NeuronMCP** | Model Context Protocol tools for AI operations | ✅ Active |
| **Webhooks** | Event-driven integrations with external systems | ✅ Active |
| **Slack/Teams** | Notification and alert integrations | ✅ Active |

### Key Technologies

- **Backend**: Go 1.24+ with Gorilla Mux, Structured logging, Prometheus metrics
- **Frontend**: Next.js 14 with TypeScript, Tailwind CSS, React Query
- **Database**: PostgreSQL 16+ with NeuronDB extension, Row-level security
- **Infrastructure**: Docker, Kubernetes, Multi-region support, HA deployment
- **Integrations**: NeuronDB, NeuronAgent, NeuronMCP, 30+ data connectors

---

## 🔗 Quick Links

### For Users
- [🚀 Getting Started Guide](getting-started.md) - Start using NeuronIP in minutes
- [📖 Feature Documentation](features/) - Learn about all features
- [🎓 Tutorials](tutorials/) - Step-by-step guides
- [🔧 Troubleshooting](troubleshooting/) - Common issues and solutions

### For Developers
- [🏗️ Architecture Overview](architecture/overview.md) - System design and components
- [💻 Development Setup](development/setup.md) - Set up your development environment
- [📝 API Reference](api/endpoints.md) - Complete API documentation
- [🤝 Contributing Guide](development/contributing.md) - How to contribute

### For Operators
- [🐳 Docker Deployment](deployment/docker.md) - Deploy with Docker
- [📦 Production Packaging](deployment/packaging.md) - How images are built and packaged
- [☸️ Production Deployment](deployment/production.md) - Production checklist
- [☸️ Kubernetes Deployment](deployment/kubernetes.md) - Kubernetes deployment guide
- [📊 Monitoring](deployment/monitoring.md) - Observability and monitoring
- [🔒 Security Guide](security/overview.md) - Security best practices

---

## 📁 Documentation Structure

```
docs/
├── README.md                    # This file - Documentation index
├── getting-started.md           # Quick start guide
├── authentication.md            # API key & login flow
├── authentication-session.md    # Session details
├── FEATURE_MAP.md               # Module → UI → API → migrations map
├── UX_MAP.md                    # Dashboard routes & page archetypes
├── implementation-status.md     # Implementation phases status
├── IMPLEMENTATION-COMPLETE.md   # Completion summary
├── ENTERPRISE_FEATURES_IMPLEMENTATION.md
├── neurondb-capability-audit.md # NeuronDB capability audit
├── architecture/                # Architecture documentation
│   ├── README.md
│   ├── overview.md
│   ├── backend.md
│   ├── frontend.md
│   ├── database.md
│   ├── data-flow.md
│   └── request-flow.md          # Request/response lifecycle & middleware order
├── api/                         # API documentation
│   ├── README.md
│   ├── overview.md
│   ├── endpoints.md
│   ├── authentication.md
│   └── rate-limiting.md
├── features/                    # Feature documentation
│   ├── semantic-search.md
│   ├── warehouse-qa.md
│   ├── support-memory.md
│   ├── compliance.md
│   ├── agent-workflows.md
│   └── (see Features section for full list)
├── development/                 # Development guides
│   ├── setup.md
│   ├── contributing.md
│   ├── coding-standards.md
│   ├── testing.md
│   ├── debugging.md
│   ├── error-handling.md
│   ├── validation.md
│   ├── wizards.md
│   └── neurondb-local-dev.md
├── deployment/                  # Deployment guides
│   ├── docker.md
│   ├── packaging.md
│   ├── production.md
│   ├── kubernetes.md
│   ├── monitoring.md
│   └── multi-region.md
├── competitive-analysis/        # Competitive analysis & status
│   └── README.md (+ comparison, gap analysis, etc.)
├── integrations/                # Integration guides
│   ├── neurondb.md
│   ├── neuronagent.md
│   ├── neuronmcp.md
│   └── custom-integrations.md
├── security/                    # Security documentation
│   ├── overview.md
│   ├── authentication.md
│   ├── authorization.md
│   └── data-protection.md
├── tutorials/                   # Tutorials and examples
│   ├── quick-start-tutorial.md
│   ├── semantic-search-tutorial.md
│   ├── warehouse-qa-tutorial.md
│   ├── agent-workflow-tutorial.md
│   └── api-integration-tutorial.md
├── troubleshooting/             # Troubleshooting guides
│   ├── common-issues.md
│   ├── performance.md
│   └── errors.md
└── reference/                   # Reference documentation
    ├── configuration.md
    ├── environment-variables.md
    ├── database-schema.md
    └── glossary.md
```

---

## 🚀 Getting Started

**Quick start:** To run the stack: [QUICK-START.md](../QUICK-START.md). To load demo data: [demo/QUICK-START.md](../demo/QUICK-START.md).

New to NeuronIP? Start here:

1. **[Quick Start Guide](getting-started.md)** - Get NeuronIP running in 5 minutes
2. **[Architecture Overview](architecture/overview.md)** - Understand the system
3. **[First Tutorial](tutorials/quick-start-tutorial.md)** - Build your first integration

### Prerequisites

- ✅ Docker and Docker Compose
- ✅ PostgreSQL 16+ with NeuronDB extension
- ✅ Go 1.24+ (for backend development)
- ✅ Node.js 18+ (for frontend development)

> 💡 **Tip**: Check the [Getting Started Guide](getting-started.md) for detailed setup instructions.

---

## 🏗️ Architecture

Understand how NeuronIP is built:

- **[System Overview](architecture/overview.md)** - High-level architecture with diagrams
- **[Backend Architecture](architecture/backend.md)** - Go services and design patterns
- **[Frontend Architecture](architecture/frontend.md)** - Next.js components and structure
- **[Database Design](architecture/database.md)** - Schema and data modeling
- **[Data Flow](architecture/data-flow.md)** - How data moves through the system
- **[Request Flow](architecture/request-flow.md)** - Request/response lifecycle and middleware order

---

## 📡 API Reference

Complete API documentation:

- **[API Overview](api/overview.md)** - Introduction to the REST API
- **[Endpoints](api/endpoints.md)** - Complete endpoint reference
- **[Authentication](api/authentication.md)** - Auth flows and security
- **[Rate Limiting](api/rate-limiting.md)** - Quotas and limits

### Quick API Example

```bash
# Health check
curl http://localhost:8082/health

# Semantic search
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"query": "What is NeuronIP?", "limit": 10}'
```

---

## ✨ Features

Detailed documentation for each feature:

### 🔍 [Semantic Knowledge Search](features/semantic-search.md)
Search your knowledge base by meaning, not just keywords. Powered by vector embeddings and semantic similarity.

**Key Capabilities:**
- Vector-based semantic search with similarity scoring
- Document chunking and indexing with metadata support
- RAG (Retrieval-Augmented Generation) pipeline
- Collection management and organization
- Semantic document embedding and retrieval
- Context-aware search results

### 💬 [Data Warehouse Q&A](features/warehouse-qa.md)
Ask natural language questions about your data warehouse and get SQL queries, visualizations, and explanations.

**Key Capabilities:**
- Natural language to SQL conversion
- Schema discovery and management
- Query execution and result visualization
- Query history and analytics
- Query optimization recommendations
- Result caching and performance tracking

### 🤖 [Customer Support Memory](features/support-memory.md)
AI-powered customer support with long-term memory and context awareness.

**Key Capabilities:**
- Ticket management with priority routing
- Conversation history and context preservation
- Similar case retrieval using semantic search
- AI agent integration with NeuronAgent
- Multi-turn conversation support
- Support analytics and metrics

### 🛡️ [Compliance & Audit Analytics](features/compliance.md)
Automated compliance checking, anomaly detection, and audit trail management.

**Key Capabilities:**
- Policy matching and enforcement
- Anomaly detection with ML-based patterns
- Comprehensive audit logging
- Compliance reporting and dashboards
- DSAR (Data Subject Access Request) automation
- PIA (Privacy Impact Assessment) workflows
- Consent management

### ⚙️ [Agent Workflows](features/agent-workflows.md)
Build and execute complex workflows with AI agents and long-term memory.

**Key Capabilities:**
- Workflow definition with YAML/JSON
- Multi-step agent orchestration
- State management and persistence
- Error recovery and retry logic
- Workflow templates and reuse
- Scheduled execution
- Monitoring and observability

### Additional feature guides

- **[Clustering](features/clustering.md)** - Multi-node clustering, Raft, sharding, task queue
- **[Streaming & CDC](features/streaming-and-cdc.md)** - Real-time streaming and change data capture
- **[ML Lifecycle & Models](features/ml-lifecycle-and-models.md)** - Model registry, governance, prompts
- **[Notion UI: Blocks & Databases](features/notion-ui-blocks-databases.md)** - Notion-like blocks and databases
- **[Workload & Data Products](features/workload-and-data-products.md)** - Workload management and data products
- **[Decision Dashboards & ITSM](features/decision-dashboards-and-itsm.md)** - Decision dashboards and ITSM integration
- **[Notebooks](features/notebooks.md)** - Collaborative notebooks
- **[Onboarding & Partners](features/onboarding-and-partners.md)** - Onboarding flows and partner ecosystem

### 📊 Data Governance & Lineage

**Data Lineage:**
- Table-level and column-level lineage tracking
- End-to-end lineage across systems
- Impact analysis and risk scoring
- Automatic lineage discovery from queries
- Transformation logic capture

**Data Quality:**
- Automated quality checks and metrics
- Trend analysis and anomaly detection
- Data profiling with statistics
- Quality dashboards and reporting
- Alerting and notification system

**Data Catalog:**
- Dataset discovery and search
- Schema evolution tracking
- Metadata management
- Ownership and stewardship
- Comments and collaboration

### 🔐 Security & Access Control

**Authentication & Authorization:**
- API key management
- JWT token-based authentication
- OAuth 2.0 and OIDC integration
- Single Sign-On (SSO)
- Two-factor authentication (2FA)
- SCIM 2.0 user provisioning

**Access Controls:**
- Role-Based Access Control (RBAC)
- Attribute-Based Access Control (ABAC)
- Column-level security policies
- Row-level security with PostgreSQL RLS
- Permission management and auditing

**Data Protection:**
- PII detection and classification
- Data masking and anonymization
- Encryption at rest and in transit
- Consent management
- Privacy compliance automation

### 📈 Observability & Monitoring

**Query Performance:**
- Query execution tracking
- Performance metrics and analytics
- Slow query identification
- Query optimization recommendations

**Metrics & Analytics:**
- Business metrics management
- Custom metric definitions
- Metric search and discovery
- Analytics dashboards

**Health & Monitoring:**
- Health check endpoints
- Component health tracking
- System metrics collection
- Prometheus integration
- Alerting and notifications

### 🔌 Data Connectors

**Supported Connectors (30+):**
- **Databases**: PostgreSQL, MySQL, SQL Server, Oracle, Snowflake, BigQuery, Redshift, Databricks, Teradata, Azure SQL, Azure Synapse
- **NoSQL**: MongoDB, Cassandra, DynamoDB, Redis, Elasticsearch
- **Data Lakes**: S3, Azure Blob Storage
- **Streaming**: Kafka
- **BI Tools**: Tableau, Power BI, Looker
- **ETL Tools**: dbt, Airflow, Fivetran, Stitch
- **CDP**: Segment, HubSpot
- **Monitoring**: Splunk

### 🧠 Knowledge Graph

- Entity extraction from text
- Relationship mapping and linking
- Graph traversal and exploration
- Entity search and discovery
- Knowledge graph visualization

### 🔄 Versioning & Evolution

- **Data Versioning**: Track data changes over time
- **Model Versioning**: ML model version management
- **Pipeline Versioning**: Data pipeline version tracking
- **Schema Evolution**: Automatic schema change detection and tracking

### 🌐 Enterprise Features

- **Multi-Tenancy**: Workspace isolation and management
- **Multi-Region**: Geographic distribution and data residency
- **Backup & Recovery**: Automated backups and disaster recovery
- **High Availability**: HA deployment with Kubernetes
- **Billing**: Usage tracking and billing integration

---

## 💻 Development

Resources for developers:

- **[Development Setup](development/setup.md)** - Set up your environment
- **[NeuronDB Local Dev](development/neurondb-local-dev.md)** - Run NeuronIP with a local NeuronDB build
- **[Contributing Guide](development/contributing.md)** - How to contribute
- **[Coding Standards](development/coding-standards.md)** - Code style guide
- **[Testing Guide](development/testing.md)** - Testing best practices
- **[Debugging Guide](development/debugging.md)** - Debugging tips

### Quick Development Commands

```bash
# Backend
cd api
go mod download
go run cmd/server/main.go

# Frontend
cd frontend
npm install
npm run dev
```

---

## 🚢 Deployment

Deployment guides and best practices:

- **[Docker Deployment](deployment/docker.md)** - Deploy with Docker Compose
- **[Production Packaging](deployment/packaging.md)** - How images are built and packaged
- **[Production Deployment](deployment/production.md)** - Production checklist
- **[Kubernetes Deployment](deployment/kubernetes.md)** - Kubernetes deployment guide
- **[Monitoring](deployment/monitoring.md)** - Observability setup

### Quick Docker Deployment

```bash
# Start all services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f neuronip-api
```

---

## 🔌 Integrations

Integration guides:

- **[NeuronDB Integration](integrations/neurondb.md)** - NeuronDB setup and usage
- **[NeuronAgent Integration](integrations/neuronagent.md)** - Agent configuration
- **[NeuronMCP Integration](integrations/neuronmcp.md)** - MCP tools setup
- **[Custom Integrations](integrations/custom-integrations.md)** - Build your own

---

## 🔒 Security

Security documentation:

- **[Security Overview](security/overview.md)** - Security architecture
- **[Authentication](security/authentication.md)** - Auth mechanisms
- **[Authorization](security/authorization.md)** - RBAC and permissions
- **[Data Protection](security/data-protection.md)** - Encryption and privacy

**Authentication docs (by topic):** [API key & login flow](authentication.md) | [Session details](authentication-session.md) | [API auth & tokens](api/authentication.md) | [Security auth](security/authentication.md)

> 🔒 **Security Note**: Always use HTTPS in production and keep your API keys secure.

---

## 🎓 Tutorials

Step-by-step tutorials:

1. **[Quick Start Tutorial](tutorials/quick-start-tutorial.md)** - Your first NeuronIP integration
2. **[Semantic Search Tutorial](tutorials/semantic-search-tutorial.md)** - Build a knowledge base
3. **[Warehouse Q&A Tutorial](tutorials/warehouse-qa-tutorial.md)** - Connect your data warehouse
4. **[Agent Workflow Tutorial](tutorials/agent-workflow-tutorial.md)** - Create an AI workflow
5. **[API Integration Tutorial](tutorials/api-integration-tutorial.md)** - Integrate with external systems

---

## 🔧 Troubleshooting

Common issues and solutions:

- **[Common Issues](troubleshooting/common-issues.md)** - Frequently encountered problems
- **[Performance](troubleshooting/performance.md)** - Optimization tips
- **[Error Reference](troubleshooting/errors.md)** - Error codes and meanings

---

## 📚 Reference

Reference documentation:

- **[Configuration](reference/configuration.md)** - All configuration options
- **[Environment Variables](reference/environment-variables.md)** - Complete env var reference
- **[Database Schema](reference/database-schema.md)** - Full schema documentation
- **[Glossary](reference/glossary.md)** - Terminology and definitions

---

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](development/contributing.md) for details. For repo-wide guidelines, see [CONTRIBUTING.md](../CONTRIBUTING.md).

### Quick Contribution Checklist

- [ ] Read the [Contributing Guide](development/contributing.md)
- [ ] Follow [Coding Standards](development/coding-standards.md)
- [ ] Write tests for new features
- [ ] Update documentation
- [ ] Submit a pull request

---

## 📞 Support

Need help?

- 📖 Check the [Troubleshooting Guide](troubleshooting/common-issues.md)
- 💬 Open an issue on GitHub
- 📧 Contact support: support@neurondb.ai
- 🔒 For security issues, see [SECURITY.md](../SECURITY.md)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with [NeuronDB](https://neurondb.ai) - AI-native database
- Powered by [NeuronAgent](https://neurondb.ai) - AI agent framework
- Integrated with [NeuronMCP](https://neurondb.ai) - Model Context Protocol

---

<div align="center">

**Made with ❤️ by the NeuronDB team**

[Documentation](.) • [GitHub](https://github.com/neurondb/NeuronIP) • [Website](https://neurondb.ai)

</div>
