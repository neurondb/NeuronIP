# NeuronIP — AI-Native Intelligence for PostgreSQL

<div align="center">
  <img src="neuronip.png" alt="NeuronIP" width="360" />

  <p>
    <a href="https://www.postgresql.org/">
      <img alt="PostgreSQL 16/17/18" src="https://img.shields.io/badge/PostgreSQL-16%2C17%2C18-blue.svg" />
    </a>
    <a href="LICENSE">
      <img alt="License" src="https://img.shields.io/badge/license-see%20LICENSE-red.svg" />
    </a>
  </p>

  <p><strong>NeuronIP is the AI-native intelligence and automation layer for PostgreSQL. It turns your database into an enterprise brain.</strong></p>

  <p>Five core capabilities unified into one platform: <strong>Semantic Knowledge Search</strong>, <strong>Data Warehouse Q&A</strong>, <strong>Customer Support Memory</strong>, <strong>Compliance & Audit Analytics</strong>, and <strong>Agent Workflows</strong>.</p>
</div>

> [!TIP]
> New here? Start with the [Quick Start](#quick-start-docker) section or jump to [What You Can Build](#-what-you-can-build).

## Hello NeuronIP (60 seconds)

Get NeuronIP running in under a minute:

```bash
# 1. Start NeuronIP with Docker Compose
docker compose up -d

# Wait for services to be healthy (about 30-60 seconds)
docker compose ps

# 2. Verify API is running
curl http://localhost:8082/health

# 3. Test semantic search (requires API key)
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "limit": 10
  }'
```

**Expected output:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

> [!SECURITY]
> The default credentials are for development only. **Always change them in production** by setting environment variables in your `.env` file. See [Service URLs & ports](#service-urls--ports) for connection details.

## Table of contents

- [What You Can Build](#-what-you-can-build)
- [What Makes NeuronIP Different](#-what-makes-neuronip-different)
- [Architecture](#architecture)
- [Installation](#installation)
  - [Quick start (Docker)](#quick-start-docker)
  - [Native install](#native-install)
  - [Manual setup](#manual-setup)
- [Service URLs & ports](#service-urls--ports)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Documentation](#documentation)
- [Repo layout](#repo-layout)
- [Contributing / security / license](#contributing--security--license)
- [Project statistics](#project-statistics)

## 🎯 What You Can Build

NeuronIP enables you to build powerful AI applications directly on top of PostgreSQL:

<details>
<summary><strong>🔍 Semantic Knowledge Search</strong></summary>

**Search your entire knowledge base by meaning, not just keywords:**

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "How do I configure authentication?",
    "collection_id": "docs",
    "limit": 5
  }'
```

**Use cases:**
- Document search with semantic understanding
- Knowledge base Q&A
- Content discovery and recommendations
- Intelligent document retrieval

</details>

<details>
<summary><strong>📊 Data Warehouse Q&A</strong></summary>

**Ask questions in natural language and get SQL + charts + explanations:**

```bash
curl -X POST http://localhost:8082/api/v1/warehouse/query \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the top 5 products by revenue this month?",
    "schema_id": "sales"
  }'
```

**Use cases:**
- Natural language to SQL conversion
- Business intelligence queries
- Data exploration and analysis
- Automated report generation

</details>

<details>
<summary><strong>🤖 Customer Support Memory</strong></summary>

**Automate support with AI agents and long-term memory:**

```bash
curl -X POST http://localhost:8082/api/v1/support/tickets \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Password reset issue",
    "description": "User cannot reset password",
    "priority": "high"
  }'
```

**Use cases:**
- Automated customer support
- Ticket routing and classification
- Similar case retrieval
- Support agent assistance

</details>

<details>
<summary><strong>🛡️ Compliance & Audit Analytics</strong></summary>

**Policy matching, anomaly detection, and semantic filtering:**

```bash
curl -X POST http://localhost:8082/api/v1/compliance/check \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_id": "gdpr-compliance",
    "data_source_id": "customer-db"
  }'
```

**Use cases:**
- Regulatory compliance checking
- Anomaly detection in data access
- Audit trail analysis
- Policy enforcement automation

</details>

<details>
<summary><strong>⚙️ Agent Workflows</strong></summary>

**Long-term memory and workflow execution powered by NeuronDB:**

```bash
curl -X POST http://localhost:8082/api/v1/workflows/{id}/execute \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "task": "analyze sales data",
      "parameters": {}
    }
  }'
```

**Use cases:**
- Automated workflow execution
- Multi-step data processing
- Agent orchestration
- Business process automation

</details>

## ⭐ What Makes NeuronIP Different

<details>
<summary><strong>📊 Feature Comparison</strong></summary>

| Feature | NeuronIP | Typical Alternatives |
|---------|----------|---------------------|
| **Semantic Search** | ✅ Built-in with NeuronDB | ❌ External vector DB required |
| **Warehouse Q&A** | ✅ Natural language to SQL | ❌ Manual SQL writing |
| **Support Memory** | ✅ Long-term agent memory | ❌ Stateless agents |
| **Compliance Analytics** | ✅ Policy matching + anomaly detection | ❌ Separate tools needed |
| **Agent Workflows** | ✅ Integrated with NeuronDB | ❌ External orchestration |
| **Unified Platform** | ✅ All features in one system | ❌ Multiple tools required |
| **PostgreSQL Native** | ✅ Direct database integration | ❌ External services |

</details>

<details>
<summary><strong>🎯 Key Advantages</strong></summary>

### 🚀 Enterprise Ready

- **Complete platform** - All five core capabilities in one system
- **PostgreSQL native** - Direct integration with your existing database
- **Production features** - Monitoring, authentication, RBAC, audit logging
- **Scalability** - Built for enterprise workloads

### 🔧 Developer Experience

- **REST API** - Clean, well-documented endpoints
- **Modern stack** - Go backend, Next.js frontend
- **Docker support** - Easy deployment and development
- **Comprehensive docs** - Clear examples and guides

### 🏢 Business Value

- **Unified intelligence** - One platform for all AI needs
- **Cost effective** - No need for multiple specialized tools
- **Fast time to value** - Quick setup and deployment
- **Extensible** - Easy to customize and extend

</details>

## Architecture

```mermaid
flowchart LR
  subgraph Frontend["NeuronIP Frontend"]
    UI["Next.js UI"]
  end
  subgraph Backend["NeuronIP API"]
    API["Go REST API"]
  end
  subgraph Database["PostgreSQL"]
    PG["PostgreSQL 16+"]
    EXT["NeuronDB Extension"]
  end
  subgraph Integrations["External Services"]
    AG["NeuronAgent"]
    MCP["NeuronMCP"]
  end
  UI -->|HTTP| API
  API -->|SQL| PG
  PG --> EXT
  API -->|REST/WebSocket| AG
  API -->|MCP Protocol| MCP
  AG -->|SQL| PG
  MCP -->|Tools| PG
```

> [!NOTE]
> NeuronIP integrates with the NeuronDB ecosystem (NeuronAgent, NeuronMCP) but can also run standalone. The `docker-compose.yml` can be configured to include or exclude these services.

## Installation

### Quick start (Docker)

**Option 1: Use Docker Compose (recommended)**

Start all NeuronIP services:

```bash
# Start all services
docker compose up -d

# Wait for services to be healthy (30-60 seconds)
docker compose ps

# Verify API is running
curl http://localhost:8082/health

# Access the application
# Frontend: http://localhost:3001
# Backend API: http://localhost:8082
```

**What you'll see:**
- 3 services starting: `neuronip-init`, `neuronip-api`, `neuronip-frontend`
- Database initialization with schema
- All services should show "healthy" status after initialization

> [!TIP]
> Make sure you have a PostgreSQL instance with NeuronDB extension running. You can use the NeuronDB Docker setup or your own PostgreSQL instance.

**Option 2: Build from source**

```bash
# Build and start all services
docker compose up -d --build

# Monitor build progress (first time takes 5-10 minutes)
docker compose logs -f

# Once built, wait for services to be healthy
docker compose ps

# Verify all services are running
curl http://localhost:8082/health
```

**Build time:** First build takes 5-10 minutes depending on your system. Subsequent starts are 30-60 seconds.

<details>
<summary><strong>Prerequisites checklist</strong></summary>

- [ ] Docker 20.10+ installed
- [ ] Docker Compose 2.0+ installed
- [ ] PostgreSQL 16+ with NeuronDB extension
- [ ] 4 GB+ RAM available
- [ ] Ports 3001, 8082 available

</details>

> [!IMPORTANT]
> **Production Security**: The default credentials shown in examples are for development only. Always use strong, unique passwords in production. Set environment variables via `.env` file or Docker secrets.

### Native install

Install NeuronIP components directly on your system without Docker.

#### Backend

**Prerequisites:**
- Go 1.24+ installed
- PostgreSQL 16+ with NeuronDB extension
- Database user with CREATE privileges

**Installation:**

```bash
cd api

# Install dependencies
go mod download

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=neuronip
export DB_PASSWORD=your_password
export DB_NAME=neuronip
export SERVER_PORT=8082

# Initialize database
psql -d neuronip -f ../neuronip.sql

# Run server
go run cmd/server/main.go
```

**Build binary:**

```bash
# Build for your platform
go build -o neuronip-api cmd/server/main.go

# Run the binary
./neuronip-api
```

#### Frontend

**Prerequisites:**
- Node.js 18+ installed
- npm or yarn

**Installation:**

```bash
cd frontend

# Install dependencies
npm install

# Set environment variables
export NEXT_PUBLIC_API_URL=http://localhost:8082/api/v1

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

### Manual setup

For detailed manual setup instructions, see the [Configuration](#configuration) section below.

## Service URLs & ports

| Service | How to reach it | Default credentials | Notes |
|---------|----------------|---------------------|-------|
| **NeuronIP Frontend** | `http://localhost:3001` | No auth (development mode) | Container: `neuronip-frontend`, Service: `neuronip-frontend` |
| **NeuronIP API** | `http://localhost:8082/health` | Health: no auth. API: API key required | Container: `neuronip-api`, Service: `neuronip-api` |
| **PostgreSQL** | `postgresql://neuronip:neuronip@localhost:5432/neuronip` | User: `neuronip`, Password: `neuronip` ⚠️ **Dev only** | Requires NeuronDB extension |

> [!WARNING]
> **Production Security**: The default credentials shown above are for development only. Always use strong, unique passwords in production. Set `POSTGRES_PASSWORD` and other secrets via environment variables or a `.env` file.

## Configuration

See `api/internal/config/config.go` for all configuration options.

### Key environment variables

**Database:**
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: neuronip)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: neuronip)

**Server:**
- `SERVER_PORT` - API server port (default: 8082)
- `LOG_LEVEL` - Logging level (default: info)
- `LOG_FORMAT` - Log format: json or text (default: json)

**NeuronDB Integration:**
- `NEURONDB_HOST` - NeuronDB host
- `NEURONDB_PORT` - NeuronDB port
- `NEURONDB_DATABASE` - NeuronDB database name

**NeuronAgent Integration:**
- `NEURONAGENT_ENDPOINT` - NeuronAgent API endpoint (default: http://localhost:8080)
- `NEURONAGENT_API_KEY` - NeuronAgent API key

**NeuronMCP Integration:**
- `NEURONMCP_BINARY_PATH` - Path to NeuronMCP binary
- `NEURONMCP_TOOL_CATEGORIES` - Comma-separated tool categories

## API Documentation

### Health Check

```bash
curl http://localhost:8082/health
```

### Authentication

Most endpoints require an API key. Include it in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8082/api/v1/...
```

### Semantic Search

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "collection_id": "docs",
    "limit": 10
  }'
```

### Warehouse Query

```bash
curl -X POST http://localhost:8082/api/v1/warehouse/query \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the top products by revenue?",
    "schema_id": "sales"
  }'
```

### Support Tickets

```bash
curl -X POST http://localhost:8082/api/v1/support/tickets \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "subject": "Issue description",
    "description": "Detailed issue description",
    "priority": "high"
  }'
```

### Compliance Check

```bash
curl -X POST http://localhost:8082/api/v1/compliance/check \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_id": "gdpr-compliance",
    "data_source_id": "customer-db"
  }'
```

### Workflow Execution

```bash
curl -X POST http://localhost:8082/api/v1/workflows/{id}/execute \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "task": "analyze data",
      "parameters": {}
    }
  }'
```

## Documentation

- **Getting Started**: This README provides a quick start guide
- **API Reference**: See [API Documentation](#api-documentation) section above
- **Configuration**: See `api/internal/config/config.go` for all configuration options
- **Database Schema**: See `neuronip.sql` for database schema
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines

### Component Documentation

- **Backend API**: Go-based REST API in `api/` directory
- **Frontend UI**: Next.js application in `frontend/` directory
- **Database**: PostgreSQL schema and migrations in `neuronip.sql` and `api/migrations/`

## Repo layout

| Component | Path | What it is |
|-----------|------|------------|
| **API Backend** | `api/` | Go-based REST API with handlers, services, middleware, and database models |
| **Frontend** | `frontend/` | Next.js 14 application with TypeScript, React components, and UI libraries |
| **Database Schema** | `neuronip.sql` | Complete database schema with tables, indexes, and functions |
| **Migrations** | `api/migrations/` | Database migration scripts |
| **Docker** | `docker-compose.yml` | Docker Compose configuration for all services |
| **Documentation** | `docs/` | Additional documentation files |

### Key Directories

**Backend (`api/`):**
- `cmd/server/` - Main application entry point
- `internal/handlers/` - HTTP request handlers
- `internal/services/` - Business logic services
- `internal/db/` - Database models and queries
- `internal/middleware/` - HTTP middleware (auth, CORS, logging, etc.)
- `internal/config/` - Configuration management

**Frontend (`frontend/`):**
- `app/` - Next.js app router pages and layouts
- `components/` - React components organized by feature
- `lib/` - Utilities, hooks, API clients, and theme configuration

## Contributing / security / license

- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md) - Guidelines for contributing to NeuronIP
- **Security**: [SECURITY.md](SECURITY.md) - Report security issues
- **License**: [LICENSE](LICENSE) - See LICENSE file for license information
- **Changelog**: [CHANGELOG.md](CHANGELOG.md) - See what's new

## Project statistics

<details>
<summary><strong>Stats snapshot</strong></summary>

- **5 core capabilities** unified in one platform
- **50+ API endpoints** across all features
- **2 main components** (Backend API + Frontend UI)
- **PostgreSQL 16+** supported
- **Go 1.24+** for backend
- **Next.js 14** for frontend
- **TypeScript** throughout frontend codebase

</details>

<details>
<summary><strong>Platform & version coverage</strong></summary>

| Category | Supported Versions |
|---------|-------------------|
| **PostgreSQL** | 16, 17, 18 |
| **Go** | 1.24+ |
| **Node.js** | 18 LTS, 20 LTS, 22 LTS |
| **Operating Systems** | Linux, macOS, Windows (via Docker) |
| **Architectures** | linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 |

</details>
