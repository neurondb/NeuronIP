# 🔧 Environment Variables Reference

<div align="center">

**Complete environment variables reference**

[← Configuration](configuration.md) • [Database Schema →](database-schema.md)

</div>

---

## 📋 Table of Contents

- [Database Variables](#database-variables)
- [Server Variables](#server-variables)
- [Logging Variables](#logging-variables)
- [CORS Variables](#cors-variables)
- [Authentication Variables](#authentication-variables)
- [Integration Variables](#integration-variables)
- [Observability Variables](#observability-variables)
- [Rate Limiting Variables](#rate-limiting-variables)

---

## 💾 Database Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | Database host | `localhost` | No |
| `DB_PORT` | Database port | `5432` | No |
| `DB_USER` | Database user | `neuronip` | No |
| `DB_PASSWORD` | Database password | `neuronip` | **Yes** |
| `DB_NAME` | Database name | `neuronip` | No |
| `DB_MAX_OPEN_CONNS` | Maximum open connections | `25` | No |
| `DB_MAX_IDLE_CONNS` | Maximum idle connections | `5` | No |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime | `5m` | No |

**Example:**
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=neuronip
DB_PASSWORD=your_secure_password
DB_NAME=neuronip
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

---

## 🖥️ Server Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` | No |
| `SERVER_PORT` | Server port | `8082` | No |
| `SERVER_READ_TIMEOUT` | Read timeout (duration) | `30s` | No |
| `SERVER_WRITE_TIMEOUT` | Write timeout (duration) | `30s` | No |

**Example:**
```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8082
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
```

---

## 📝 Logging Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` | No |
| `LOG_FORMAT` | Log format (`json`, `text`) | `json` | No |
| `LOG_OUTPUT` | Log output (`stdout`, `stderr`, or file path) | `stdout` | No |

**Example:**
```bash
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

---

## 🌐 CORS Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `CORS_ALLOWED_ORIGINS` | Allowed origins (comma-separated) | `*` | No |
| `CORS_ALLOWED_METHODS` | Allowed methods (comma-separated) | `GET,POST,PUT,DELETE,OPTIONS` | No |
| `CORS_ALLOWED_HEADERS` | Allowed headers (comma-separated) | `Content-Type,Authorization` | No |

**Example:**
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.neurondb.ai
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With
```

---

## 🔐 Authentication Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET` | JWT signing secret | (empty) | **Yes** (for JWT auth) |
| `ENABLE_API_KEYS` | Enable API key authentication | `true` | No |
| `SCIM_SECRET` | SCIM 2.0 secret for user provisioning | (empty) | No (for SCIM) |

**Example:**
```bash
JWT_SECRET=your_jwt_secret_key_here
ENABLE_API_KEYS=true
SCIM_SECRET=your_scim_secret
```

> ⚠️ **Security Warning**: Always use strong, randomly generated secrets in production. Never commit secrets to version control.

---

## 🔌 Integration Variables

### NeuronDB

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NEURONDB_HOST` | NeuronDB host | `localhost` | No |
| `NEURONDB_PORT` | NeuronDB port | `5433` | No |
| `NEURONDB_DATABASE` | NeuronDB database name | `neurondb` | No |
| `NEURONDB_USER` | NeuronDB user | `neurondb` | No |
| `NEURONDB_PASSWORD` | NeuronDB password | `neurondb` | No |

**Example:**
```bash
NEURONDB_HOST=localhost
NEURONDB_PORT=5433
NEURONDB_DATABASE=neurondb
NEURONDB_USER=neurondb
NEURONDB_PASSWORD=your_neurondb_password
```

### NeuronAgent

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NEURONAGENT_ENDPOINT` | NeuronAgent API endpoint | `http://localhost:8080` | No |
| `NEURONAGENT_API_KEY` | NeuronAgent API key | (empty) | **Yes** (for NeuronAgent) |
| `NEURONAGENT_ENABLE_SESSIONS` | Enable session management | `true` | No |
| `NEURONAGENT_ENABLE_WORKFLOWS` | Enable workflow execution | `true` | No |
| `NEURONAGENT_ENABLE_EVALUATION` | Enable agent evaluation | `true` | No |
| `NEURONAGENT_ENABLE_REPLAY` | Enable workflow replay | `true` | No |
| `NEURONAGENT_SESSION_TIMEOUT` | Session timeout duration | `30m` | No |
| `NEURONAGENT_WORKFLOW_TIMEOUT` | Workflow timeout duration | `1h` | No |

**Example:**
```bash
NEURONAGENT_ENDPOINT=http://localhost:8080
NEURONAGENT_API_KEY=your_agent_api_key
NEURONAGENT_ENABLE_SESSIONS=true
NEURONAGENT_ENABLE_WORKFLOWS=true
NEURONAGENT_SESSION_TIMEOUT=30m
NEURONAGENT_WORKFLOW_TIMEOUT=1h
```

### NeuronMCP

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NEURONMCP_BINARY_PATH` | Path to NeuronMCP binary | `/usr/local/bin/neurondb-mcp` | No |
| `NEURONMCP_TOOL_CATEGORIES` | Tool categories (comma-separated) | `vector,embedding,rag,ml,analytics,postgresql` | No |
| `NEURONMCP_ENABLE_VECTOR_OPS` | Enable vector operations | `true` | No |
| `NEURONMCP_ENABLE_ML_TOOLS` | Enable ML tools | `true` | No |
| `NEURONMCP_ENABLE_RAG_TOOLS` | Enable RAG tools | `true` | No |
| `NEURONMCP_ENABLE_POSTGRES_TOOLS` | Enable PostgreSQL tools | `true` | No |
| `NEURONMCP_TIMEOUT` | MCP operation timeout | `30s` | No |

**Example:**
```bash
NEURONMCP_BINARY_PATH=/usr/local/bin/neurondb-mcp
NEURONMCP_TOOL_CATEGORIES=vector,embedding,rag,ml,analytics,postgresql
NEURONMCP_ENABLE_VECTOR_OPS=true
NEURONMCP_ENABLE_ML_TOOLS=true
NEURONMCP_ENABLE_RAG_TOOLS=true
NEURONMCP_ENABLE_POSTGRES_TOOLS=true
NEURONMCP_TIMEOUT=30s
```

---

## 👁️ Observability Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENABLE_TRACING` | Enable distributed tracing | `false` | No |

**Example:**
```bash
ENABLE_TRACING=true
```

---

## ⏱️ Rate Limiting Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `RATE_LIMIT_ENABLED` | Enable rate limiting | `true` | No |
| `RATE_LIMIT_MAX_REQUESTS` | Maximum requests per window | `1000` | No |
| `RATE_LIMIT_WINDOW` | Rate limit window duration | `1h` | No |

**Example:**
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=1000
RATE_LIMIT_WINDOW=1h
```

---

## 📝 Complete Configuration Example

Here's a complete `.env` file example for production:

```bash
# Database Configuration
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=neuronip
DB_PASSWORD=secure_password_here
DB_NAME=neuronip
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8082
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://app.neurondb.ai,https://admin.neurondb.ai
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Authentication Configuration
JWT_SECRET=your_secure_jwt_secret_minimum_32_characters
ENABLE_API_KEYS=true
SCIM_SECRET=your_scim_secret_for_user_provisioning

# NeuronDB Integration
NEURONDB_HOST=neurondb.example.com
NEURONDB_PORT=5433
NEURONDB_DATABASE=neurondb
NEURONDB_USER=neurondb
NEURONDB_PASSWORD=neurondb_password

# NeuronAgent Integration
NEURONAGENT_ENDPOINT=https://agent.neurondb.ai
NEURONAGENT_API_KEY=your_agent_api_key
NEURONAGENT_ENABLE_SESSIONS=true
NEURONAGENT_ENABLE_WORKFLOWS=true
NEURONAGENT_SESSION_TIMEOUT=30m
NEURONAGENT_WORKFLOW_TIMEOUT=1h

# NeuronMCP Integration
NEURONMCP_BINARY_PATH=/usr/local/bin/neurondb-mcp
NEURONMCP_TOOL_CATEGORIES=vector,embedding,rag,ml,analytics,postgresql
NEURONMCP_ENABLE_VECTOR_OPS=true
NEURONMCP_ENABLE_ML_TOOLS=true
NEURONMCP_ENABLE_RAG_TOOLS=true
NEURONMCP_ENABLE_POSTGRES_TOOLS=true
NEURONMCP_TIMEOUT=30s

# Observability
ENABLE_TRACING=true

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX_REQUESTS=1000
RATE_LIMIT_WINDOW=1h
```

---

## 🔒 Security Best Practices

1. **Never commit secrets** to version control
2. **Use strong, randomly generated secrets** (minimum 32 characters)
3. **Use environment-specific configuration** (dev, staging, production)
4. **Rotate secrets regularly** in production
5. **Use secret management tools** (AWS Secrets Manager, HashiCorp Vault, etc.)
6. **Restrict database access** to only necessary IPs
7. **Enable SSL/TLS** for database connections in production
8. **Use HTTPS** for all API endpoints in production

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
