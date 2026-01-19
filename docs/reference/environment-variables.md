# 🔧 Environment Variables Reference

<div align="center">

**Complete environment variables reference**

[← Configuration](configuration.md) • [Database Schema →](database-schema.md)

</div>

---

## 📋 Table of Contents

- [Database Variables](#database-variables)
- [Server Variables](#server-variables)
- [Authentication Variables](#authentication-variables)
- [Integration Variables](#integration-variables)

---

## 💾 Database Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `neuronip` |
| `DB_PASSWORD` | Database password | `neuronip` |
| `DB_NAME` | Database name | `neuronip` |

---

## 🖥️ Server Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8082` |
| `SERVER_READ_TIMEOUT` | Read timeout | `30s` |
| `SERVER_WRITE_TIMEOUT` | Write timeout | `30s` |

---

## 🔐 Authentication Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | JWT signing secret | (required) |
| `ENABLE_API_KEYS` | Enable API key auth | `true` |

---

## 🔌 Integration Variables

### NeuronDB

| Variable | Description | Default |
|----------|-------------|---------|
| `NEURONDB_HOST` | NeuronDB host | `localhost` |
| `NEURONDB_PORT` | NeuronDB port | `5433` |
| `NEURONDB_DATABASE` | NeuronDB database | `neurondb` |

### NeuronAgent

| Variable | Description | Default |
|----------|-------------|---------|
| `NEURONAGENT_ENDPOINT` | NeuronAgent endpoint | `http://localhost:8080` |
| `NEURONAGENT_API_KEY` | NeuronAgent API key | (required) |

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
