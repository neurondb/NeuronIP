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

NeuronIP is a comprehensive enterprise intelligence platform that combines five core capabilities into a unified system:

| Feature | Description | Status |
|---------|-------------|--------|
| 🔍 **Semantic Knowledge Search** | Search your entire knowledge base by meaning | ✅ Active |
| 💬 **Data Warehouse Q&A** | Ask questions and get SQL + charts + explanation | ✅ Active |
| 🤖 **Customer Support Memory** | Automate support with AI agents and long-term memory | ✅ Active |
| 🛡️ **Compliance & Audit Analytics** | Policy matching, anomaly detection, semantic filtering | ✅ Active |
| ⚙️ **Agent Workflows** | Long-term memory and workflow execution powered by NeuronDB | ✅ Active |

### Key Technologies

- **Backend**: Go 1.24+ with Gorilla Mux
- **Frontend**: Next.js 14 with TypeScript
- **Database**: PostgreSQL 16+ with NeuronDB extension
- **Integrations**: NeuronDB, NeuronAgent, NeuronMCP

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
├── getting-started.md          # Quick start guide
├── architecture/              # Architecture documentation
│   ├── README.md
│   ├── overview.md
│   ├── backend.md
│   ├── frontend.md
│   ├── database.md
│   └── data-flow.md
├── api/                       # API documentation
│   ├── README.md
│   ├── overview.md
│   ├── endpoints.md
│   ├── authentication.md
│   └── rate-limiting.md
├── features/                  # Feature documentation
│   ├── semantic-search.md
│   ├── warehouse-qa.md
│   ├── support-memory.md
│   ├── compliance.md
│   └── agent-workflows.md
├── development/               # Development guides
│   ├── setup.md
│   ├── contributing.md
│   ├── coding-standards.md
│   ├── testing.md
│   └── debugging.md
├── deployment/                # Deployment guides
│   ├── docker.md
│   ├── packaging.md
│   ├── production.md
│   ├── kubernetes.md
│   └── monitoring.md
├── integrations/              # Integration guides
│   ├── neurondb.md
│   ├── neuronagent.md
│   ├── neuronmcp.md
│   └── custom-integrations.md
├── security/                   # Security documentation
│   ├── overview.md
│   ├── authentication.md
│   ├── authorization.md
│   └── data-protection.md
├── tutorials/                 # Tutorials and examples
│   ├── quick-start-tutorial.md
│   ├── semantic-search-tutorial.md
│   ├── warehouse-qa-tutorial.md
│   ├── agent-workflow-tutorial.md
│   └── api-integration-tutorial.md
├── troubleshooting/           # Troubleshooting guides
│   ├── common-issues.md
│   ├── performance.md
│   └── errors.md
└── reference/                  # Reference documentation
    ├── configuration.md
    ├── environment-variables.md
    ├── database-schema.md
    └── glossary.md
```

---

## 🚀 Getting Started

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
- Vector-based semantic search
- Document chunking and indexing
- RAG (Retrieval-Augmented Generation) pipeline
- Collection management

### 💬 [Data Warehouse Q&A](features/warehouse-qa.md)
Ask natural language questions about your data warehouse and get SQL queries, visualizations, and explanations.

**Key Capabilities:**
- Natural language to SQL conversion
- Schema discovery and management
- Query execution and result visualization
- Query history and analytics

### 🤖 [Customer Support Memory](features/support-memory.md)
AI-powered customer support with long-term memory and context awareness.

**Key Capabilities:**
- Ticket management
- Conversation history
- Similar case retrieval
- AI agent integration

### 🛡️ [Compliance & Audit Analytics](features/compliance.md)
Automated compliance checking, anomaly detection, and audit trail management.

**Key Capabilities:**
- Policy matching
- Anomaly detection
- Audit logging
- Compliance reporting

### ⚙️ [Agent Workflows](features/agent-workflows.md)
Build and execute complex workflows with AI agents and long-term memory.

**Key Capabilities:**
- Workflow definition and execution
- Agent orchestration
- State management
- Error recovery

---

## 💻 Development

Resources for developers:

- **[Development Setup](development/setup.md)** - Set up your environment
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

We welcome contributions! See our [Contributing Guide](development/contributing.md) for details.

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
