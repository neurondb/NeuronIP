# 📡 API Documentation

Complete REST API reference for NeuronIP.

## 📚 Contents

- **[Overview](overview.md)** - API introduction and authentication
- **[Endpoints](endpoints.md)** - Complete endpoint reference
- **[Authentication](authentication.md)** - Auth flows and security
- **[Rate Limiting](rate-limiting.md)** - Quotas and limits

## 🚀 Quick Start

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

[← Back to Documentation](../README.md)
