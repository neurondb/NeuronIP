# 🚀 Getting Started with NeuronIP

<div align="center">

**Get NeuronIP up and running in minutes**

[Installation](#-installation) • [Configuration](#-configuration) • [First Steps](#-first-steps) • [Next Steps](#-next-steps)

</div>

---

## 📋 Table of Contents

- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
  - [Docker Compose (Recommended)](#docker-compose-recommended)
  - [Manual Installation](#manual-installation)
- [Configuration](#-configuration)
- [Verification](#-verification)
- [First Steps](#-first-steps)
- [Next Steps](#-next-steps)
- [Troubleshooting](#-troubleshooting)

---

## ✅ Prerequisites

Before you begin, ensure you have the following installed:

### Required Software

| Software | Version | Purpose | Download |
|----------|---------|---------|----------|
| **Docker** | 20.10+ | Container runtime | [Get Docker](https://docs.docker.com/get-docker/) |
| **Docker Compose** | 2.0+ | Multi-container orchestration | [Get Docker Compose](https://docs.docker.com/compose/install/) |
| **PostgreSQL** | 16+ | Database (if not using Docker) | [Get PostgreSQL](https://www.postgresql.org/download/) |
| **NeuronDB Extension** | Latest | AI-native database extension | [Get NeuronDB](https://neurondb.ai) |

### Optional (for Development)

| Software | Version | Purpose |
|----------|---------|---------|
| **Go** | 1.24+ | Backend development |
| **Node.js** | 18+ | Frontend development |
| **Git** | Latest | Version control |

### System Requirements

- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB+ recommended
- **Disk**: 10GB+ free space
- **OS**: Linux, macOS, or Windows (with WSL2)

---

## 📦 Installation

### Docker Compose (Recommended)

The easiest way to get started is using Docker Compose. This method automatically sets up all services including the database.

#### Step 1: Clone the Repository

```bash
git clone https://github.com/neurondb/NeuronIP.git
cd NeuronIP
```

#### Step 2: Configure Environment Variables

Create a `.env` file in the root directory:

```bash
# Database Configuration
POSTGRES_USER=neurondb
POSTGRES_PASSWORD=your_secure_password_here
DB_NAME=neuronip

# NeuronAgent Configuration
NEURONAGENT_ENDPOINT=http://neuronagent:8080
NEURONAGENT_API_KEY=your_api_key_here

# Server Configuration
SERVER_PORT=8082
```

> ⚠️ **Security Warning**: Change the default passwords before deploying to production!

#### Step 3: Start Services

```bash
# Start all services
docker compose up -d

# Check service status
docker compose ps

# View logs
docker compose logs -f
```

#### Step 4: Verify Installation

```bash
# Check API health
curl http://localhost:8082/health

# Expected response:
# {"status":"healthy","timestamp":"2024-01-01T00:00:00Z"}
```

> ✅ **Success!** If you see a healthy status, NeuronIP is running correctly.

---

### Manual Installation

If you prefer to run services manually or need more control:

#### Backend Setup

```bash
# Navigate to API directory
cd api

# Install dependencies
go mod download

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=neuronip
export DB_PASSWORD=neuronip
export DB_NAME=neuronip
export SERVER_PORT=8082

# Initialize database
psql -d neuronip -f ../neuronip.sql

# Run server
go run cmd/server/main.go
```

#### Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Set environment variables
export NEXT_PUBLIC_API_URL=http://localhost:8082/api/v1

# Run development server
npm run dev
```

---

## ⚙️ Configuration

### Environment Variables

NeuronIP uses environment variables for configuration. Here's a complete reference:

<details>
<summary><b>📋 Complete Environment Variables List</b></summary>

#### Database Configuration

```bash
DB_HOST=localhost              # Database host
DB_PORT=5432                   # Database port
DB_USER=neuronip               # Database user
DB_PASSWORD=neuronip           # Database password
DB_NAME=neuronip               # Database name
DB_MAX_OPEN_CONNS=25           # Maximum open connections
DB_MAX_IDLE_CONNS=5            # Maximum idle connections
DB_CONN_MAX_LIFETIME=5m        # Connection max lifetime
```

#### Server Configuration

```bash
SERVER_HOST=0.0.0.0            # Server bind address
SERVER_PORT=8082               # Server port
SERVER_READ_TIMEOUT=30s        # Read timeout
SERVER_WRITE_TIMEOUT=30s       # Write timeout
```

#### Logging Configuration

```bash
LOG_LEVEL=info                 # Log level (debug, info, warn, error)
LOG_FORMAT=json                # Log format (json, text)
LOG_OUTPUT=stdout              # Log output (stdout, stderr, file path)
```

#### CORS Configuration

```bash
CORS_ALLOWED_ORIGINS=*         # Allowed origins (comma-separated)
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

#### Authentication Configuration

```bash
JWT_SECRET=your_jwt_secret     # JWT signing secret
ENABLE_API_KEYS=true           # Enable API key authentication
```

#### NeuronDB Configuration

```bash
NEURONDB_HOST=localhost        # NeuronDB host
NEURONDB_PORT=5433             # NeuronDB port
NEURONDB_DATABASE=neurondb     # NeuronDB database
NEURONDB_USER=neurondb         # NeuronDB user
NEURONDB_PASSWORD=neurondb     # NeuronDB password
```

#### NeuronAgent Configuration

```bash
NEURONAGENT_ENDPOINT=http://localhost:8080
NEURONAGENT_API_KEY=your_api_key
NEURONAGENT_ENABLE_SESSIONS=true
NEURONAGENT_ENABLE_WORKFLOWS=true
NEURONAGENT_SESSION_TIMEOUT=30m
NEURONAGENT_WORKFLOW_TIMEOUT=1h
```

#### NeuronMCP Configuration

```bash
NEURONMCP_BINARY_PATH=/usr/local/bin/neurondb-mcp
NEURONMCP_TOOL_CATEGORIES=vector,embedding,rag,ml,analytics,postgresql
NEURONMCP_ENABLE_VECTOR_OPS=true
NEURONMCP_ENABLE_ML_TOOLS=true
NEURONMCP_ENABLE_RAG_TOOLS=true
NEURONMCP_ENABLE_POSTGRES_TOOLS=true
NEURONMCP_TIMEOUT=30s
```

</details>

### Configuration Files

For production deployments, you may want to use configuration files. See [Configuration Reference](reference/configuration.md) for details.

---

## ✅ Verification

After installation, verify that all components are working:

### 1. Health Check

```bash
curl http://localhost:8082/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "database": "connected",
  "neurondb": "connected"
}
```

### 2. API Endpoints

```bash
# Test semantic search endpoint
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"query": "test query", "limit": 5}'
```

### 3. Frontend Access

Open your browser and navigate to:
- **Frontend**: http://localhost:3001
- **API**: http://localhost:8082
- **API Docs**: http://localhost:8082/api/v1/docs (if enabled)

### 4. Database Connection

```bash
# Connect to database
psql -h localhost -U neuronip -d neuronip

# Check tables
\dt neuronip.*

# Check NeuronDB extension
SELECT * FROM pg_extension WHERE extname = 'neurondb';
```

---

## 🎯 First Steps

Now that NeuronIP is running, let's do something useful:

### Step 1: Create an API Key

```bash
# Create API key (example - actual endpoint may vary)
curl -X POST http://localhost:8082/api/v1/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First API Key",
    "permissions": ["read", "write"]
  }'
```

### Step 2: Add Your First Document

```bash
curl -X POST http://localhost:8082/api/v1/semantic/documents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "title": "Getting Started Guide",
    "content": "This is my first document in NeuronIP...",
    "content_type": "document",
    "collection_id": null
  }'
```

### Step 3: Perform Your First Search

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "query": "getting started",
    "limit": 10
  }'
```

### Step 4: Explore the Frontend

1. Open http://localhost:3001 in your browser
2. Log in (or create an account)
3. Navigate to the **Semantic Search** section
4. Try searching for your document

---

## 🎓 Next Steps

Now that you have NeuronIP running, explore these resources:

### Tutorials

1. **[Quick Start Tutorial](tutorials/quick-start-tutorial.md)** - Build your first integration
2. **[Semantic Search Tutorial](tutorials/semantic-search-tutorial.md)** - Create a knowledge base
3. **[Warehouse Q&A Tutorial](tutorials/warehouse-qa-tutorial.md)** - Connect your data warehouse

### Documentation

- **[Architecture Overview](architecture/overview.md)** - Understand the system
- **[API Reference](api/endpoints.md)** - Explore all endpoints
- **[Feature Guides](features/)** - Deep dive into features

### Development

- **[Development Setup](development/setup.md)** - Set up for development
- **[Contributing Guide](development/contributing.md)** - Contribute to NeuronIP

---

## 🔧 Troubleshooting

### Common Issues

<details>
<summary><b>❌ Service won't start</b></summary>

**Problem**: Docker containers fail to start

**Solutions**:
1. Check Docker is running: `docker ps`
2. Check ports are available: `netstat -an | grep 8082`
3. View logs: `docker compose logs neuronip-api`
4. Check environment variables are set correctly

</details>

<details>
<summary><b>❌ Database connection failed</b></summary>

**Problem**: Cannot connect to database

**Solutions**:
1. Verify database is running: `docker compose ps neurondb`
2. Check connection string in environment variables
3. Verify database exists: `psql -l | grep neuronip`
4. Check firewall settings

</details>

<details>
<summary><b>❌ API returns 401 Unauthorized</b></summary>

**Problem**: Authentication failing

**Solutions**:
1. Verify API key is correct
2. Check `Authorization` header format: `Bearer YOUR_API_KEY`
3. Ensure API key hasn't expired
4. Check authentication is enabled: `ENABLE_API_KEYS=true`

</details>

<details>
<summary><b>❌ Frontend can't connect to API</b></summary>

**Problem**: Frontend shows connection errors

**Solutions**:
1. Verify `NEXT_PUBLIC_API_URL` is set correctly
2. Check CORS settings allow frontend origin
3. Verify API is running: `curl http://localhost:8082/health`
4. Check browser console for errors

</details>

### Getting Help

If you're still stuck:

1. 📖 Check the [Troubleshooting Guide](troubleshooting/common-issues.md)
2. 🔍 Search [GitHub Issues](https://github.com/neurondb/NeuronIP/issues)
3. 💬 Ask in [Discussions](https://github.com/neurondb/NeuronIP/discussions)
4. 📧 Contact support: support@neurondb.ai

---

## 📚 Additional Resources

- [Architecture Documentation](architecture/overview.md)
- [API Reference](api/endpoints.md)
- [Configuration Reference](reference/configuration.md)
- [Security Guide](security/overview.md)

---

<div align="center">

**Ready to build?** Check out our [tutorials](tutorials/) to get started!

[← Back to Documentation](README.md) • [Next: Architecture →](architecture/overview.md)

</div>
