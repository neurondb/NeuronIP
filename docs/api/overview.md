# 📡 API Overview

<div align="center">

**NeuronIP REST API Introduction**

[Endpoints →](endpoints.md) • [Authentication →](authentication.md)

</div>

---

## 📋 Table of Contents

- [Base URL](#base-url)
- [API Versioning](#api-versioning)
- [Request Format](#request-format)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Authentication](#authentication)

---

## 🌐 Base URL

### Development
```
http://localhost:8082/api/v1
```

### Production
```
https://api.neurondb.ai/v1
```

---

## 🔢 API Versioning

NeuronIP uses URL-based versioning:

- **Current Version**: `v1`
- **Version Format**: `/api/v1/...`

Future versions will be available at `/api/v2/...`, etc.

---

## 📤 Request Format

### Content-Type

All requests must include the `Content-Type` header:

```
Content-Type: application/json
```

### Example Request

```bash
curl -X POST http://localhost:8082/api/v1/semantic/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "query": "What is NeuronIP?",
    "limit": 10
  }'
```

---

## 📥 Response Format

### Success Response

```json
{
  "results": [...],
  "count": 10
}
```

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Query is required",
    "details": {
      "field": "query",
      "reason": "missing_required_field"
    }
  }
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request |
| `401` | Unauthorized |
| `403` | Forbidden |
| `404` | Not Found |
| `429` | Too Many Requests |
| `500` | Internal Server Error |

---

## ⚠️ Error Handling

### Error Response Structure

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "additional_info"
    }
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `BAD_REQUEST` | Invalid request format |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `NOT_FOUND` | Resource not found |
| `VALIDATION_FAILED` | Request validation failed |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

---

## 🔒 Authentication

### API Key Authentication

Include your API key in the `Authorization` header:

```
Authorization: Bearer YOUR_API_KEY
```

### Getting an API Key

1. Log in to the NeuronIP dashboard
2. Navigate to **Settings** → **API Keys**
3. Click **Create API Key**
4. Copy and store the key securely

> ⚠️ **Security**: API keys are only shown once. Store them securely.

See [Authentication Guide](authentication.md) for details.

---

## ⏱️ Rate Limiting

### Default Limits

- **Free Tier**: 100 requests/hour
- **Pro Tier**: 1,000 requests/hour
- **Enterprise**: Custom limits

### Rate Limit Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

See [Rate Limiting Guide](rate-limiting.md) for details.

---

## 📚 Related Documentation

- [Endpoints Reference](endpoints.md) - Complete API reference
- [Authentication Guide](authentication.md) - Auth details
- [Rate Limiting](rate-limiting.md) - Limits and quotas

---

<div align="center">

[← Back to API Docs](README.md) • [Next: Endpoints →](endpoints.md)

</div>
