# ⚙️ Configuration Reference

<div align="center">

**All configuration options for NeuronIP**

[Environment Variables →](environment-variables.md) • [Database Schema →](database-schema.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Configuration Methods](#configuration-methods)
- [Environment Variables](#environment-variables)
- [Configuration Categories](#configuration-categories)
- [Best Practices](#best-practices)

---

## 🎯 Overview

NeuronIP uses environment variables for all configuration. The configuration system supports:

- **Flexible configuration** via environment variables
- **Default values** for all non-critical settings
- **Validation** on startup to catch configuration errors early
- **Type safety** with proper parsing of durations, integers, and lists

All configuration is loaded at application startup from environment variables. See [Environment Variables](environment-variables.md) for the complete reference.

---

## ⚙️ Configuration Methods

### Method 1: Environment Variables (Recommended)

Set environment variables directly:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_PASSWORD=your_password
```

### Method 2: `.env` File

Create a `.env` file in the root directory:

```bash
# .env
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD=your_password
```

Then load it:

```bash
# With docker-compose (automatic)
docker compose up

# Manually
export $(cat .env | xargs)
```

### Method 3: Docker Compose Environment

In `docker-compose.yml`:

```yaml
services:
  neuronip-api:
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_PASSWORD=${DB_PASSWORD}
    env_file:
      - .env
```

### Method 4: Kubernetes ConfigMap/Secrets

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: neuronip-config
data:
  DB_HOST: "postgres"
  DB_PORT: "5432"
  SERVER_PORT: "8082"
```

---

## 📚 Configuration Categories

### Database Configuration

Control database connection settings, pooling, and connection lifecycle.

**Key Variables:**
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`

See [Environment Variables](environment-variables.md#database-variables) for details.

### Server Configuration

Configure HTTP server settings including host, port, and timeouts.

**Key Variables:**
- `SERVER_HOST`, `SERVER_PORT`
- `SERVER_READ_TIMEOUT`, `SERVER_WRITE_TIMEOUT`

See [Environment Variables](environment-variables.md#server-variables) for details.

### Authentication & Security

Configure authentication mechanisms, API keys, JWT, and SCIM.

**Key Variables:**
- `JWT_SECRET`, `ENABLE_API_KEYS`, `SCIM_SECRET`

See [Environment Variables](environment-variables.md#authentication-variables) for details.

### Integration Configuration

Configure NeuronDB, NeuronAgent, and NeuronMCP integrations.

**Key Variables:**
- `NEURONDB_HOST`, `NEURONDB_PORT`, `NEURONDB_DATABASE`
- `NEURONAGENT_ENDPOINT`, `NEURONAGENT_API_KEY`
- `NEURONMCP_BINARY_PATH`, `NEURONMCP_TOOL_CATEGORIES`

See [Environment Variables](environment-variables.md#integration-variables) for details.

### Observability Configuration

Configure logging, tracing, and monitoring.

**Key Variables:**
- `LOG_LEVEL`, `LOG_FORMAT`, `LOG_OUTPUT`
- `ENABLE_TRACING`

See [Environment Variables](environment-variables.md#logging-variables) for details.

### Rate Limiting Configuration

Configure rate limiting behavior.

**Key Variables:**
- `RATE_LIMIT_ENABLED`, `RATE_LIMIT_MAX_REQUESTS`, `RATE_LIMIT_WINDOW`

See [Environment Variables](environment-variables.md#rate-limiting-variables) for details.

### CORS Configuration

Configure Cross-Origin Resource Sharing.

**Key Variables:**
- `CORS_ALLOWED_ORIGINS`, `CORS_ALLOWED_METHODS`, `CORS_ALLOWED_HEADERS`

See [Environment Variables](environment-variables.md#cors-variables) for details.

---

## ✅ Best Practices

### 1. Use Environment-Specific Configuration

**Development:**
```bash
DB_HOST=localhost
LOG_LEVEL=debug
```

**Production:**
```bash
DB_HOST=postgres.production.internal
LOG_LEVEL=info
```

### 2. Secure Sensitive Values

- **Never commit secrets** to version control
- **Use secret management** tools (AWS Secrets Manager, HashiCorp Vault)
- **Rotate secrets regularly**
- **Use strong passwords** (minimum 32 characters for JWT secrets)

### 3. Validate Configuration

Configuration is validated on startup. Invalid configuration will prevent the server from starting:

```bash
Configuration validation failed: DB_PASSWORD is required
```

### 4. Use Defaults Wisely

Default values are provided for development convenience. Always override defaults in production:

- `DB_PASSWORD` - Always set a secure password
- `JWT_SECRET` - Required for JWT authentication
- `NEURONAGENT_API_KEY` - Required if using NeuronAgent

### 5. Monitor Configuration Changes

- Use configuration management tools
- Document all configuration changes
- Test configuration changes in staging first
- Keep configuration in version control (without secrets)

---

## 📖 Complete Reference

For complete details on all configuration options, see:

- **[Environment Variables Reference](environment-variables.md)** - Complete list of all environment variables with defaults and descriptions

---

## 🔧 Configuration Examples

### Minimal Development Configuration

```bash
DB_PASSWORD=dev_password
JWT_SECRET=dev_jwt_secret
```

### Production Configuration

See [Environment Variables](environment-variables.md#complete-configuration-example) for a complete production example.

---

<div align="center">

[← Back to Documentation](../README.md) • [Environment Variables →](environment-variables.md)

</div>
